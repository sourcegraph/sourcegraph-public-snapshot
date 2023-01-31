package repos

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/conf/reposource"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/auth"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/github"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/internal/jsonc"
	"github.com/sourcegraph/sourcegraph/internal/lazyregexp"
	"github.com/sourcegraph/sourcegraph/internal/ratelimit"
	"github.com/sourcegraph/sourcegraph/internal/timeutil"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/schema"
)

// A GitHubSource yields repositories from a single GitHub connection configured
// in Sourcegraph via the external services configuration.
type GitHubSource struct {
	svc             *types.ExternalService
	config          *schema.GitHubConnection
	exclude         excludeFunc
	excludeArchived bool
	excludeForks    bool
	githubDotCom    bool
	baseURL         *url.URL
	v3Client        *github.V3Client
	v4Client        *github.V4Client
	// searchClient is for using the GitHub search API, which has an independent
	// rate limit much lower than non-search API requests.
	searchClient *github.V3Client

	// originalHostname is the hostname of config.Url (differs from client APIURL, whose host is api.github.com
	// for an originalHostname of github.com).
	originalHostname string

	// useGitHubApp indicate whether clients are authenticated through GitHub App,
	// which may need to hit different API endpoints from regular RESTful API.
	useGitHubApp bool

	logger log.Logger
}

var (
	_ Source                     = &GitHubSource{}
	_ UserSource                 = &GitHubSource{}
	_ AffiliatedRepositorySource = &GitHubSource{}
	_ VersionSource              = &GitHubSource{}
)

// NewGithubSource returns a new GitHubSource from the given external service.
func NewGithubSource(ctx context.Context, logger log.Logger, externalServicesStore database.ExternalServiceStore, svc *types.ExternalService, cf *httpcli.Factory) (*GitHubSource, error) {
	rawConfig, err := svc.Config.Decrypt(ctx)
	if err != nil {
		return nil, errors.Errorf("external service id=%d config error: %s", svc.ID, err)
	}
	var c schema.GitHubConnection
	if err := jsonc.Unmarshal(rawConfig, &c); err != nil {
		return nil, errors.Errorf("external service id=%d config error: %s", svc.ID, err)
	}
	return newGithubSource(logger, externalServicesStore, svc, &c, cf)
}

var githubRemainingGauge = promauto.NewGaugeVec(prometheus.GaugeOpts{
	// _v2 since we have an older metric defined in github-proxy
	Name: "src_github_rate_limit_remaining_v2",
	Help: "Number of calls to GitHub's API remaining before hitting the rate limit.",
}, []string{"resource", "name"})

var githubRatelimitWaitCounter = promauto.NewCounterVec(prometheus.CounterOpts{
	Name: "src_github_rate_limit_wait_duration_seconds",
	Help: "The amount of time spent waiting on the rate limit",
}, []string{"resource", "name"})

