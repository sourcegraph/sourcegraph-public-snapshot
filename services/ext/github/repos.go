package github

import (
	"encoding/json"
	"fmt"
	"strconv"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	"context"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sourcegraph/go-github/github"
	gogithub "github.com/sourcegraph/go-github/github"
	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/conf"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/githubutil"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/rcache"
	"sourcegraph.com/sqs/pbtypes"
)

var (
	reposGithubPublicCache        = rcache.New("gh_pub", conf.GetenvIntOrDefault("SG_REPOS_GITHUB_PUBLIC_CACHE_TTL_SECONDS", 600))
	reposGithubPublicCacheCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "src",
		Subsystem: "repos",
		Name:      "github_cache_hit",
		Help:      "Counts cache hits and misses for public github repo metadata.",
	}, []string{"type"})
	reposGithubConcurrent = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "src",
		Subsystem: "github",
		Name:      "repo_concurrent",
		Help:      "Number of concurrent calls to GitHub's Repo API.",
	})
)

func init() {
	prometheus.MustRegister(reposGithubPublicCacheCounter)
	prometheus.MustRegister(reposGithubConcurrent)
}

type Repos interface {
	Get(context.Context, string) (*sourcegraph.Repo, error)
	GetByID(context.Context, int) (*sourcegraph.Repo, error)
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

// Bound the number of parallel requests to github. This is a hotfix to avoid
// triggering github abuse. Default is really low to try to recover from abuse detection.
var sem = make(chan bool, conf.GetenvIntOrDefault("SG_GITHUB_CONCURRENT", 1))

func lock() {
	reposGithubConcurrent.Inc()
	sem <- true
}

func unlock() {
	<-sem
	reposGithubConcurrent.Dec()
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

	if cached := getFromPublicCache(ctx, repo); cached != nil {
		reposGithubPublicCacheCounter.WithLabelValues("hit").Inc()
		if cached.NotFound {
			// The repo is in the cache but not available. If the user is authenticated,
			// request the repo from the GitHub API (but do not add it to the cache).
			if client(ctx).isAuthedUser {
				reposGithubPublicCacheCounter.WithLabelValues("authed").Inc()
				lock()
				defer unlock()
				return getFromAPI(ctx, owner, repoName)
			}
			return nil, grpc.Errorf(codes.NotFound, "github repo not found: %s", repo)
		}
		return &cached.Repo, nil
	}

	lock()
	defer unlock()
	remoteRepo, err := getFromAPI(ctx, owner, repoName)
	if grpc.Code(err) == codes.NotFound {
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
	lock()
	defer unlock()
	ghrepo, resp, err := client(ctx).repos.GetByID(id)
	if err != nil {
		return nil, checkResponse(ctx, resp, err, fmt.Sprintf("github.Repos.GetByID #%d", id))
	}
	return toRepo(ghrepo), nil
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
		ts := pbtypes.NewTimestamp(ghrepo.UpdatedAt.Time)
		repo.UpdatedAt = &ts
	}
	if ghrepo.PushedAt != nil {
		ts := pbtypes.NewTimestamp(ghrepo.PushedAt.Time)
		repo.PushedAt = &ts
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
	lock()
	defer unlock()
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
	lock()
	defer unlock()
	owner, repoName, err := githubutil.SplitRepoURI(repo)
	if err != nil {
		return grpc.Errorf(codes.NotFound, "github repo not found: %s", repo)
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
