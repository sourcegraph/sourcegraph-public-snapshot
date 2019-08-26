package repos

import (
	"context"
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"sync"

	multierror "github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/conf/reposource"
	"github.com/sourcegraph/sourcegraph/pkg/extsvc/bitbucketserver"
	"github.com/sourcegraph/sourcegraph/pkg/httpcli"
	"github.com/sourcegraph/sourcegraph/pkg/jsonc"
	"github.com/sourcegraph/sourcegraph/schema"
	log15 "gopkg.in/inconshreveable/log15.v2"
)

// A BitbucketServerSource yields repositories from a single BitbucketServer connection configured
// in Sourcegraph via the external services configuration.
type BitbucketServerSource struct {
	svc             *ExternalService
	config          *schema.BitbucketServerConnection
	exclude         map[string]bool
	excludePatterns []*regexp.Regexp
	client          *bitbucketserver.Client
}

// NewBitbucketServerSource returns a new BitbucketServerSource from the given external service.
func NewBitbucketServerSource(svc *ExternalService, cf *httpcli.Factory) (*BitbucketServerSource, error) {
	var c schema.BitbucketServerConnection
	if err := jsonc.Unmarshal(svc.Config, &c); err != nil {
		return nil, fmt.Errorf("external service id=%d config error: %s", svc.ID, err)
	}
	return newBitbucketServerSource(svc, &c, cf)
}

func newBitbucketServerSource(svc *ExternalService, c *schema.BitbucketServerConnection, cf *httpcli.Factory) (*BitbucketServerSource, error) {
	baseURL, err := url.Parse(c.Url)
	if err != nil {
		return nil, err
	}
	baseURL = NormalizeBaseURL(baseURL)

	if cf == nil {
		cf = NewHTTPClientFactory()
	}

	var opts []httpcli.Opt
	if c.Certificate != "" {
		pool, err := newCertPool(c.Certificate)
		if err != nil {
			return nil, err
		}
		opts = append(opts, httpcli.NewCertPoolOpt(pool))
	}

	cli, err := cf.Doer(opts...)
	if err != nil {
		return nil, err
	}

	exclude := make(map[string]bool, len(c.Exclude))
	var excludePatterns []*regexp.Regexp
	for _, r := range c.Exclude {
		if r.Name != "" {
			exclude[strings.ToLower(r.Name)] = true
		}

		if r.Id != 0 {
			exclude[strconv.Itoa(r.Id)] = true
		}

		if r.Pattern != "" {
			re, err := regexp.Compile(r.Pattern)
			if err != nil {
				return nil, err
			}
			excludePatterns = append(excludePatterns, re)
		}
	}

	client := bitbucketserver.NewClient(baseURL, cli)
	client.Token = c.Token
	client.Username = c.Username
	client.Password = c.Password

	return &BitbucketServerSource{
		svc:             svc,
		config:          c,
		exclude:         exclude,
		excludePatterns: excludePatterns,
		client:          client,
	}, nil
}

// ListRepos returns all BitbucketServer repositories accessible to all connections configured
// in Sourcegraph via the external services configuration.
func (s BitbucketServerSource) ListRepos(ctx context.Context, results chan *SourceResult) {
	s.listAllRepos(ctx, results)
}

// ExternalServices returns a singleton slice containing the external service.
func (s BitbucketServerSource) ExternalServices() ExternalServices {
	return ExternalServices{s.svc}
}

