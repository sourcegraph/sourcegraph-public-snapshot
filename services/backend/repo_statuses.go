package backend

import (
	"context"

	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/store"
	"sourcegraph.com/sqs/pbtypes"
)

var RepoStatuses sourcegraph.RepoStatusesServer = &repoStatuses{}

type repoStatuses struct{}

var _ sourcegraph.RepoStatusesServer = (*repoStatuses)(nil)

func (s *repoStatuses) GetCombined(ctx context.Context, repoRev *sourcegraph.RepoRevSpec) (*sourcegraph.CombinedStatus, error) {
	return store.RepoStatusesFromContext(ctx).GetCombined(ctx, repoRev.Repo, repoRev.CommitID)
}

func (s *repoStatuses) GetCoverage(ctx context.Context, _ *pbtypes.Void) (*sourcegraph.RepoStatusList, error) {
	return store.RepoStatusesFromContext(ctx).GetCoverage(ctx)
}

func (s *repoStatuses) Create(ctx context.Context, op *sourcegraph.RepoStatusesCreateOp) (*sourcegraph.RepoStatus, error) {
	repoRev := op.Repo
	status := &op.Status

	if err := store.RepoStatusesFromContext(ctx).Create(ctx, repoRev.Repo, repoRev.CommitID, status); err != nil {
		return nil, err
	}
	return status, nil
}
