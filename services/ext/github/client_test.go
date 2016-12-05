package github

import (
	"context"

	"github.com/sourcegraph/go-github/github"
)

func testContext(client *minimalClient) context.Context {
	return newContext(context.Background(), client)
}

type mockGitHubRepos struct {
	Get_              func(owner, repo string) (*github.Repository, *github.Response, error)
	GetByID_          func(id int) (*github.Repository, *github.Response, error)
	List_             func(user string, opt *github.RepositoryListOptions) ([]*github.Repository, *github.Response, error)
	ListContributors_ func(owner string, repository string, opt *github.ListContributorsOptions) ([]*github.Contributor, *github.Response, error)
	CreateHook_       func(owner, repo string, hook *github.Hook) (*github.Hook, *github.Response, error)
}

var _ githubRepos = (*mockGitHubRepos)(nil)

func (s mockGitHubRepos) Get(owner, repo string) (*github.Repository, *github.Response, error) {
	return s.Get_(owner, repo)
}

func (s mockGitHubRepos) GetByID(id int) (*github.Repository, *github.Response, error) {
	return s.GetByID_(id)
}

func (s mockGitHubRepos) List(user string, opt *github.RepositoryListOptions) ([]*github.Repository, *github.Response, error) {
	return s.List_(user, opt)
}

func (s mockGitHubRepos) ListContributors(owner string, repository string, opt *github.ListContributorsOptions) ([]*github.Contributor, *github.Response, error) {
	return s.ListContributors_(owner, repository, opt)
}

func (s mockGitHubRepos) CreateHook(owner, repo string, hook *github.Hook) (*github.Hook, *github.Response, error) {
	return s.CreateHook_(owner, repo, hook)
}

type mockGitHubAuthorizations struct {
	Revoke_ func(clientID, token string) (*github.Response, error)
}

var _ githubAuthorizations = (*mockGitHubAuthorizations)(nil)

func (s mockGitHubAuthorizations) Revoke(clientID, token string) (*github.Response, error) {
	return s.Revoke_(clientID, token)
}
