package repos

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/sourcegraph/log"
	"golang.org/x/sync/errgroup"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/conf/reposource"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/auth"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/bitbucketserver"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/internal/jsonc"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/schema"
)

// A BitbucketServerSource yields repositories from a single BitbucketServer connection configured
// in Sourcegraph via the external services configuration.
type BitbucketServerSource struct {
	svc     *types.ExternalService
	config  *schema.BitbucketServerConnection
	exclude excludeFunc
	client  *bitbucketserver.Client
	logger  log.Logger
}

var _ Source = &BitbucketServerSource{}
var _ UserSource = &BitbucketServerSource{}
var _ VersionSource = &BitbucketServerSource{}

// NewBitbucketServerSource returns a new BitbucketServerSource from the given external service.
// rl is optional
func NewBitbucketServerSource(ctx context.Context, logger log.Logger, svc *types.ExternalService, cf *httpcli.Factory) (*BitbucketServerSource, error) {
	rawConfig, err := svc.Config.Decrypt(ctx)
	if err != nil {
		return nil, errors.Errorf("external service id=%d config error: %s", svc.ID, err)
	}
	var c schema.BitbucketServerConnection
	if err := jsonc.Unmarshal(rawConfig, &c); err != nil {
		return nil, errors.Errorf("external service id=%d config error: %s", svc.ID, err)
	}
	return newBitbucketServerSource(logger, svc, &c, cf)
}

func newBitbucketServerSource(logger log.Logger, svc *types.ExternalService, c *schema.BitbucketServerConnection, cf *httpcli.Factory) (*BitbucketServerSource, error) {
	if cf == nil {
		cf = httpcli.ExternalClientFactory
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
		eb.Pattern(r.Pattern)
	}
	exclude, err := eb.Build()
	if err != nil {
		return nil, err
	}

	client, err := bitbucketserver.NewClient(svc.URN(), c, cli)
	if err != nil {
		return nil, err
	}

	return &BitbucketServerSource{
		svc:     svc,
		config:  c,
		exclude: exclude,
		client:  client,
		logger:  logger,
	}, nil
}

// ListRepos returns all BitbucketServer repositories accessible to all connections configured
// in Sourcegraph via the external services configuration.
func (s BitbucketServerSource) ListRepos(ctx context.Context, results chan SourceResult) {
	s.listAllRepos(ctx, results)
}

func (s BitbucketServerSource) WithAuthenticator(a auth.Authenticator) (Source, error) {
	switch a.(type) {
	case *auth.OAuthBearerToken,
		*auth.OAuthBearerTokenWithSSH,
		*auth.BasicAuth,
		*auth.BasicAuthWithSSH,
		*bitbucketserver.SudoableOAuthClient:
		break

	default:
		return nil, newUnsupportedAuthenticatorError("BitbucketServerSource", a)
	}

	sc := s
	sc.client = sc.client.WithAuthenticator(a)

	return &sc, nil
}

// ExternalServices returns a singleton slice containing the external service.
func (s BitbucketServerSource) ExternalServices() types.ExternalServices {
	return types.ExternalServices{s.svc}
}