func newGithubSource(
	logger log.Logger,
	externalServicesStore database.ExternalServiceStore,
	svc *types.ExternalService,
	c *schema.GitHubConnection,
	cf *httpcli.Factory,
) (*GitHubSource, error) {
	baseURL, err := url.Parse(c.Url)
	if err != nil {
		return nil, err
	}
	baseURL = extsvc.NormalizeBaseURL(baseURL)
	originalHostname := baseURL.Hostname()

	apiURL, githubDotCom := github.APIRoot(baseURL)

	if cf == nil {
		cf = httpcli.ExternalClientFactory
	}

	opts := []httpcli.Opt{
		// Use a 30s timeout to avoid running into EOF errors, because GitHub
		// closes idle connections after 60s
		httpcli.NewIdleConnTimeoutOpt(30 * time.Second),
	}

	if c.Certificate != "" {
		opts = append(opts, httpcli.NewCertPoolOpt(c.Certificate))
	}

	cli, err := cf.Doer(opts...)
	if err != nil {
		return nil, err
	}

	var (
		eb              excludeBuilder
		excludeArchived bool
		excludeForks    bool
	)

	for _, r := range c.Exclude {
		eb.Exact(r.Name)
		eb.Exact(r.Id)
		eb.Pattern(r.Pattern)

		if r.Archived {
			excludeArchived = true
		}

		if r.Forks {
			excludeForks = true
		}
	}

	exclude, err := eb.Build()
	if err != nil {
		return nil, err
	}
	token := &auth.OAuthBearerToken{Token: c.Token}
	urn := svc.URN()

	var (
		v3ClientLogger = log.Scoped("source", "github client for github source")
		v3Client       = github.NewV3Client(v3ClientLogger, urn, apiURL, token, cli)
		v4Client       = github.NewV4Client(urn, apiURL, token, cli)

		searchClientLogger = log.Scoped("search", "github client for search")
		searchClient       = github.NewV3SearchClient(searchClientLogger, urn, apiURL, token, cli)
	)

	useGitHubApp := false
	config, err := conf.GitHubAppConfig()
	if err != nil {
		return nil, err
	}
	if c.GithubAppInstallationID != "" && config.Configured() {
		installationID, err := strconv.ParseInt(c.GithubAppInstallationID, 10, 64)
		if err != nil {
			return nil, errors.Wrap(err, "parse installation ID")
		}

		installationAuther, err := database.BuildGitHubAppInstallationAuther(externalServicesStore, config.AppID, config.PrivateKey, urn, apiURL, cli, installationID, svc)
		if err != nil {
			return nil, errors.Wrap(err, "creating GitHub App installation authenticator")
		}

		v3Client = github.NewV3Client(v3ClientLogger, urn, apiURL, installationAuther, cli)
		v4Client = github.NewV4Client(urn, apiURL, installationAuther, cli)
		useGitHubApp = true
	}

	for resource, monitor := range map[string]*ratelimit.Monitor{
		"rest":    v3Client.RateLimitMonitor(),
		"graphql": v4Client.RateLimitMonitor(),
		"search":  searchClient.RateLimitMonitor(),
	} {
		// Copy the resource or funcs below will use the last one seen while iterating
		// the map
		resource := resource
		// Copy displayName so that the funcs below don't capture the svc pointer
		displayName := svc.DisplayName
		monitor.SetCollector(&ratelimit.MetricsCollector{
			Remaining: func(n float64) {
				githubRemainingGauge.WithLabelValues(resource, displayName).Set(n)
			},
			WaitDuration: func(n time.Duration) {
				githubRatelimitWaitCounter.WithLabelValues(resource, displayName).Add(n.Seconds())
			},
		})
	}

	return &GitHubSource{
		svc:              svc,
		config:           c,
		exclude:          exclude,
		excludeArchived:  excludeArchived,
		excludeForks:     excludeForks,
		baseURL:          baseURL,
		githubDotCom:     githubDotCom,
		v3Client:         v3Client,
		v4Client:         v4Client,
		searchClient:     searchClient,
		originalHostname: originalHostname,
		useGitHubApp:     useGitHubApp,
		logger: logger.With(
			log.Object("GitHubSource",
				log.Bool("excludeForks", excludeForks),
				log.Bool("githubDotCom", githubDotCom),
				log.String("originalHostname", originalHostname),
				log.Bool("useGitHubApp", useGitHubApp),
			),
		),
	}, nil
}

func (s *GitHubSource) WithAuthenticator(a auth.Authenticator) (Source, error) {
	switch a.(type) {
	case *auth.OAuthBearerToken,
		*auth.OAuthBearerTokenWithSSH:
		break

	default:
		return nil, newUnsupportedAuthenticatorError("GitHubSource", a)
	}

	sc := *s
	sc.v3Client = sc.v3Client.WithAuthenticator(a)
	sc.v4Client = sc.v4Client.WithAuthenticator(a)
	sc.searchClient = sc.searchClient.WithAuthenticator(a)

	return &sc, nil
}

type githubResult struct {
	err  error
	repo *github.Repository
}

func (s *GitHubSource) ValidateAuthenticator(ctx context.Context) error {
	var err error
	if s.config.GithubAppInstallationID != "" {
		// GitHub App does not have an affiliated user, use another
		// request instead.
		_, err = s.v3Client.GetAuthenticatedOAuthScopes(ctx)
	} else {
		_, err = s.v3Client.GetAuthenticatedUser(ctx)
	}
	return err
}

func (s *GitHubSource) Version(ctx context.Context) (string, error) {
	return s.v3Client.GetVersion(ctx)
}

