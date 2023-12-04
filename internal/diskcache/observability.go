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

func newOperations(observationCtx *observation.Context, component string) *operations {
	var m *metrics.REDMetrics
	if observationCtx.Registerer != nil {
		m = metrics.NewREDMetrics(
			observationCtx.Registerer,
			"diskcache",
			metrics.WithLabels("component", "op"),
		)
	}

	// we dont enable tracing for evict, as it doesnt take a context.Context
	// so it cannot be part of a trace
	evictobservationCtx := &observation.Context{
		Logger:       observationCtx.Logger,
		Registerer:   observationCtx.Registerer,
		HoneyDataset: observationCtx.HoneyDataset,
	}

	op := func(name string, ctx *observation.Context) *observation.Operation {
		return ctx.Operation(observation.Op{
			Name:              fmt.Sprintf("diskcache.%s", name),
			Metrics:           m,
			MetricLabelValues: []string{component, name},
		})
	}

	return &operations{
		cachedFetch:     op("Cached Fetch", observationCtx),
		evict:           op("Evict", evictobservationCtx),
		backgroundFetch: op("Background Fetch", observationCtx),
	}
}
