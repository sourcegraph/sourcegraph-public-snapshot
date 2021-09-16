package indexing

import (
	"fmt"
	"sync"

	"github.com/sourcegraph/sourcegraph/internal/metrics"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type schedulerOperations struct {
	HandleIndexScheduler *observation.Operation
	QueueRepository      *observation.Operation
}

type dependencyReposOperations struct {
	InsertCloneableDependencyRepo *observation.Operation
}

var (
	schedulerOps       *schedulerOperations
	dependencyReposOps *dependencyReposOperations
	once               sync.Once
)

func newOperations(observationContext *observation.Context) *schedulerOperations {
	once.Do(func() {
		m := metrics.NewOperationMetrics(
			observationContext.Registerer,
			"codeintel_index_scheduler",
			metrics.WithLabels("op"),
			metrics.WithCountHelp("Total number of method invocations."),
		)

		op := func(prefix, name string) *observation.Operation {
			return observationContext.Operation(observation.Op{
				Name:              fmt.Sprintf("codeintel.%s.%s", prefix, name),
				MetricLabelValues: []string{name},
				Metrics:           m,
			})
		}

		schedulerOps = &schedulerOperations{
			HandleIndexScheduler: op("indexing", "HandleIndexSchedule"),
			QueueRepository:      op("indexing", "QueueRepository"),
		}

		m = metrics.NewOperationMetrics(
			observationContext.Registerer,
			"codeintel_dependency_repos",
			metrics.WithLabels("op", "scheme", "new"),
		)

		dependencyReposOps = &dependencyReposOperations{
			InsertCloneableDependencyRepo: op("dependencyrepos", "InsertCloneableDependencyRepo"),
		}
	})
	return schedulerOps
}
