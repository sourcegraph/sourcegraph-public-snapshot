package background

import (
	"github.com/inconshreveable/log15"
	"github.com/opentracing/opentracing-go"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
)

type codeMonitorsMetrics struct {
	workerMetrics workerutil.WorkerMetrics
	resets        prometheus.Counter
	resetFailures prometheus.Counter
	errors        prometheus.Counter
}

func newMetricsForTriggerQueries() codeMonitorsMetrics {
	observationContext := &observation.Context{
		Logger:     log15.Root(),
		Tracer:     &trace.Tracer{Tracer: opentracing.GlobalTracer()},
		Registerer: prometheus.DefaultRegisterer,
	}

	resetFailures := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "src_codemonitors_query_reset_failures_total",
		Help: "The number of reset failures.",
	})
	observationContext.Registerer.MustRegister(resetFailures)

	resets := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "src_codemonitors_query_resets_total",
		Help: "The number of records reset.",
	})
	observationContext.Registerer.MustRegister(resets)

	errors := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "src_codemonitors_query_errors_total",
		Help: "The number of errors that occur during job.",
	})
	observationContext.Registerer.MustRegister(errors)

	return codeMonitorsMetrics{
		workerMetrics: workerutil.NewMetrics(observationContext, "code_monitors_trigger_queries", nil),
		resets:        resets,
		resetFailures: resetFailures,
		errors:        errors,
	}
}

func newActionMetrics() codeMonitorsMetrics {
	observationContext := &observation.Context{
		Logger:     log15.Root(),
		Tracer:     &trace.Tracer{Tracer: opentracing.GlobalTracer()},
		Registerer: prometheus.DefaultRegisterer,
	}

	resetFailures := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "src_codemonitors_action_reset_failures_total",
		Help: "The number of reset failures.",
	})
	observationContext.Registerer.MustRegister(resetFailures)

	resets := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "src_codemonitors_action_resets_total",
		Help: "The number of records reset.",
	})
	observationContext.Registerer.MustRegister(resets)

	errors := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "src_codemonitors_action_errors_total",
		Help: "The number of errors that occur during job.",
	})
	observationContext.Registerer.MustRegister(errors)

	return codeMonitorsMetrics{
		workerMetrics: workerutil.NewMetrics(observationContext, "code_monitors_actions", nil),
		resets:        resets,
		resetFailures: resetFailures,
		errors:        errors,
	}
}
