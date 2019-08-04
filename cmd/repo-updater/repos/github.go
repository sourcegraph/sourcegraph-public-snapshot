package repos

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"regexp"
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
	svc             *ExternalService
	config          *schema.GitHubConnection
	exclude         map[string]bool
	excludePatterns []*regexp.Regexp
	githubDotCom    bool
	baseURL         *url.URL
	client          *github.Client
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

	opts := []httpcli.Opt{
		// Use a 30s timeout to avoid running into EOF errors, because GitHub
		// closes idle connections after 60s
		httpcli.NewIdleConnTimeoutOpt(30 * time.Second),
	}

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

		if r.Id != "" {
			exclude[r.Id] = true
		}

		if r.Pattern != "" {
			re, err := regexp.Compile(r.Pattern)
			if err != nil {
				return nil, err
			}
			excludePatterns = append(excludePatterns, re)
		}
	}

	return &GithubSource{
		svc:              svc,
		config:           c,
		exclude:          exclude,
		excludePatterns:  excludePatterns,
		baseURL:          baseURL,
		githubDotCom:     githubDotCom,
		client:           github.NewClient(apiURL, c.Token, cli),
		searchClient:     github.NewClient(apiURL, c.Token, cli),
		originalHostname: originalHostname,
	}, nil
}

