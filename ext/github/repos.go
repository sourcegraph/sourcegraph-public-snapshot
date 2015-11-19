package github

import (
	"errors"
	"net/http"
	"os"
	"strings"

	"github.com/sourcegraph/go-github/github"
	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"sourcegraph.com/sqs/pbtypes"
	"src.sourcegraph.com/sourcegraph/ext"
	"src.sourcegraph.com/sourcegraph/ext/github/githubcli"
	"src.sourcegraph.com/sourcegraph/store"
	"src.sourcegraph.com/sourcegraph/util/githubutil"
)

// Repos is a GitHub-backed implementation of the Repos store.
type Repos struct{}

var _ store.Repos = (*Repos)(nil)

func (s *Repos) Get(ctx context.Context, repo string) (*sourcegraph.Repo, error) {
	owner, repoName, err := githubutil.SplitGitHubRepoURI(repo)
	if err != nil {
		return nil, err
	}

	ghrepo, resp, err := client(ctx).repos.Get(owner, repoName)
	if err != nil {
		if resp != nil && resp.StatusCode == http.StatusNotFound {
			return nil, &store.RepoNotFoundError{Repo: repo}
		}
		if resp != nil && resp.StatusCode == http.StatusForbidden {
			return nil, &os.PathError{Op: "Repos.Get", Path: repo, Err: os.ErrPermission}
		}
		return nil, err
	}

	return repoFromGitHub(ghrepo), nil
}

