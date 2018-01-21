package github

import (
	"encoding/json"
	"net/http"

	"context"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sourcegraph/go-github/github"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/api/legacyerr"
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
	github.Repository

	// NotFound indicates that the GitHub API returned a 404 when
	// using an Unauthed or Authed request (repo may be exist privately for another authed user).
	NotFound bool
}

var GetRepoMock func(ctx context.Context, repo string) (*github.Repository, error)

func MockGetRepo_Return(returns *github.Repository) {
	GetRepoMock = func(context.Context, string) (*github.Repository, error) {
		return returns, nil
	}
}

func GetRepo(ctx context.Context, repo string) (*github.Repository, error) {
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
			return nil, legacyerr.Errorf(legacyerr.NotFound, "github repo not found: %s", repo)
		}
		return &cached.Repository, nil
	}

	ghrepo, err := getFromAPI(ctx, owner, repoName)
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

	ghrepoCopy := *ghrepo
	addToPublicCache(repo, &cachedRepo{Repository: ghrepoCopy})
	reposGithubPublicCacheCounter.WithLabelValues("miss").Inc()

	return ghrepo, nil
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
func getFromAPI(ctx context.Context, owner, repoName string) (*github.Repository, error) {
	ghrepo, resp, err := UnauthedClient(ctx).Repositories.Get(ctx, owner, repoName)
	if err == nil {
		return ghrepo, nil
	}
	if resp != nil && resp.StatusCode == http.StatusNotFound {
		return nil, legacyerr.Errorf(legacyerr.NotFound, "github repo not found: %s", repoName)
	}
	return nil, err
}
