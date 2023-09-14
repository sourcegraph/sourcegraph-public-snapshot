package db

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func NewNoopDB() VectorDB {
	return noopDB{}
}

var _ VectorDB = noopDB{}

type noopDB struct{}

func (noopDB) Search(context.Context, SearchParams) ([]ChunkResult, error) {
	return nil, nil
}
func (noopDB) PrepareUpdate(ctx context.Context, modelID string, modelDims uint64) error {
	return nil
}
func (noopDB) HasIndex(ctx context.Context, modelID string, repoID api.RepoID, revision api.CommitID) (bool, error) {
	return false, nil
}
func (noopDB) InsertChunks(context.Context, InsertParams) error {
	return nil
}
func (noopDB) FinalizeUpdate(context.Context, FinalizeUpdateParams) error {
	return nil
}

var ErrDisabled = errors.New("Qdrant is disabled. Enable by setting QDRANT_ENDPOINT")

func NewDisabledDB() VectorDB {
	return disabledDB{}
}

var _ VectorDB = disabledDB{}

type disabledDB struct{}

func (disabledDB) Search(context.Context, SearchParams) ([]ChunkResult, error) {
	return nil, ErrDisabled
}
func (disabledDB) PrepareUpdate(ctx context.Context, modelID string, modelDims uint64) error {
	return ErrDisabled
}
func (disabledDB) HasIndex(ctx context.Context, modelID string, repoID api.RepoID, revision api.CommitID) (bool, error) {
	return false, ErrDisabled
}
func (disabledDB) InsertChunks(context.Context, InsertParams) error {
	return ErrDisabled
}
func (disabledDB) FinalizeUpdate(context.Context, FinalizeUpdateParams) error {
	return ErrDisabled
}
