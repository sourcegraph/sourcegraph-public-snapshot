package mocks

import (
	"golang.org/x/net/context"

	gogithub "github.com/sourcegraph/go-github/github"
	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
)

type GitHubRepoGetter struct {
	Get_            func(context.Context, string) (*sourcegraph.Repo, error)
	GetByID_        func(context.Context, int) (*sourcegraph.Repo, error)
	ListAccessible_ func(context.Context, *gogithub.RepositoryListOptions) ([]*sourcegraph.Repo, error)
}

func (s *GitHubRepoGetter) Get(ctx context.Context, repo string) (*sourcegraph.Repo, error) {
	return s.Get_(ctx, repo)
}

func (s *GitHubRepoGetter) GetByID(ctx context.Context, id int) (*sourcegraph.Repo, error) {
	return s.GetByID_(ctx, id)
}

func (s *GitHubRepoGetter) ListAccessible(ctx context.Context, opt *gogithub.RepositoryListOptions) ([]*sourcegraph.Repo, error) {
	return s.ListAccessible_(ctx, opt)
}

func (s *GitHubRepoGetter) MockGet_Return(ctx context.Context, returns *sourcegraph.Repo) {
	s.Get_ = func(context.Context, string) (*sourcegraph.Repo, error) {
		return returns, nil
	}
}

func (s *GitHubRepoGetter) MockListAccessible_IDs(ctx context.Context, ids []string) (called *bool) {
	called = new(bool)
	s.ListAccessible_ = func(ctx context.Context, opt *gogithub.RepositoryListOptions) ([]*sourcegraph.Repo, error) {
		*called = true

		if opt != nil && opt.Page > 1 {
			return nil, nil
		}

		repos := make([]*sourcegraph.Repo, len(ids))
		for i, id := range ids {
			repos[i] = &sourcegraph.Repo{Origin: &sourcegraph.Origin{ID: id}}
		}
		return repos, nil
	}
	return
}
