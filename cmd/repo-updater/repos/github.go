package repos

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"

	multierror "github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/pkg/conf/reposource"
	"github.com/sourcegraph/sourcegraph/pkg/extsvc/github"
	"github.com/sourcegraph/sourcegraph/pkg/httpcli"
	"github.com/sourcegraph/sourcegraph/pkg/jsonc"
	"github.com/sourcegraph/sourcegraph/schema"
	log15 "gopkg.in/inconshreveable/log15.v2"
)

// A GithubSource yields repositories from a single Github connection configured
// in Sourcegraph via the external services configuration.
type GithubSource struct {
	svc          *ExternalService
	config       *schema.GitHubConnection
	exclude      map[string]bool
	githubDotCom bool
	baseURL      *url.URL
	client       *github.Client
	// searchClient is for using the GitHub search API, which has an independent
	// rate limit much lower than non-search API requests.
	searchClient *github.Client

	// originalHostname is the hostname of config.Url (differs from client APIURL, whose host is api.github.com
	// for an originalHostname of github.com).
	originalHostname string
}

// NewGithubSource returns a new GithubSource from the given external service.
func NewGithubSource(svc *ExternalService, cf *httpcli.Factory) (*GithubSource, error) {
	var c schema.GitHubConnection
	if err := jsonc.Unmarshal(svc.Config, &c); err != nil {
		return nil, fmt.Errorf("external service id=%d config error: %s", svc.ID, err)
	}
	return newGithubSource(svc, &c, cf)
}

func newGithubSource(svc *ExternalService, c *schema.GitHubConnection, cf *httpcli.Factory) (*GithubSource, error) {
	baseURL, err := url.Parse(c.Url)
	if err != nil {
		return nil, err
	}
	baseURL = NormalizeBaseURL(baseURL)
	originalHostname := baseURL.Hostname()

	apiURL, githubDotCom := github.APIRoot(baseURL)

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
	for _, r := range c.Exclude {
		if r.Name != "" {
			exclude[strings.ToLower(r.Name)] = true
		}

		if r.Id != "" {
			exclude[r.Id] = true
		}
	}

	return &GithubSource{
		svc:              svc,
		config:           c,
		exclude:          exclude,
		baseURL:          baseURL,
		githubDotCom:     githubDotCom,
		client:           github.NewClient(apiURL, c.Token, cli),
		searchClient:     github.NewClient(apiURL, c.Token, cli),
		originalHostname: originalHostname,
	}, nil
}

// ListRepos returns all Github repositories accessible to all connections configured
// in Sourcegraph via the external services configuration.
func (s GithubSource) ListRepos(ctx context.Context) (repos []*Repo, err error) {
	rs, err := s.listAllRepositories(ctx)
	for _, r := range rs {
		repos = append(repos, s.makeRepo(r))
	}
	return repos, err
}

// ExternalServices returns a singleton slice containing the external service.
func (s GithubSource) ExternalServices() ExternalServices {
	return ExternalServices{s.svc}
}

// GetRepo returns the Github repository with the given name and owner
// ("org/repo-name")
func (s GithubSource) GetRepo(ctx context.Context, nameWithOwner string) (*Repo, error) {
	r, err := s.getRepository(ctx, nameWithOwner)
	if err != nil {
		return nil, err
	}
	return s.makeRepo(r), nil
}

func (s GithubSource) makeRepo(r *github.Repository) *Repo {
	urn := s.svc.URN()
	return &Repo{
		Name: string(reposource.GitHubRepoName(
			s.config.RepositoryPathPattern,
			s.originalHostname,
			r.NameWithOwner,
		)),
		URI: string(reposource.GitHubRepoName(
			"",
			s.originalHostname,
			r.NameWithOwner,
		)),
		ExternalRepo: *github.ExternalRepoSpec(r, *s.baseURL),
		Description:  r.Description,
		Fork:         r.IsFork,
		Enabled:      true,
		Archived:     r.IsArchived,
		Sources: map[string]*SourceInfo{
			urn: {
				ID:       urn,
				CloneURL: s.authenticatedRemoteURL(r),
			},
		},
		Metadata: r,
	}
}

