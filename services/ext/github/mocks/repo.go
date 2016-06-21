package mocks

import (
	"golang.org/x/net/context"

	gogithub "github.com/sourcegraph/go-github/github"
	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
)

type GitHubRepoGetter struct {
	Get_            func(context.Context, string) (*sourcegraph.RemoteRepo, error)
	GetByID_        func(context.Context, int) (*sourcegraph.RemoteRepo, error)
	ListAccessible_ func(context.Context, *gogithub.RepositoryListOptions) ([]*sourcegraph.RemoteRepo, error)
}

func (s *GitHubRepoGetter) Get(ctx context.Context, repo string) (*sourcegraph.RemoteRepo, error) {
	return s.Get_(ctx, repo)
}

func (s *GitHubRepoGetter) GetByID(ctx context.Context, id int) (*sourcegraph.RemoteRepo, error) {
	return s.GetByID_(ctx, id)
}

func (s *GitHubRepoGetter) ListAccessible(ctx context.Context, opt *gogithub.RepositoryListOptions) ([]*sourcegraph.RemoteRepo, error) {
	return s.ListAccessible_(ctx, opt)
}

func (s *GitHubRepoGetter) MockGet_Return(ctx context.Context, returns *sourcegraph.RemoteRepo) {
	s.Get_ = func(context.Context, string) (*sourcegraph.RemoteRepo, error) {
		return returns, nil
	}
}
