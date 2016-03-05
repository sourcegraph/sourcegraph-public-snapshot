package github

import "github.com/sourcegraph/go-github/github"

// minimalClient contains the minimal set of GitHub API methods needed
// by this package.
type minimalClient struct {
	repos githubRepos
	orgs  githubOrgs
}

func newMinimalClient(client *github.Client) *minimalClient {
	return &minimalClient{
		repos: client.Repositories,
		orgs:  client.Organizations,
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