func (s *GitHubSource) CheckConnection(ctx context.Context) error {
	_, err := s.v3Client.GetAuthenticatedUser(ctx)
	if err != nil {
		return errors.Wrap(err, "connection check failed. could not fetch authenticated user")
	}
	return nil
}

// ListRepos returns all Github repositories accessible to all connections configured
// in Sourcegraph via the external services configuration.
func (s *GitHubSource) ListRepos(ctx context.Context, results chan SourceResult) {
	unfiltered := make(chan *githubResult)
	go func() {
		s.listAllRepositories(ctx, unfiltered)
		close(unfiltered)
	}()

	seen := make(map[int64]bool)
	for res := range unfiltered {
		if res.err != nil {
			results <- SourceResult{Source: s, Err: res.err}
			continue
		}

		s.logger.Debug("unfiltered", log.String("repo", res.repo.NameWithOwner))
		if !seen[res.repo.DatabaseID] && !s.excludes(res.repo) {
			results <- SourceResult{Source: s, Repo: s.makeRepo(res.repo)}
			s.logger.Debug("sent to result", log.String("repo", res.repo.NameWithOwner))
			seen[res.repo.DatabaseID] = true
		}
	}
}

// ExternalServices returns a singleton slice containing the external service.
func (s *GitHubSource) ExternalServices() types.ExternalServices {
	return types.ExternalServices{s.svc}
}

// GetRepo returns the GitHub repository with the given name and owner
// ("org/repo-name")
func (s *GitHubSource) GetRepo(ctx context.Context, nameWithOwner string) (*types.Repo, error) {
	r, err := s.getRepository(ctx, nameWithOwner)
	if err != nil {
		return nil, err
	}
	return s.makeRepo(r), nil
}

func (s *GitHubSource) makeRepo(r *github.Repository) *types.Repo {
	urn := s.svc.URN()
	metadata := *r
	// This field flip flops depending on which token was used to retrieve the repo
	// so we don't want to store it.
	metadata.ViewerPermission = ""
	return &types.Repo{
		Name: reposource.GitHubRepoName(
			s.config.RepositoryPathPattern,
			s.originalHostname,
			r.NameWithOwner,
		),
		URI: string(reposource.GitHubRepoName(
			"",
			s.originalHostname,
			r.NameWithOwner,
		)),
		ExternalRepo: github.ExternalRepoSpec(r, s.baseURL),
		Description:  r.Description,
		Fork:         r.IsFork,
		Archived:     r.IsArchived,
		Stars:        r.StargazerCount,
		Private:      r.IsPrivate,
		Sources: map[string]*types.SourceInfo{
			urn: {
				ID:       urn,
				CloneURL: s.remoteURL(r),
			},
		},
		Metadata: &metadata,
	}
}

// remoteURL returns the repository's Git remote URL
//
// note: this used to contain credentials but that is no longer the case
// if you need to get an authenticated clone url use repos.CloneURL
func (s *GitHubSource) remoteURL(repo *github.Repository) string {
	if s.config.GitURLType == "ssh" {
		assembledURL := fmt.Sprintf("git@%s:%s.git", s.originalHostname, repo.NameWithOwner)
		return assembledURL
	}

	return repo.URL
}

func (s *GitHubSource) excludes(r *github.Repository) bool {
	if r.IsLocked || r.IsDisabled {
		return true
	}

	if s.exclude(r.NameWithOwner) || s.exclude(r.ID) {
		return true
	}

	if s.excludeArchived && r.IsArchived {
		return true
	}

	if s.excludeForks && r.IsFork {
		return true
	}

	return false
}

// repositoryPager is a function that returns repositories on a given `page`.
// It also returns:
// - `hasNext` bool: if there is a next page
// - `cost` int: rate limit cost used to determine recommended wait before next call
// - `err` error: if something goes wrong
type repositoryPager func(page int) (repos []*github.Repository, hasNext bool, cost int, err error)

