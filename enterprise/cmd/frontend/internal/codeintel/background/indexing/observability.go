package indexing

import (
	"github.com/prometheus/client_golang/prometheus"

	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type operations struct {
	numErrors prometheus.Counter
}

var NewOperations = newOperations

func newOperations(observationContext *observation.Context) *operations {
	counter := func(name, help string) prometheus.Counter {
		counter := prometheus.NewCounter(prometheus.CounterOpts{
			Name: name,
			Help: help,
		})

		observationContext.Registerer.MustRegister(counter)
		return counter
	}

	numErrors := counter(
		"src_codeintel_background_index_scheduler_errors_total",
		"The number of errors that occur during a codeintel background index scheduling job.",
	)

	return &operations{
		numErrors: numErrors,
	}
}
