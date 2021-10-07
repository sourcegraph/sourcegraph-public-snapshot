package dbstore

import (
	"context"
	"database/sql"

	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/metrics"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type Store struct {
	*basestore.Store
	operations *Operations
}

func NewWithDB(db dbutil.DB, observationContext *observation.Context, metrics *metrics.OperationMetrics) *Store {
	if metrics == nil {
		metrics = NewOperationsMetrics(observationContext)
	}

	return &Store{
		Store:      basestore.NewWithDB(db, sql.TxOptions{}),
		operations: NewOperationsFromMetrics(observationContext, metrics),
	}
}

func (s *Store) With(other basestore.ShareableStore) *Store {
	return &Store{
		Store:      s.Store.With(other),
		operations: s.operations,
	}
}

func (s *Store) Transact(ctx context.Context) (*Store, error) {
	txBase, err := s.Store.Transact(ctx)
	if err != nil {
		return nil, err
	}

	return &Store{
		Store:      txBase,
		operations: s.operations,
	}, nil
}
