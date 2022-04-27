package store

import (
	"fmt"

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
	withMigrationLog  *observation.Operation
}

func NewOperations(observationContext *observation.Context) *Operations {
	metrics := metrics.NewREDMetrics(
		observationContext.Registerer,
		"migrations",
		metrics.WithLabels("op"),
		metrics.WithCountHelp("Total number of method invocations."),
	)

	op := func(name string) *observation.Operation {
		return observationContext.Operation(observation.Op{
			Name:              fmt.Sprintf("migrations.%s", name),
			MetricLabelValues: []string{name},
			Metrics:           metrics,
		})
	}

	return &Operations{
		describe:          op("Describe"),
		down:              op("Down"),
		ensureSchemaTable: op("EnsureSchemaTable"),
		indexStatus:       op("IndexStatus"),
		tryLock:           op("TryLock"),
		up:                op("Up"),
		versions:          op("Versions"),
		withMigrationLog:  op("WithMigrationLog"),
	}
}
