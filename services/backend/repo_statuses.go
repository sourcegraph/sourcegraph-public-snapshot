package backend

import (
	"fmt"

	"golang.org/x/net/context"
	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/store"
	"sourcegraph.com/sourcegraph/sourcegraph/services/svc"
	"sourcegraph.com/sqs/pbtypes"
)

var RepoStatuses sourcegraph.RepoStatusesServer = &repoStatuses{}

type repoStatuses struct{}

var _ sourcegraph.RepoStatusesServer = (*repoStatuses)(nil)

func (s *repoStatuses) GetCombined(ctx context.Context, repoRev *sourcegraph.RepoRevSpec) (*sourcegraph.CombinedStatus, error) {
	if repoRev == nil {
		return nil, fmt.Errorf("nil repo rev")
	}
	repo, err := svc.Repos(ctx).Get(ctx, &sourcegraph.RepoSpec{URI: repoRev.Repo})
	if err != nil {
		return nil, err
	}
	return store.RepoStatusesFromContext(ctx).GetCombined(ctx, repo.ID, repoRev.CommitID)
}

func (s *repoStatuses) GetCoverage(ctx context.Context, _ *pbtypes.Void) (*sourcegraph.RepoStatusList, error) {
	return store.RepoStatusesFromContext(ctx).GetCoverage(ctx)
}

func (s *repoStatuses) Create(ctx context.Context, op *sourcegraph.RepoStatusesCreateOp) (*sourcegraph.RepoStatus, error) {
	repoRev := op.Repo
	status := &op.Status

	repo, err := svc.Repos(ctx).Get(ctx, &sourcegraph.RepoSpec{URI: repoRev.Repo})
	if err != nil {
		return nil, err
	}

	if err := store.RepoStatusesFromContext(ctx).Create(ctx, repo.ID, repoRev.CommitID, status); err != nil {
		return nil, err
	}
	return status, nil
}
