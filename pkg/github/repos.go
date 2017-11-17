package github

import (
	"encoding/json"
	"math"
	"net/http"

	"context"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sourcegraph/go-github/github"
	gogithub "github.com/sourcegraph/go-github/github"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/actor"
	sourcegraph "sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/api/legacyerr"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/conf/feature"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/githubutil"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/rcache"
)

var (
	reposGithubPublicCache        = rcache.NewWithTTL("gh_pub", 600)
	reposGithubPublicCacheCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "src",
		Subsystem: "repos",
		Name:      "github_cache_hit",
		Help:      "Counts cache hits and misses for public github repo metadata.",
	}, []string{"type"})
)

func init() {
	prometheus.MustRegister(reposGithubPublicCacheCounter)
}

type cachedRepo struct {
	sourcegraph.Repo

	// NotFound indicates that the GitHub API returned a 404 when
	// using an Unauthed or Authed request (repo may be exist privately for another authed user).
	NotFound bool
}

var GetRepoMock func(ctx context.Context, repo string) (*sourcegraph.Repo, error)

func MockGetRepo_Return(returns *sourcegraph.Repo) {
	GetRepoMock = func(context.Context, string) (*sourcegraph.Repo, error) {
		return returns, nil
	}
}

func GetRepo(ctx context.Context, repo string) (*sourcegraph.Repo, error) {
	if GetRepoMock != nil {
		return GetRepoMock(ctx, repo)
	}

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

	// specialCasePrivate indicates whether we need special handling
	// for private repos. After Sep 20, all repos may be treated identically
	// for caching purposes (there should be no distinction between "public" and "private").
	// It is still possible in an on-prem server to have private GitHub repos in the cache.
	// In this world, the "public repo cache" contains repos which are public to the cluster.
	specialCasePrivate := !feature.Features.Sep20Auth

	if cached := getFromPublicCache(ctx, repo); cached != nil {
		reposGithubPublicCacheCounter.WithLabelValues("hit").Inc()
		if cached.NotFound {
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

	addRepoToCache := func() {
		remoteRepoCopy := *remoteRepo
		addToPublicCache(repo, &cachedRepo{Repo: remoteRepoCopy})
		reposGithubPublicCacheCounter.WithLabelValues("miss").Inc()
	}
	if specialCasePrivate {
		// We are only allowed to cache public repos.
		if !remoteRepo.Private {
			addRepoToCache()
		} else {
			reposGithubPublicCacheCounter.WithLabelValues("private").Inc()
		}
	} else {
		// No special casing; always add resolved repos to the cache.
		addRepoToCache()
	}
	return remoteRepo, nil
}

var SearchRepoMock func(ctx context.Context, query string, op *github.SearchOptions) ([]*sourcegraph.Repo, error)

func MockSearch_Return(returns []*sourcegraph.Repo) (called *bool) {
	called = new(bool)
	SearchRepoMock = func(ctx context.Context, query string, op *gogithub.SearchOptions) ([]*sourcegraph.Repo, error) {
		*called = true
		return returns, nil
	}
	return
}

func SearchRepo(ctx context.Context, query string, op *github.SearchOptions) ([]*sourcegraph.Repo, error) {
	if SearchRepoMock != nil {
		return SearchRepoMock(ctx, query, op)
	}

	res, _, err := UnauthedClient(ctx).Search.Repositories(ctx, query, op)
	if err != nil {
		return nil, err
	}
	repos := make([]*sourcegraph.Repo, 0, len(res.Repositories))
	for _, ghrepo := range res.Repositories {
		if !feature.Features.Sep20Auth {
			// 🚨 SECURITY: these search results may contain repos that a user shouldn't 🚨
			// have access to within one of their installations. Filter out all private
			// repos (private repos can be obtained with ListAllGitHubRepos)
			if *ghrepo.Private {
				continue
			}
		}
		repos = append(repos, ToRepo(&ghrepo))
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

// getFromAPI attempts to fetch a public or private repo from the GitHub API
// without use of the redis cache.
func getFromAPI(ctx context.Context, owner, repoName string) (*sourcegraph.Repo, error) {
	ghrepo, resp, err := UnauthedClient(ctx).Repositories.Get(ctx, owner, repoName)
	if err == nil {
		return ToRepo(ghrepo), nil
	}
	if resp != nil && resp.StatusCode == http.StatusNotFound {
		return nil, legacyerr.Errorf(legacyerr.NotFound, "github repo not found: %s", repoName)
	}
	return nil, err
}

func ToRepo(ghrepo *github.Repository) *sourcegraph.Repo {
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
	uintv := func(v *int) *uint {
		if v == nil || *v > math.MaxUint32 {
			return nil
		}
		u := uint(*v)
		return &u
	}
	repo := sourcegraph.Repo{
		URI:           "github.com/" + *ghrepo.FullName,
		DefaultBranch: strv(ghrepo.DefaultBranch),
		Description:   strv(ghrepo.Description),
		Language:      strv(ghrepo.Language),
		Private:       boolv(ghrepo.Private),
		Fork:          boolv(ghrepo.Fork),
		StarsCount:    uintv(ghrepo.StargazersCount),
		ForksCount:    uintv(ghrepo.ForksCount),
	}
	if ghrepo.CreatedAt != nil {
		repo.CreatedAt = &ghrepo.CreatedAt.Time
	}
	if ghrepo.UpdatedAt != nil {
		repo.UpdatedAt = &ghrepo.UpdatedAt.Time
	}
	if ghrepo.PushedAt != nil {
		repo.PushedAt = &ghrepo.PushedAt.Time
	}
	return &repo
}

var ListAccessibleReposMock func(ctx context.Context) ([]*sourcegraph.Repo, error)

func MockListAccessibleRepos_Return(returns []*sourcegraph.Repo) (called *bool) {
	called = new(bool)
	ListAccessibleReposMock = func(ctx context.Context) ([]*sourcegraph.Repo, error) {
		*called = true
		return returns, nil
	}
	return
}

// ListAccessibleRepos lists repos that are accessible to the authenticated
// user.
//
// See https://developer.github.com/v3/repos/#list-your-repositories
// for more information.
func ListAccessibleRepos(ctx context.Context) ([]*sourcegraph.Repo, error) {
	if ListAccessibleReposMock != nil {
		return ListAccessibleReposMock(ctx)
	}
	// Note: When GitHubApps is enabled a list of all repositories that are
	// accessible to a user via their installations are returned. This API does not
	// support RepositoryListOptions.
	installs, err := ListAllAccessibleInstallations(ctx)
	if err != nil {
		return nil, err
	}
	repos := []*sourcegraph.Repo{}
	for _, ins := range installs {
		ghRepos, err := ListAllAccessibleReposForInstallation(ctx, *ins.ID)
		if err != nil {
			return nil, err
		}
		for _, r := range ghRepos {
			repos = append(repos, ToRepo(r))
		}
	}
	return repos, nil
}

func ListStarredRepos(ctx context.Context, opt *gogithub.ActivityListStarredOptions) ([]*sourcegraph.Repo, error) {
	var ghRepos []*gogithub.StarredRepository
	var resp *gogithub.Response
	var err error
	// We can't get access to private starred repo's with the API. This only returns public starred repos.
	ghRepos, resp, err = UnauthedClient(ctx).Activity.ListStarred(ctx, actor.FromContext(ctx).Login, opt)
	if err != nil {
		return nil, err
	}
	if err != nil {
		return nil, checkResponse(ctx, resp, err, "github.activity.ListStarred")
	}
	var repos []*sourcegraph.Repo
	for _, ghRepo := range ghRepos {
		repos = append(repos, ToRepo(ghRepo.Repository))
	}
	return repos, nil
}

// TODO(john): the name of this method is now misleading. Before Sep 20,
// this will *only* list public repos. After Sep 20, this method may list
// private repos if the github-proxy uses a personal access token which can
// access private repos. Rename "ListReposForUser".
func ListPublicReposForUser(ctx context.Context, login string) ([]*sourcegraph.Repo, error) {
	if ListPublicReposMock != nil {
		return ListPublicReposMock(ctx, login)
	}

	beforeSep20 := !feature.Features.Sep20Auth
	var user string
	if beforeSep20 {
		// List only repos owned by the specified user.
		user = login
	}
	// Else, when user is empty, the GitHub API will list all repos visible to
	// the "currently authed user" (including public, private, & org repos
	// not owned by the user). The github-proxy will only make an authenticated
	// request if a personal access token is provided.

	ghRepos, _, err := UnauthedClient(ctx).Repositories.List(ctx, user, nil)
	if err != nil {
		return nil, err
	}
	repos := []*sourcegraph.Repo{}
	for _, r := range ghRepos {
		repos = append(repos, ToRepo(r))
	}
	return repos, nil
}

var ListPublicReposMock func(ctx context.Context, login string) ([]*sourcegraph.Repo, error)

func MockListPublicRepos_Return(returns []*sourcegraph.Repo) (called *bool) {
	called = new(bool)
	ListPublicReposMock = func(ctx context.Context, login string) ([]*sourcegraph.Repo, error) {
		*called = true
		return returns, nil
	}
	return
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
		repos, err := ListAccessibleRepos(ctx)
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