func (s BitbucketServerSource) makeRepo(repo *bitbucketserver.Repo, isArchived bool) *types.Repo {
	host, err := url.Parse(s.config.Url)
	if err != nil {
		// This should never happen
		panic(errors.Errorf("malformed bitbucket config, invalid URL: %q, error: %s", s.config.Url, err))
	}
	host = extsvc.NormalizeBaseURL(host)

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
			cloneURL = setUserinfoBestEffort(l.Href, s.config.Username, "")
			// No break, so that we fallback to http in case of ssh missing
			// with GitURLType == "ssh"
		}
	}

	urn := s.svc.URN()

	return &types.Repo{
		Name: reposource.BitbucketServerRepoName(
			s.config.RepositoryPathPattern,
			host.Hostname(),
			project,
			repo.Slug,
		),
		URI: string(reposource.BitbucketServerRepoName(
			"",
			host.Hostname(),
			project,
			repo.Slug,
		)),
		ExternalRepo: api.ExternalRepoSpec{
			ID:          strconv.Itoa(repo.ID),
			ServiceType: extsvc.TypeBitbucketServer,
			ServiceID:   host.String(),
		},
		Description: repo.Name,
		Fork:        repo.Origin != nil,
		Archived:    isArchived,
		Private:     !repo.Public,
		Sources: map[string]*types.SourceInfo{
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
		s.exclude(name) ||
		s.exclude(strconv.Itoa(r.ID)) ||
		(s.config.ExcludePersonalRepositories && r.IsPersonalRepository()) {
		return true
	}

	return false
}

func (s *BitbucketServerSource) listAllRepos(ctx context.Context, results chan SourceResult) {
	// "archived" label is a convention used at some customers for indicating a
	// repository is archived (like github's archived state). This is not returned in
	// the normal repository listing endpoints, so we need to fetch it separately.
	archived, err := s.listAllLabeledRepos(ctx, "archived")
	if err != nil {
		results <- SourceResult{Source: s, Err: errors.Wrap(err, "failed to list repos with archived label")}
		return
	}

	ch := make(chan []*bitbucketserver.Repo)

	g, ctx := errgroup.WithContext(ctx)

	// g.SetLimit will not apply a limit for a negative value, but a value of 0 will effectively
	// disable this method. To avoid operator error, we make an explicit check even though it may
	// appear redundant - it could save us from a stressful issue in the future.
	mc := s.config.RateLimit.MaxConcurrency
	if s.config.RateLimit.MaxConcurrency > 0 {
		s.logger.Info("bitbucketserver.listAllRepos setting max concurrency limit", log.Int("maxConcurrency", mc))

		// When SetLimit sets the maximum concurrency to a positive integer N, the N+1 call to g.Go
		// will block unless a slot is available. As a result we **must** invoke g.Go inside a
		// goroutine to avoid blocking. Otherwise, we will never reach the code path to read from
		// the channel blocking the entire method call indefinitely.
		g.SetLimit(mc)
	}

	// Start a new goroutine 'worker" for all repos explicitly added to the code host config.
	go func() {
		g.Go(func() error {
			// Admins normally add to end of lists, so end of list most likely has new repos
			// => stream them first.
			for i := len(s.config.Repos) - 1; i >= 0; i-- {
				name := s.config.Repos[i]
				ps := strings.SplitN(name, "/", 2)
				if len(ps) != 2 {
					// This is an invalid repo name but not a good enough reason to stop the sync
					// completely. As a result log it and continue.
					s.logger.Error(fmt.Sprintf("bitbucketserver.listAllRepos: invalid repo name %q", name))
				}

				projectKey, repoSlug := ps[0], ps[1]
				repo, err := s.client.Repo(ctx, projectKey, repoSlug)
				if err != nil {
					// TODO(tsenart): When implementing dry-run, reconsider alternatives to return
					// 404 errors on external service config validation.
					if bitbucketserver.IsNotFound(err) {
						s.logger.Warn("bitbucketserver.listAllRepos - skipping missing bitbucketserver.repos entry:", log.String("name", name), log.Error(err))
						continue
					}
				}

				select {
				case ch <- []*bitbucketserver.Repo{repo}:
				case <-ctx.Done():
					return errors.Wrap(ctx.Err(), "explicit repos worker")
				}
			}

			return nil
		})
	}()

	// Start a new goroutine "worker" for every line in the repositoryQuery field of the code host
	// config.
	for _, q := range s.config.RepositoryQuery {
		switch q {
		case "none":
			continue
		case "all":
			q = "" // No filters.
		}

		// Copy to avoid the infamouse go-closure bug.
		query := q
		go func() {
			g.Go(func() error {
				next := &bitbucketserver.PageToken{Limit: 1000}
				for next.HasMore() {
					repos, page, err := s.client.Repos(ctx, next, query)
					if err != nil {
						return errors.Wrapf(err, "bitbucketserver.repositoryQuery: query=%q, page=%+v", query, next)
					}

					select {
					case ch <- repos:
						next = page
					case <-ctx.Done():
						return errors.Wrap(ctx.Err(), "repositoryQuery worker")
					}
				}

				return nil
			})
		}()
	}

	// Start a new goroutine "worker" for every project in the projectKeys field of the code host
	// config.
	for _, k := range s.config.ProjectKeys {
		// Copy to avoid the infamouse go-closure bug.
		key := k
		go func() {
			g.Go(func() error {
				repos, err := s.client.ProjectRepos(ctx, key)
				if err != nil {
					// Getting a "fatal" error for a single project key is not a strong enough reason to
					// stop syncing, instead log this error as a warning and write an empty repo list to the table.
					//
					// We cannot return this error since it will terminate all other goroutines in this
					// group.
					s.logger.Warn("bitbucketserver.listAllRepos: error with projectKey", log.String("projectKey", key), log.Error(err))
				}

				select {
				case ch <- repos:
				case <-ctx.Done():
					return errors.Wrap(ctx.Err(), "projectKeys worker")
				}

				return nil
			})
		}()
	}

	// Wait for all the goroutines and close the channel so that the reader below can exit.
	go func() {
		if err := g.Wait(); err != nil {
			s.logger.Error("bitbucketserver.listAllRepos failed", log.Error(err))
		}

		close(ch)
	}()

	// Block until all sync goroutines have completed.
	seen := make(map[int]bool)
	for repos := range ch {
		for _, repo := range repos {
			if !seen[repo.ID] && !s.excludes(repo) {
				_, isArchived := archived[repo.ID]
				results <- SourceResult{Source: s, Repo: s.makeRepo(repo, isArchived)}
				seen[repo.ID] = true
			}
		}
	}

	s.logger.Info("bitbucket.listAllRepos completed")
}

func (s *BitbucketServerSource) listAllLabeledRepos(ctx context.Context, label string) (map[int]struct{}, error) {
	ids := map[int]struct{}{}
	next := &bitbucketserver.PageToken{Limit: 1000}
	for next.HasMore() {
		repos, page, err := s.client.LabeledRepos(ctx, next, label)
		if err != nil {
			// If the instance doesn't have the label then no repos are
			// labeled. Older versions of bitbucket do not support labels, so
			// they too have no labelled repos.
			if bitbucketserver.IsNoSuchLabel(err) || bitbucketserver.IsNotFound(err) {
				// treat as empty
				return ids, nil
			}
			return nil, err
		}

		for _, r := range repos {
			ids[r.ID] = struct{}{}
		}

		next = page
	}
	return ids, nil
}

// AuthenticatedUsername uses the underlying bitbucketserver.Client to get the
// username belonging to the credentials associated with the
// BitbucketServerSource.
func (s *BitbucketServerSource) AuthenticatedUsername(ctx context.Context) (string, error) {
	return s.client.AuthenticatedUsername(ctx)
}

func (s *BitbucketServerSource) ValidateAuthenticator(ctx context.Context) error {
	_, err := s.client.AuthenticatedUsername(ctx)
	return err
}

func (s *BitbucketServerSource) Version(ctx context.Context) (string, error) {
	return s.client.GetVersion(ctx)
}
