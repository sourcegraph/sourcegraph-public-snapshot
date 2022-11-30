package dependencies

import (
	"fmt"

	"github.com/sourcegraph/sourcegraph/internal/memo"
	"github.com/sourcegraph/sourcegraph/internal/metrics"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type operations struct {
	listDependencyRepos       *observation.Operation
	upsertDependencyRepos     *observation.Operation
	deleteDependencyReposByID *observation.Operation
}

var m = memo.NewMemoizedConstructorWithArg(func(observationContext *observation.Context) (*metrics.REDMetrics, error) {
	return metrics.NewREDMetrics(
		observationContext.Registerer,
		"codeintel_dependencies",
		metrics.WithLabels("op"),
		metrics.WithCountHelp("Total number of method invocations."),
	), nil
})

func newOperations(observationContext *observation.Context) *operations {
	m, _ := m.Init(observationContext)

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
