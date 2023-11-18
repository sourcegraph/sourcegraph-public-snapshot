package background

import (
	"github.com/prometheus/client_golang/prometheus"

	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
)

type codeMonitorsMetrics struct {
	workerMetrics workerutil.WorkerObservability
	resets        prometheus.Counter
	resetFailures prometheus.Counter
	errors        prometheus.Counter
}

func newMetricsForTriggerQueries(observationCtx *observation.Context) codeMonitorsMetrics {
	observationCtx = observation.ContextWithLogger(observationCtx.Logger.Scoped("triggers"), observationCtx)

	resetFailures := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "src_codemonitors_query_reset_failures_total",
		Help: "The number of reset failures.",
	})
	observationCtx.Registerer.MustRegister(resetFailures)

	resets := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "src_codemonitors_query_resets_total",
		Help: "The number of records reset.",
	})
	observationCtx.Registerer.MustRegister(resets)

	errors := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "src_codemonitors_query_errors_total",
		Help: "The number of errors that occur during job.",
	})
	observationCtx.Registerer.MustRegister(errors)

	return codeMonitorsMetrics{
		workerMetrics: workerutil.NewMetrics(observationCtx, "code_monitors_trigger_queries"),
		resets:        resets,
		resetFailures: resetFailures,
		errors:        errors,
	}
}

func newActionMetrics(observationCtx *observation.Context) codeMonitorsMetrics {
	observationCtx = observation.ContextWithLogger(observationCtx.Logger.Scoped("actions"), observationCtx)

	resetFailures := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "src_codemonitors_action_reset_failures_total",
		Help: "The number of reset failures.",
	})
	observationCtx.Registerer.MustRegister(resetFailures)

	resets := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "src_codemonitors_action_resets_total",
		Help: "The number of records reset.",
	})
	observationCtx.Registerer.MustRegister(resets)

	errors := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "src_codemonitors_action_errors_total",
		Help: "The number of errors that occur during job.",
	})
	observationCtx.Registerer.MustRegister(errors)

	return codeMonitorsMetrics{
		workerMetrics: workerutil.NewMetrics(observationCtx, "code_monitors_actions"),
		resets:        resets,
		resetFailures: resetFailures,
		errors:        errors,
	}
}
