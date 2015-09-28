package github

import (
	"github.com/sourcegraph/go-github/github"
	"golang.org/x/net/context"
)

func testContext(client *minimalClient) context.Context {
	return newContext(context.Background(), client)
}

type mockGitHubRepos struct {
	Get_               func(owner, repo string) (*github.Repository, *github.Response, error)
	List_              func(user string, opt *github.RepositoryListOptions) ([]github.Repository, *github.Response, error)
	ListAll_           func(opt *github.RepositoryListAllOptions) ([]github.Repository, *github.Response, error)
	GetCombinedStatus_ func(owner, repo, commit string, opt *github.ListOptions) (*github.CombinedStatus, *github.Response, error)
	CreateStatus_      func(owner, repo, commit string, status *github.RepoStatus) (*github.RepoStatus, *github.Response, error)
}

var _ githubRepos = (*mockGitHubRepos)(nil)

func (s mockGitHubRepos) Get(owner, repo string) (*github.Repository, *github.Response, error) {
	return s.Get_(owner, repo)
}

func (s mockGitHubRepos) List(user string, opt *github.RepositoryListOptions) ([]github.Repository, *github.Response, error) {
	return s.List_(user, opt)
}

func (s mockGitHubRepos) ListAll(opt *github.RepositoryListAllOptions) ([]github.Repository, *github.Response, error) {
	return s.ListAll_(opt)
}

func (s mockGitHubRepos) GetCombinedStatus(owner, repo, commit string, opt *github.ListOptions) (*github.CombinedStatus, *github.Response, error) {
	return s.GetCombinedStatus_(owner, repo, commit, opt)
}

func (s mockGitHubRepos) CreateStatus(owner, repo, commit string, status *github.RepoStatus) (*github.RepoStatus, *github.Response, error) {
	return s.CreateStatus_(owner, repo, commit, status)
}

type mockGitHubReposSearch struct {
	Repositories_ func(query string, opt *github.SearchOptions) (*github.RepositoriesSearchResult, *github.Response, error)
}

var _ githubReposSearch = (*mockGitHubReposSearch)(nil)

func (s mockGitHubReposSearch) Repositories(query string, opt *github.SearchOptions) (*github.RepositoriesSearchResult, *github.Response, error) {
	return s.Repositories_(query, opt)
}
