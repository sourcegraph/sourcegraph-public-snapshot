package github

import (
	"encoding/json"
	"fmt"
	"strconv"

	"context"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sourcegraph/go-github/github"
	gogithub "github.com/sourcegraph/go-github/github"
	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph/legacyerr"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/githubutil"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/rcache"
)

var (
	reposGithubPublicCache        = rcache.New("gh_pub", 600)
	reposGithubPublicCacheCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "src",
		Subsystem: "repos",
		Name:      "github_cache_hit",
		Help:      "Counts cache hits and misses for public github repo metadata.",
	}, []string{"type"})
	reposGitHubRequestsCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "src",
		Subsystem: "repos",
		Name:      "github_unauthed_api_requests",
		Help:      "Counts uncached requests to the GitHub API, and information on their origin if available.",
	}, []string{"source"})
)

func init() {
	prometheus.MustRegister(reposGithubPublicCacheCounter)
	prometheus.MustRegister(reposGitHubRequestsCounter)
}

type Repos interface {
	Get(context.Context, string) (*sourcegraph.Repo, error)
	GetByID(context.Context, int) (*sourcegraph.Repo, error)
	Search(ctx context.Context, query string, op *github.SearchOptions) ([]*sourcegraph.Repo, error)
	ListAccessible(context.Context, *github.RepositoryListOptions) ([]*sourcegraph.Repo, error)
	CreateHook(context.Context, string, *github.Hook) error
}

type repos struct{}

type cachedRepo struct {
	sourcegraph.Repo

	// NotFound indicates that the GitHub API returned a 404 when
	// using an Unauthed or Authed request (repo may be exist privately for another authed user).
	NotFound bool
}

var _ Repos = (*repos)(nil)

func (s *repos) Get(ctx context.Context, repo string) (*sourcegraph.Repo, error) {
	// This function is called a lot, especially on popular public
	// repos. For public repos we have the same result for everyone, so it
	// is cacheable. (Permissions can change, but we no longer store that.) But
	// for the purpose of avoiding rate limits, we set all public repos to
	// read-only permissions.
	//
	// First parse the repo url before even trying (redis) cache, since this can
	// invalide the request more quickly and cheaply.
	owner, repoName, err := githubutil.SplitRepoURI(repo)
	if err != nil {
		reposGithubPublicCacheCounter.WithLabelValues("local-error").Inc()
		return nil, legacyerr.Errorf(legacyerr.NotFound, "github repo not found: %s", repo)
	}

	if cached := getFromPublicCache(ctx, repo); cached != nil {
		reposGithubPublicCacheCounter.WithLabelValues("hit").Inc()
		if cached.NotFound {
			// The repo is in the cache but not available. If the user is authenticated,
			// request the repo from the GitHub API (but do not add it to the cache).
			if client(ctx).isAuthedUser {
				reposGithubPublicCacheCounter.WithLabelValues("authed").Inc()
				return getFromAPI(ctx, owner, repoName)
			}
			return nil, legacyerr.Errorf(legacyerr.NotFound, "github repo not found: %s", repo)
		}
		return &cached.Repo, nil
	}

	remoteRepo, err := getFromAPI(ctx, owner, repoName)
	if legacyerr.ErrCode(err) == legacyerr.NotFound {
		// Before we do anything, ensure we cache NotFound responses.
		// Do this if client is unauthed or authed, it's okay since we're only caching not found responses here.
		addToPublicCache(repo, &cachedRepo{NotFound: true})
		reposGithubPublicCacheCounter.WithLabelValues("public-notfound").Inc()
	}
	if err != nil {
		reposGithubPublicCacheCounter.WithLabelValues("error").Inc()
		return nil, err
	}

	// We are only allowed to cache public repos.
	if !remoteRepo.Private {
		remoteRepoCopy := *remoteRepo
		if client(ctx).isAuthedUser {
			// Repos' Permissions fields may differ for the different authed users.
			// When adding to public cache, reset them to defaults.
			remoteRepoCopy.Permissions = &sourcegraph.RepoPermissions{Pull: true, Push: false, Admin: false}
		}
		addToPublicCache(repo, &cachedRepo{Repo: remoteRepoCopy})
		reposGithubPublicCacheCounter.WithLabelValues("miss").Inc()
	} else {
		reposGithubPublicCacheCounter.WithLabelValues("private").Inc()
	}
	return remoteRepo, nil
}

