package janitor

import (
	"github.com/prometheus/client_golang/prometheus"

	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type metrics struct {
	numVMsRemoved prometheus.Counter
	numErrors     prometheus.Counter
}

var NewMetrics = newMetrics

func newMetrics(observationCtx *observation.Context) *metrics {
	counter := func(name, help string) prometheus.Counter {
		counter := prometheus.NewCounter(prometheus.CounterOpts{
			Name: name,
			Help: help,
		})

		observationCtx.Registerer.MustRegister(counter)
		return counter
	}

	numVMsRemoved := counter(
		"src_executor_orphaned_vms_removed_total",
		"The number of orphaned virtual machines removed from the host.",
	)
	numErrors := counter(
		"src_executor_janitor_errors_total",
		"The number of errors that occur during the janitor job.",
	)

	return &metrics{
		numVMsRemoved: numVMsRemoved,
		numErrors:     numErrors,
	}
}
