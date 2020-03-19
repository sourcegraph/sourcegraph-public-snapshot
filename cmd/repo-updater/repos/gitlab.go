package repos

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/internal/conf/reposource"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/gitlab"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/internal/jsonc"
	"github.com/sourcegraph/sourcegraph/schema"
	"gopkg.in/inconshreveable/log15.v2"
)

// A GitLabSource yields repositories from a single GitLab connection configured
// in Sourcegraph via the external services configuration.
type GitLabSource struct {
	svc                 *ExternalService
	config              *schema.GitLabConnection
	exclude             excludeFunc
	baseURL             *url.URL // URL with path /api/v4 (no trailing slash)
	nameTransformations reposource.NameTransformations
	client              *gitlab.Client
}

// NewGitLabSource returns a new GitLabSource from the given external service.
func NewGitLabSource(svc *ExternalService, cf *httpcli.Factory) (*GitLabSource, error) {
	var c schema.GitLabConnection
	if err := jsonc.Unmarshal(svc.Config, &c); err != nil {
		return nil, fmt.Errorf("external service id=%d config error: %s", svc.ID, err)
	}
	return newGitLabSource(svc, &c, cf)
}

func newGitLabSource(svc *ExternalService, c *schema.GitLabConnection, cf *httpcli.Factory) (*GitLabSource, error) {
	baseURL, err := url.Parse(c.Url)
	if err != nil {
		return nil, err
	}
	baseURL = extsvc.NormalizeBaseURL(baseURL)

	if cf == nil {
		cf = httpcli.NewExternalHTTPClientFactory()
	}

	var opts []httpcli.Opt
	if c.Certificate != "" {
		opts = append(opts, httpcli.NewCertPoolOpt(c.Certificate))
	}

	cli, err := cf.Doer(opts...)
	if err != nil {
		return nil, err
	}

	var eb excludeBuilder
	for _, r := range c.Exclude {
		eb.Exact(r.Name)
		eb.Exact(strconv.Itoa(r.Id))
	}
	exclude, err := eb.Build()
	if err != nil {
		return nil, err
	}

	// Validate and cache user-defined name transformations.
	nts, err := reposource.CompileGitLabNameTransformations(c.NameTransformations)
	if err != nil {
		return nil, err
	}

	return &GitLabSource{
		svc:                 svc,
		config:              c,
		exclude:             exclude,
		baseURL:             baseURL,
		nameTransformations: nts,
		client:              gitlab.NewClientProvider(baseURL, cli).GetPATClient(c.Token, ""),
	}, nil
}

// ListRepos returns all GitLab repositories accessible to all connections configured
// in Sourcegraph via the external services configuration.
func (s GitLabSource) ListRepos(ctx context.Context, results chan SourceResult) {
	s.listAllProjects(ctx, results)
}

// GetRepo returns the GitLab repository with the given pathWithNamespace.
func (s GitLabSource) GetRepo(ctx context.Context, pathWithNamespace string) (*Repo, error) {
	proj, err := s.client.GetProject(ctx, gitlab.GetProjectOp{
		PathWithNamespace: pathWithNamespace,
		CommonOp:          gitlab.CommonOp{NoCache: true},
	})

	if err != nil {
		return nil, err
	}

	return s.makeRepo(proj), nil
}

// ExternalServices returns a singleton slice containing the external service.
func (s GitLabSource) ExternalServices() ExternalServices {
	return ExternalServices{s.svc}
}

func (s GitLabSource) makeRepo(proj *gitlab.Project) *Repo {
	urn := s.svc.URN()
	return &Repo{
		Name: string(reposource.GitLabRepoName(
			s.config.RepositoryPathPattern,
			s.baseURL.Hostname(),
			proj.PathWithNamespace,
			s.nameTransformations,
		)),
		URI: string(reposource.GitLabRepoName(
			"",
			s.baseURL.Hostname(),
			proj.PathWithNamespace,
			s.nameTransformations,
		)),
		ExternalRepo: gitlab.ExternalRepoSpec(proj, *s.baseURL),
		Description:  proj.Description,
		Fork:         proj.ForkedFromProject != nil,
		Archived:     proj.Archived,
		Private:      proj.Visibility == "private",
		Sources: map[string]*SourceInfo{
			urn: {
				ID:       urn,
				CloneURL: s.authenticatedRemoteURL(proj),
			},
		},
		Metadata: proj,
	}
}

