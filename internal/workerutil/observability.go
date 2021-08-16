package workerutil

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/sourcegraph/sourcegraph/internal/metrics"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type WorkerMetrics struct {
	operations *operations
	numJobs    prometheus.Gauge
}

type operations struct {
	handle *observation.Operation
}

// NewMetrics creates and registers the following metrics for a generic worker instance.
//
//   - {prefix}_duration_seconds_bucket: handler operation latency histogram
//   - {prefix}_total: number of handler operations
//   - {prefix}_error_total: number of handler operations resulting in an error
//   - {prefix}_handlers: the number of active handler routines
//
// The given labels are emitted on each metric.
func NewMetrics(observationContext *observation.Context, prefix string, labels map[string]string) WorkerMetrics {
	keys := make([]string, 0, len(labels))
	values := make([]string, 0, len(labels))
	for key, value := range labels {
		keys = append(keys, key)
		values = append(values, value)
	}

	gauge := func(name, help string) prometheus.Gauge {
		gaugeVec := prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: fmt.Sprintf("src_%s_%s", prefix, name),
			Help: help,
		}, keys)

		observationContext.Registerer.MustRegister(gaugeVec)
		return gaugeVec.WithLabelValues(values...)
	}

	numJobs := gauge(
		"handlers",
		"The number of active handlers.",
	)

	return WorkerMetrics{
		operations: newOperations(observationContext, prefix, keys, values),
		numJobs:    numJobs,
	}
}

func newOperations(observationContext *observation.Context, prefix string, keys, values []string) *operations {
	metrics := metrics.NewOperationMetrics(
		observationContext.Registerer,
		prefix,
		metrics.WithLabels(append(keys, "op")...),
		metrics.WithCountHelp("Total number of method invocations."),
	)

	op := func(name string) *observation.Operation {
		return observationContext.Operation(observation.Op{
			Name:              fmt.Sprintf("worker.%s", name),
			MetricLabelValues: append(append([]string{}, values...), name),
			Metrics:           metrics,
		})
	}

	return &operations{
		handle: op("Handle"),
	}
}
