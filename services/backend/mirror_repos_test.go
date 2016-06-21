package backend

import (
	"errors"
	"testing"

	"golang.org/x/net/context"

	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/store"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/vcs"
	vcstest "sourcegraph.com/sourcegraph/sourcegraph/pkg/vcs/testing"
)

func TestRefreshVCS(t *testing.T) {
	ctx, mock := testContext()
	var updatedEverything bool
	mock.servers.Repos.MockGet(t, 1)
	mock.stores.RepoVCS.MockOpen(t, 1, vcstest.MockRepository{
		Branches_: func(_ vcs.BranchesOptions) ([]*vcs.Branch, error) {
			return []*vcs.Branch{}, nil
		},
		UpdateEverything_: func(_ vcs.RemoteOpts) (*vcs.UpdateResult, error) {
			updatedEverything = true
			return &vcs.UpdateResult{Changes: []vcs.Change{}}, nil
		},
	})
	mock.servers.Auth.GetExternalToken_ = func(v0 context.Context, v1 *sourcegraph.ExternalTokenSpec) (*sourcegraph.ExternalToken, error) {
		return nil, errors.New("mock")
	}
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
	var cloned, built bool
	mock.servers.Repos.MockGet(t, 1)
	mock.servers.Repos.MockResolveRev_NoCheck(t, "deadbeef")
	mock.stores.RepoVCS.MockOpen(t, 1, vcstest.MockRepository{
		Branches_: func(_ vcs.BranchesOptions) ([]*vcs.Branch, error) {
			return nil, vcs.RepoNotExistError{}
		},
	})
	mock.stores.RepoVCS.Clone_ = func(_ context.Context, _ int32, _ *store.CloneInfo) error {
		cloned = true
		return nil
	}
	mock.servers.Builds.Create_ = func(_ context.Context, _ *sourcegraph.BuildsCreateOp) (*sourcegraph.Build, error) {
		built = true
		return &sourcegraph.Build{}, nil
	}
	mock.servers.Auth.GetExternalToken_ = func(v0 context.Context, v1 *sourcegraph.ExternalTokenSpec) (*sourcegraph.ExternalToken, error) {
		return nil, errors.New("mock")
	}
	calledInternalUpdate := mock.stores.Repos.MockInternalUpdate(t)

	_, err := MirrorRepos.RefreshVCS(ctx, &sourcegraph.MirrorReposRefreshVCSOp{Repo: 1})
	if !cloned {
		t.Error("RefreshVCS did not clone missing repo")
	}
	if !built {
		t.Error("RefreshVCS did not build repo")
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
	var built bool
	mock.servers.Repos.MockGet(t, 1)
	mock.servers.Repos.MockResolveRev_NoCheck(t, "deadbeef")
	mock.stores.RepoVCS.MockOpen(t, 1, vcstest.MockRepository{
		Branches_: func(_ vcs.BranchesOptions) ([]*vcs.Branch, error) {
			return nil, vcs.RepoNotExistError{}
		},
	})
	mock.stores.RepoVCS.Clone_ = func(_ context.Context, _ int32, _ *store.CloneInfo) error {
		return vcs.ErrRepoExist
	}
	mock.servers.Builds.Create_ = func(_ context.Context, _ *sourcegraph.BuildsCreateOp) (*sourcegraph.Build, error) {
		built = true
		return &sourcegraph.Build{}, nil
	}
	mock.servers.Auth.GetExternalToken_ = func(v0 context.Context, v1 *sourcegraph.ExternalTokenSpec) (*sourcegraph.ExternalToken, error) {
		return nil, errors.New("mock")
	}
	calledInternalUpdate := mock.stores.Repos.MockInternalUpdate(t)

	_, err := MirrorRepos.RefreshVCS(ctx, &sourcegraph.MirrorReposRefreshVCSOp{Repo: 1})
	if !built {
		t.Error("RefreshVCS did not build repo")
	}
	if err != nil {
		t.Fatalf("RefreshVCS call failed: %s", err)
	}
	if !*calledInternalUpdate {
		t.Error("!calledInternalUpdate")
	}
}
