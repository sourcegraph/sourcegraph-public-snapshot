package mocks

import (
	"context"

	gogithub "github.com/sourcegraph/go-github/github"
	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
)

type GitHubRepoGetter struct {
	Get_            func(context.Context, string) (*sourcegraph.Repo, error)
	ListAccessible_ func(context.Context, *gogithub.RepositoryListOptions) ([]*sourcegraph.Repo, error)
	CreateHook_     func(context.Context, string, *gogithub.Hook) error
	Search_         func(ctx context.Context, query string, op *gogithub.SearchOptions) ([]*sourcegraph.Repo, error)
}

func (s *GitHubRepoGetter) Get(ctx context.Context, repo string) (*sourcegraph.Repo, error) {
	return s.Get_(ctx, repo)
}

func (s *GitHubRepoGetter) ListAccessible(ctx context.Context, opt *gogithub.RepositoryListOptions) ([]*sourcegraph.Repo, error) {
	return s.ListAccessible_(ctx, opt)
}

func (s *GitHubRepoGetter) CreateHook(ctx context.Context, repo string, hook *gogithub.Hook) error {
	return s.CreateHook_(ctx, repo, hook)
}

func (s *GitHubRepoGetter) MockGet_Return(ctx context.Context, returns *sourcegraph.Repo) {
	s.Get_ = func(context.Context, string) (*sourcegraph.Repo, error) {
		return returns, nil
	}
}

func (s *GitHubRepoGetter) MockListAccessible(ctx context.Context, repos []*sourcegraph.Repo) (called *bool) {
	called = new(bool)
	s.ListAccessible_ = func(ctx context.Context, opt *gogithub.RepositoryListOptions) ([]*sourcegraph.Repo, error) {
		*called = true

		if opt != nil && opt.Page > 1 {
			return nil, nil
		}

		return repos, nil
	}
	return
}

func (s *GitHubRepoGetter) Search(ctx context.Context, query string, op *gogithub.SearchOptions) ([]*sourcegraph.Repo, error) {
	return s.Search_(ctx, query, op)
}

func (s *GitHubRepoGetter) MockSearch(ctx context.Context, repos []*sourcegraph.Repo) (called *bool) {
	called = new(bool)
	s.Search_ = func(ctx context.Context, query string, op *gogithub.SearchOptions) ([]*sourcegraph.Repo, error) {
		*called = true
		return repos, nil
	}
	return
}
