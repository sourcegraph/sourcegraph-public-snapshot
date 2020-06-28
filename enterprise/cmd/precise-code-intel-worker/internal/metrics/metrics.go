package metrics

import (
	"github.com/sourcegraph/sourcegraph/internal/metrics"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type WorkerMetrics struct {
	Processor *metrics.OperationMetrics // TODO - get rid of this

	//
	//

	ProcessOperation             *observation.Operation
	RepoStateOperation           *observation.Operation
	CorrelateOperation           *observation.Operation
	CanonicalizeOperation        *observation.Operation
	PruneOperation               *observation.Operation
	GroupBundleDataOperation     *observation.Operation
	WriteOperation               *observation.Operation
	UpdateXrepoDatabaseOperation *observation.Operation
	SendDBOperation              *observation.Operation
}

func NewWorkerMetrics(observationContext *observation.Context) WorkerMetrics {
	processor := metrics.NewOperationMetrics(observationContext.Registerer, "upload_queue_processor")

	// metrics := singletonMetrics.Get(func() *metrics.OperationMetrics {
	// 	return metrics.NewOperationMetrics(
	// 		observationContext.Registerer,
	// 		"bundle_reader",
	// 		metrics.WithLabels("op"),
	// 		metrics.WithCountHelp("Total number of results returned"),
	// 	)
	// })

	return WorkerMetrics{
		Processor: processor,
		ProcessOperation: observationContext.Operation(observation.Op{
			Name:         "Processor.Process",
			MetricLabels: []string{"process"},
			// Metrics:      metrics,
		}),
		RepoStateOperation: observationContext.Operation(observation.Op{
			Name:         "Processor.RepoState",
			MetricLabels: []string{"repo_state"},
			// Metrics:      metrics,
		}),
		CorrelateOperation: observationContext.Operation(observation.Op{
			Name:         "Processor.Correlate",
			MetricLabels: []string{"correlate"},
			// Metrics:      metrics,
		}),
		CanonicalizeOperation: observationContext.Operation(observation.Op{
			Name:         "Processor.Canonicalize",
			MetricLabels: []string{"canonicalize"},
			// Metrics:      metrics,
		}),
		PruneOperation: observationContext.Operation(observation.Op{
			Name:         "Processor.Prune",
			MetricLabels: []string{"prune"},
			// Metrics:      metrics,
		}),
		GroupBundleDataOperation: observationContext.Operation(observation.Op{
			Name:         "Processor.GroupBundleData",
			MetricLabels: []string{"groupBundleData"},
			// Metrics:      metrics,
		}),
		WriteOperation: observationContext.Operation(observation.Op{
			Name:         "Processor.Write",
			MetricLabels: []string{"write"},
			// Metrics:      metrics,
		}),
		UpdateXrepoDatabaseOperation: observationContext.Operation(observation.Op{
			Name:         "Processor.UpdateXrepoDatabase",
			MetricLabels: []string{"update_xrepo_database"},
			// Metrics:      metrics,
		}),
		SendDBOperation: observationContext.Operation(observation.Op{
			Name:         "Processor.SendDB",
			MetricLabels: []string{"send_db"},
			// Metrics:      metrics,
		}),
	}
}
