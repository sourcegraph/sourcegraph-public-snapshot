package pgsql

import (
	"golang.org/x/net/context"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/server/accesscontrol"
	"src.sourcegraph.com/sourcegraph/store"
)

type repoStatuses struct{}

var _ store.RepoStatuses = (*repoStatuses)(nil)

func (s *repoStatuses) GetCombined(ctx context.Context, repoRev sourcegraph.RepoRevSpec) (*sourcegraph.CombinedStatus, error) {
	if err := accesscontrol.VerifyUserHasReadAccess(ctx, "RepoStatuses.GetCombined", repoRev.URI); err != nil {
		return nil, err
	}
	// Not yet implemented
	return &sourcegraph.CombinedStatus{}, nil
}

func (s *repoStatuses) Create(ctx context.Context, repoRev sourcegraph.RepoRevSpec, status *sourcegraph.RepoStatus) error {
	if err := accesscontrol.VerifyUserHasWriteAccess(ctx, "RepoStatuses.Create", repoRev.URI); err != nil {
		return err
	}
	// Not yet implemented (no-op)
	return nil
}