func (s *repos) GetByID(ctx context.Context, id int) (*sourcegraph.Repo, error) {
	ghrepo, resp, err := client(ctx).repos.GetByID(id)
	if err != nil {
		return nil, checkResponse(ctx, resp, err, fmt.Sprintf("github.Repos.GetByID #%d", id))
	}
	return toRepo(ghrepo), nil
}

func (s *repos) Search(ctx context.Context, query string, op *github.SearchOptions) ([]*sourcegraph.Repo, error) {
	res, _, err := client(ctx).search.Repositories(query, op)
	if err != nil {
		return nil, err
	}
	repos := make([]*sourcegraph.Repo, 0, len(res.Repositories))
	for _, ghrepo := range res.Repositories {
		repos = append(repos, toRepo(&ghrepo))
	}
	return repos, nil
}

// getFromPublicCache attempts to get a response from the redis cache.
// It returns nil error for cache-hit condition and non-nil error for cache-miss.
func getFromPublicCache(ctx context.Context, repo string) *cachedRepo {
	b, ok := reposGithubPublicCache.Get(repo)
	if !ok {
		return nil
	}

	var cached cachedRepo
	if err := json.Unmarshal(b, &cached); err != nil {
		return nil
	}

	return &cached
}

// addToPublicCache will cache the value for repo.
func addToPublicCache(repo string, c *cachedRepo) {
	b, err := json.Marshal(c)
	if err != nil {
		return
	}
	reposGithubPublicCache.Set(repo, b)
}

var GitHubTrackingContextKey = &struct{ name string }{"GitHubTrackingSource"}

// getFromAPI attempts to get a response from the GitHub API without use of
// the redis cache.
func getFromAPI(ctx context.Context, owner, repoName string) (*sourcegraph.Repo, error) {
	ghrepo, resp, err := client(ctx).repos.Get(owner, repoName)
	if err != nil {
		return nil, checkResponse(ctx, resp, err, fmt.Sprintf("github.Repos.Get %q", githubutil.RepoURI(owner, repoName)))
	}
	// Temporary: Track where anonymous requests are coming from that don't hit the cache.
	if _, ok := resp.Header["X-From-Cache"]; !client(ctx).isAuthedUser && !ok {
		src, ok := ctx.Value(GitHubTrackingContextKey).(string)
		if !ok {
			src = "unknown"
		}
		reposGitHubRequestsCounter.WithLabelValues(src).Inc()
	}
	return toRepo(ghrepo), nil
}

func toRepo(ghrepo *github.Repository) *sourcegraph.Repo {
	strv := func(s *string) string {
		if s == nil {
			return ""
		}
		return *s
	}
	boolv := func(b *bool) bool {
		if b == nil {
			return false
		}
		return *b
	}
	repo := sourcegraph.Repo{
		URI: "github.com/" + *ghrepo.FullName,
		Origin: &sourcegraph.Origin{
			ID:         strconv.Itoa(*ghrepo.ID),
			Service:    sourcegraph.Origin_GitHub,
			APIBaseURL: "https://api.github.com",
		},
		Name:          *ghrepo.Name,
		HTTPCloneURL:  strv(ghrepo.CloneURL),
		DefaultBranch: strv(ghrepo.DefaultBranch),
		Description:   strv(ghrepo.Description),
		Language:      strv(ghrepo.Language),
		Private:       boolv(ghrepo.Private),
		Fork:          boolv(ghrepo.Fork),
		Mirror:        ghrepo.MirrorURL != nil,
	}
	if ghrepo.Owner != nil {
		repo.Owner = strv(ghrepo.Owner.Login)
	}
	if ghrepo.UpdatedAt != nil {
		repo.UpdatedAt = &ghrepo.UpdatedAt.Time
	}
	if ghrepo.PushedAt != nil {
		repo.PushedAt = &ghrepo.PushedAt.Time
	}
	if pp := ghrepo.Permissions; pp != nil {
		p := *pp
		repo.Permissions = &sourcegraph.RepoPermissions{
			Pull:  p["pull"],
			Push:  p["push"],
			Admin: p["admin"],
		}
	}
	return &repo
}

