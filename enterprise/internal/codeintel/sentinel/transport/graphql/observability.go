package graphql

import (
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type operations struct {
}

func newOperations(observationCtx *observation.Context) *operations {
	// m := metrics.NewREDMetrics(
	// 	observationCtx.Registerer,
	// 	"codeintel_sentinel_transport_graphql",
	// 	metrics.WithLabels("op"),
	// 	metrics.WithCountHelp("Total number of method invocations."),
	// )

	// op := func(name string) *observation.Operation {
	// 	return observationCtx.Operation(observation.Op{
	// 		Name:              fmt.Sprintf("codeintel.sentinel.transport.graphql.%s", name),
	// 		MetricLabelValues: []string{name},
	// 		Metrics:           m,
	// 	})
	// }

	return &operations{}
}
