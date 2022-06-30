package lsifstore

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/stores"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type LsifStore interface {
	Clear(ctx context.Context, bundleIDs ...int) (err error)
}

type store struct {
	db         *basestore.Store
	operations *operations
}

func New(db stores.CodeIntelDB, observationContext *observation.Context) LsifStore {
	return &store{
		db:         basestore.NewWithHandle(db.Handle()),
		operations: newOperations(observationContext),
	}
}

func (s *store) Transact(ctx context.Context) (*store, error) {
	tx, err := s.db.Transact(ctx)
	if err != nil {
		return nil, err
	}

	return &store{
		db:         tx,
		operations: s.operations,
	}, nil
}