func repoFromGitHub(ghrepo *github.Repository) *sourcegraph.Repo {
	gitHubHost := githubcli.Config.Host()
	repo := sourcegraph.Repo{
		URI:         gitHubHost + "/" + *ghrepo.FullName,
		Permissions: convertGitHubRepoPerms(ghrepo),
		Mirror:      true,
	}

	if ghrepo.CloneURL != nil {
		repo.HTTPCloneURL = *ghrepo.CloneURL
	}

	// GitHub's SSHURL field is of the form
	// "git@github.com:owner/repo.git", but we want
	// "ssh://git@github.com/owner/repo.git."
	if ghrepo.SSHURL != nil {
		origHostStr := "git@" + gitHubHost + ":"
		newHostStr := "ssh://git@" + gitHubHost + "/"
		repo.SSHCloneURL = strings.Replace(*ghrepo.SSHURL, origHostStr, newHostStr, 1)
	}

	repo.Name = *ghrepo.Name
	if ghrepo.Description != nil {
		repo.Description = *ghrepo.Description
	}
	repo.VCS = sourcegraph.Git
	if ghrepo.DefaultBranch != nil {
		repo.DefaultBranch = *ghrepo.DefaultBranch
	}
	if ghrepo.Homepage != nil {
		repo.HomepageURL = *ghrepo.Homepage
	}
	if ghrepo.Language != nil {
		repo.Language = *ghrepo.Language
	}
	if ghrepo.Fork != nil {
		repo.Fork = *ghrepo.Fork
	}
	if ghrepo.Private != nil {
		repo.Private = *ghrepo.Private
	}
	if ghrepo.CreatedAt != nil {
		repo.CreatedAt = pbtypes.NewTimestamp(ghrepo.CreatedAt.Time)
	}
	if ghrepo.UpdatedAt != nil {
		repo.UpdatedAt = pbtypes.NewTimestamp(ghrepo.UpdatedAt.Time)
	}
	if ghrepo.PushedAt != nil {
		repo.PushedAt = pbtypes.NewTimestamp(ghrepo.PushedAt.Time)
	}

	// Look for "DEPRECATED" in the description. If it's removed from the
	// description, the repo won't be un-deprecated. This allows us to manually
	// deprecate repos that don't contain "DEPRECATED" in their description.
	if ghrepo.Description != nil && strings.Contains(*ghrepo.Description, "DEPRECATED") {
		repo.Deprecated = true
	}

	repo.GitHub = &sourcegraph.GitHubRepo{}
	if ghrepo.WatchersCount != nil {
		repo.GitHub.Stars = int32(*ghrepo.WatchersCount)
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

// TODO(public-release): This function has been commented out mostly for
// performance reasons. It is fine for now, but MUST be implemented before we
// have any code that needs to know the user's auth with respect to a GitHub
// repo.
func (s *Repos) GetPerms(ctx context.Context, repo string) (*sourcegraph.RepoPermissions, error) {
	// TODO(sqs): optimization: if the GitHub client is anonymous,
	// then there are no permissions, so we can just return nil here
	// instead of incurring the HTTP request.

	//r, err := s.Get(ctx, repo)
	//if err != nil {
	//return nil, err
	//}
	//if r.Permissions == nil {
	//return &sourcegraph.RepoPermissions{Read: true}, nil
	//}
	//return r.Permissions, nil
	return &sourcegraph.RepoPermissions{Read: true}, nil
}

func (s *Repos) List(ctx context.Context, opt *sourcegraph.RepoListOptions) ([]*sourcegraph.Repo, error) {
	if opt == nil {
		opt = &sourcegraph.RepoListOptions{}
	}

	listOpt := github.ListOptions{
		Page:    opt.ListOptions.PageOrDefault(),
		PerPage: opt.ListOptions.PerPageOrDefault(),
	}

	var (
		ghRepos []github.Repository
		err     error
	)
	githubHost := githubcli.Config.Host()
	if opt.Owner != "" {
		ghRepos, _, err = client(ctx).repos.List(opt.Owner, &github.RepositoryListOptions{
			Sort:        opt.Sort,
			Direction:   opt.Direction,
			ListOptions: listOpt,
		})
	} else if opt.Query != "" {
		repoQuery := strings.TrimSpace(strings.TrimPrefix(opt.Query, githubHost+"/"))
		parts := strings.Split(repoQuery, "/")
		if len(parts) == 2 && parts[1] == "" {
			repoQuery = "user:" + parts[0]
		} else if len(parts) == 1 {
			repoQuery = parts[0]
		} else {
			repoQuery = "user:" + parts[0] + " " + parts[1]
		}

		var results *github.RepositoriesSearchResult
		results, _, err = client(ctx).search.Repositories("in:name "+repoQuery, &github.SearchOptions{ListOptions: listOpt})
		for _, repo := range results.Repositories {
			ghRepos = append(ghRepos, repo)
		}
	} else {
		ghRepos, _, err = client(ctx).repos.ListAll(&github.RepositoryListAllOptions{ListOptions: listOpt})
	}
	if err != nil {
		return nil, err
	}

	repos := make([]*sourcegraph.Repo, len(ghRepos))
	for i, ghrepo := range ghRepos {
		repos[i] = repoFromGitHub(&ghrepo)
	}
	return repos, nil
}

// TODO: rename or consolidate this method, since it can be used to list private
// as well as public repos.
func (s *Repos) ListPrivate(ctx context.Context) ([]*sourcegraph.Repo, error) {
	tokenStore := &ext.AccessTokens{}
	token, err := tokenStore.Get(ctx, githubcli.Config.Host())
	if err != nil {
		return nil, err
	}

	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	tc := oauth2.NewClient(oauth2.NoContext, ts)

	client := github.NewClient(tc)
	repoType := "private"
	if githubcli.Config.IsGitHubEnterprise() {
		client.BaseURL = githubcli.Config.APIBaseURL()
		client.UploadURL = githubcli.Config.UploadURL()
		repoType = "" // import both public and private repos from GHE.
	}

	var repos []*sourcegraph.Repo
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
			repos = append(repos, repoFromGitHub(&ghrepo))
		}
		if resp.NextPage == 0 {
			break
		}
		repoOpts.ListOptions.Page = resp.NextPage
	}

	return repos, nil
}

func (s *Repos) Create(ctx context.Context, newRepo *sourcegraph.Repo) (*sourcegraph.Repo, error) {
	return nil, errors.New("GitHub repo creation is not implemented")
}

func (s *Repos) Update(ctx context.Context, op *sourcegraph.ReposUpdateOp) error {
	return errors.New("GitHub repo updating is not implemented")
}

func (s *Repos) Delete(ctx context.Context, repo string) error {
	return errors.New("GitHub repo deletion is not implemented")
}

func (s *Repos) CreateChangeset(ctx context.Context, repo string, cs *sourcegraph.Changeset) error {
	return errors.New("GitHub PR creation is not implemented")
}

func (s *Repos) GetChangeset(ctx context.Context, repo string, ID int64) (*sourcegraph.Changeset, error) {
	return nil, errors.New("GitHub PR get is not implemented")
}
