package backend

import (
	"context"

	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/services/backend/internal/localstore"
	"sourcegraph.com/sqs/pbtypes"
)

var RepoStatuses = &repoStatuses{}

type repoStatuses struct{}

func (s *repoStatuses) GetCombined(ctx context.Context, repoRev *sourcegraph.RepoRevSpec) (res *sourcegraph.CombinedStatus, err error) {
	if Mocks.RepoStatuses.GetCombined != nil {
		return Mocks.RepoStatuses.GetCombined(ctx, repoRev)
	}

	ctx, done := trace(ctx, "RepoStatuses", "GetCombined", repoRev, &err)
	defer done()

	return localstore.RepoStatuses.GetCombined(ctx, repoRev.Repo, repoRev.CommitID)
}

func (s *repoStatuses) GetCoverage(ctx context.Context, _ *pbtypes.Void) (res *sourcegraph.RepoStatusList, err error) {
	if Mocks.RepoStatuses.GetCoverage != nil {
		return Mocks.RepoStatuses.GetCoverage(ctx, &pbtypes.Void{})
	}

	ctx, done := trace(ctx, "RepoStatuses", "GetCoverage", nil, &err)
	defer done()

	return localstore.RepoStatuses.GetCoverage(ctx)
}

func (s *repoStatuses) Create(ctx context.Context, op *sourcegraph.RepoStatusesCreateOp) (res *sourcegraph.RepoStatus, err error) {
	if Mocks.RepoStatuses.Create != nil {
		return Mocks.RepoStatuses.Create(ctx, op)
	}

	ctx, done := trace(ctx, "RepoStatuses", "Create", op, &err)
	defer done()

	repoRev := op.Repo
	status := &op.Status

	if err := localstore.RepoStatuses.Create(ctx, repoRev.Repo, repoRev.CommitID, status); err != nil {
		return nil, err
	}
	return status, nil
}

type MockRepoStatuses struct {
	GetCombined func(v0 context.Context, v1 *sourcegraph.RepoRevSpec) (*sourcegraph.CombinedStatus, error)
	GetCoverage func(v0 context.Context, v1 *pbtypes.Void) (*sourcegraph.RepoStatusList, error)
	Create      func(v0 context.Context, v1 *sourcegraph.RepoStatusesCreateOp) (*sourcegraph.RepoStatus, error)
}