func (s BitbucketServerSource) makeRepo(repo *bitbucketserver.Repo) *Repo {
	host, err := url.Parse(s.config.Url)
	if err != nil {
		// This should never happen
		panic(errors.Errorf("malformed bitbucket config, invalid URL: %q, error: %s", s.config.Url, err))
	}
	host = NormalizeBaseURL(host)

	// Name
	project := "UNKNOWN"
	if repo.Project != nil {
		project = repo.Project.Key
	}

	// Clone URL
	var cloneURL string
	for _, l := range repo.Links.Clone {
		if l.Name == "ssh" && s.config.GitURLType == "ssh" {
			cloneURL = l.Href
			break
		}
		if l.Name == "http" {
			var password string
			if s.config.Token != "" {
				password = s.config.Token // prefer personal access token
			} else {
				password = s.config.Password
			}
			cloneURL = setUserinfoBestEffort(l.Href, s.config.Username, password)
			// No break, so that we fallback to http in case of ssh missing
			// with GitURLType == "ssh"
		}
	}

	// Repo Links
	// var links *protocol.RepoLinks
	// for _, l := range repo.Links.Self {
	// 	root := strings.TrimSuffix(l.Href, "/browse")
	// 	links = &protocol.RepoLinks{
	// 		Root:   l.Href,
	// 		Tree:   root + "/browse/{path}?at={rev}",
	// 		Blob:   root + "/browse/{path}?at={rev}",
	// 		Commit: root + "/commits/{commit}",
	// 	}
	// 	break
	// }

	urn := s.svc.URN()

	return &Repo{
		Name: string(reposource.BitbucketServerRepoName(
			s.config.RepositoryPathPattern,
			host.Hostname(),
			project,
			repo.Slug,
		)),
		URI: string(reposource.BitbucketServerRepoName(
			"",
			host.Hostname(),
			project,
			repo.Slug,
		)),
		ExternalRepo: api.ExternalRepoSpec{
			ID:          strconv.Itoa(repo.ID),
			ServiceType: bitbucketserver.ServiceType,
			ServiceID:   host.String(),
		},
		Description: repo.Name,
		Fork:        repo.Origin != nil,
		Enabled:     true,
		Sources: map[string]*SourceInfo{
			urn: {
				ID:       urn,
				CloneURL: cloneURL,
			},
		},
		Metadata: repo,
	}
}

func (s *BitbucketServerSource) excludes(r *bitbucketserver.Repo) bool {
	name := r.Slug
	if r.Project != nil {
		name = r.Project.Key + "/" + name
	}
	if r.State != "AVAILABLE" ||
		s.exclude[strings.ToLower(name)] ||
		s.exclude[strconv.Itoa(r.ID)] ||
		(s.config.ExcludePersonalRepositories && r.IsPersonalRepository()) {
		return true
	}

	for _, re := range s.excludePatterns {
		if re.MatchString(name) {
			return true
		}
	}
	return false
}

func (s *BitbucketServerSource) listAllRepos(ctx context.Context, results chan *SourceResult) {
	type batch struct {
		repos []*bitbucketserver.Repo
		err   error
	}

	ch := make(chan batch)

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()

		repos := make([]*bitbucketserver.Repo, 0, len(s.config.Repos))
		errs := new(multierror.Error)

		for _, name := range s.config.Repos {
			ps := strings.SplitN(name, "/", 2)
			if len(ps) != 2 {
				errs = multierror.Append(errs,
					errors.Errorf("bitbucketserver.repos: name=%q", name))
				continue
			}

			projectKey, repoSlug := ps[0], ps[1]
			repo, err := s.client.Repo(ctx, projectKey, repoSlug)
			if err != nil {
				// TODO(tsenart): When implementing dry-run, reconsider alternatives to return
				// 404 errors on external service config validation.
				if bitbucketserver.IsNotFound(err) {
					log15.Warn("skipping missing bitbucketserver.repos entry:", "name", name, "err", err)
					continue
				}
				errs = multierror.Append(errs,
					errors.Wrapf(err, "bitbucketserver.repos: name: %q", name))
			} else {
				repos = append(repos, repo)
			}
		}

		ch <- batch{repos: repos, err: errs.ErrorOrNil()}
	}()

	for _, q := range s.config.RepositoryQuery {
		switch q {
		case "none":
			continue
		case "all":
			q = "" // No filters.
		}

		wg.Add(1)
		go func(q string) {
			defer wg.Done()

			next := &bitbucketserver.PageToken{Limit: 1000}
			for next.HasMore() {
				repos, page, err := s.client.Repos(ctx, next, q)
				if err != nil {
					ch <- batch{err: errors.Wrapf(err, "bitbucketserver.repositoryQuery: query=%q, page=%+v", q, next)}
					break
				}

				ch <- batch{repos: repos}
				next = page
			}
		}(q)
	}

	go func() {
		wg.Wait()
		close(ch)
	}()

	seen := make(map[int]bool)
	for r := range ch {
		if r.err != nil {
			results <- &SourceResult{Source: s, Err: r.err}
			continue
		}

		for _, repo := range r.repos {
			if !seen[repo.ID] && !s.excludes(repo) {
				results <- &SourceResult{Source: s, Repo: s.makeRepo(repo)}
				seen[repo.ID] = true
			}
		}

	}
}