// paginate returns all the repositories from the given repositoryPager.
// It repeatedly calls `pager` with incrementing page count until it
// returns false for hasNext.
func (s *GitHubSource) paginate(ctx context.Context, results chan *githubResult, pager repositoryPager) {
	hasNext := true
	for page := 1; hasNext; page++ {
		if err := ctx.Err(); err != nil {
			results <- &githubResult{err: err}
			return
		}

		var pageRepos []*github.Repository
		var cost int
		var err error
		pageRepos, hasNext, cost, err = pager(page)
		if err != nil {
			results <- &githubResult{err: err}
			return
		}

		for _, r := range pageRepos {
			if err := ctx.Err(); err != nil {
				results <- &githubResult{err: err}
				return
			}

			results <- &githubResult{repo: r}
		}

		if hasNext && cost > 0 {
			// 0-duration sleep unless nearing rate limit exhaustion, or
			// shorter if context has been canceled (next iteration of loop
			// will then return `ctx.Err()`).
			timeutil.SleepWithContext(ctx, s.v3Client.RateLimitMonitor().RecommendedWaitForBackgroundOp(cost))
		}
	}
}

// listOrg handles the `org` config option.
// It returns all the repositories belonging to the given organization
// by hitting the /orgs/:org/repos endpoint.
//
// It returns an error if the request fails on the first page.
func (s *GitHubSource) listOrg(ctx context.Context, org string, results chan *githubResult) {
	dedupC := make(chan *githubResult)

	// Currently, the Github API doesn't return internal repos
	// when calling it with the "all" type.
	// We need to call it twice, once with the "all" type and
	// once with the "internal" type.
	// However, since we don't have any guarantee that this behavior
	// will always remain the same and that Github will never fix this issue,
	// we need to deduplicate the results before sending them to the results channel.

	getReposByType := func(tp string) error {
		var oerr error

		s.paginate(ctx, dedupC, func(page int) (repos []*github.Repository, hasNext bool, cost int, err error) {
			defer func() {
				if page == 1 {
					var e *github.APIError
					if errors.As(err, &e) && e.Code == 404 {
						oerr = errors.Errorf("organisation %q (specified in configuration) not found", org)
						err = nil
					}
				}

				remaining, reset, retry, _ := s.v3Client.RateLimitMonitor().Get()
				s.logger.Debug(
					"github sync: ListOrgRepositories",
					log.Int("repos", len(repos)),
					log.Int("rateLimitCost", cost),
					log.Int("rateLimitRemaining", remaining),
					log.Duration("rateLimitReset", reset),
					log.Duration("retryAfter", retry),
					log.String("type", tp),
				)
			}()

			return s.v3Client.ListOrgRepositories(ctx, org, page, tp)
		})

		return oerr
	}

	go func() {
		defer close(dedupC)

		err := getReposByType("all")
		// Handle 404 from org repos endpoint by trying user repos endpoint
		if err != nil && ctx.Err() == nil {
			if s.listUser(ctx, org, dedupC) != nil {
				dedupC <- &githubResult{err: err}
			}
			return
		}

		if err := ctx.Err(); err != nil {
			dedupC <- &githubResult{err: err}
			return
		}

		// if the first call succeeded,
		// call the same endpoint with the "internal" type
		if err = getReposByType("internal"); err != nil {
			dedupC <- &githubResult{err: err}
		}
	}()

	seen := make(map[string]bool)

	for res := range dedupC {
		if res.err == nil {
			if seen[res.repo.ID] {
				continue
			}

			seen[res.repo.ID] = true
		}

		results <- res
	}
}

// listUser returns all the repositories belonging to the given user
// by hitting the /users/:user/repos endpoint.
//
// It returns an error if the request fails on the first page.
func (s *GitHubSource) listUser(ctx context.Context, user string, results chan *githubResult) (fail error) {
	s.paginate(ctx, results, func(page int) (repos []*github.Repository, hasNext bool, cost int, err error) {
		defer func() {
			if err != nil && page == 1 {
				fail, err = err, nil
			}

			remaining, reset, retry, _ := s.v3Client.RateLimitMonitor().Get()
			s.logger.Debug(
				"github sync: ListUserRepositories",
				log.Int("repos", len(repos)),
				log.Int("rateLimitCost", cost),
				log.Int("rateLimitRemaining", remaining),
				log.Duration("rateLimitReset", reset),
				log.Duration("retryAfter", retry),
			)
		}()
		return s.v3Client.ListUserRepositories(ctx, user, page)
	})
	return
}

