package github

import (
	"errors"
	"fmt"
	"strconv"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sourcegraph/go-github/github"
	"golang.org/x/net/context"
	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/conf"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/githubutil"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/rcache"
	"sourcegraph.com/sqs/pbtypes"
)

var (
	reposGithubPublicCacheTTL     = conf.GetenvIntOrDefault("SG_REPOS_GITHUB_PUBLIC_CACHE_TTL_SECONDS", 600)
	reposGithubPublicCache        = rcache.New("gh_pub")
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

type Repos interface {
	Get(context.Context, string) (*sourcegraph.Repo, error)
	GetByID(context.Context, int) (*sourcegraph.Repo, error)
	ListAccessible(context.Context, *github.RepositoryListOptions) ([]*sourcegraph.Repo, error)
}

type repos struct{}

type cachedRepo struct {
	sourcegraph.Repo

	// PublicNotFound indicates that the GitHub API returned a 404 when
	// using an Unauthed request (repo may be exist privately).
	PublicNotFound bool
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
		return nil, grpc.Errorf(codes.NotFound, "github repo not found: %s", repo)
	}

	if cached := getFromCache(ctx, repo); cached != nil {
		reposGithubPublicCacheCounter.WithLabelValues("hit").Inc()
		if cached.PublicNotFound {
			return nil, grpc.Errorf(codes.NotFound, "github repo not found: %s", repo)
		}
		return &cached.Repo, nil
	}

	remoteRepo, err := getFromAPI(ctx, owner, repoName)
	if grpc.Code(err) == codes.NotFound {
		// Before we do anything, ensure we cache NotFound responses.
		// Do this if client is unauthed or authed, it's okay since we're only caching not found responses here.
		addToCache(repo, &cachedRepo{PublicNotFound: true})
		reposGithubPublicCacheCounter.WithLabelValues("public-notfound").Inc()
	}
	if err != nil {
		reposGithubPublicCacheCounter.WithLabelValues("error").Inc()
		return nil, err
	}

	// We are allowed to cache public repos
	if !remoteRepo.Private {
		addToCache(repo, &cachedRepo{Repo: *remoteRepo})
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

var errInapplicableCache = errors.New("cached value cannot be used in this scenario")

// getFromCache attempts to get a response from the redis cache.
// It returns nil error for cache-hit condition and non-nil error for cache-miss.
func getFromCache(ctx context.Context, repo string) *cachedRepo {
	var cached cachedRepo
	err := reposGithubPublicCache.Get(repo, &cached)
	if err != nil {
		return nil
	}

	// Do not use a cached NotFound if we are an authed user, since it may
	// exist as a private repo for the user.
	if client(ctx).isAuthedUser && cached.PublicNotFound {
		return nil
	}

	return &cached
}

// addToCache will cache the value for repo.
func addToCache(repo string, c *cachedRepo) {
	_ = reposGithubPublicCache.Add(repo, c, reposGithubPublicCacheTTL)
}

// getFromAPI attempts to get a response from the GitHub API without use of
// the redis cache.
func getFromAPI(ctx context.Context, owner, repoName string) (*sourcegraph.Repo, error) {
	ghrepo, resp, err := client(ctx).repos.Get(owner, repoName)
	if err != nil {
		return nil, checkResponse(ctx, resp, err, fmt.Sprintf("github.Repos.Get %q", githubutil.RepoURI(owner, repoName)))
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
			ID:      strconv.Itoa(*ghrepo.ID),
			Service: sourcegraph.Origin_GitHub,
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
		ts := pbtypes.NewTimestamp(ghrepo.UpdatedAt.Time)
		repo.UpdatedAt = &ts
	}
	if ghrepo.PushedAt != nil {
		ts := pbtypes.NewTimestamp(ghrepo.PushedAt.Time)
		repo.PushedAt = &ts
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
		repos = append(repos, toRepo(&ghRepo))
	}
	return repos, nil
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
