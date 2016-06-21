package mockstore

import (
	"testing"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/store"
)

func (s *Repos) MockGet(t *testing.T, wantRepo int32) (called *bool) {
	called = new(bool)
	s.Get_ = func(ctx context.Context, repo int32) (*sourcegraph.Repo, error) {
		*called = true
		if repo != wantRepo {
			t.Errorf("got repo %d, want %d", repo, wantRepo)
			return nil, grpc.Errorf(codes.NotFound, "repo %v not found", wantRepo)
		}
		return &sourcegraph.Repo{ID: repo}, nil
	}
	return
}

func (s *Repos) MockGet_Path(t *testing.T, wantRepo int32, repoPath string) (called *bool) {
	called = new(bool)
	s.Get_ = func(ctx context.Context, repo int32) (*sourcegraph.Repo, error) {
		*called = true
		if repo != wantRepo {
			t.Errorf("got repo %d, want %d", repo, wantRepo)
			return nil, grpc.Errorf(codes.NotFound, "repo %v not found", wantRepo)
		}
		return &sourcegraph.Repo{ID: repo, URI: repoPath}, nil
	}
	return
}

func (s *Repos) MockUpdate(t *testing.T, wantRepo int32) (called *bool) {
	called = new(bool)
	s.Update_ = func(ctx context.Context, repoUpdate store.RepoUpdate) error {
		*called = true
		if repoUpdate.ReposUpdateOp.Repo != wantRepo {
			t.Errorf("got repo %q, want %q", repoUpdate.ReposUpdateOp.Repo, wantRepo)
			return grpc.Errorf(codes.NotFound, "repo %v not found", wantRepo)
		}
		return nil
	}
	return
}

func (s *Repos) MockGet_Return(t *testing.T, returns *sourcegraph.Repo) (called *bool) {
	called = new(bool)
	s.Get_ = func(ctx context.Context, repo int32) (*sourcegraph.Repo, error) {
		*called = true
		if repo != returns.ID {
			t.Errorf("got repo %d, want %d", repo, returns.ID)
			return nil, grpc.Errorf(codes.NotFound, "repo %v (%d) not found", returns.URI, returns.ID)
		}
		return returns, nil
	}
	return
}

func (s *Repos) MockGetByURI(t *testing.T, wantURI string, repoID int32) (called *bool) {
	called = new(bool)
	s.GetByURI_ = func(ctx context.Context, uri string) (*sourcegraph.Repo, error) {
		*called = true
		if uri != wantURI {
			t.Errorf("got repo URI %q, want %q", uri, wantURI)
			return nil, grpc.Errorf(codes.NotFound, "repo %v not found", uri)
		}
		return &sourcegraph.Repo{ID: repoID, URI: uri}, nil
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

func (s *Repos) MockInternalUpdate(t *testing.T) (called *bool) {
	called = new(bool)
	s.InternalUpdate_ = func(ctx context.Context, repo int32, op store.InternalRepoUpdate) error {
		*called = true
		return nil
	}
	return
}