// listRepos returns the valid repositories from the given list of repository names.
// This is done by hitting the /repos/:owner/:name endpoint for each of the given
// repository names.
func (s *GitHubSource) listRepos(ctx context.Context, repos []string, results chan *githubResult) {
	if err := s.fetchAllRepositoriesInBatches(ctx, results); err == nil {
		return
	} else {
		if err := ctx.Err(); err != nil {
			return
		}
		// The way we fetch repositories in batches through the GraphQL API -
		// using aliases to query multiple repositories in one query - is
		// currently "undefined behaviour". Very rarely but unreproducibly it
		// resulted in EOF errors while testing. And since we rely on fetching
		// to work, we fall back to the (slower) sequential fetching in case we
		// run into an GraphQL API error
		s.logger.Warn("github sync: fetching in batches failed, falling back to sequential fetch", log.Error(err))
	}

	// Admins normally add to end of lists, so end of list most likely has new
	// repos => stream them first.
	for i := len(repos) - 1; i >= 0; i-- {
		nameWithOwner := repos[i]
		if err := ctx.Err(); err != nil {
			results <- &githubResult{err: errors.Wrapf(err, "context error for repository: namewithOwner=%s", nameWithOwner)}
			return
		}

		owner, name, err := github.SplitRepositoryNameWithOwner(nameWithOwner)
		if err != nil {
			results <- &githubResult{err: errors.Newf("Invalid GitHub repository: nameWithOwner=%s", nameWithOwner)}
			return
		}
		var repo *github.Repository
		repo, err = s.v3Client.GetRepository(ctx, owner, name)
		if err != nil {
			// TODO(tsenart): When implementing dry-run, reconsider alternatives to return
			// 404 errors on external service config validation.
			if github.IsNotFound(err) {
				s.logger.Warn("skipping missing github.repos entry:", log.String("name", nameWithOwner), log.Error(err))
			} else {
				results <- &githubResult{err: errors.Wrapf(err, "Error getting GitHub repository: nameWithOwner=%s", nameWithOwner)}
			}
			continue
		}
		s.logger.Debug("github sync: GetRepository", log.String("repo", repo.NameWithOwner))

		results <- &githubResult{repo: repo}

		// If there is another iteration of the loop: 0-duration sleep unless
		// nearing rate limit exhaustion, or shorter if context has been
		// canceled. If context has been canceled, the `ctx.Err()` will be
		// returned by next iteration.
		if i > 0 {
			timeutil.SleepWithContext(ctx, s.v3Client.RateLimitMonitor().RecommendedWaitForBackgroundOp(1))
		}
	}
}

// listPublic handles the `public` keyword of the `repositoryQuery` config option.
// It returns the public repositories listed on the /repositories endpoint.
func (s *GitHubSource) listPublic(ctx context.Context, results chan *githubResult) {
	if s.githubDotCom {
		results <- &githubResult{err: errors.New(`unsupported configuration "public" for "repositoryQuery" for github.com`)}
		return
	}

	// The regular Github API endpoint for listing public repos doesn't return whether the repo is archived, so we have to list
	// all of the public archived repos first so we know if a repo is archived or not.
	// TODO: Remove querying for archived repos first when https://github.com/orgs/community/discussions/12554 gets resolved
	archivedReposChan := make(chan *githubResult)
	archivedRepos := make(map[string]struct{})
	archivedReposCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	go func() {
		s.listPublicArchivedRepos(archivedReposCtx, archivedReposChan)
		close(archivedReposChan)
	}()

	for res := range archivedReposChan {
		if res.err != nil {
			results <- &githubResult{err: errors.Wrap(res.err, "failed to list public archived Github repositories")}
			return
		}
		archivedRepos[res.repo.ID] = struct{}{}
	}

	var sinceRepoID int64
	for {
		if err := ctx.Err(); err != nil {
			results <- &githubResult{err: err}
			return
		}

		repos, hasNextPage, err := s.v3Client.ListPublicRepositories(ctx, sinceRepoID)
		if err != nil {
			results <- &githubResult{err: errors.Wrapf(err, "failed to list public repositories: sinceRepoID=%d", sinceRepoID)}
			return
		}
		if !hasNextPage {
			return
		}
		s.logger.Debug("github sync public", log.Int("repos", len(repos)), log.Error(err))
		for _, r := range repos {
			_, isArchived := archivedRepos[r.ID]
			r.IsArchived = isArchived
			if err := ctx.Err(); err != nil {
				results <- &githubResult{err: err}
				return
			}

			results <- &githubResult{repo: r}
			if sinceRepoID < r.DatabaseID {
				sinceRepoID = r.DatabaseID
			}
		}
	}
}

