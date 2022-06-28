package indexing

import (
	"fmt"
	"sync"

	"github.com/sourcegraph/sourcegraph/internal/metrics"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type dependencyReposOperations struct {
	InsertCloneableDependencyRepo *observation.Operation
}

var (
	once               sync.Once
	dependencyReposOps *dependencyReposOperations
)

func newOperations(observationContext *observation.Context) *dependencyReposOperations {
	once.Do(func() {
		m := metrics.NewREDMetrics(
			observationContext.Registerer,
			"codeintel_dependency_repos",
			metrics.WithLabels("op", "scheme", "new"),
		)

		op := func(prefix, name string) *observation.Operation {
			return observationContext.Operation(observation.Op{
				Name:              fmt.Sprintf("codeintel.%s.%s", prefix, name),
				MetricLabelValues: []string{name},
				Metrics:           m,
			})
		}

		dependencyReposOps = &dependencyReposOperations{
			InsertCloneableDependencyRepo: op("dependencyrepos", "InsertCloneableDependencyRepo"),
		}
	})
	return dependencyReposOps
}
