package definitions

import (
	"github.com/sourcegraph/sourcegraph/monitoring/definitions/shared"
	"github.com/sourcegraph/sourcegraph/monitoring/monitoring"
)

func ExecutorQueue() *monitoring.Container {
	const (
		containerName      = "executor-queue"
		queueContainerName = "(executor|sourcegraph-code-intel-indexers|executor-batches|executor-queue)"
	)

	return &monitoring.Container{
		Name:        "executor-queue",
		Title:       "Executor Queue",
		Description: "Coordinates the executor work queues.",
		Groups: []monitoring.Group{
			// src_executor_total
			// src_executor_processor_total
			shared.Queue.NewGroup(queueContainerName, monitoring.ObservableOwnerCodeIntel, shared.QueueSizeGroupOptions{
				GroupConstructorOptions: shared.GroupConstructorOptions{
					Namespace:       "executor",
					DescriptionRoot: "Executor jobs",

					ObservableConstructorOptions: shared.ObservableConstructorOptions{
						MetricNameRoot:        "executor",
						MetricDescriptionRoot: "unprocessed executor job",
						By:                    []string{"queue"},
					},
				},

				QueueSize: shared.NoAlertsOption("none"),
				QueueGrowthRate: shared.NoAlertsOption(`
					This value compares the rate of enqueues against the rate of finished jobs for the selected queue.

						- A value < than 1 indicates that process rate > enqueue rate
						- A value = than 1 indicates that process rate = enqueue rate
						- A value > than 1 indicates that process rate < enqueue rate
				`),
			}),

			// src_workerutil_dbworker_store_codeintel_index_total
			// src_workerutil_dbworker_store_codeintel_index_duration_seconds_bucket
			// src_workerutil_dbworker_store_codeintel_index_errors_total
			shared.Observation.NewGroup(containerName, monitoring.ObservableOwnerCodeIntel, shared.ObservationGroupOptions{
				GroupConstructorOptions: shared.GroupConstructorOptions{
					Namespace:       "workerutil",
					DescriptionRoot: "dbworker/store stats (db=frontend, table=lsif_indexes)",
					Hidden:          true,

					ObservableConstructorOptions: shared.ObservableConstructorOptions{
						MetricNameRoot:        "workerutil_dbworker_store_codeintel_index",
						MetricDescriptionRoot: "store",
					},
				},

				Total:    shared.NoAlertsOption("none"),
				Duration: shared.NoAlertsOption("none"),
				Errors:   shared.NoAlertsOption("none"),
			}),

			// Resource monitoring
			shared.NewFrontendInternalAPIErrorResponseMonitoringGroup(containerName, monitoring.ObservableOwnerCodeIntel, nil),
			shared.NewDatabaseConnectionsMonitoringGroup(containerName),
			shared.NewContainerMonitoringGroup(containerName, monitoring.ObservableOwnerCodeIntel, nil),
			shared.NewProvisioningIndicatorsGroup(containerName, monitoring.ObservableOwnerCodeIntel, nil),
			shared.NewGolangMonitoringGroup(containerName, monitoring.ObservableOwnerCodeIntel, nil),
			shared.NewKubernetesMonitoringGroup(containerName, monitoring.ObservableOwnerCodeIntel, nil),
		},
	}
}