// authenticatedRemoteURL returns the repository's Git remote URL with the configured
// GitHub personal access token inserted in the URL userinfo.
func (s *GithubSource) authenticatedRemoteURL(repo *github.Repository) string {
	if s.config.GitURLType == "ssh" {
		url := fmt.Sprintf("git@%s:%s.git", s.originalHostname, repo.NameWithOwner)
		return url
	}

	if s.config.Token == "" {
		return repo.URL
	}
	u, err := url.Parse(repo.URL)
	if err != nil {
		log15.Warn("Error adding authentication to GitHub repository Git remote URL.", "url", repo.URL, "error", err)
		return repo.URL
	}
	u.User = url.User(s.config.Token)
	return u.String()
}

func (s *GithubSource) excludes(r *github.Repository) bool {
	return s.exclude[strings.ToLower(r.NameWithOwner)] || s.exclude[r.ID]
}

// RepositoryPager is a function that takes in a page number and returns
type RepositoryPager func(page int) (repos []*github.Repository, hasNextPage bool, wait int, err error)

// exhaustRepositoryPager takes in a RepositoryPager and accumulates all the repositories from every page.
func (s *GithubSource) exhaustRepositoryPager(ctx context.Context, pager RepositoryPager) (set map[int64]*github.Repository, err error) {
	set = make(map[int64]*github.Repository)

	hasNextPage := true
	for page := 1; hasNextPage; page++ {
		if err = ctx.Err(); err != nil {
			return
		}

		var pageRepos []*github.Repository
		var wait int
		pageRepos, hasNextPage, wait, err = pager(page)
		if err != nil {
			return
		}

		for _, r := range pageRepos {
			set[r.DatabaseID] = r
		}

		if hasNextPage && wait > 0 {
			time.Sleep(s.client.RateLimit.RecommendedWaitForBackgroundOp(wait))
		}
	}

	return
}

// listOrgRepositories handles the `orgs` option.
func (s *GithubSource) listOrgRepositories(ctx context.Context, org string) (map[int64]*github.Repository, error) {
	return s.exhaustRepositoryPager(
		ctx,
		func(page int) (repos []*github.Repository, hasNextPage bool, rateLimitCost int, err error) {
			repos, hasNextPage, rateLimitCost, err = s.client.ListOrgRepositories(ctx, org, page)
			rateLimitRemaining, rateLimitReset, rateLimitRetry, _ := s.client.RateLimit.Get()
			log15.Debug(
				"github sync: ListOrgRepositories",
				"repos", len(repos),
				"rateLimitCost", rateLimitCost,
				"rateLimitRemaining", rateLimitRemaining,
				"rateLimitReset", rateLimitReset,
				"retryAfter", rateLimitRetry,
			)
			return
		},
	)
}

// listSelectedRepositories handles the `repos` option.
func (s *GithubSource) listSelectedRepositories(ctx context.Context, names []string) (set map[int64]*github.Repository, err error) {
	set = make(map[int64]*github.Repository)
	for _, nameWithOwner := range s.config.Repos {
		if err = ctx.Err(); err != nil {
			return
		}

		var owner, name string
		owner, name, err = github.SplitRepositoryNameWithOwner(nameWithOwner)
		if err != nil {
			err = errors.New("Invalid GitHub repository: nameWithOwner=" + nameWithOwner)
			return
		}
		var repo *github.Repository
		repo, err = s.client.GetRepository(ctx, owner, name)
		if err != nil {
			// TODO(tsenart): When implementing dry-run, reconsider alternatives to return
			// 404 errors on external service config validation.
			if github.IsNotFound(err) {
				log15.Warn("skipping missing github.repos entry:", "name", nameWithOwner, "err", err)
				err = nil
				continue
			}
			err = errors.Wrapf(err, "Error getting GitHub repository: nameWithOwner=%s", nameWithOwner)
			return
		}
		log15.Debug("github sync: GetRepository", "repo", repo.NameWithOwner)
		set[repo.DatabaseID] = repo
		time.Sleep(s.client.RateLimit.RecommendedWaitForBackgroundOp(1)) // 0-duration sleep unless nearing rate limit exhaustion
	}
	return
}

