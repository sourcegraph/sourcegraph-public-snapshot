package dbstore

import (
	"context"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/metrics"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type Store struct {
	logger log.Logger
	*basestore.Store
	operations *operations
}

func NewWithDB(db database.DB, observationContext *observation.Context) *Store {
	operationsMetrics := metrics.NewREDMetrics(
		observationContext.Registerer,
		"codeintel_dbstore",
		metrics.WithLabels("op"),
		metrics.WithCountHelp("Total number of method invocations."),
	)

	return &Store{
		logger:     log.Scoped("dbstore", ""),
		Store:      basestore.NewWithHandle(db.Handle()),
		operations: newOperations(observationContext, operationsMetrics),
	}
}

func (s *Store) With(other basestore.ShareableStore) *Store {
	return &Store{
		Store:      s.Store.With(other),
		operations: s.operations,
	}
}

func (s *Store) Transact(ctx context.Context) (*Store, error) {
	return s.transact(ctx)
}

func (s *Store) transact(ctx context.Context) (*Store, error) {
	txBase, err := s.Store.Transact(ctx)
	if err != nil {
		return nil, err
	}

	return &Store{
		logger:     s.logger,
		Store:      txBase,
		operations: s.operations,
	}, nil
}
