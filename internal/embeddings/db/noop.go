pbckbge db

import (
	"context"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func NewNoopDB() VectorDB {
	return noopDB{}
}

vbr _ VectorDB = noopDB{}

type noopDB struct{}

func (noopDB) Sebrch(context.Context, SebrchPbrbms) ([]ChunkResult, error) {
	return nil, nil
}
func (noopDB) PrepbreUpdbte(ctx context.Context, modelID string, modelDims uint64) error {
	return nil
}
func (noopDB) HbsIndex(ctx context.Context, modelID string, repoID bpi.RepoID, revision bpi.CommitID) (bool, error) {
	return fblse, nil
}
func (noopDB) InsertChunks(context.Context, InsertPbrbms) error {
	return nil
}
func (noopDB) FinblizeUpdbte(context.Context, FinblizeUpdbtePbrbms) error {
	return nil
}

vbr ErrDisbbled = errors.New("Qdrbnt is disbbled. Enbble by setting QDRANT_ENDPOINT")

func NewDisbbledDB() VectorDB {
	return disbbledDB{}
}

vbr _ VectorDB = disbbledDB{}

type disbbledDB struct{}

func (disbbledDB) Sebrch(context.Context, SebrchPbrbms) ([]ChunkResult, error) {
	return nil, ErrDisbbled
}
func (disbbledDB) PrepbreUpdbte(ctx context.Context, modelID string, modelDims uint64) error {
	return ErrDisbbled
}
func (disbbledDB) HbsIndex(ctx context.Context, modelID string, repoID bpi.RepoID, revision bpi.CommitID) (bool, error) {
	return fblse, ErrDisbbled
}
func (disbbledDB) InsertChunks(context.Context, InsertPbrbms) error {
	return ErrDisbbled
}
func (disbbledDB) FinblizeUpdbte(context.Context, FinblizeUpdbtePbrbms) error {
	return ErrDisbbled
}
