package background

import (
	"github.com/prometheus/client_golang/prometheus"

	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type mergerMetrics struct {
	numRepositoriesUpdated prometheus.Counter
	numInputRowsProcessed  prometheus.Counter
}

func newMergerMetrics(observationContext *observation.Context) *mergerMetrics {
	counter := func(name, help string) prometheus.Counter {
		counter := prometheus.NewCounter(prometheus.CounterOpts{
			Name: name,
			Help: help,
		})

		observationContext.Registerer.MustRegister(counter)
		return counter
	}

	numRepositoriesUpdated := counter(
		"src_codeintel_ranking_repositories_updated_total",
		"The number of updates to document scores of any repository.",
	)
	numInputRowsProcessed := counter(
		"src_codeintel_ranking_input_rows_processed_total",
		"The number of input row records merged into document scores for a single repo.",
	)

	return &mergerMetrics{
		numRepositoriesUpdated: numRepositoriesUpdated,
		numInputRowsProcessed:  numInputRowsProcessed,
	}
}
