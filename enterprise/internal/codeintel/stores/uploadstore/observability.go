package uploadstore

import (
	"github.com/sourcegraph/sourcegraph/internal/metrics"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type operations struct {
	get     *observation.Operation
	upload  *observation.Operation
	compose *observation.Operation
	delete  *observation.Operation
}

func makeOperations(observationContext *observation.Context) *operations {
	metrics := metrics.NewOperationMetrics(
		observationContext.Registerer,
		"codeintel_uploadstore",
		metrics.WithLabels("op"),
		metrics.WithCountHelp("Total number of method invocations."),
	)

	op := func(name string) *observation.Operation {
		return observationContext.Operation(observation.Op{
			Name:         "codeintel.uploadstore.%s",
			MetricLabels: []string{name},
			Metrics:      metrics,
		})
	}

	return &operations{
		get:     op("Get"),
		upload:  op("Upload"),
		compose: op("Compose"),
		delete:  op("Delete"),
	}
}