// listPublicArchivedRepos returns all of the public archived repositories listed on the /search/repositories endpoint.
// NOTE: There is a limitation on the search API that this uses, if there are more than 1000 public archived repos that
// were created in the same time (to the second), this list will miss any repos that lie outside of the first 1000.
func (s *GitHubSource) listPublicArchivedRepos(ctx context.Context, results chan *githubResult) {
	s.listSearch(ctx, "archived:true is:public", results)
}

// listAffiliated handles the `affiliated` keyword of the `repositoryQuery` config option.
// It returns the repositories affiliated with the client token by hitting the /user/repos
// endpoint.
//
// Affiliation is present if the user: (1) owns the repo, (2) is apart of an org that
// the repo belongs to, or (3) is a collaborator.
func (s *GitHubSource) listAffiliated(ctx context.Context, results chan *githubResult) {
	s.paginate(ctx, results, func(page int) (repos []*github.Repository, hasNext bool, cost int, err error) {
		defer func() {
			remaining, reset, retry, _ := s.v3Client.RateLimitMonitor().Get()
			s.logger.Debug(
				"github sync: ListAffiliated",
				log.Int("repos", len(repos)),
				log.Int("rateLimitCost", cost),
				log.Int("rateLimitRemaining", remaining),
				log.Duration("rateLimitReset", reset),
				log.Duration("retryAfter", retry),
			)
		}()
		if s.useGitHubApp {
			return s.v3Client.ListInstallationRepositories(ctx, page)
		}
		return s.v3Client.ListAffiliatedRepositories(ctx, github.VisibilityAll, page)
	})
}

// listSearch handles the `repositoryQuery` config option when a keyword is not present.
// It returns the repositories matching a GitHub's advanced repository search query
// via the GraphQL API.
func (s *GitHubSource) listSearch(ctx context.Context, q string, results chan *githubResult) {
	(&repositoryQuery{Query: q, Searcher: s.v4Client, Logger: s.logger}).Do(ctx, results)
}

// GitHub was founded on February 2008, so this minimum date covers all repos
// created on it.
var minCreated = time.Date(2007, time.June, 1, 0, 0, 0, 0, time.UTC)

type dateRange struct{ From, To, OriginalTo time.Time }

func (r dateRange) String() string {
	const dateFormat = "2006-01-02T15:04:05-07:00"

	return fmt.Sprintf("%s..%s",
		r.From.Format(dateFormat),
		r.To.Format(dateFormat),
	)
}

func (r dateRange) Size() time.Duration { return r.To.Sub(r.From) }

type repositoryQuery struct {
	Query    string
	Created  *dateRange
	Cursor   github.Cursor
	First    int
	Limit    int
	Searcher *github.V4Client
	Logger   log.Logger
}

func (q *repositoryQuery) Do(ctx context.Context, results chan *githubResult) {
	if q.First == 0 {
		q.First = 100
	}

	if q.Limit == 0 {
		// Default GitHub API search results limit per search.
		q.Limit = 1000
	}

	// This is kind of like a modified binary search algorithm
	// 1) We search for all repos matching the query, if we get <1000 repos we return them
	// 2) If we get >1000 repos we slap a created filter to the query, searching for all repos that match the query between From (2007) and To (Now())
	// 3) If the repos are still >1000 move the To pointer back half of the distance towards From
	// 4) Repeat step 3 until results are <1000
	// 5) At this point we have scanned all results between 2007 -> To, move the From and To pointers to the remaining unscanned timeslice (To+1 -> Now()), repeat from step 3
	// 6) Once all the repos created from 2007 to Now() are found, return
	for {
		if err := ctx.Err(); err != nil {
			results <- &githubResult{err: ctx.Err()}
			return
		}

		res, err := q.Searcher.SearchRepos(ctx, github.SearchReposParams{
			Query: q.String(),
			First: q.First,
			After: q.Cursor,
		})
		if err != nil {
			select {
			case <-ctx.Done():
			case results <- &githubResult{err: errors.Wrapf(err, "failed to search GitHub repositories with %q", q)}:
			}
			return
		}

		if res.TotalCount > q.Limit {
			q.Logger.Info(
				"repositoryQuery matched more than limit, refining",
				log.Int("limit", q.Limit),
				log.Int("resultCount", res.TotalCount),
				log.String("query", q.String()),
			)

			if q.Created == nil {
				timeNow := time.Now().UTC()
				q.Created = &dateRange{From: minCreated, To: timeNow, OriginalTo: timeNow}
			}

			if q.Refine() {
				q.Logger.Info("repositoryQuery refined", log.String("query", q.String()))
				continue
			}

			select {
			case <-ctx.Done():
			case results <- &githubResult{err: errors.Errorf("repositoryQuery %q couldn't be refined further, results would be missed", q)}:
			}
			return
		}
		q.Logger.Info("repositoryQuery matched", log.String("query", q.String()), log.Int("total", res.TotalCount), log.Int("page", len(res.Repos)))
		for i := range res.Repos {
			select {
			case <-ctx.Done():
				return
			case results <- &githubResult{repo: &res.Repos[i]}:
			}
		}

		if res.EndCursor != "" {
			q.Cursor = res.EndCursor
		} else if !q.Next() {
			return
		}
	}
}

