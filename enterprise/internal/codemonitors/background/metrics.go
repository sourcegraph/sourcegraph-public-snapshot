package background

import (
	"github.com/inconshreveable/log15"
	"github.com/opentracing/opentracing-go"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sourcegraph/sourcegraph/internal/metrics"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/trace"
)

type codeMonitorsMetrics struct {
	handleOperation *observation.Operation
	resets          prometheus.Counter
	resetFailures   prometheus.Counter
	errors          prometheus.Counter
}

func newMetricsForTriggerQueries() codeMonitorsMetrics {
	observationContext := &observation.Context{
		Logger:     log15.Root(),
		Tracer:     &trace.Tracer{Tracer: opentracing.GlobalTracer()},
		Registerer: prometheus.DefaultRegisterer,
	}

	metrics := metrics.NewOperationMetrics(
		observationContext.Registerer,
		"code_monitors_trigger_queries",
		metrics.WithLabels("op"),
		metrics.WithCountHelp("Total number of queries run"),
	)

	handleOperation := observationContext.Operation(observation.Op{
		Name:         "Query.Run",
		MetricLabels: []string{"process"},
		Metrics:      metrics,
	})

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
		handleOperation: handleOperation,
		resets:          resets,
		resetFailures:   resetFailures,
		errors:          errors,
	}
}

func newActionMetrics() codeMonitorsMetrics {
	observationContext := &observation.Context{
		Logger:     log15.Root(),
		Tracer:     &trace.Tracer{Tracer: opentracing.GlobalTracer()},
		Registerer: prometheus.DefaultRegisterer,
	}

	metrics := metrics.NewOperationMetrics(
		observationContext.Registerer,
		"code_monitors_actions",
		metrics.WithLabels("op"),
		metrics.WithCountHelp("Total number of actions run"),
	)

	handleOperation := observationContext.Operation(observation.Op{
		Name:         "Action.Run",
		MetricLabels: []string{"process"},
		Metrics:      metrics,
	})

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
		handleOperation: handleOperation,
		resets:          resets,
		resetFailures:   resetFailures,
		errors:          errors,
	}
}