// ListAccessible lists repos that are accessible to the authenticated
// user.
//
// See https://developer.github.com/v3/repos/#list-your-repositories
// for more information.
func (s *repos) ListAccessible(ctx context.Context, opt *github.RepositoryListOptions) ([]*sourcegraph.Repo, error) {
	ghRepos, resp, err := client(ctx).repos.List("", opt)
	if err != nil {
		return nil, checkResponse(ctx, resp, err, "github.Repos.ListAccessible")
	}

	var repos []*sourcegraph.Repo
	for _, ghRepo := range ghRepos {
		repos = append(repos, toRepo(ghRepo))
	}
	return repos, nil
}

// CreateHook creates a Hook for the specified repository.
//
// See http://developer.github.com/v3/repos/hooks/#create-a-hook
// for more information.
func (s *repos) CreateHook(ctx context.Context, repo string, hook *github.Hook) error {
	owner, repoName, err := githubutil.SplitRepoURI(repo)
	if err != nil {
		return legacyerr.Errorf(legacyerr.NotFound, "github repo not found: %s", repo)
	}
	_, resp, err := client(ctx).repos.CreateHook(owner, repoName, hook)
	if err != nil {
		return checkResponse(ctx, resp, err, fmt.Sprintf("github.Repos.CreateHook %q", githubutil.RepoURI(owner, repoName)))
	}
	return nil
}

// WithRepos returns a copy of parent with the given GitHub Repos service.
func WithRepos(parent context.Context, s Repos) context.Context {
	return context.WithValue(parent, reposKey, s)
}

// ReposFromContext gets the context's GitHub Repos service.
// If the value is not present, it creates a temporary one.
func ReposFromContext(ctx context.Context) Repos {
	s, ok := ctx.Value(reposKey).(Repos)
	if !ok || s == nil {
		return &repos{}
	}
	return s
}

func ListStarredRepos(ctx context.Context, opt *gogithub.ActivityListStarredOptions) ([]*sourcegraph.Repo, error) {
	ghRepos, resp, err := client(ctx).activity.ListStarred("", opt)
	if err != nil {
		return nil, checkResponse(ctx, resp, err, "github.activity.ListStarred")
	}
	var repos []*sourcegraph.Repo
	for _, ghRepo := range ghRepos {
		repos = append(repos, toRepo(ghRepo.Repository))
	}
	return repos, nil
}

// ListAllGitHubRepos lists all GitHub repositories that fit the
// criteria that are accessible to the currently authenticated user.
// It's a convenience wrapper around Repos.ListAccessible, since there
// are a few places where we want a list of *all* repositories
// accessible to a user.
func ListAllGitHubRepos(ctx context.Context, op_ *gogithub.RepositoryListOptions) ([]*sourcegraph.Repo, error) {
	const perPage = 100
	const maxPage = 1000
	op := *op_
	op.PerPage = perPage

	if !HasAuthedUser(ctx) {
		return nil, nil
	}
	var allRepos []*sourcegraph.Repo
	for page := 1; page <= maxPage; page++ {
		op.Page = page
		repos, err := ReposFromContext(ctx).ListAccessible(ctx, &op)
		if err != nil {
			return nil, err
		}
		allRepos = append(allRepos, repos...)
		if len(repos) < perPage {
			break
		}
	}
	return allRepos, nil
}

func ListGitHubContributors(ctx context.Context, repo *sourcegraph.Repo, opt *gogithub.ListContributorsOptions) ([]*sourcegraph.Contributor, error) {
	ghContributors, resp, err := client(ctx).repos.ListContributors(repo.Owner, repo.Name, opt)
	if err != nil {
		return nil, checkResponse(ctx, resp, err, "github.repos.ListContributors")
	}

	var contribs []*sourcegraph.Contributor
	for _, ghContrib := range ghContributors {
		contribs = append(contribs, toContributor(ghContrib))
	}

	return contribs, nil
}

func toContributor(ghContrib *github.Contributor) *sourcegraph.Contributor {
	strv := func(s *string) string {
		if s == nil {
			return ""
		}
		return *s
	}
	intv := func(i *int) int {
		if i == nil {
			return 0
		}
		return *i
	}

	contrib := sourcegraph.Contributor{
		Login:         strv(ghContrib.Login),
		AvatarURL:     strv(ghContrib.AvatarURL),
		Contributions: intv(ghContrib.Contributions),
	}

	return &contrib
}