// authenticatedRemoteURL returns the GitLab projects's Git remote URL with the configured GitLab personal access
// token inserted in the URL userinfo, for repositories needing authentication.
func (s *GitLabSource) authenticatedRemoteURL(proj *gitlab.Project) string {
	if s.config.GitURLType == "ssh" {
		return proj.SSHURLToRepo // SSH authentication must be provided out-of-band
	}
	if s.config.Token == "" || !proj.RequiresAuthentication() {
		return proj.HTTPURLToRepo
	}
	u, err := url.Parse(proj.HTTPURLToRepo)
	if err != nil {
		log15.Warn("Error adding authentication to GitLab repository Git remote URL.", "url", proj.HTTPURLToRepo, "error", err)
		return proj.HTTPURLToRepo
	}
	// Any username works; "git" is not special.
	u.User = url.UserPassword("git", s.config.Token)
	return u.String()
}

func (s *GitLabSource) excludes(p *gitlab.Project) bool {
	return s.exclude(p.PathWithNamespace) || s.exclude(strconv.Itoa(p.ID))
}

func (s *GitLabSource) listAllProjects(ctx context.Context, results chan SourceResult) {
	type batch struct {
		projs []*gitlab.Project
		err   error
	}

	ch := make(chan batch)

	var wg sync.WaitGroup

	projch := make(chan *schema.GitLabProject)
	for i := 0; i < 5; i++ { // 5 concurrent requests
		wg.Add(1)
		go func() {
			defer wg.Done()
			for p := range projch {
				proj, err := s.client.GetProject(ctx, gitlab.GetProjectOp{
					ID:                p.Id,
					PathWithNamespace: p.Name,
					CommonOp:          gitlab.CommonOp{NoCache: true},
				})

				if err != nil {
					// TODO(tsenart): When implementing dry-run, reconsider alternatives to return
					// 404 errors on external service config validation.
					if gitlab.IsNotFound(err) {
						log15.Warn("skipping missing gitlab.projects entry:", "name", p.Name, "id", p.Id, "err", err)
						continue
					}
					ch <- batch{err: errors.Wrapf(err, "gitlab.projects: id: %d, name: %q", p.Id, p.Name)}
				} else {
					ch <- batch{projs: []*gitlab.Project{proj}}
				}

				time.Sleep(s.client.RateLimit.RecommendedWaitForBackgroundOp(1))
			}
		}()
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		defer close(projch)
		// Admins normally add to end of lists, so end of list most likely has
		// new repos => stream them first.
		for i := len(s.config.Projects) - 1; i >= 0; i-- {
			select {
			case projch <- s.config.Projects[i]:
			case <-ctx.Done():
				return
			}
		}
	}()

	for _, projectQuery := range s.config.ProjectQuery {
		if projectQuery == "none" {
			continue
		}

		const perPage = 100
		wg.Add(1)
		go func(projectQuery string) {
			defer wg.Done()

			url, err := projectQueryToURL(projectQuery, perPage) // first page URL
			if err != nil {
				ch <- batch{err: errors.Wrapf(err, "invalid GitLab projectQuery=%q", projectQuery)}
				return
			}

			for {
				if err := ctx.Err(); err != nil {
					ch <- batch{err: err}
					return
				}
				projects, nextPageURL, err := s.client.ListProjects(ctx, url)
				if err != nil {
					ch <- batch{err: errors.Wrapf(err, "error listing GitLab projects: url=%q", url)}
					return
				}
				ch <- batch{projs: projects}
				if nextPageURL == nil {
					return
				}
				url = *nextPageURL

				// 0-duration sleep unless nearing rate limit exhaustion
				time.Sleep(s.client.RateLimit.RecommendedWaitForBackgroundOp(1))
			}
		}(projectQuery)
	}

	go func() {
		wg.Wait()
		close(ch)
	}()

	seen := make(map[int]bool)
	for b := range ch {
		if b.err != nil {
			results <- SourceResult{Source: s, Err: b.err}
			continue
		}

		for _, proj := range b.projs {
			if !seen[proj.ID] && !s.excludes(proj) {
				results <- SourceResult{Source: s, Repo: s.makeRepo(proj)}
				seen[proj.ID] = true
			}
		}
	}
}

var schemeOrHostNotEmptyErr = errors.New("scheme and host should be empty")

func projectQueryToURL(projectQuery string, perPage int) (string, error) {
	// If all we have is the URL query, prepend "projects"
	if strings.HasPrefix(projectQuery, "?") {
		projectQuery = "projects" + projectQuery
	} else if projectQuery == "" {
		projectQuery = "projects"
	}

	u, err := url.Parse(projectQuery)
	if err != nil {
		return "", err
	}
	if u.Scheme != "" || u.Host != "" {
		return "", schemeOrHostNotEmptyErr
	}
	q := u.Query()
	q.Set("per_page", strconv.Itoa(perPage))
	u.RawQuery = q.Encode()

	return u.String(), nil
}
