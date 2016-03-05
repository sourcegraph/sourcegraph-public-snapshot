package github

import (
	"github.com/sourcegraph/go-github/github"
	"golang.org/x/net/context"
)

func testContext(client *minimalClient) context.Context {
	return newContext(context.Background(), client)
}

type mockGitHubRepos struct {
	Get_     func(owner, repo string) (*github.Repository, *github.Response, error)
	GetByID_ func(id int) (*github.Repository, *github.Response, error)
	List_    func(user string, opt *github.RepositoryListOptions) ([]github.Repository, *github.Response, error)
}

var _ githubRepos = (*mockGitHubRepos)(nil)

func (s mockGitHubRepos) Get(owner, repo string) (*github.Repository, *github.Response, error) {
	return s.Get_(owner, repo)
}

func (s mockGitHubRepos) GetByID(id int) (*github.Repository, *github.Response, error) {
	return s.GetByID_(id)
}

func (s mockGitHubRepos) List(user string, opt *github.RepositoryListOptions) ([]github.Repository, *github.Response, error) {
	return s.List_(user, opt)
}
