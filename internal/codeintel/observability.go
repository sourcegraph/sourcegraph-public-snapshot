package codeintel

import (
	"fmt"

	"github.com/sourcegraph/sourcegraph/internal/metrics"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type dependencyServiceOperations struct {
	dependencies *observation.Operation
}

func newDependencyServiceOperations(observationContext *observation.Context) *dependencyServiceOperations {
	metrics := metrics.NewREDMetrics(
		observationContext.Registerer,
		"codeintel_dependencies",
		metrics.WithLabels("op"),
		metrics.WithCountHelp("Total number of method invocations."),
	)

	op := func(name string) *observation.Operation {
		return observationContext.Operation(observation.Op{
			Name:              fmt.Sprintf("codeintel.dependencies.%s", name),
			MetricLabelValues: []string{name},
			Metrics:           metrics,
		})
	}

	return &dependencyServiceOperations{
		dependencies: op("Dependencies"),
	}
}
