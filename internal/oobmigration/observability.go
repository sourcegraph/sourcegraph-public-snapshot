package oobmigration

import (
	"fmt"
	"strconv"

	"github.com/sourcegraph/sourcegraph/internal/metrics"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type operations struct {
	upForMigration   func(migrationID int) *observation.Operation
	downForMigration func(migrationID int) *observation.Operation
}

func newOperations(observationContext *observation.Context) *operations {
	metrics := metrics.NewOperationMetrics(
		observationContext.Registerer,
		"oobmigration",
		metrics.WithLabels("op", "migration"),
		metrics.WithCountHelp("Total number of migrator invocations."),
	)

	opForMigration := func(name string) func(migrationID int) *observation.Operation {
		return func(migrationID int) *observation.Operation {
			return observationContext.Operation(observation.Op{
				Name:              fmt.Sprintf("oobmigration.%s", name),
				MetricLabelValues: []string{name, strconv.Itoa(migrationID)},
				Metrics:           metrics,
			})
		}
	}

	return &operations{
		upForMigration:   opForMigration("up"),
		downForMigration: opForMigration("down"),
	}
}
