package pgsql

import (
	"golang.org/x/net/context"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/store"
)

type RepoStatuses struct{}

var _ store.RepoStatuses = (*RepoStatuses)(nil)

func (s *RepoStatuses) GetCombined(ctx context.Context, repoRev sourcegraph.RepoRevSpec) (*sourcegraph.CombinedStatus, error) {
	// Not yet implemented
	return &sourcegraph.CombinedStatus{}, nil
}

func (s *RepoStatuses) Create(ctx context.Context, repoRev sourcegraph.RepoRevSpec, status *sourcegraph.RepoStatus) error {
	// Not yet implemented (no-op)
	return nil
}
