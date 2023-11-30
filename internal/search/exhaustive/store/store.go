package store

import (
	"context"
	"fmt"

	"github.com/sourcegraph/log"
	"go.opentelemetry.io/otel/attribute"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/metrics"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// ErrNoResults is returned by Store method calls that found no results.
var ErrNoResults = errors.New("no results")

// Store exposes methods to read and write to the DB for exhaustive searches.
type Store struct {
	logger log.Logger
	db     database.DB
	*basestore.Store
	operations     *operations
	observationCtx *observation.Context
}

// New returns a new Store backed by the given database.
func New(db database.DB, observationCtx *observation.Context) *Store {
	return &Store{
		logger:         observationCtx.Logger,
		db:             db,
		Store:          basestore.NewWithHandle(db.Handle()),
		operations:     newOperations(observationCtx),
		observationCtx: observationCtx,
	}
}

// Transact creates a new transaction.
// It's required to implement this method and wrap the Transact method of the
// underlying basestore.Store.
func (s *Store) Transact(ctx context.Context) (*Store, error) {
	txBase, err := s.Store.Transact(ctx)
	if err != nil {
		return nil, err
	}
	return &Store{
		logger:         s.logger,
		db:             s.db,
		Store:          txBase,
		operations:     s.operations,
		observationCtx: s.observationCtx,
	}, nil
}

func opAttrs(attrs ...attribute.KeyValue) observation.Args {
	return observation.Args{Attrs: attrs}
}

type operations struct {
	createExhaustiveSearchJob *observation.Operation
	cancelSearchJob           *observation.Operation
	getExhaustiveSearchJob    *observation.Operation
	userHasAccess             *observation.Operation
	listExhaustiveSearchJobs  *observation.Operation
	deleteExhaustiveSearchJob *observation.Operation

	createExhaustiveSearchRepoJob         *observation.Operation
	createExhaustiveSearchRepoRevisionJob *observation.Operation
	getAggregateRepoRevState              *observation.Operation
}

var m = new(metrics.SingletonREDMetrics)

func newOperations(observationCtx *observation.Context) *operations {
	redMetrics := m.Get(func() *metrics.REDMetrics {
		return metrics.NewREDMetrics(
			observationCtx.Registerer,
			"searchjobs_store",
			metrics.WithLabels("op"),
			metrics.WithCountHelp("Total number of method invocations."),
		)
	})

	op := func(name string) *observation.Operation {
		return observationCtx.Operation(observation.Op{
			Name:              fmt.Sprintf("searchjobs.store.%s", name),
			MetricLabelValues: []string{name},
			Metrics:           redMetrics,
		})
	}

	return &operations{
		createExhaustiveSearchJob: op("CreateExhaustiveSearchJob"),
		cancelSearchJob:           op("CancelSearchJob"),
		getExhaustiveSearchJob:    op("GetExhaustiveSearchJob"),
		userHasAccess:             op("UserHasAccess"),
		listExhaustiveSearchJobs:  op("ListExhaustiveSearchJobs"),
		deleteExhaustiveSearchJob: op("DeleteExhaustiveSearchJob"),

		createExhaustiveSearchRepoJob:         op("CreateExhaustiveSearchRepoJob"),
		createExhaustiveSearchRepoRevisionJob: op("CreateExhaustiveSearchRepoRevisionJob"),
		getAggregateRepoRevState:              op("GetAggregateRepoRevState"),
	}
}
