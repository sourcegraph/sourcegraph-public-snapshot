package github

import (
	"fmt"
	"net/http"
	"os"

	"github.com/sourcegraph/go-github/github"
	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"sourcegraph.com/sqs/pbtypes"
	"src.sourcegraph.com/sourcegraph/ext/github/githubcli"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/store"
	"src.sourcegraph.com/sourcegraph/util/githubutil"
)

type Repos struct{}

func (s *Repos) Get(ctx context.Context, repo string) (*sourcegraph.RemoteRepo, error) {
	owner, repoName, err := githubutil.SplitGitHubRepoURI(repo)
	if err != nil {
		return nil, err
	}

	ghrepo, resp, err := client(ctx).repos.Get(owner, repoName)
	if err != nil {
		return nil, s.checkResponse(repo, resp, err)
	}

	return toRemoteRepo(ghrepo), nil
}

func (s *Repos) GetByID(ctx context.Context, id int) (*sourcegraph.RemoteRepo, error) {
	ghrepo, resp, err := client(ctx).repos.GetByID(id)
	if err != nil {
		return nil, s.checkResponse(fmt.Sprintf("GitHub repo #%d", id), resp, err)
	}
	return toRemoteRepo(ghrepo), nil
}

func (s *Repos) checkResponse(repo string, resp *github.Response, err error) error {
	if err == nil {
		return nil
	}
	if resp != nil && resp.StatusCode == http.StatusNotFound {
		return &store.RepoNotFoundError{Repo: repo}
	}
	if resp != nil && resp.StatusCode == http.StatusForbidden {
		return &os.PathError{Op: "Repos.Get", Path: repo, Err: os.ErrPermission}
	}
	return err
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

// ListWithToken lists repos from GitHub that are visible in the given auth
// token's scope.
func (s *Repos) ListWithToken(ctx context.Context, token string) ([]*sourcegraph.RemoteRepo, error) {
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	tc := oauth2.NewClient(oauth2.NoContext, ts)

	client := github.NewClient(tc)
	repoType := ""
	if githubcli.Config.IsGitHubEnterprise() {
		client.BaseURL = githubcli.Config.APIBaseURL()
		client.UploadURL = githubcli.Config.UploadURL()
	}

	var repos []*sourcegraph.RemoteRepo
	repoOpts := &github.RepositoryListOptions{
		Type: repoType,
		ListOptions: github.ListOptions{
			PerPage: 100, // 100 is the max page size for GitHub's API
		},
	}
	// List responses on GitHub are paginated; continue fetching private repos
	// until each page has been obtained.
	for {
		userRepos, resp, err := client.Repositories.List("", repoOpts)
		if err != nil {
			return nil, err
		}
		for _, ghrepo := range userRepos {
			repos = append(repos, toRemoteRepo(&ghrepo))
		}
		if resp.NextPage == 0 {
			break
		}
		repoOpts.ListOptions.Page = resp.NextPage
	}

	return repos, nil
}
