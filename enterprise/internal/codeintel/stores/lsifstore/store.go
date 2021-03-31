package lsifstore

import (
	"context"
	"database/sql"

	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type Store struct {
	*basestore.Store
	serializer *Serializer
	operations *operations
}

func NewStore(db dbutil.DB, observationContext *observation.Context) *Store {
	return &Store{
		Store:      basestore.NewWithHandle(basestore.NewHandleWithDB(db, sql.TxOptions{})),
		serializer: NewSerializer(),
		operations: newOperations(observationContext),
	}
}

func (s *Store) Transact(ctx context.Context) (*Store, error) {
	tx, err := s.Store.Transact(ctx)
	if err != nil {
		return nil, err
	}

	return &Store{
		Store:      tx,
		serializer: s.serializer,
		operations: s.operations,
	}, nil
}

func (s *Store) Done(err error) error {
	return s.Store.Done(err)
}
