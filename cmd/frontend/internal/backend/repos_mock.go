package backend

import (
	"context"
	"sync"
	"testing"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/inventory"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/errcode"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

type MockRepos struct {
	Get               func(v0 context.Context, id api.RepoID) (*types.Repo, error)
	GetByName         func(v0 context.Context, name api.RepoName) (*types.Repo, error)
	List              func(v0 context.Context, v1 database.ReposListOptions) ([]*types.Repo, error)
	ResolveRev        func(v0 context.Context, repo api.RepoName, rev string) (api.CommitID, error)
	GetInventory      func(v0 context.Context, repo api.RepoName, commitID api.CommitID) (*inventory.Inventory, error)
	RecloneRepository func(v0 context.Context, name api.RepoID) error
}

var errRepoNotFound = &errcode.Mock{
	Message:    "repo not found",
	IsNotFound: true,
}

func (s *MockRepos) MockGet(t *testing.T, wantRepo api.RepoID) (called *bool) {
	called = new(bool)
	s.Get = func(_ context.Context, repo api.RepoID) (*types.Repo, error) {
		*called = true
		if repo != wantRepo {
			t.Errorf("got repo %d, want %d", repo, wantRepo)
			return nil, errRepoNotFound
		}
		return &types.Repo{ID: repo}, nil
	}
	return
}

func (s *MockRepos) MockGetByName(t *testing.T, wantName api.RepoName, repo api.RepoID) (called *bool) {
	called = new(bool)
	s.GetByName = func(_ context.Context, name api.RepoName) (*types.Repo, error) {
		*called = true
		if name != wantName {
			t.Errorf("got repo name %q, want %q", name, wantName)
			return nil, errRepoNotFound
		}
		return &types.Repo{ID: repo, Name: name}, nil
	}
	return
}

func (s *MockRepos) MockGet_Return(t *testing.T, returns *types.Repo) (called *bool) {
	called = new(bool)
	s.Get = func(_ context.Context, repo api.RepoID) (*types.Repo, error) {
		*called = true
		if repo != returns.ID {
			t.Errorf("got repo %d, want %d", repo, returns.ID)
			return nil, errRepoNotFound
		}
		return returns, nil
	}
	return
}

func (s *MockRepos) MockList(t *testing.T, wantRepos ...api.RepoName) (called *bool) {
	called = new(bool)
	s.List = func(_ context.Context, opt database.ReposListOptions) ([]*types.Repo, error) {
		*called = true
		repos := make([]*types.Repo, len(wantRepos))
		for i, repo := range wantRepos {
			repos[i] = &types.Repo{Name: repo}
		}
		return repos, nil
	}
	return
}

func (s *MockRepos) MockRecloneRepository(t *testing.T, wantRepo api.RepoID) (called *bool) {
	called = new(bool)
	s.RecloneRepository = func(_ context.Context, repo api.RepoID) error {
		*called = repo == wantRepo
		return nil
	}
	return
}

func (s *MockRepos) MockResolveRev_NoCheck(t *testing.T, commitID api.CommitID) (called *bool) {
	var once sync.Once
	called = new(bool)
	s.ResolveRev = func(ctx context.Context, repo api.RepoName, rev string) (api.CommitID, error) {
		once.Do(func() {
			*called = true
		})
		return commitID, nil
	}
	return
}

func (s *MockRepos) MockResolveRev_NotFound(t *testing.T, wantRepo api.RepoName, wantRev string) (called *bool) {
	called = new(bool)
	s.ResolveRev = func(ctx context.Context, repo api.RepoName, rev string) (api.CommitID, error) {
		*called = true
		if repo != wantRepo {
			t.Errorf("got repo %v, want %v", repo, wantRepo)
		}
		if rev != wantRev {
			t.Errorf("got rev %v, want %v", rev, wantRev)
		}
		return "", &gitdomain.RevisionNotFoundError{Repo: repo, Spec: rev}
	}
	return
}

func (s *MockRepos) MockGetCommit_Return_NoCheck(t *testing.T, commit *gitdomain.Commit) (called *bool) {
	called = new(bool)
	return
}
