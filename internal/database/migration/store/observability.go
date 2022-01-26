package store

import (
	"fmt"

	"github.com/sourcegraph/sourcegraph/internal/metrics"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type Operations struct {
	down              *observation.Operation
	ensureSchemaTable *observation.Operation
	lock              *observation.Operation
	tryLock           *observation.Operation
	up                *observation.Operation
	version           *observation.Operation
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
		down:              op("Down"),
		ensureSchemaTable: op("EnsureSchemaTable"),
		lock:              op("Lock"),
		tryLock:           op("TryLock"),
		up:                op("Up"),
		version:           op("Version"),
	}
}
