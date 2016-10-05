package backend

import (
	"testing"

	"context"

	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/vcs"
	vcstest "sourcegraph.com/sourcegraph/sourcegraph/pkg/vcs/testing"
	"sourcegraph.com/sourcegraph/sourcegraph/services/backend/internal/localstore"
)

func TestRefreshVCS(t *testing.T) {
	ctx := testContext()
	var updatedEverything bool
	localstore.Mocks.Repos.MockGet(t, 1)
	localstore.Mocks.RepoVCS.MockOpen(t, 1, vcstest.MockRepository{
		Branches_: func(ctx context.Context, _ vcs.BranchesOptions) ([]*vcs.Branch, error) {
			return []*vcs.Branch{}, nil
		},
		UpdateEverything_: func(ctx context.Context, _ vcs.RemoteOpts) (*vcs.UpdateResult, error) {
			updatedEverything = true
			return &vcs.UpdateResult{Changes: []vcs.Change{}}, nil
		},
	})
	calledInternalUpdate := localstore.Mocks.Repos.MockInternalUpdate(t)

	err := MirrorRepos.RefreshVCS(ctx, &sourcegraph.MirrorReposRefreshVCSOp{Repo: 1})
	if !updatedEverything {
		t.Error("Did not call UpdateEverything")
	}
	if err != nil {
		t.Fatalf("RefreshVCS call failed: %s", err)
	}
	if !*calledInternalUpdate {
		t.Error("!calledInternalUpdate")
	}
}

func TestRefreshVCS_cloneRepo(t *testing.T) {
	skipCloneRepoAsyncSteps = true
	ctx := testContext()
	var cloned bool
	localstore.Mocks.Repos.MockGet(t, 1)
	Mocks.Repos.MockResolveRev_NoCheck(t, "deadbeef")
	localstore.Mocks.RepoVCS.MockOpen(t, 1, vcstest.MockRepository{
		Branches_: func(ctx context.Context, _ vcs.BranchesOptions) ([]*vcs.Branch, error) {
			return nil, vcs.RepoNotExistError{}
		},
	})
	localstore.Mocks.RepoVCS.Clone = func(_ context.Context, _ int32, _ *localstore.CloneInfo) error {
		cloned = true
		return nil
	}
	Mocks.Async.RefreshIndexes = func(v0 context.Context, v1 *sourcegraph.AsyncRefreshIndexesOp) error {
		return nil
	}
	calledInternalUpdate := localstore.Mocks.Repos.MockInternalUpdate(t)

	err := MirrorRepos.RefreshVCS(ctx, &sourcegraph.MirrorReposRefreshVCSOp{Repo: 1})
	if !cloned {
		t.Error("RefreshVCS did not clone missing repo")
	}
	if err != nil {
		t.Fatalf("RefreshVCS call failed: %s", err)
	}
	if !*calledInternalUpdate {
		t.Error("!calledInternalUpdate")
	}
}

func TestRefreshVCS_cloneRepoExists(t *testing.T) {
	skipCloneRepoAsyncSteps = true
	ctx := testContext()
	localstore.Mocks.Repos.MockGet(t, 1)
	Mocks.Repos.MockResolveRev_NoCheck(t, "deadbeef")
	localstore.Mocks.RepoVCS.MockOpen(t, 1, vcstest.MockRepository{
		Branches_: func(ctx context.Context, _ vcs.BranchesOptions) ([]*vcs.Branch, error) {
			return nil, vcs.RepoNotExistError{}
		},
	})
	localstore.Mocks.RepoVCS.Clone = func(_ context.Context, _ int32, _ *localstore.CloneInfo) error {
		return vcs.ErrRepoExist
	}
	Mocks.Async.RefreshIndexes = func(v0 context.Context, v1 *sourcegraph.AsyncRefreshIndexesOp) error {
		return nil
	}
	calledInternalUpdate := localstore.Mocks.Repos.MockInternalUpdate(t)

	err := MirrorRepos.RefreshVCS(ctx, &sourcegraph.MirrorReposRefreshVCSOp{Repo: 1})
	if err != nil {
		t.Fatalf("RefreshVCS call failed: %s", err)
	}
	if !*calledInternalUpdate {
		t.Error("!calledInternalUpdate")
	}
}
