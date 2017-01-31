package backend

import (
	"context"

	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/services/backend/internal/localstore"
)

var Pkgs = &pkgs{}

type pkgs struct{}

// UnsafeRefreshIndex refreshes the package index for the specified repository.
// It is safe to invoke on both public and private repositories, as read access
// is verified after query time (i.e. in localstore.Pkgs.ListPackages).
//
// SECURITY: It is the caller's responsibility to ensure that invoking this
// function does not leak existence of a private repository. For example,
// returning error or success to a user would cause a security issue. Also
// waiting for this method to complete before returning to the user leaks
// existence via timing information alone. Generally, only the indexer should
// invoke this method.
func (p *pkgs) UnsafeRefreshIndex(ctx context.Context, op *sourcegraph.DefsRefreshIndexOp) (err error) {
	if Mocks.Pkgs.UnsafeRefreshIndex != nil {
		return Mocks.Pkgs.UnsafeRefreshIndex(ctx, op)
	}

	ctx, done := trace(ctx, "Pkgs", "UnsafeRefreshIndex", op, &err)
	defer done()

	inv, err := Repos.GetInventory(ctx, &sourcegraph.RepoRevSpec{Repo: op.RepoID, CommitID: op.CommitID})
	if err != nil {
		return err
	}
	return localstore.Pkgs.UnsafeRefreshIndex(ctx, op, inv.Languages)
}

func (p *pkgs) ListPackages(ctx context.Context, op *sourcegraph.ListPackagesOp) (pkgs []sourcegraph.PackageInfo, err error) {
	if Mocks.Pkgs.ListPackages != nil {
		return Mocks.Pkgs.ListPackages(ctx, op)
	}
	return localstore.Pkgs.ListPackages(ctx, op)
}

type MockPkgs struct {
	UnsafeRefreshIndex func(ctx context.Context, op *sourcegraph.DefsRefreshIndexOp) error
	ListPackages       func(ctx context.Context, op *sourcegraph.ListPackagesOp) (pkgs []sourcegraph.PackageInfo, err error)
}
