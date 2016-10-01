package backend

import (
	"testing"

	"context"

	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/store"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/vcs"
	vcstest "sourcegraph.com/sourcegraph/sourcegraph/pkg/vcs/testing"
	"sourcegraph.com/sqs/pbtypes"
)

func TestRefreshVCS(t *testing.T) {
	ctx, mock := testContext()
	var updatedEverything bool
	mock.stores.Repos.MockGet(t, 1)
	mock.stores.RepoVCS.MockOpen(t, 1, vcstest.MockRepository{
		Branches_: func(_ vcs.BranchesOptions) ([]*vcs.Branch, error) {
			return []*vcs.Branch{}, nil
		},
		UpdateEverything_: func(_ vcs.RemoteOpts) (*vcs.UpdateResult, error) {
			updatedEverything = true
			return &vcs.UpdateResult{Changes: []vcs.Change{}}, nil
		},
	})
	calledInternalUpdate := mock.stores.Repos.MockInternalUpdate(t)

	_, err := MirrorRepos.RefreshVCS(ctx, &sourcegraph.MirrorReposRefreshVCSOp{Repo: 1})
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
	ctx, mock := testContext()
	var cloned bool
	mock.stores.Repos.MockGet(t, 1)
	mock.servers.Repos.MockResolveRev_NoCheck(t, "deadbeef")
	mock.stores.RepoVCS.MockOpen(t, 1, vcstest.MockRepository{
		Branches_: func(_ vcs.BranchesOptions) ([]*vcs.Branch, error) {
			return nil, vcs.RepoNotExistError{}
		},
	})
	mock.stores.RepoVCS.Clone = func(_ context.Context, _ int32, _ *store.CloneInfo) error {
		cloned = true
		return nil
	}
	mock.servers.Async.RefreshIndexes_ = func(v0 context.Context, v1 *sourcegraph.AsyncRefreshIndexesOp) (*pbtypes.Void, error) {
		return &pbtypes.Void{}, nil
	}
	calledInternalUpdate := mock.stores.Repos.MockInternalUpdate(t)

	_, err := MirrorRepos.RefreshVCS(ctx, &sourcegraph.MirrorReposRefreshVCSOp{Repo: 1})
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
	ctx, mock := testContext()
	mock.stores.Repos.MockGet(t, 1)
	mock.servers.Repos.MockResolveRev_NoCheck(t, "deadbeef")
	mock.stores.RepoVCS.MockOpen(t, 1, vcstest.MockRepository{
		Branches_: func(_ vcs.BranchesOptions) ([]*vcs.Branch, error) {
			return nil, vcs.RepoNotExistError{}
		},
	})
	mock.stores.RepoVCS.Clone = func(_ context.Context, _ int32, _ *store.CloneInfo) error {
		return vcs.ErrRepoExist
	}
	mock.servers.Async.RefreshIndexes_ = func(v0 context.Context, v1 *sourcegraph.AsyncRefreshIndexesOp) (*pbtypes.Void, error) {
		return &pbtypes.Void{}, nil
	}
	calledInternalUpdate := mock.stores.Repos.MockInternalUpdate(t)

	_, err := MirrorRepos.RefreshVCS(ctx, &sourcegraph.MirrorReposRefreshVCSOp{Repo: 1})
	if err != nil {
		t.Fatalf("RefreshVCS call failed: %s", err)
	}
	if !*calledInternalUpdate {
		t.Error("!calledInternalUpdate")
	}
}
