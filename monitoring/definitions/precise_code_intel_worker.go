package definitions

import (
	"github.com/sourcegraph/sourcegraph/monitoring/definitions/shared"
	"github.com/sourcegraph/sourcegraph/monitoring/monitoring"
)

func PreciseCodeIntelWorker() *monitoring.Container {
	const containerName = "precise-code-intel-worker"

	return &monitoring.Container{
		Name:        "precise-code-intel-worker",
		Title:       "Precise Code Intel Worker",
		Description: "Handles conversion of uploaded precise code intelligence bundles.",
		Groups: []monitoring.Group{
			// src_codeintel_upload_total
			// src_codeintel_upload_processor_total
			shared.Queue.NewGroup(containerName, monitoring.ObservableOwnerCodeIntel, shared.QueueSizeGroupOptions{
				GroupConstructorOptions: shared.GroupConstructorOptions{
					Namespace:       "codeintel",
					DescriptionRoot: "LSIF uploads",

					ObservableConstructorOptions: shared.ObservableConstructorOptions{
						MetricNameRoot:        "codeintel_upload",
						MetricDescriptionRoot: "unprocessed upload record",
					},
				},

				QueueSize: shared.NoAlertsOption("none"),
				QueueGrowthRate: shared.NoAlertsOption(`
					This value compares the rate of enqueues against the rate of finished jobs.

						- A value < than 1 indicates that process rate > enqueue rate
						- A value = than 1 indicates that process rate = enqueue rate
						- A value > than 1 indicates that process rate < enqueue rate
				`),
			}),

			// src_codeintel_upload_processor_total
			// src_codeintel_upload_processor_duration_seconds_bucket
			// src_codeintel_upload_processor_errors_total
			// src_codeintel_upload_processor_handlers
			shared.Workerutil.NewGroup(containerName, monitoring.ObservableOwnerCodeIntel, shared.WorkerutilGroupOptions{
				GroupConstructorOptions: shared.GroupConstructorOptions{
					Namespace:       "codeintel",
					DescriptionRoot: "LSIF uploads",

					ObservableConstructorOptions: shared.ObservableConstructorOptions{
						MetricNameRoot:        "codeintel_upload",
						MetricDescriptionRoot: "handler",
					},
				},

				Total:    shared.NoAlertsOption("none"),
				Duration: shared.NoAlertsOption("none"),
				Errors:   shared.NoAlertsOption("none"),
				Handlers: shared.NoAlertsOption("none"),
			}),

			// src_codeintel_dbstore_total
			// src_codeintel_dbstore_duration_seconds_bucket
			// src_codeintel_dbstore_errors_total
			shared.Observation.NewGroup(containerName, monitoring.ObservableOwnerCodeIntel, shared.ObservationGroupOptions{
				GroupConstructorOptions: shared.GroupConstructorOptions{
					Namespace:       "codeintel",
					DescriptionRoot: "dbstore stats (db=frontend)",
					Hidden:          true,

					ObservableConstructorOptions: shared.ObservableConstructorOptions{
						MetricNameRoot:        "codeintel_dbstore",
						MetricDescriptionRoot: "store",
					},
				},

				Total:    shared.NoAlertsOption("none"),
				Duration: shared.NoAlertsOption("none"),
				Errors:   shared.NoAlertsOption("none"),
			}),

			// src_codeintel_lsifstore_total
			// src_codeintel_lsifstore_duration_seconds_bucket
			// src_codeintel_lsifstore_errors_total
			shared.Observation.NewGroup(containerName, monitoring.ObservableOwnerCodeIntel, shared.ObservationGroupOptions{
				GroupConstructorOptions: shared.GroupConstructorOptions{
					Namespace:       "codeintel",
					DescriptionRoot: "lsifstore stats (db=codeintel-db)",
					Hidden:          true,

					ObservableConstructorOptions: shared.ObservableConstructorOptions{
						MetricNameRoot:        "codeintel_lsifstore",
						MetricDescriptionRoot: "store",
					},
				},

				Total:    shared.NoAlertsOption("none"),
				Duration: shared.NoAlertsOption("none"),
				Errors:   shared.NoAlertsOption("none"),
			}),

			// src_workerutil_dbworker_store_codeintel_upload_total
			// src_workerutil_dbworker_store_codeintel_upload_duration_seconds_bucket
			// src_workerutil_dbworker_store_codeintel_upload_errors_total
			shared.Observation.NewGroup(containerName, monitoring.ObservableOwnerCodeIntel, shared.ObservationGroupOptions{
				GroupConstructorOptions: shared.GroupConstructorOptions{
					Namespace:       "workerutil",
					DescriptionRoot: "dbworker/store stats (db=frontend, table=lsif_uploads)",
					Hidden:          true,

					ObservableConstructorOptions: shared.ObservableConstructorOptions{
						MetricNameRoot:        "workerutil_dbworker_store_codeintel_upload",
						MetricDescriptionRoot: "store",
					},
				},

				Total:    shared.NoAlertsOption("none"),
				Duration: shared.NoAlertsOption("none"),
				Errors:   shared.NoAlertsOption("none"),
			}),

			// src_codeintel_gitserver_total
			// src_codeintel_gitserver_duration_seconds_bucket
			// src_codeintel_gitserver_errors_total
			shared.Observation.NewGroup(containerName, monitoring.ObservableOwnerCodeIntel, shared.ObservationGroupOptions{
				GroupConstructorOptions: shared.GroupConstructorOptions{
					Namespace:       "codeintel",
					DescriptionRoot: "gitserver client",
					Hidden:          true,

					ObservableConstructorOptions: shared.ObservableConstructorOptions{
						MetricNameRoot:        "codeintel_gitserver",
						MetricDescriptionRoot: "client",
					},
				},

				Total:    shared.NoAlertsOption("none"),
				Duration: shared.NoAlertsOption("none"),
				Errors:   shared.NoAlertsOption("none"),
			}),

			// src_codeintel_uploadstore_total
			// src_codeintel_uploadstore_duration_seconds_bucket
			// src_codeintel_uploadstore_errors_total
			shared.Observation.NewGroup(containerName, monitoring.ObservableOwnerCodeIntel, shared.ObservationGroupOptions{
				GroupConstructorOptions: shared.GroupConstructorOptions{
					Namespace:       "codeintel",
					DescriptionRoot: "uploadstore stats (queries GCS/S3/MinIO)",
					Hidden:          true,

					ObservableConstructorOptions: shared.ObservableConstructorOptions{
						MetricNameRoot:        "codeintel_uploadstore",
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