func (s *repositoryQuery) Next() bool {
	// We are good to exit under the following conditions:
	// 1) s.Created == nil: If s.Created is nil, this means that we never refined i.e. we found all the repos matching the query.
	// 2) !s.Created.To.Before(s.Created.OriginalTo): In our search we move the To and keep From the same, To should always be sometime before it was originally,
	//    unless we finished searching, in which case we can return.
	// 3) !s.Created.To.After(s.Created.From): Sanity, we should never hit this, but in case To is somehow before From, we should exit else we will be stuck here forever.
	if s.Created == nil || !s.Created.To.Before(s.Created.OriginalTo) || !s.Created.To.After(s.Created.From) {
		return false
	}

	s.Cursor = ""

	// We just finished the timeslice of From -> To, at this point we have scanned everything in the range og:
	// minCreated -> To, now we want to find the next time slice between To+1 and Now{}
	s.Created.From = s.Created.To.Add(time.Second)
	s.Created.To = s.Created.OriginalTo

	return true
}

// Refine does one pass at refining the query to match <= 1000 repos in order
// to avoid hitting the GitHub search API 1000 results limit, which would cause
// use to miss matches.
func (s *repositoryQuery) Refine() bool {

	if s.Created.Size() < 2*time.Second {
		// Can't refine further than 1 second
		return false
	}

	// We found too many results, move the slice:
	// From -> To  ----> From -> To - (To-From)/2
	s.Created.To = s.Created.To.Add(-(s.Created.Size() / 2))
	return true
}

func (s repositoryQuery) String() string {
	q := s.Query
	if s.Created != nil {
		q += " created:" + s.Created.String()
	}
	return q
}

// regOrg is a regular expression that matches the pattern `org:<org-name>`
// `<org-name>` follows the GitHub username convention:
// - only single hyphens and alphanumeric characters allowed.
// - cannot begin/end with hyphen.
// - up to 38 characters.
var regOrg = lazyregexp.New(`^org:([a-zA-Z0-9](?:-?[a-zA-Z0-9]){0,38})$`)

// matchOrg extracts the org name from the pattern `org:<org-name>` if it exists.
func matchOrg(q string) string {
	match := regOrg.FindStringSubmatch(q)
	if len(match) != 2 {
		return ""
	}
	return match[1]
}

// listRepositoryQuery handles the `repositoryQuery` config option.
// The supported keywords to select repositories are:
// - `public`: public repositories (from endpoint: /repositories)
// - `affiliated`: repositories affiliated with client token (from endpoint: /user/repos)
// - `none`: disables `repositoryQuery`
// Inputs other than these three keywords will be queried using
// GitHub advanced repository search (endpoint: /search/repositories)
func (s *GitHubSource) listRepositoryQuery(ctx context.Context, query string, results chan *githubResult) {
	switch query {
	case "public":
		s.listPublic(ctx, results)
		return
	case "affiliated":
		s.listAffiliated(ctx, results)
		return
	case "none":
		// nothing
		return
	}

	// Special-casing for `org:<org-name>`
	// to directly use GitHub's org repo
	// list API instead of the limited
	// search API.
	//
	// If the org repo list API fails, we
	// try the user repo list API.
	if org := matchOrg(query); org != "" {
		s.listOrg(ctx, org, results)
		return
	}

	// Run the query as a GitHub advanced repository search
	// (https://github.com/search/advanced).
	s.listSearch(ctx, query, results)
}

