package backend

import (
	"sync"
	"testing"

	"context"

	sourcegraph "sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/api/legacyerr"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/inventory"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/vcs"
)

type MockRepos struct {
	Get                  func(v0 context.Context, v1 *sourcegraph.RepoSpec) (*sourcegraph.Repo, error)
	GetByURI             func(v0 context.Context, v1 string) (*sourcegraph.Repo, error)
	List                 func(v0 context.Context, v1 *sourcegraph.RepoListOptions) (*sourcegraph.RepoList, error)
	GetCommit            func(v0 context.Context, v1 *sourcegraph.RepoRevSpec) (*vcs.Commit, error)
	ResolveRev           func(v0 context.Context, v1 *sourcegraph.ReposResolveRevOp) (*sourcegraph.ResolvedRev, error)
	ListCommits          func(v0 context.Context, v1 *sourcegraph.ReposListCommitsOp) (*sourcegraph.CommitList, error)
	ListDeps             func(v0 context.Context, v1 *sourcegraph.URIList) (*sourcegraph.URIList, error)
	ListCommitters       func(v0 context.Context, v1 *sourcegraph.ReposListCommittersOp) (*sourcegraph.CommitterList, error)
	GetInventory         func(v0 context.Context, v1 *sourcegraph.RepoRevSpec) (*inventory.Inventory, error)
	GetInventoryUncached func(ctx context.Context, repoRev *sourcegraph.RepoRevSpec) (*inventory.Inventory, error)
	RefreshIndex         func(ctx context.Context, repo string) (err error)
}

func (s *MockRepos) MockGet(t *testing.T, wantRepo int32) (called *bool) {
	called = new(bool)
	s.Get = func(ctx context.Context, repo *sourcegraph.RepoSpec) (*sourcegraph.Repo, error) {
		*called = true
		if repo.ID != wantRepo {
			t.Errorf("got repo %d, want %d", repo.ID, wantRepo)
			return nil, legacyerr.Errorf(legacyerr.NotFound, "repo %d not found", wantRepo)
		}
		return &sourcegraph.Repo{ID: repo.ID}, nil
	}
	return
}

func (s *MockRepos) MockGetByURI(t *testing.T, wantURI string, repoID int32) (called *bool) {
	called = new(bool)
	s.GetByURI = func(ctx context.Context, uri string) (*sourcegraph.Repo, error) {
		*called = true
		if uri != wantURI {
			t.Errorf("got repo URI %q, want %q", uri, wantURI)
			return nil, legacyerr.Errorf(legacyerr.NotFound, "repo %v not found", uri)
		}
		return &sourcegraph.Repo{ID: repoID, URI: uri}, nil
	}
	return
}

func (s *MockRepos) MockGet_Return(t *testing.T, returns *sourcegraph.Repo) (called *bool) {
	called = new(bool)
	s.Get = func(ctx context.Context, repo *sourcegraph.RepoSpec) (*sourcegraph.Repo, error) {
		*called = true
		if repo.ID != returns.ID {
			t.Errorf("got repo %d, want %d", repo.ID, returns.ID)
			return nil, legacyerr.Errorf(legacyerr.NotFound, "repo %d not found", returns.ID)
		}
		return returns, nil
	}
	return
}

func (s *MockRepos) MockList(t *testing.T, wantRepos ...string) (called *bool) {
	called = new(bool)
	s.List = func(ctx context.Context, opt *sourcegraph.RepoListOptions) (*sourcegraph.RepoList, error) {
		*called = true
		repos := make([]*sourcegraph.Repo, len(wantRepos))
		for i, repo := range wantRepos {
			repos[i] = &sourcegraph.Repo{URI: repo}
		}
		return &sourcegraph.RepoList{Repos: repos}, nil
	}
	return
}

func (s *MockRepos) MockListCommits(t *testing.T, wantCommitIDs ...vcs.CommitID) (called *bool) {
	called = new(bool)
	s.ListCommits = func(ctx context.Context, op *sourcegraph.ReposListCommitsOp) (*sourcegraph.CommitList, error) {
		*called = true
		commits := make([]*vcs.Commit, len(wantCommitIDs))
		for i, commit := range wantCommitIDs {
			commits[i] = &vcs.Commit{ID: commit}
		}
		return &sourcegraph.CommitList{Commits: commits}, nil
	}
	return
}

func (s *MockRepos) MockResolveRev_NoCheck(t *testing.T, commitID vcs.CommitID) (called *bool) {
	var once sync.Once
	called = new(bool)
	s.ResolveRev = func(ctx context.Context, op *sourcegraph.ReposResolveRevOp) (*sourcegraph.ResolvedRev, error) {
		once.Do(func() {
			*called = true
		})
		return &sourcegraph.ResolvedRev{CommitID: string(commitID)}, nil
	}
	return
}

func (s *MockRepos) MockResolveRev_NotFound(t *testing.T, repo int32, rev string) (called *bool) {
	called = new(bool)
	s.ResolveRev = func(ctx context.Context, op *sourcegraph.ReposResolveRevOp) (*sourcegraph.ResolvedRev, error) {
		*called = true
		if op.Repo != repo || op.Rev != rev {
			t.Errorf("got %+v, want %+v", op, &sourcegraph.ReposResolveRevOp{Repo: repo, Rev: rev})
		}
		return nil, legacyerr.Errorf(legacyerr.NotFound, "revision not found")
	}
	return
}

func (s *MockRepos) MockGetCommit_Return_NoCheck(t *testing.T, commit *vcs.Commit) (called *bool) {
	called = new(bool)
	s.GetCommit = func(ctx context.Context, repoRev *sourcegraph.RepoRevSpec) (*vcs.Commit, error) {
		*called = true
		return commit, nil
	}
	return
}
