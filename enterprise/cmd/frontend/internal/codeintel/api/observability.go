package api

import (
	"fmt"

	"github.com/sourcegraph/sourcegraph/internal/metrics"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type operations struct {
	definitions      *observation.Operation
	diagnostics      *observation.Operation
	findClosestDumps *observation.Operation
	hover            *observation.Operation
	ranges           *observation.Operation
	references       *observation.Operation
}

func makeOperations(observationContext *observation.Context) *operations {
	metrics := metrics.NewOperationMetrics(
		observationContext.Registerer,
		"codeintel_api",
		metrics.WithLabels("op"),
		metrics.WithCountHelp("Total number of method invocations."),
	)

	op := func(name string) *observation.Operation {
		return observationContext.Operation(observation.Op{
			Name:         fmt.Sprintf("codeintel.api.%s", name),
			MetricLabels: []string{name},
			Metrics:      metrics,
		})
	}

	return &operations{
		definitions:      op("Definitions"),
		diagnostics:      op("Diagnostics"),
		findClosestDumps: op("FindClosestDumps"),
		hover:            op("Hover"),
		ranges:           op("Ranges"),
		references:       op("References"),
	}
}
