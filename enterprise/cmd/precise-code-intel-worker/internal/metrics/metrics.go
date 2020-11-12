package metrics

import (
	"github.com/sourcegraph/sourcegraph/internal/metrics"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type WorkerMetrics struct {
	ProcessOperation *observation.Operation
}

func NewWorkerMetrics(observationContext *observation.Context) WorkerMetrics {
	metrics := metrics.NewOperationMetrics(
		observationContext.Registerer,
		"upload_queue_processor",
		metrics.WithLabels("op"),
		metrics.WithCountHelp("Total number of records processed"),
	)

	return WorkerMetrics{
		ProcessOperation: observationContext.Operation(observation.Op{
			Name:         "Processor.Process",
			MetricLabels: []string{"process"},
			Metrics:      metrics,
		}),
	}
}
