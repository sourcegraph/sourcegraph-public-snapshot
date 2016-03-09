package github

import (
	"fmt"

	"github.com/sourcegraph/go-github/github"
	"golang.org/x/net/context"
	"sourcegraph.com/sqs/pbtypes"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/store"
	"src.sourcegraph.com/sourcegraph/util/githubutil"
)

type Repos struct{}

func (s *Repos) Get(ctx context.Context, repo string) (*sourcegraph.RemoteRepo, error) {
	owner, repoName, err := githubutil.SplitGitHubRepoURI(repo)
	if err != nil {
		return nil, &store.RepoNotFoundError{Repo: repo}
	}

	ghrepo, resp, err := client(ctx).repos.Get(owner, repoName)
	if err != nil {
		return nil, checkResponse(resp, err, fmt.Sprintf("github.Repos.Get %q", repo))
	}

	return toRemoteRepo(ghrepo), nil
}

func (s *Repos) GetByID(ctx context.Context, id int) (*sourcegraph.RemoteRepo, error) {
	ghrepo, resp, err := client(ctx).repos.GetByID(id)
	if err != nil {
		return nil, checkResponse(resp, err, fmt.Sprintf("github.Repos.GetByID #%d", id))
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
		Permissions:   convertGitHubRepoPerms(ghrepo),
	}
	if ghrepo.Owner != nil {
		repo.Owner = strv(ghrepo.Owner.Login)
		repo.OwnerIsOrg = strv(ghrepo.Owner.Type) == "Organization"
	}
	if ghrepo.UpdatedAt != nil {
		ts := pbtypes.NewTimestamp(ghrepo.UpdatedAt.Time)
		repo.UpdatedAt = &ts
	}
	if ghrepo.WatchersCount != nil {
		repo.Stars = int32(*ghrepo.WatchersCount)
	}
	return &repo
}

func convertGitHubRepoPerms(ghrepo *github.Repository) *sourcegraph.RepoPermissions {
	if ghrepo.Permissions == nil {
		return nil
	}
	gp := *ghrepo.Permissions
	rp := &sourcegraph.RepoPermissions{}
	rp.Read = gp["pull"]
	rp.Write = gp["push"]
	rp.Admin = gp["admin"]
	return rp
}

// ListAccessible lists repos that are accessible to the authenticated
// user.
//
// See https://developer.github.com/v3/repos/#list-your-repositories
// for more information.
func (s *Repos) ListAccessible(ctx context.Context, opt *github.RepositoryListOptions) ([]*sourcegraph.RemoteRepo, error) {
	ghRepos, resp, err := client(ctx).repos.List("", opt)
	if err != nil {
		return nil, checkResponse(resp, err, "github.Repos.ListAccessible")
	}

	var repos []*sourcegraph.RemoteRepo
	for _, ghRepo := range ghRepos {
		repos = append(repos, toRemoteRepo(&ghRepo))
	}
	return repos, nil
}
