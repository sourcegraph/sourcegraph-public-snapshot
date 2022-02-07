package janitor

import (
	"github.com/prometheus/client_golang/prometheus"

	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type Metrics struct {
	cacheSizeBytes prometheus.Gauge
	evictions      prometheus.Counter
	errors         prometheus.Counter
}

func NewMetrics(observationContext *observation.Context) *Metrics {
	cacheSizeBytes := prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "src",
		Name:      "codeintel_symbols_store_cache_size_bytes",
		Help:      "The total size of items in the on disk cache.",
	})
	observationContext.Registerer.MustRegister(cacheSizeBytes)

	evictions := prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "src",
		Name:      "codeintel_symbols_store_evictions_total",
		Help:      "The total number of items evicted from the cache.",
	})
	observationContext.Registerer.MustRegister(evictions)

	errors := prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "src",
		Name:      "codeintel_symbols_store_errors_total",
		Help:      "The total number of failures evicting items from the cache.",
	})
	observationContext.Registerer.MustRegister(errors)

	return &Metrics{
		cacheSizeBytes: cacheSizeBytes,
		evictions:      evictions,
		errors:         errors,
	}
}
