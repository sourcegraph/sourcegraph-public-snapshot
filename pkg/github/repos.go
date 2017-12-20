package github

import (
	"encoding/json"
	"math"
	"net/http"

	"context"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sourcegraph/go-github/github"
	sourcegraph "sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
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

	remoteRepoCopy := *remoteRepo
	addToPublicCache(repo, &cachedRepo{Repo: remoteRepoCopy})
	reposGithubPublicCacheCounter.WithLabelValues("miss").Inc()

	return remoteRepo, nil
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
		URI:         "github.com/" + *ghrepo.FullName,
		Description: strv(ghrepo.Description),
		Language:    strv(ghrepo.Language),
		Private:     boolv(ghrepo.Private),
		Fork:        boolv(ghrepo.Fork),
		StarsCount:  uintv(ghrepo.StargazersCount),
		ForksCount:  uintv(ghrepo.ForksCount),
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
