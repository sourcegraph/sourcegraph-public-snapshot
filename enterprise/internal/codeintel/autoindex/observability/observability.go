package observability

import (
	"fmt"
	"sync"

	"github.com/sourcegraph/sourcegraph/internal/metrics"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type Operations struct {
	QueueIndex                *observation.Operation
	HandleIndexabilityUpdater *observation.Operation
	HandleIndexScheduler      *observation.Operation
	QueueRepository           *observation.Operation
}

var (
	singletonOperations *Operations
	once                sync.Once
)

func NewOperations(observationContext *observation.Context) *Operations {
	once.Do(func() {
		metrics := metrics.NewOperationMetrics(
			observationContext.Registerer,
			"codeintel_indexing",
			metrics.WithLabels("op"),
			metrics.WithCountHelp("Total number of method invocations."),
		)

		op := func(name string) *observation.Operation {
			return observationContext.Operation(observation.Op{
				Name:         fmt.Sprintf("codeintel.indexing.%s", name),
				MetricLabels: []string{name},
				Metrics:      metrics,
			})
		}

		singletonOperations = &Operations{
			HandleIndexabilityUpdater: op("HandleIndexabilityUpdate"),
			HandleIndexScheduler:      op("HandleIndexSchedule"),
			QueueRepository:           op("QueueRepository"),
			QueueIndex:                op("QueueIndex"),
		}
	})
	return singletonOperations
}
