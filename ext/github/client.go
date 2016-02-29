package github

import "github.com/sourcegraph/go-github/github"

// minimalClient contains the minimal set of GitHub API methods needed
// by this package.
type minimalClient struct {
	repos githubRepos
	orgs  githubOrgs

	isAuthedUser bool // whether the client is using a GitHub user's auth token
}

func newMinimalClient(client *github.Client, isAuthedUser bool) *minimalClient {
	return &minimalClient{
		repos: client.Repositories,
		orgs:  client.Organizations,

		isAuthedUser: isAuthedUser,
	}
}

type githubRepos interface {
	Get(owner, repo string) (*github.Repository, *github.Response, error)
	List(user string, opt *github.RepositoryListOptions) ([]github.Repository, *github.Response, error)
}

type githubOrgs interface {
	ListMembers(org string, opt *github.ListMembersOptions) ([]github.User, *github.Response, error)
	List(member string, opt *github.ListOptions) ([]github.Organization, *github.Response, error)
}