// listQueryRepositories handles the `repositoryQuery` option.
func (s *GithubSource) listQueryRepositories(ctx context.Context, query string) (set map[int64]*github.Repository, err error) {
	switch query {
	case "public":
		set = make(map[int64]*github.Repository)
		if s.githubDotCom {
			err = errors.New(`unsupported configuration "public" for "repositoryQuery" for github.com`)
			return
		}
		var sinceRepoID int64
		for {
			if err = ctx.Err(); err != nil {
				return
			}

			var repos []*github.Repository
			repos, err = s.client.ListPublicRepositories(ctx, sinceRepoID)
			if err != nil {
				err = errors.Wrapf(err, "failed to list public repositories: sinceRepoID=%d", sinceRepoID)
				return
			}
			if len(repos) == 0 {
				return
			}
			log15.Debug("github sync public", "repos", len(repos), "error", err)
			for _, r := range repos {
				fmt.Println(r.NameWithOwner)
				set[r.DatabaseID] = r
				if sinceRepoID < r.DatabaseID {
					sinceRepoID = r.DatabaseID
				}
			}
		}
	case "affiliated":
		set, err = s.exhaustRepositoryPager(
			ctx,
			func(page int) (repos []*github.Repository, hasNextPage bool, rateLimitCost int, err error) {
				repos, hasNextPage, rateLimitCost, err = s.client.ListUserRepositories(ctx, page)
				rateLimitRemaining, rateLimitReset, rateLimitRetry, _ := s.client.RateLimit.Get()
				log15.Debug(
					"github sync: ListUserRepositories",
					"repos", len(repos),
					"rateLimitCost", rateLimitCost,
					"rateLimitRemaining", rateLimitRemaining,
					"rateLimitReset", rateLimitReset,
					"retryAfter", rateLimitRetry,
				)
				return
			},
		)
	case "none":
		// nothing
	default:
		// Run the query as a GitHub advanced repository search
		// (https://github.com/search/advanced).
		set, err = s.exhaustRepositoryPager(
			ctx,
			func(page int) (repos []*github.Repository, hasNextPage bool, rateLimitCost int, err error) {
				var reposPage github.RepositoryListPage
				reposPage, err = s.searchClient.ListRepositoriesForSearch(ctx, query, page)
				if err != nil {
					err = errors.Wrapf(err, "failed to list GitHub repositories for search: page=%d, searchString=%q", page, query)
					return
				}

				if reposPage.TotalCount > 1000 {
					// GitHub's advanced repository search will only
					// return 1000 results. We specially handle this case
					// to ensure the admin gets a detailed error
					// message. https://github.com/sourcegraph/sourcegraph/issues/2562
					err = errors.Errorf(`repositoryQuery %q would return %d results. GitHub's Search API only returns up to 1000 results. Please adjust your repository query into multiple queries such that each returns less than 1000 results. For example: {"repositoryQuery": %s}`, query, reposPage.TotalCount, exampleRepositoryQuerySplit(query))
					return
				}

				hasNextPage = reposPage.HasNextPage
				repos = reposPage.Repos

				rateLimitRemaining, rateLimitReset, rateLimitRetry, ok := s.searchClient.RateLimit.Get()
				log15.Debug(
					"github sync: ListRepositoriesForSearch",
					"searchString", query,
					"repos", len(repos),
					"rateLimitRemaining", rateLimitRemaining,
					"rateLimitReset", rateLimitReset,
					"retryAfter", rateLimitRetry,
				)

				// GitHub search has vastly different rate limits to
				// the normal GitHub API (30req/m vs
				// 5000req/h). RecommendedWaitForBackgroundOp has
				// heuristics tuned for the normal API, part of which
				// is to not sleep if we have ample rate limit left.
				//
				// So we only let the heuristic kick in if we have
				// less than 5 requests left.
				if rateLimitRetry > 0 || (ok && rateLimitRemaining < 5) {
					rateLimitCost = 1
				}

				return
			},
		)
	}
	return
}

