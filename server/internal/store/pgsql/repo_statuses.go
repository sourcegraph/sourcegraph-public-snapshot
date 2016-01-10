package pgsql

import (
	"golang.org/x/net/context"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/store"
)

type repoStatuses struct{}

var _ store.RepoStatuses = (*repoStatuses)(nil)

func (s *repoStatuses) GetCombined(ctx context.Context, repoRev sourcegraph.RepoRevSpec) (*sourcegraph.CombinedStatus, error) {
	// Not yet implemented
	return &sourcegraph.CombinedStatus{}, nil
}

func (s *repoStatuses) Create(ctx context.Context, repoRev sourcegraph.RepoRevSpec, status *sourcegraph.RepoStatus) error {
	// Not yet implemented (no-op)
	return nil
}