func (s GithubSource) Client() *github.Client { return s.client }

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
		ExternalRepo: github.ExternalRepoSpec(r, *s.baseURL),
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
	if s.exclude[strings.ToLower(r.NameWithOwner)] || s.exclude[r.ID] {
		return true
	}

	for _, re := range s.excludePatterns {
		if re.MatchString(r.NameWithOwner) {
			return true
		}
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
func (s *GithubSource) paginate(ctx context.Context, pager repositoryPager) (map[int64]*github.Repository, error) {
	set := make(map[int64]*github.Repository)

	hasNext := true
	for page := 1; hasNext; page++ {
		if err := ctx.Err(); err != nil {
			return set, err
		}

		var pageRepos []*github.Repository
		var cost int
		var err error
		pageRepos, hasNext, cost, err = pager(page)
		if err != nil {
			return set, err
		}

		for _, r := range pageRepos {
			set[r.DatabaseID] = r
		}

		if hasNext && cost > 0 {
			time.Sleep(s.client.RateLimit.RecommendedWaitForBackgroundOp(cost))
		}
	}

	return set, nil
}

// listOrg handles the `org` config option.
// It returns all the repositories belonging to the given organization
// by hitting the /orgs/:org/repos endpoint.
func (s *GithubSource) listOrg(ctx context.Context, org string) (map[int64]*github.Repository, error) {
	return s.paginate(ctx, func(page int) (repos []*github.Repository, hasNext bool, cost int, err error) {
		defer func() {
			remaining, reset, retry, _ := s.client.RateLimit.Get()
			log15.Debug(
				"github sync: ListOrgRepositories",
				"repos", len(repos),
				"rateLimitCost", cost,
				"rateLimitRemaining", remaining,
				"rateLimitReset", reset,
				"retryAfter", retry,
			)
		}()
		return s.client.ListOrgRepositories(ctx, org, page)
	})
}

// listRepos returns the valid repositories from the given list of repository names.
// This is done by hitting the /repos/:owner/:name endpoint for each of the given
// repository names.
func (s *GithubSource) listRepos(ctx context.Context, repos []string) (map[int64]*github.Repository, error) {
	set := make(map[int64]*github.Repository)
	if err := s.fetchAllRepositoriesInBatches(ctx, set); err == nil {
		return set, nil
	} else {
		// The way we fetch repositories in batches through the GraphQL API -
		// using aliases to query multiple repositories in one query - is
		// currently "undefined behaviour". Very rarely but unreproducibly it
		// resulted in EOF errors while testing. And since we rely on fetching
		// to work, we fall back to the (slower) sequential fetching in case we
		// run into an GraphQL API error
		log15.Warn("github sync: fetching in batches failed. falling back to sequential fetch", "error", err)
	}

	for _, nameWithOwner := range repos {
		if err := ctx.Err(); err != nil {
			return set, err
		}

		owner, name, err := github.SplitRepositoryNameWithOwner(nameWithOwner)
		if err != nil {
			return set, errors.New("Invalid GitHub repository: nameWithOwner=" + nameWithOwner)
		}
		var repo *github.Repository
		repo, err = s.client.GetRepository(ctx, owner, name)
		if err != nil {
			// TODO(tsenart): When implementing dry-run, reconsider alternatives to return
			// 404 errors on external service config validation.
			if github.IsNotFound(err) {
				log15.Warn("skipping missing github.repos entry:", "name", nameWithOwner, "err", err)
				continue
			}
			return set, errors.Wrapf(err, "Error getting GitHub repository: nameWithOwner=%s", nameWithOwner)
		}
		log15.Debug("github sync: GetRepository", "repo", repo.NameWithOwner)
		set[repo.DatabaseID] = repo
		time.Sleep(s.client.RateLimit.RecommendedWaitForBackgroundOp(1)) // 0-duration sleep unless nearing rate limit exhaustion
	}

	return set, nil
}

// listPublic handles the `public` keyword of the `repositoryQuery` config option.
// It returns the public repositories listed on the /repositories endpoint.
func (s *GithubSource) listPublic(ctx context.Context) (map[int64]*github.Repository, error) {
	set := make(map[int64]*github.Repository)
	if s.githubDotCom {
		return set, errors.New(`unsupported configuration "public" for "repositoryQuery" for github.com`)
	}
	var sinceRepoID int64
	for {
		if err := ctx.Err(); err != nil {
			return set, err
		}

		repos, err := s.client.ListPublicRepositories(ctx, sinceRepoID)
		if err != nil {
			return set, errors.Wrapf(err, "failed to list public repositories: sinceRepoID=%d", sinceRepoID)
		}
		if len(repos) == 0 {
			return set, nil
		}
		log15.Debug("github sync public", "repos", len(repos), "error", err)
		for _, r := range repos {
			set[r.DatabaseID] = r
			if sinceRepoID < r.DatabaseID {
				sinceRepoID = r.DatabaseID
			}
		}
	}
}

// listAffiliated handles the `affiliated` keyword of the `repositoryQuery` config option.
// It returns the repositories affiliated with the client token by hitting the /user/repos
// endpoint.
//
// Affiliation is present if the user: (1) owns the repo, (2) is apart of an org that
// the repo belongs to, or (3) is a collaborator.
func (s *GithubSource) listAffiliated(ctx context.Context) (map[int64]*github.Repository, error) {
	return s.paginate(ctx, func(page int) (repos []*github.Repository, hasNext bool, cost int, err error) {
		defer func() {
			remaining, reset, retry, _ := s.client.RateLimit.Get()
			log15.Debug(
				"github sync: ListAffiliated",
				"repos", len(repos),
				"rateLimitCost", cost,
				"rateLimitRemaining", remaining,
				"rateLimitReset", reset,
				"retryAfter", retry,
			)
		}()
		return s.client.ListUserRepositories(ctx, page)
	})
}

// listSearch handles the `repositoryQuery` config option when a keyword is not present.
// It returns the repositories resulting from from GitHub's advanced repository search
// by hitting the /search/repositories endpoint.
func (s *GithubSource) listSearch(ctx context.Context, query string) (map[int64]*github.Repository, error) {
	return s.paginate(ctx, func(page int) ([]*github.Repository, bool, int, error) {
		reposPage, err := s.searchClient.ListRepositoriesForSearch(ctx, query, page)
		if err != nil {
			return nil, false, 0, errors.Wrapf(err, "failed to list GitHub repositories for search: page=%d, searchString=%q", page, query)
		}

		if reposPage.TotalCount > 1000 {
			// GitHub's advanced repository search will only
			// return 1000 results. We specially handle this case
			// to ensure the admin gets a detailed error
			// message. https://github.com/sourcegraph/sourcegraph/issues/2562
			return nil, false, 0, errors.Errorf(`repositoryQuery %q would return %d results. GitHub's Search API only returns up to 1000 results. Please adjust your repository query into multiple queries such that each returns less than 1000 results. For example: {"repositoryQuery": %s}`, query, reposPage.TotalCount, exampleRepositoryQuerySplit(query))
		}

		repos, hasNext := reposPage.Repos, reposPage.HasNextPage
		remaining, reset, retry, ok := s.searchClient.RateLimit.Get()
		log15.Debug(
			"github sync: ListRepositoriesForSearch",
			"searchString", query,
			"repos", len(repos),
			"rateLimitRemaining", remaining,
			"rateLimitReset", reset,
			"retryAfter", retry,
		)

		// GitHub search has vastly different rate limits to
		// the normal GitHub API (30req/m vs
		// 5000req/h). RecommendedWaitForBackgroundOp has
		// heuristics tuned for the normal API, part of which
		// is to not sleep if we have ample rate limit left.
		//
		// So we only let the heuristic kick in if we have
		// less than 5 requests left.
		var cost int
		if retry > 0 || (ok && remaining < 5) {
			cost = 1
		}

		return repos, hasNext, cost, nil
	})
}

// regOrg is a regular expression that matches the pattern `org:<org-name>`
// `<org-name>` follows the GitHub username convention:
// - only single hyphens and alphanumeric characters allowed.
// - cannot begin/end with hyphen.
// - up to 38 characters.
var regOrg = regexp.MustCompile(`^org:([a-zA-Z0-9](?:-?[a-zA-Z0-9]){0,38})$`)

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
func (s *GithubSource) listRepositoryQuery(ctx context.Context, query string) (map[int64]*github.Repository, error) {
	switch query {
	case "public":
		return s.listPublic(ctx)
	case "affiliated":
		return s.listAffiliated(ctx)
	case "none":
		// nothing
		return nil, nil
	}

	// Special-casing for `org:<org-name>`
	// to directly use GitHub's org repo
	// list API instead of the limited
	// search API.
	if org := matchOrg(query); org != "" {
		return s.listOrg(ctx, org)
	}

	// Run the query as a GitHub advanced repository search
	// (https://github.com/search/advanced).
	return s.listSearch(ctx, query)
}

// listAllRepositories returns the repositories from the given `orgs`, `repos`, and
// `repositoryQuery` config options excluding the ones specified by `exclude`.
func (s *GithubSource) listAllRepositories(ctx context.Context) ([]*github.Repository, error) {
	set := make(map[int64]*github.Repository)
	errs := new(multierror.Error)

	for _, query := range s.config.RepositoryQuery {
		list, err := s.listRepositoryQuery(ctx, query)
		if err != nil {
			errs = multierror.Append(errs, err)
			continue
		}
		for id, r := range list {
			set[id] = r
		}
	}

	list, err := s.listRepos(ctx, s.config.Repos)
	if err != nil {
		errs = multierror.Append(errs, err)
	}
	for id, r := range list {
		set[id] = r
	}

	for _, org := range s.config.Orgs {
		list, err := s.listOrg(ctx, org)
		if err != nil {
			errs = multierror.Append(errs, errors.Wrapf(err, "failed to list organization %s repos", org))
			continue
		}
		for id, r := range list {
			set[id] = r
		}
	}

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
