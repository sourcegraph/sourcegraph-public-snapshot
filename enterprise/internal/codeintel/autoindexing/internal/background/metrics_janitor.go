package background

import (
	"github.com/prometheus/client_golang/prometheus"

	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type janitorMetrics struct {
	// Data retention metrics
	numErrors              prometheus.Counter
	numIndexRecordsRemoved prometheus.Counter
}

func newJanitorMetrics(observationContext *observation.Context) *janitorMetrics {
	counter := func(name, help string) prometheus.Counter {
		counter := prometheus.NewCounter(prometheus.CounterOpts{
			Name: name,
			Help: help,
		})

		observationContext.Registerer.MustRegister(counter)
		return counter
	}

	numErrors := counter(
		"src_codeintel_autoindexing_background_cleanup_errors_total",
		"The number of errors that occur during a codeintel expiration job.",
	)
	numIndexRecordsRemoved := counter(
		"src_codeintel_background_index_records_removed_total",
		"The number of codeintel index records removed.",
	)

	return &janitorMetrics{
		numErrors:              numErrors,
		numIndexRecordsRemoved: numIndexRecordsRemoved,
	}
}
