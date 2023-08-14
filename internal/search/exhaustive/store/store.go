package store

import (
	"fmt"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/metrics"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

// Store exposes methods to read and write to the DB for exhaustive searches.
type Store struct {
	logger log.Logger
	*basestore.Store
	operations     *operations
	observationCtx *observation.Context
}

// New returns a new Store backed by the given database.
func New(db database.DB, observationCtx *observation.Context) *Store {
	return &Store{
		logger:         observationCtx.Logger,
		Store:          basestore.NewWithHandle(db.Handle()),
		operations:     newOperations(observationCtx),
		observationCtx: observationCtx,
	}
}

type operations struct {
	createExhaustiveSearchJob *observation.Operation

	createExhaustiveSearchRepoJob *observation.Operation

	createExhaustiveSearchRepoRevisionJob *observation.Operation
}

var m = new(metrics.SingletonREDMetrics)

func newOperations(observationCtx *observation.Context) *operations {
	redMetrics := m.Get(func() *metrics.REDMetrics {
		return metrics.NewREDMetrics(
			observationCtx.Registerer,
			"exhaustive_search",
			metrics.WithLabels("op"),
			metrics.WithCountHelp("Total number of method invocations."),
		)
	})

	op := func(name string) *observation.Operation {
		return observationCtx.Operation(observation.Op{
			Name:              fmt.Sprintf("search.exhaustive.%s", name),
			MetricLabelValues: []string{name},
			Metrics:           redMetrics,
		})
	}

	return &operations{
		createExhaustiveSearchJob: op("CreateExhaustiveSearchJob"),

		createExhaustiveSearchRepoJob: op("CreateExhaustiveSearchRepoJob"),

		createExhaustiveSearchRepoRevisionJob: op("CreateExhaustiveSearchRepoRevisionJob"),
	}
}
