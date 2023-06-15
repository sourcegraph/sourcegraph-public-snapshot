package store

import (
	"fmt"
	"sync"

	"github.com/sourcegraph/sourcegraph/internal/metrics"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type Operations struct {
	describe          *observation.Operation
	down              *observation.Operation
	ensureSchemaTable *observation.Operation
	indexStatus       *observation.Operation
	tryLock           *observation.Operation
	up                *observation.Operation
	versions          *observation.Operation
	runDDLStatements  *observation.Operation
	withMigrationLog  *observation.Operation
}

var (
	once sync.Once
	ops  *Operations
)

func NewOperations(observationCtx *observation.Context) *Operations {
	once.Do(func() {
		redMetrics := metrics.NewREDMetrics(
			observationCtx.Registerer,
			"migrations",
			metrics.WithLabels("op"),
			metrics.WithCountHelp("Total number of method invocations."),
		)

		op := func(name string) *observation.Operation {
			return observationCtx.Operation(observation.Op{
				Name:              fmt.Sprintf("migrations.%s", name),
				MetricLabelValues: []string{name},
				Metrics:           redMetrics,
			})
		}

		ops = &Operations{
			describe:          op("Describe"),
			down:              op("Down"),
			ensureSchemaTable: op("EnsureSchemaTable"),
			indexStatus:       op("IndexStatus"),
			tryLock:           op("TryLock"),
			up:                op("Up"),
			versions:          op("Versions"),
			runDDLStatements:  op("RunDDLStatements"),
			withMigrationLog:  op("WithMigrationLog"),
		}
	})
	return ops
}
