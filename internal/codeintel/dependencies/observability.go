package dependencies

import (
	"fmt"

	"github.com/sourcegraph/sourcegraph/internal/metrics"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type operations struct {
	dependencies                           *observation.Operation
	resolveLockfileDependenciesFromArchive *observation.Operation
	resolveLockfileDependenciesFromStore   *observation.Operation
}

func newOperations(observationContext *observation.Context) *operations {
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

	return &operations{
		dependencies:                           op("Dependencies"),
		resolveLockfileDependenciesFromArchive: op("resolveLockfileDependenciesFromArchive"),
		resolveLockfileDependenciesFromStore:   op("resolveLockfileDependenciesFromStore"),
	}
}
