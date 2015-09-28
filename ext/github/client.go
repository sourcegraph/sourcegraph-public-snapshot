package github

import "github.com/sourcegraph/go-github/github"

// minimalClient contains the minimal set of GitHub API methods needed
// to implement the stores in this package.
type minimalClient struct {
	repos  githubRepos
	search githubReposSearch
	users  githubUsers
	orgs   githubOrgs
}

func newMinimalClient(client *github.Client) *minimalClient {
	return &minimalClient{
		repos:  client.Repositories,
		search: client.Search,
		users:  client.Users,
		orgs:   client.Organizations,
	}
}

type githubRepos interface {
	Get(owner, repo string) (*github.Repository, *github.Response, error)
	List(user string, opt *github.RepositoryListOptions) ([]github.Repository, *github.Response, error)
	ListAll(opt *github.RepositoryListAllOptions) ([]github.Repository, *github.Response, error)
	GetCombinedStatus(owner, repo, commit string, opt *github.ListOptions) (*github.CombinedStatus, *github.Response, error)
	CreateStatus(owner, repo, commit string, status *github.RepoStatus) (*github.RepoStatus, *github.Response, error)
}

type githubReposSearch interface {
	Repositories(query string, opt *github.SearchOptions) (*github.RepositoriesSearchResult, *github.Response, error)
}

type githubUsers interface {
	Get(login string) (*github.User, *github.Response, error)
	GetByID(id int) (*github.User, *github.Response, error)
	ListAll(opt *github.UserListOptions) ([]github.User, *github.Response, error)
}

type githubOrgs interface {
	ListMembers(org string, opt *github.ListMembersOptions) ([]github.User, *github.Response, error)
	List(member string, opt *github.ListOptions) ([]github.Organization, *github.Response, error)
}
