package batches

import (
	"fmt"

	"github.com/sourcegraph/sourcegraph/internal/metrics"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type operations struct {
	get *observation.Operation
}

func newOperations(observationContext *observation.Context) *operations {
	m := metrics.NewREDMetrics(
		observationContext.Registerer,
		"apiworker_apiclient_batches",
		metrics.WithLabels("op"),
		metrics.WithCountHelp("Total number of method invocations."),
	)

	op := func(name string) *observation.Operation {
		return observationContext.Operation(observation.Op{
			Name:              fmt.Sprintf("apiworker.apiclient.batches.%s", name),
			MetricLabelValues: []string{name},
			Metrics:           m,
		})
	}

	return &operations{
		get: op("Get"),
	}
}