func (s *GithubSource) listAllRepositories(ctx context.Context) ([]*github.Repository, error) {
	set := make(map[int64]*github.Repository)
	errs := new(multierror.Error)

	// repositoryQuery
	for _, repositoryQuery := range s.config.RepositoryQuery {
		repos, err := s.listQueryRepositories(ctx, repositoryQuery)
		if err != nil {
			errs = multierror.Append(errs, err)
			continue
		}
		for id, r := range repos {
			set[id] = r
		}
	}

	// repos
	err := s.fetchAllRepositoriesInBatches(ctx, set)
	if err != nil {
		// The way we fetch repositories in batches through the GraphQL API -
		// using aliases to query multiple repositories in one query - is
		// currently "undefined behaviour". Very rarely but unreproducibly it
		// resulted in EOF errors while testing. And since we rely on fetching
		// to work, we fall back to the (slower) sequential fetching in case we
		// run into an GraphQL API error
		log15.Warn("github sync: fetching in batches failed. falling back to sequential fetch", "error", err)

		repos, err := s.listSelectedRepositories(s.config.Repos, ctx)
		if err != nil {
			errs = multierror.Append(errs, err)
		}

		for id, r := range repos {
			set[id] = r
		}
	}

	// orgs
	for _, org := range s.config.Orgs {
		repos, err := s.listOrgRepositories(ctx, org)
		if err != nil {
			errs = multierror.Append(errs, errors.Wrapf(err, "failed to list organization %s repos", org))
			continue
		}

		for id, r := range repos {
			set[id] = r
		}
	}

	// exclude
	repos := make([]*github.Repository, 0, len(set))
	for _, repo := range set {
		if !s.excludes(repo) {
			repos = append(repos, repo)
		}
	}

	return repos, errs.ErrorOrNil()
}

func (s *GithubSource) getRepository(ctx context.Context, nameWithOwner string) (*github.Repository, error) {
	owner, name, err := github.SplitRepositoryNameWithOwner(nameWithOwner)
	if err != nil {
		return nil, errors.Wrapf(err, "Invalid GitHub repository: nameWithOwner="+nameWithOwner)
	}

	repo, err := s.client.GetRepository(ctx, owner, name)
	if err != nil {
		return nil, err
	}

	return repo, nil
}

// fetchAllRepositoriesInBatches fetches the repositories configured in
// config.Repos in batches and adds them to the supplied set
func (s *GithubSource) fetchAllRepositoriesInBatches(ctx context.Context, set map[int64]*github.Repository) error {
	const batchSize = 30

	for i := 0; i < len(s.config.Repos); i += batchSize {
		if err := ctx.Err(); err != nil {
			return err
		}

		start := i
		end := i + batchSize
		if end > len(s.config.Repos) {
			end = len(s.config.Repos)
		}
		batch := s.config.Repos[start:end]

		repos, err := s.client.GetReposByNameWithOwner(ctx, batch...)
		if err != nil {
			return err
		}

		log15.Debug("github sync: GetGetReposByNameWithOwner", "repos", batch)
		for _, r := range repos {
			set[r.DatabaseID] = r
		}

		time.Sleep(s.client.RateLimit.RecommendedWaitForBackgroundOp(1)) // 0-duration sleep unless nearing rate limit exhaustion
	}

	return nil
}

func exampleRepositoryQuerySplit(q string) string {
	var qs []string
	for _, suffix := range []string{"created:>=2019", "created:2018", "created:2016..2017", "created:<2016"} {
		qs = append(qs, fmt.Sprintf("%s %s", q, suffix))
	}
	// Avoid escaping < and >
	var b bytes.Buffer
	enc := json.NewEncoder(&b)
	enc.SetEscapeHTML(false)
	_ = enc.Encode(qs)
	return strings.TrimSpace(b.String())
}
