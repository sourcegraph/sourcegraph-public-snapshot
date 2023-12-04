package dependencies

import (
	"github.com/prometheus/client_golang/prometheus"

	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type resetterMetrics struct {
	numIndexResets                  prometheus.Counter
	numIndexResetFailures           prometheus.Counter
	numIndexResetErrors             prometheus.Counter
	numDependencyIndexResets        prometheus.Counter
	numDependencyIndexResetFailures prometheus.Counter
	numDependencyIndexResetErrors   prometheus.Counter
}

func NewResetterMetrics(observationCtx *observation.Context) *resetterMetrics {
	counter := func(name, help string) prometheus.Counter {
		counter := prometheus.NewCounter(prometheus.CounterOpts{
			Name: name,
			Help: help,
		})

		observationCtx.Registerer.MustRegister(counter)
		return counter
	}

	numIndexResets := counter(
		"src_codeintel_background_index_record_resets_total",
		"The number of index records reset.",
	)
	numIndexResetFailures := counter(
		"src_codeintel_background_index_record_reset_failures_total",
		"The number of dependency index reset failures.",
	)
	numIndexResetErrors := counter(
		"src_codeintel_background_index_record_reset_errors_total",
		"The number of errors that occur during index records reset.",
	)

	numDependencyIndexResets := counter(
		"src_codeintel_background_dependency_index_record_resets_total",
		"The number of dependency index records reset.",
	)
	numDependencyIndexResetFailures := counter(
		"src_codeintel_background_dependency_index_record_reset_failures_total",
		"The number of index reset failures.",
	)
	numDependencyIndexResetErrors := counter(
		"src_codeintel_background_dependency_index_record_reset_errors_total",
		"The number of errors that occur during dependency index records reset.",
	)

	return &resetterMetrics{
		numIndexResets:                  numIndexResets,
		numIndexResetFailures:           numIndexResetFailures,
		numIndexResetErrors:             numIndexResetErrors,
		numDependencyIndexResets:        numDependencyIndexResets,
		numDependencyIndexResetFailures: numDependencyIndexResetFailures,
		numDependencyIndexResetErrors:   numDependencyIndexResetErrors,
	}
}
