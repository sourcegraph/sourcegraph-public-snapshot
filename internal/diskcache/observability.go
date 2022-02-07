package diskcache

import (
	"fmt"

	"github.com/sourcegraph/sourcegraph/internal/metrics"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type operations struct {
	cachedFetch     *observation.Operation
	evict           *observation.Operation
	backgroundFetch *observation.Operation
}

func newOperations(observationContext *observation.Context, component string) *operations {
	var m *metrics.REDMetrics
	if observationContext.Registerer != nil {
		m = metrics.NewREDMetrics(
			observationContext.Registerer,
			"diskcache",
			metrics.WithLabels("component", "op"),
		)
	}

	// we dont enable tracing for evict, as it doesnt take a context.Context
	// so it cannot be part of a trace
	evictObservationContext := &observation.Context{
		Logger:       observationContext.Logger,
		Registerer:   observationContext.Registerer,
		HoneyDataset: observationContext.HoneyDataset,
	}

	op := func(name string, ctx *observation.Context) *observation.Operation {
		return ctx.Operation(observation.Op{
			Name:              fmt.Sprintf("diskcache.%s", name),
			Metrics:           m,
			MetricLabelValues: []string{component, name},
		})
	}

	return &operations{
		cachedFetch:     op("Cached Fetch", observationContext),
		evict:           op("Evict", evictObservationContext),
		backgroundFetch: op("Background Fetch", observationContext),
	}
}
