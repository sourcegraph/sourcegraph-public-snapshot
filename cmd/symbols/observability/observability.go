package observability

import (
	"fmt"

	"github.com/sourcegraph/sourcegraph/internal/metrics"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type Operations struct {
	Search *observation.Operation
}

func NewOperations(observationCtx *observation.Context) *Operations {
	redMetrics := metrics.NewREDMetrics(
		observationCtx.Registerer,
		"codeintel_symbols_api",
		metrics.WithLabels("op", "parseAmount"),
		metrics.WithCountHelp("Total number of method invocations."),
		metrics.WithDurationBuckets([]float64{1, 2, 5, 10, 30, 60}),
	)

	op := func(name string) *observation.Operation {
		return observationCtx.Operation(observation.Op{
			Name:              fmt.Sprintf("codeintel.symbols.api.%s", name),
			MetricLabelValues: []string{name},
			Metrics:           redMetrics,
		})
	}

	return &Operations{
		Search: op("Search"),
	}
}
