package dbstore

import (
	"context"

	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/migration/schemas"
	"github.com/sourcegraph/sourcegraph/internal/metrics"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type Store struct {
	*basestore.Store[schemas.Production]
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
		Store:      basestore.NewWithHandle(db.Handle()),
		operations: newOperations(observationContext, operationsMetrics),
	}
}

func (s *Store) With(other basestore.ShareableStore[schemas.Production]) *Store {
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
		Store:      txBase,
		operations: s.operations,
	}, nil
}

// intsToQueries converts a slice of ints into a slice of queries.
func intsToQueries(values []int) []*sqlf.Query {
	var queries []*sqlf.Query
	for _, value := range values {
		queries = append(queries, sqlf.Sprintf("%d", value))
	}

	return queries
}
