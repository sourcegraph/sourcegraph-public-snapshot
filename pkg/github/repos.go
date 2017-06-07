package github

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

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
	// PreferRawGit, when set to true, will make this package prefer using raw git functionality
	// when fetching single repositories, rather than hitting the GitHub API. Currently used by the
	// indexer, which otherwise would likely hit the GitHub API limit and doesn't even have GitHub
	// auth credentials. We will soon replace `getFromAPI` with `getFromGit`, and this package
	// variable can be removed when that's done.
	PreferRawGit bool

	reposGithubPublicCache        = rcache.NewWithTTL("gh_pub", 600)
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

	// DangerouslySkipAPIPermissionCheck, if true, eliminates most (but not necessarily all) calls to the GitHub API.
	// WARNING: this will also disable permissions checks that depend on the GitHub API.
	// This should only be set to true in the local installer case.
	DangerouslySkipAPIPermissionCheck bool
)

func init() {
	prometheus.MustRegister(reposGithubPublicCacheCounter)
	prometheus.MustRegister(reposGitHubRequestsCounter)
	if noGitHubAPI, err := strconv.ParseBool(os.Getenv("NO_GITHUB_API")); err == nil {
		log.Printf("detected NO_GITHUB_API=%v", noGitHubAPI)
		DangerouslySkipAPIPermissionCheck = noGitHubAPI
	}
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

	if cached := getFromPublicCache(ctx, repo); cached != nil {
		reposGithubPublicCacheCounter.WithLabelValues("hit").Inc()
		if cached.NotFound {
			// The repo is in the cache but not available. If the user is authenticated,
			// request the repo from the GitHub API (but do not add it to the cache).
			if HasAuthedUser(ctx) {
				reposGithubPublicCacheCounter.WithLabelValues("authed").Inc()
				return getFromAPI(ctx, owner, repoName)
			}
			return nil, legacyerr.Errorf(legacyerr.NotFound, "github repo not found: %s", repo)
		}
		return &cached.Repo, nil
	}

	var remoteRepo *sourcegraph.Repo
	if !PreferRawGit { // normally prefer getting from GitHub API if we have authed credentials
		remoteRepo, err = getFromAPI(ctx, owner, repoName)
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
	} else { // fall back to getting repo via raw git
		remoteRepo, err = getFromGit(ctx, owner, repoName)
		if err != nil {
			reposGithubPublicCacheCounter.WithLabelValues("error").Inc()
			return nil, err
		}
	}

	// We are only allowed to cache public repos.
	if !remoteRepo.Private {
		remoteRepoCopy := *remoteRepo
		addToPublicCache(repo, &cachedRepo{Repo: remoteRepoCopy})
		reposGithubPublicCacheCounter.WithLabelValues("miss").Inc()
	} else {
		reposGithubPublicCacheCounter.WithLabelValues("private").Inc()
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

	if feature.Features.GitHubApps {
		installs, err := ListAllAccessibleInstallations(ctx)
		if err != nil {
			return nil, err
		}
		for _, ins := range installs {
			cl, err := InstallationClient(ctx, *ins.ID)
			if err != nil {
				return nil, err
			}
			res, _, err := cl.Search.Repositories(ctx, query, op)
			if err != nil {
				return nil, err
			}
			repos := make([]*sourcegraph.Repo, 0, len(res.Repositories))
			for _, ghrepo := range res.Repositories {
				// 🚨 SECURITY: these search results may contain repos that a user shouldn't 🚨
				// have access to within one of their installations. Filter out all private
				// repos (private repos can be obtained with ListAllGitHubRepos)
				if *ghrepo.Private {
					continue
				}
				repos = append(repos, ToRepo(&ghrepo))
			}
			return repos, nil
		}
	}
	res, _, err := Client(ctx).Search.Repositories(ctx, query, op)
	if err != nil {
		return nil, err
	}
	repos := make([]*sourcegraph.Repo, 0, len(res.Repositories))
	for _, ghrepo := range res.Repositories {
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

var lsRemoteRefMatcher = regexp.MustCompile(`^ref:\s+refs/heads/([^\s]+)\s+HEAD\n`)

// getFromGit fetches a remote GitHub repository using git operations only. Curently this only works
// for publicly accessible repositories.  At some future point, we may consider deleting gitFromAPI
// and using getFromGit exclusively, as it works for any generic git repository and doesn't count
// against the GitHub API rate limit.
func getFromGit(ctx context.Context, owner, repoName string) (*sourcegraph.Repo, error) {
	cmd := exec.CommandContext(ctx, "git", "ls-remote", "--symref", fmt.Sprintf("https://github.com/%s/%s", owner, repoName), "HEAD")
	cmd.Stdin = nil
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	m := lsRemoteRefMatcher.FindStringSubmatch(string(out))
	if len(m) < 2 {
		return nil, fmt.Errorf("couldn't parse HEAD ref from: %q", string(out))
	}
	defaultBranch := m[1]
	return &sourcegraph.Repo{
		URI:           fmt.Sprintf("github.com/%s/%s", owner, repoName),
		Owner:         owner,
		Name:          repoName,
		DefaultBranch: defaultBranch,
		Private:       false,
	}, nil
}

// getFromAPI attempts to get a response from the GitHub API without use of
// the redis cache.
func getFromAPI(ctx context.Context, owner, repoName string) (*sourcegraph.Repo, error) {
	if feature.Features.GitHubApps {
		// The current GitHub App API only allows users to access their repos by
		// listing them via their installations. Check each installation and find the
		// repo we're looking for.
		installs, err := ListAllAccessibleInstallations(ctx)
		if err != nil {
			return nil, err
		}
		for _, ins := range installs {
			repos, err := ListAllAccessibleReposForInstallation(ctx, *ins.ID)
			if err != nil {
				return nil, err
			}
			for _, r := range repos {
				if *r.Name == repoName && *r.Owner.Login == owner {
					return ToRepo(r), nil
				}
			}
		}
		return nil, errors.New("repo not found")
	} else {
		ghrepo, resp, err := Client(ctx).Repositories.Get(ctx, owner, repoName)
		if err != nil {
			return nil, checkResponse(ctx, resp, err, fmt.Sprintf("github.Repos.Get %q", githubutil.RepoURI(owner, repoName)))
		}
		// Temporary: Track where anonymous requests are coming from that don't hit the cache.
		if _, ok := resp.Header["X-From-Cache"]; !HasAuthedUser(ctx) && !ok {
			src, ok := ctx.Value(GitHubTrackingContextKey).(string)
			if !ok {
				src = "unknown"
			}
			reposGitHubRequestsCounter.WithLabelValues(src).Inc()
		}
		return ToRepo(ghrepo), nil
	}
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
	repo := sourcegraph.Repo{
		URI:           "github.com/" + *ghrepo.FullName,
		Name:          *ghrepo.Name,
		DefaultBranch: strv(ghrepo.DefaultBranch),
		Description:   strv(ghrepo.Description),
		Language:      strv(ghrepo.Language),
		Private:       boolv(ghrepo.Private),
		Fork:          boolv(ghrepo.Fork),
	}
	if ghrepo.Owner != nil {
		repo.Owner = strv(ghrepo.Owner.Login)
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

var ListAccessibleReposMock func(ctx context.Context, opt *github.RepositoryListOptions) ([]*sourcegraph.Repo, error)

func MockListAccessibleRepos_Return(returns []*sourcegraph.Repo) (called *bool) {
	called = new(bool)
	ListAccessibleReposMock = func(ctx context.Context, opt *gogithub.RepositoryListOptions) ([]*sourcegraph.Repo, error) {
		*called = true

		if opt != nil && opt.Page > 1 {
			return nil, nil
		}

		return returns, nil
	}
	return
}

// ListAccessibleRepos lists repos that are accessible to the authenticated
// user.
//
// See https://developer.github.com/v3/repos/#list-your-repositories
// for more information.
func ListAccessibleRepos(ctx context.Context, opt *github.RepositoryListOptions) ([]*sourcegraph.Repo, error) {
	// Note: When GitHubApps is enabled a list of all repositories that are
	// accessible to a user via their installations are returned. This API does not
	// support RepositoryListOptions.
	//
	// TODO: remove unused "opt" agrument when removing this feature flag.
	if feature.Features.GitHubApps {
		var repos []*sourcegraph.Repo
		installs, err := ListAllAccessibleInstallations(ctx)
		if err != nil {
			return nil, err
		}
		for _, ins := range installs {
			ghRepos, err := ListAllAccessibleReposForInstallation(ctx, *ins.ID)
			if err != nil {
				return nil, err
			}
			repos = make([]*sourcegraph.Repo, 0, len(ghRepos))
			for _, r := range ghRepos {
				repos = append(repos, ToRepo(r))
			}
		}
		return repos, nil
	}
	if ListAccessibleReposMock != nil {
		return ListAccessibleReposMock(ctx, opt)
	}

	ghRepos, resp, err := Client(ctx).Repositories.List(ctx, "", opt)
	if err != nil {
		return nil, checkResponse(ctx, resp, err, "github.Repos.ListAccessible")
	}

	var repos []*sourcegraph.Repo
	for _, ghRepo := range ghRepos {
		repos = append(repos, ToRepo(ghRepo))
	}
	return repos, nil
}

func ListStarredRepos(ctx context.Context, opt *gogithub.ActivityListStarredOptions) ([]*sourcegraph.Repo, error) {
	var ghRepos []*gogithub.StarredRepository
	var resp *gogithub.Response
	var err error
	if feature.Features.GitHubApps {
		// We can't get access to private starred repo's with the API. This only returns public starred repos.
		ghRepos, resp, err = gogithub.NewClient(nil).Activity.ListStarred(ctx, actor.FromContext(ctx).Login, opt)
		if err != nil {
			return nil, err
		}
	} else {
		ghRepos, resp, err = Client(ctx).Activity.ListStarred(ctx, "", opt)
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
		repos, err := ListAccessibleRepos(ctx, &op)
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

// IsRepoAndShouldCheckPermissions returns true if we can infer from the repository URI that the
// repository is hosted on github.com (and DangerouslySkipAPIPermissionCheck is false).
func IsRepoAndShouldCheckPermissions(uri string) bool {
	return strings.HasPrefix(strings.ToLower(uri), "github.com/") && !DangerouslySkipAPIPermissionCheck
}
