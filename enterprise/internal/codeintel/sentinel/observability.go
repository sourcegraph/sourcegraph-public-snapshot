package sentinel

import (
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type operations struct {
}

// var m = new(metrics.SingletonREDMetrics)

func newOperations(observationCtx *observation.Context) *operations {
	// redMetrics := m.Get(func() *metrics.REDMetrics {
	// 	return metrics.NewREDMetrics(
	// 		observationCtx.Registerer,
	// 		"codeintel_sentinel",
	// 		metrics.WithLabels("op"),
	// 		metrics.WithCountHelp("Total number of method invocations."),
	// 	)
	// })

	// op := func(name string) *observation.Operation {
	// 	return observationCtx.Operation(observation.Op{
	// 		Name:              fmt.Sprintf("codeintel.sentinel.%s", name),
	// 		MetricLabelValues: []string{name},
	// 		Metrics:           redMetrics,
	// 	})
	// }

	return &operations{}
}
