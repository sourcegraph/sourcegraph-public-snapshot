package dependencies

import (
	"fmt"

	"github.com/sourcegraph/sourcegraph/internal/metrics"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type operations struct {
	listDependencyRepos       *observation.Operation
	upsertDependencyRepos     *observation.Operation
	deleteDependencyReposByID *observation.Operation
}

func newOperations(observationContext *observation.Context) *operations {
	m := metrics.NewREDMetrics(
		observationContext.Registerer,
		"codeintel_dependencies",
		metrics.WithLabels("op"),
		metrics.WithCountHelp("Total number of method invocations."),
	)

	op := func(name string) *observation.Operation {
		return observationContext.Operation(observation.Op{
			Name:              fmt.Sprintf("codeintel.dependencies.%s", name),
			MetricLabelValues: []string{name},
			Metrics:           m,
		})
	}

	return &operations{
		listDependencyRepos:       op("ListDependencyRepos"),
		upsertDependencyRepos:     op("UpsertDependencyRepos"),
		deleteDependencyReposByID: op("DeleteDependencyReposByID"),
	}
}
