package db

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/api"
)

func NewNoopInserter() VectorInserter {
	return noopInserter{}
}

var _ VectorDB = noopInserter{}

type noopInserter struct{}

func (noopInserter) Search(context.Context, SearchParams) ([]ChunkResult, error) {
	return nil, nil
}
func (noopInserter) PrepareUpdate(ctx context.Context, modelID string, modelDims uint64) error {
	return nil
}
func (noopInserter) HasIndex(ctx context.Context, modelID string, repoID api.RepoID, revision api.CommitID) (bool, error) {
	return false, nil
}
func (noopInserter) InsertChunks(context.Context, InsertParams) error {
	return nil
}
func (noopInserter) FinalizeUpdate(context.Context, FinalizeUpdateParams) error {
	return nil
}
