package github

import (
	"net/http"
	"os"
	"strings"

	"github.com/sourcegraph/go-github/github"
	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"sourcegraph.com/sqs/pbtypes"
	"src.sourcegraph.com/sourcegraph/auth/authutil"
	"src.sourcegraph.com/sourcegraph/ext/github/githubcli"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/store"
	"src.sourcegraph.com/sourcegraph/util/githubutil"
)

type Repos struct{}

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
		ts := pbtypes.NewTimestamp(ghrepo.CreatedAt.Time)
		repo.CreatedAt = &ts
	}
	if ghrepo.UpdatedAt != nil {
		ts := pbtypes.NewTimestamp(ghrepo.UpdatedAt.Time)
		repo.UpdatedAt = &ts
	}
	if ghrepo.PushedAt != nil {
		ts := pbtypes.NewTimestamp(ghrepo.PushedAt.Time)
		repo.PushedAt = &ts
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

// ListWithToken lists repos from GitHub that are visible in the given auth
// token's scope.
func (s *Repos) ListWithToken(ctx context.Context, token string) ([]*sourcegraph.RemoteRepo, error) {
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	tc := oauth2.NewClient(oauth2.NoContext, ts)

	client := github.NewClient(tc)
	repoType := "private"
	if githubcli.Config.IsGitHubEnterprise() {
		client.BaseURL = githubcli.Config.APIBaseURL()
		client.UploadURL = githubcli.Config.UploadURL()
		repoType = "" // import both public and private repos from GHE.
	}
	if authutil.ActiveFlags.PrivateMirrors {
		repoType = "" // import both public and private repos for PrivateMirrors.
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
			remoteRepo := &sourcegraph.RemoteRepo{
				Repo: *repoFromGitHub(&ghrepo),
			}
			if ghrepo.Owner != nil {
				remoteRepo.Owner = userFromGitHub(ghrepo.Owner)
			}
			if ghrepo.Size != nil {
				remoteRepo.RepoSize = int32(*ghrepo.Size)
			}
			if ghrepo.WatchersCount != nil {
				remoteRepo.Watchers = int32(*ghrepo.WatchersCount)
			}
			if ghrepo.SubscribersCount != nil {
				remoteRepo.Subscribers = int32(*ghrepo.SubscribersCount)
			}
			if ghrepo.StargazersCount != nil {
				remoteRepo.Stars = int32(*ghrepo.StargazersCount)
			}
			if ghrepo.OpenIssuesCount != nil {
				remoteRepo.OpenIssues = int32(*ghrepo.OpenIssuesCount)
			}
			if ghrepo.ForksCount != nil {
				remoteRepo.Forks = int32(*ghrepo.ForksCount)
			}
			repos = append(repos, remoteRepo)
		}
		if resp.NextPage == 0 {
			break
		}
		repoOpts.ListOptions.Page = resp.NextPage
	}

	return repos, nil
}
