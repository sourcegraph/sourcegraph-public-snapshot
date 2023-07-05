package store

import (
	"github.com/sourcegraph/sourcegraph/internal/metrics"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type operations struct {
	// getFullSCIPNameByDescriptor *observation.Operation
}

var m = new(metrics.SingletonREDMetrics)

func newOperations(observationCtx *observation.Context) *operations {
	// m := m.Get(func() *metrics.REDMetrics {
	// 	return metrics.NewREDMetrics(
	// 		observationCtx.Registerer,
	// 		"codeintel_context_store",
	// 		metrics.WithLabels("op"),
	// 		metrics.WithCountHelp("Total number of method invocations."),
	// 	)
	// })

	// op := func(name string) *observation.Operation {
	// 	return observationCtx.Operation(observation.Op{
	// 		Name:              fmt.Sprintf("codeintel.context.store.%s", name),
	// 		MetricLabelValues: []string{name},
	// 		Metrics:           m,
	// 	})
	// }

	return &operations{
		// getFullSCIPNameByDescriptor: op("GetFullSCIPNameByDescriptor"),
	}
}
