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

var m = new(metrics.SingletonREDMetrics)

func newOperations(observationCtx *observation.Context) *operations {
	m := m.Get(func() *metrics.REDMetrics {
		return metrics.NewREDMetrics(
			observationCtx.Registerer,
			"codeintel_dependencies",
			metrics.WithLabels("op"),
			metrics.WithCountHelp("Total number of method invocations."),
		)
	})

	op := func(name string) *observation.Operation {
		return observationCtx.Operation(observation.Op{
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
