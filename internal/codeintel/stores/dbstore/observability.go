package dbstore

import (
	"fmt"

	"github.com/sourcegraph/sourcegraph/internal/metrics"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type Operations struct {
	repoName           *observation.Operation
	getJVMDependencies *observation.Operation
}

func NewOperationsMetrics(observationContext *observation.Context) *metrics.OperationMetrics {
	return metrics.NewOperationMetrics(
		observationContext.Registerer,
		"codeintel_dbstore",
		metrics.WithLabels("op"),
		metrics.WithCountHelp("Total number of method invocations."),
	)
}

func NewOperationsFromMetrics(observationContext *observation.Context, metrics *metrics.OperationMetrics) *Operations {
	op := func(name string) *observation.Operation {
		return observationContext.Operation(observation.Op{
			Name:         fmt.Sprintf("codeintel.dbstore.%s", name),
			MetricLabels: []string{name},
			Metrics:      metrics,
		})
	}

	return &Operations{
		repoName:           op("RepoName"),
		getJVMDependencies: op("GetJVMDependencies"),
	}
}
