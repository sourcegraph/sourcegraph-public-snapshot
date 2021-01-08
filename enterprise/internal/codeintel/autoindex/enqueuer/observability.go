package enqueuer

import (
	"fmt"
	"sync"

	"github.com/sourcegraph/sourcegraph/internal/metrics"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

// must be a singletone
type operations struct {
	queueIndex *observation.Operation
}

var (
	singletonOperations *operations
	once                sync.Once
)

func newOperations(observationContext *observation.Context) *operations {
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

		singletonOperations = &operations{
			queueIndex: op("QueueIndex"),
		}
	})
	return singletonOperations
}
