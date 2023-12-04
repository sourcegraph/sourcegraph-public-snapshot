package internal

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/sourcegraph/sourcegraph/internal/metrics"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type operations struct {
	batchLogSemaphoreWait prometheus.Histogram
	batchLog              *observation.Operation
	batchLogSingle        *observation.Operation
}

func newOperations(observationCtx *observation.Context) *operations {
	batchLogSemaphoreWait := prometheus.NewHistogram(prometheus.HistogramOpts{
		Namespace: "src",
		Name:      "batch_log_semaphore_wait_duration_seconds",
		Help:      "Time in seconds spent waiting for the global batch log semaphore",
		Buckets:   prometheus.DefBuckets,
	})
	observationCtx.Registerer.MustRegister(batchLogSemaphoreWait)

	redMetrics := metrics.NewREDMetrics(
		observationCtx.Registerer,
		"gitserver_api",
		metrics.WithLabels("op"),
		metrics.WithCountHelp("Total number of method invocations."),
	)

	op := func(name string) *observation.Operation {
		return observationCtx.Operation(observation.Op{
			Name:              fmt.Sprintf("gitserver.api.%s", name),
			MetricLabelValues: []string{name},
			Metrics:           redMetrics,
		})
	}

	// suboperations do not have their own metrics but do have their own spans.
	// This allows us to more granularly track the latency for parts of a
	// request without noising up Prometheus.
	subOp := func(name string) *observation.Operation {
		return observationCtx.Operation(observation.Op{
			Name: fmt.Sprintf("gitserver.api.%s", name),
		})
	}

	return &operations{
		batchLogSemaphoreWait: batchLogSemaphoreWait,
		batchLog:              op("BatchLog"),
		batchLogSingle:        subOp("batchLogSingle"),
	}
}
