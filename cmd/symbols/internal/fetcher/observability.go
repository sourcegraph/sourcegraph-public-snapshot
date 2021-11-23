package fetcher

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/sourcegraph/sourcegraph/internal/metrics"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type operations struct {
	fetching               prometheus.Gauge
	fetchQueueSize         prometheus.Gauge
	fetchRepositoryArchive *observation.Operation
}

func newOperations(observationContext *observation.Context) *operations {
	fetching := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "codeintel_symbols_fetching",
		Help: "The number of fetches currently running.",
	})
	observationContext.Registerer.MustRegister(fetching)

	fetchQueueSize := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "codeintel_symbols_fetch_queue_size",
		Help: "The number of fetch jobs enqueued.",
	})
	observationContext.Registerer.MustRegister(fetchQueueSize)

	operationMetrics := metrics.NewREDMetrics(
		observationContext.Registerer,
		"codeintel_symbols_repository_fetcher",
		metrics.WithLabels("op"),
		metrics.WithCountHelp("Total number of method invocations."),
	)

	op := func(name string) *observation.Operation {
		return observationContext.Operation(observation.Op{
			Name:              fmt.Sprintf("codeintel.symbols.parser.%s", name),
			MetricLabelValues: []string{name},
			Metrics:           operationMetrics,
		})
	}

	return &operations{
		fetching:               fetching,
		fetchQueueSize:         fetchQueueSize,
		fetchRepositoryArchive: op("FetchRepositoryArchive"),
	}
}
