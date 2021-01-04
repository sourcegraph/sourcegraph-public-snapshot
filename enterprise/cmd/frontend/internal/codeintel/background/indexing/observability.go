package indexing

import (
	"fmt"

	"github.com/sourcegraph/sourcegraph/internal/metrics"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type operations struct {
	handleIndexabilityUpdater *observation.Operation
	handleIndexScheduler      *observation.Operation
	queueRepository           *observation.Operation
	queueIndex                *observation.Operation
}

var NewOperations = newOperations

func newOperations(observationContext *observation.Context) *operations {
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

	return &operations{
		handleIndexabilityUpdater: op("HandleIndexabilityUpdate"),
		handleIndexScheduler:      op("HandleIndexSchedule"),
		queueRepository:           op("QueueRepository"),
		queueIndex:                op("QueueIndex"),
	}
}
