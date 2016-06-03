package github

import (
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	"gopkg.in/inconshreveable/log15.v2"

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

type Repos struct{}

func (s *Repos) Get(ctx context.Context, repo string) (*sourcegraph.RemoteRepo, error) {
	// This function is called a lot, especially on popular public
	// repos. For public repos we have the same result for everyone, so it
	// is cacheable. (Permissions can change, but we no longer store that.) But
	// for the purpose of avoiding rate limits, we set all public repos to
	// read-only permissions.
	var cachedRemoteRepo sourcegraph.RemoteRepo
	if err := reposGithubPublicCache.Get(repo, &cachedRemoteRepo); err == nil {
		reposGithubPublicCacheCounter.WithLabelValues("hit").Inc()
		return &cachedRemoteRepo, nil
	} else if err != rcache.ErrNotFound {
		log15.Error("github cache-get error", "err", err)
	}

	owner, repoName, err := githubutil.SplitRepoURI(repo)
	if err != nil {
		reposGithubPublicCacheCounter.WithLabelValues("local-error").Inc()
		return nil, grpc.Errorf(codes.NotFound, "github repo not found: %s", repo)
	}

	ghrepo, resp, err := client(ctx).repos.Get(owner, repoName)
	if err != nil {
		reposGithubPublicCacheCounter.WithLabelValues("error").Inc()
		return nil, checkResponse(ctx, resp, err, fmt.Sprintf("github.Repos.Get %q", repo))
	}
	remoteRepo := toRemoteRepo(ghrepo)
	if ghrepo.Private != nil && !*ghrepo.Private {
		reposGithubPublicCache.Add(repo, remoteRepo, reposGithubPublicCacheTTL)
		reposGithubPublicCacheCounter.WithLabelValues("miss").Inc()
	} else {
		reposGithubPublicCacheCounter.WithLabelValues("private").Inc()
	}

	return remoteRepo, nil
}

func (s *Repos) GetByID(ctx context.Context, id int) (*sourcegraph.RemoteRepo, error) {
	ghrepo, resp, err := client(ctx).repos.GetByID(id)
	if err != nil {
		return nil, checkResponse(ctx, resp, err, fmt.Sprintf("github.Repos.GetByID #%d", id))
	}
	return toRemoteRepo(ghrepo), nil
}

func toRemoteRepo(ghrepo *github.Repository) *sourcegraph.RemoteRepo {
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
	repo := sourcegraph.RemoteRepo{
		GitHubID:      int32(*ghrepo.ID),
		Name:          *ghrepo.Name,
		VCS:           "git",
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
		repo.OwnerIsOrg = strv(ghrepo.Owner.Type) == "Organization"
	}
	if ghrepo.UpdatedAt != nil {
		ts := pbtypes.NewTimestamp(ghrepo.UpdatedAt.Time)
		repo.UpdatedAt = &ts
	}
	if ghrepo.PushedAt != nil {
		ts := pbtypes.NewTimestamp(ghrepo.PushedAt.Time)
		repo.PushedAt = &ts
	}
	if ghrepo.WatchersCount != nil {
		repo.Stars = int32(*ghrepo.WatchersCount)
	}
	return &repo
}

// ListAccessible lists repos that are accessible to the authenticated
// user.
//
// See https://developer.github.com/v3/repos/#list-your-repositories
// for more information.
func (s *Repos) ListAccessible(ctx context.Context, opt *github.RepositoryListOptions) ([]*sourcegraph.RemoteRepo, error) {
	ghRepos, resp, err := client(ctx).repos.List("", opt)
	if err != nil {
		return nil, checkResponse(ctx, resp, err, "github.Repos.ListAccessible")
	}

	var repos []*sourcegraph.RemoteRepo
	for _, ghRepo := range ghRepos {
		repos = append(repos, toRemoteRepo(&ghRepo))
	}
	return repos, nil
}