// listAllRepositories returns the repositories from the given `orgs`, `repos`, and
// `repositoryQuery` config options excluding the ones specified by `exclude`.
func (s *GitHubSource) listAllRepositories(ctx context.Context, results chan *githubResult) {
	s.listRepos(ctx, s.config.Repos, results)

	// Admins normally add to end of lists, so end of list most likely has new
	// repos => stream them first.
	for i := len(s.config.RepositoryQuery) - 1; i >= 0; i-- {
		s.listRepositoryQuery(ctx, s.config.RepositoryQuery[i], results)
	}

	for i := len(s.config.Orgs) - 1; i >= 0; i-- {
		s.listOrg(ctx, s.config.Orgs[i], results)
	}
}

func (s *GitHubSource) getRepository(ctx context.Context, nameWithOwner string) (*github.Repository, error) {
	owner, name, err := github.SplitRepositoryNameWithOwner(nameWithOwner)
	if err != nil {
		return nil, errors.Wrapf(err, "Invalid GitHub repository: nameWithOwner="+nameWithOwner)
	}

	repo, err := s.v3Client.GetRepository(ctx, owner, name)
	if err != nil {
		return nil, err
	}

	return repo, nil
}

// fetchAllRepositoriesInBatches fetches the repositories configured in
// config.Repos in batches and adds them to the supplied set
func (s *GitHubSource) fetchAllRepositoriesInBatches(ctx context.Context, results chan *githubResult) error {
	const batchSize = 30

	// Admins normally add to end of lists, so end of list most likely has new
	// repos => stream them first.
	s.logger.Debug("fetching list of repos", log.Int("len", len(s.config.Repos)))
	for end := len(s.config.Repos); end > 0; end -= batchSize {
		if err := ctx.Err(); err != nil {
			return err
		}

		start := end - batchSize
		if start < 0 {
			start = 0
		}
		batch := s.config.Repos[start:end]

		repos, err := s.v4Client.GetReposByNameWithOwner(ctx, batch...)
		if err != nil {
			return errors.Wrap(err, "GetReposByNameWithOwner failed")
		}

		s.logger.Debug("github sync: GetReposByNameWithOwner", log.Strings("repos", batch))
		for _, r := range repos {
			if err := ctx.Err(); err != nil {
				if r != nil {
					err = errors.Wrapf(err, "context error for repository: %s", r.NameWithOwner)
				}

				results <- &githubResult{err: err}
				return err
			}

			results <- &githubResult{repo: r}
			s.logger.Debug("sent repo to result", log.String("repo", fmt.Sprintf("%+v", r)))
		}
	}

	return nil
}

func (s *GitHubSource) AffiliatedRepositories(ctx context.Context) ([]types.CodeHostRepository, error) {
	var (
		repos []*github.Repository
		page  = 1
		cost  int
		err   error
	)
	defer func() {
		remaining, reset, retry, _ := s.v3Client.RateLimitMonitor().Get()
		s.logger.Debug(
			"github sync: ListAffiliated",
			log.Int("repos", len(repos)),
			log.Int("rateLimitCost", cost),
			log.Int("rateLimitRemaining", remaining),
			log.Duration("rateLimitReset", reset),
			log.Duration("retryAfter", retry),
		)
	}()
	out := make([]types.CodeHostRepository, 0)
	hasNextPage := true
	for hasNextPage {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		var repos []*github.Repository
		if s.useGitHubApp {
			repos, hasNextPage, _, err = s.v3Client.ListInstallationRepositories(ctx, page)
		} else {
			repos, hasNextPage, _, err = s.v3Client.ListAffiliatedRepositories(ctx, github.VisibilityAll, page)
		}
		if err != nil {
			return nil, err
		}

		for _, repo := range repos {
			out = append(out, types.CodeHostRepository{
				Name:       repo.NameWithOwner,
				Private:    repo.IsPrivate,
				CodeHostID: s.svc.ID,
			})
		}
		page++
	}
	return out, nil
}
