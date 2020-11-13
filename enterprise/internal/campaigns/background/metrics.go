package background

import (
	"github.com/inconshreveable/log15"
	"github.com/opentracing/opentracing-go"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sourcegraph/sourcegraph/internal/metrics"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/trace"
)

type campaignsMetrics struct {
	handleOperation *observation.Operation
	resets          prometheus.Counter
	resetFailures   prometheus.Counter
	errors          prometheus.Counter
}

func newMetrics() campaignsMetrics {
	observationContext := &observation.Context{
		Logger:     log15.Root(),
		Tracer:     &trace.Tracer{Tracer: opentracing.GlobalTracer()},
		Registerer: prometheus.DefaultRegisterer,
	}

	metrics := metrics.NewOperationMetrics(
		observationContext.Registerer,
		"campaigns_reconciler",
		metrics.WithLabels("op"),
		metrics.WithCountHelp("Total number of changesets reconciled"),
	)

	handleOperation := observationContext.Operation(observation.Op{
		Name:         "Reconciler.Process",
		MetricLabels: []string{"process"},
		Metrics:      metrics,
	})

	resetFailures := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "src_campaigns_background_reconciler_reset_failures_total",
		Help: "The number of reconciler reset failures.",
	})
	observationContext.Registerer.MustRegister(resetFailures)

	resets := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "src_campaigns_background_reconciler_resets_total",
		Help: "The number of reconciler records reset.",
	})
	observationContext.Registerer.MustRegister(resets)

	errors := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "src_campaigns_background_errors_total",
		Help: "The number of errors that occur during a campaigns background job.",
	})
	observationContext.Registerer.MustRegister(errors)

	return campaignsMetrics{
		handleOperation: handleOperation,
		resets:          resets,
		resetFailures:   resetFailures,
		errors:          errors,
	}
}
