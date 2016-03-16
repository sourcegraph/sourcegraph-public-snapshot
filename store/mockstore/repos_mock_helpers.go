package mockstore

import (
	"testing"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"sourcegraph.com/sourcegraph/sourcegraph/go-sourcegraph/sourcegraph"
)

func (s *Repos) MockGet(t *testing.T, wantRepo string) (called *bool) {
	called = new(bool)
	s.Get_ = func(ctx context.Context, repo string) (*sourcegraph.Repo, error) {
		*called = true
		if repo != wantRepo {
			t.Errorf("got repo %q, want %q", repo, wantRepo)
			return nil, grpc.Errorf(codes.NotFound, "repo %v not found", wantRepo)
		}
		return &sourcegraph.Repo{URI: repo}, nil
	}
	return
}

func (s *Repos) MockGet_Return(t *testing.T, returns *sourcegraph.Repo) (called *bool) {
	called = new(bool)
	s.Get_ = func(ctx context.Context, repo string) (*sourcegraph.Repo, error) {
		*called = true
		if repo != returns.URI {
			t.Errorf("got repo %q, want %q", repo, returns.URI)
			return nil, grpc.Errorf(codes.NotFound, "repo %v not found", returns.URI)
		}
		return returns, nil
	}
	return
}

func (s *Repos) MockList(t *testing.T, wantRepos ...string) (called *bool) {
	called = new(bool)
	s.List_ = func(ctx context.Context, opt *sourcegraph.RepoListOptions) ([]*sourcegraph.Repo, error) {
		*called = true
		repos := make([]*sourcegraph.Repo, len(wantRepos))
		for i, repo := range wantRepos {
			repos[i] = &sourcegraph.Repo{URI: repo}
		}
		return repos, nil
	}
	return
}
