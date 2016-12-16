package backend

import (
	"sync"
	"testing"

	"context"

	gogithub "github.com/sourcegraph/go-github/github"
	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph/legacyerr"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/inventory"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/vcs"
)

type MockRepos struct {
	Get                         func(v0 context.Context, v1 *sourcegraph.RepoSpec) (*sourcegraph.Repo, error)
	Resolve                     func(v0 context.Context, v1 *sourcegraph.RepoResolveOp) (*sourcegraph.RepoResolution, error)
	List                        func(v0 context.Context, v1 *sourcegraph.RepoListOptions) (*sourcegraph.RepoList, error)
	ListStarredRepos            func(v0 context.Context, v1 *gogithub.ActivityListStarredOptions) (*sourcegraph.RepoList, error)
	ListContributors            func(v0 context.Context, v1 *gogithub.ListContributorsOptions) ([]*sourcegraph.Contributor, error)
	Create                      func(v0 context.Context, v1 *sourcegraph.ReposCreateOp) (*sourcegraph.Repo, error)
	Update                      func(v0 context.Context, v1 *sourcegraph.ReposUpdateOp) error
	GetConfig                   func(v0 context.Context, v1 *sourcegraph.RepoSpec) (*sourcegraph.RepoConfig, error)
	GetCommit                   func(v0 context.Context, v1 *sourcegraph.RepoRevSpec) (*vcs.Commit, error)
	ResolveRev                  func(v0 context.Context, v1 *sourcegraph.ReposResolveRevOp) (*sourcegraph.ResolvedRev, error)
	ListCommits                 func(v0 context.Context, v1 *sourcegraph.ReposListCommitsOp) (*sourcegraph.CommitList, error)
	ListBranches                func(v0 context.Context, v1 *sourcegraph.ReposListBranchesOp) (*sourcegraph.BranchList, error)
	ListTags                    func(v0 context.Context, v1 *sourcegraph.ReposListTagsOp) (*sourcegraph.TagList, error)
	ListDeps                    func(v0 context.Context, v1 *sourcegraph.URIList) (*sourcegraph.URIList, error)
	ListCommitters              func(v0 context.Context, v1 *sourcegraph.ReposListCommittersOp) (*sourcegraph.CommitterList, error)
	GetSrclibDataVersionForPath func(v0 context.Context, v1 *sourcegraph.TreeEntrySpec) (*sourcegraph.SrclibDataVersion, error)
	GetInventory                func(v0 context.Context, v1 *sourcegraph.RepoRevSpec) (*inventory.Inventory, error)
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

func (s *MockRepos) MockGet_Path(t *testing.T, wantRepo int32, repoPath string) (called *bool) {
	called = new(bool)
	s.Get = func(ctx context.Context, repo *sourcegraph.RepoSpec) (*sourcegraph.Repo, error) {
		*called = true
		if repo.ID != wantRepo {
			t.Errorf("got repo %d, want %d", repo.ID, wantRepo)
			return nil, legacyerr.Errorf(legacyerr.NotFound, "repo %d not found", wantRepo)
		}
		return &sourcegraph.Repo{ID: repo.ID, URI: repoPath}, nil
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

func (s *MockRepos) MockResolve_Local(t *testing.T, wantPath string, repoID int32) (called *bool) {
	called = new(bool)
	s.Resolve = func(ctx context.Context, op *sourcegraph.RepoResolveOp) (*sourcegraph.RepoResolution, error) {
		*called = true
		if op.Path != wantPath {
			t.Errorf("got repo %q, want %q", op.Path, wantPath)
			return nil, legacyerr.Errorf(legacyerr.NotFound, "repo path %s resolution failed", wantPath)
		}
		return &sourcegraph.RepoResolution{Repo: repoID, CanonicalPath: wantPath}, nil
	}
	return
}

func (s *MockRepos) MockResolve_Remote(t *testing.T, wantPath string, resolved *sourcegraph.Repo) (called *bool) {
	called = new(bool)
	s.Resolve = func(ctx context.Context, op *sourcegraph.RepoResolveOp) (*sourcegraph.RepoResolution, error) {
		*called = true
		if op.Path != wantPath {
			t.Errorf("got repo %q, want %q", op.Path, wantPath)
			return nil, legacyerr.Errorf(legacyerr.NotFound, "repo path %s resolution failed", wantPath)
		}
		return &sourcegraph.RepoResolution{RemoteRepo: resolved}, nil
	}
	return
}

func (s *MockRepos) MockResolve_NotFound(t *testing.T, wantPath string) (called *bool) {
	called = new(bool)
	s.Resolve = func(ctx context.Context, op *sourcegraph.RepoResolveOp) (*sourcegraph.RepoResolution, error) {
		*called = true
		if op.Path != wantPath {
			t.Errorf("got repo %q, want %q", op.Path, wantPath)
		}
		return nil, legacyerr.Errorf(legacyerr.NotFound, "repo path %s resolution failed", wantPath)
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

func (s *MockRepos) MockListStarredRepos(t *testing.T, wantRepos ...string) (called *bool) {
	called = new(bool)
	s.ListStarredRepos = func(ctx context.Context, opt *gogithub.ActivityListStarredOptions) (*sourcegraph.RepoList, error) {
		*called = true
		repos := make([]*sourcegraph.Repo, len(wantRepos))
		for i, repo := range wantRepos {
			repos[i] = &sourcegraph.Repo{URI: repo}
		}
		return &sourcegraph.RepoList{Repos: repos}, nil
	}
	return
}

func (s *MockRepos) MockListContributors(t *testing.T, wantContributors ...string) (called *bool) {
	called = new(bool)
	s.ListContributors = func(ctx context.Context, v1 *gogithub.ListContributorsOptions) ([]*sourcegraph.Contributor, error) {
		*called = true
		contributors := make([]*sourcegraph.Contributor, len(wantContributors))
		for i, contributor := range wantContributors {
			contributors[i] = &sourcegraph.Contributor{Login: contributor}
		}
		return contributors, nil

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

func (s *MockRepos) MockGetSrclibDataVersionForPath_Current(t *testing.T) (called *bool) {
	called = new(bool)
	s.GetSrclibDataVersionForPath = func(ctx context.Context, entry *sourcegraph.TreeEntrySpec) (*sourcegraph.SrclibDataVersion, error) {
		*called = true
		return &sourcegraph.SrclibDataVersion{CommitID: entry.RepoRev.CommitID}, nil
	}
	return
}
