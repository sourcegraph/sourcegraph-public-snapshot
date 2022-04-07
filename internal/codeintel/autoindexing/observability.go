package autoindexing

import (
	"fmt"

	"github.com/sourcegraph/sourcegraph/internal/metrics"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type operations struct {
	list     *observation.Operation
	get      *observation.Operation
	getBatch *observation.Operation
	delete   *observation.Operation
	enqueue  *observation.Operation
	infer    *observation.Operation
}

func newOperations(observationContext *observation.Context) *operations {
	metrics := metrics.NewREDMetrics(
		observationContext.Registerer,
		"codeintel_autoindexing",
		metrics.WithLabels("op"),
		metrics.WithCountHelp("Total number of method invocations."),
	)

	op := func(name string) *observation.Operation {
		return observationContext.Operation(observation.Op{
			Name:              fmt.Sprintf("codeintel.autoindexing.%s", name),
			MetricLabelValues: []string{name},
			Metrics:           metrics,
		})
	}

	return &operations{
		list:     op("List"),
		get:      op("Get"),
		getBatch: op("GetBatch"),
		delete:   op("Delete"),
		enqueue:  op("Enqueue"),
		infer:    op("Infer"),
	}
}
