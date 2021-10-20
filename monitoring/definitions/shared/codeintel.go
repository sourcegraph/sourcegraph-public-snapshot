package shared

import (
	"fmt"

	"github.com/sourcegraph/sourcegraph/monitoring/monitoring"
)

// CodeIntelligence exports available shared observable and group constructors related to
// the code intelligence team. Some of these panels are useful from multiple container
// contexts, so we maintain this struct as a place of authority over team alert definitions.
var CodeIntelligence codeIntelligence

// codeIntelligence provides `CodeIntelligence` implementations.
type codeIntelligence struct{}

// src_codeintel_resolvers_total
// src_codeintel_resolvers_duration_seconds_bucket
// src_codeintel_resolvers_errors_total
func (codeIntelligence) NewResolversGroup(containerName string) monitoring.Group {
	return Observation.NewGroup(containerName, monitoring.ObservableOwnerCodeIntel, ObservationGroupOptions{
		GroupConstructorOptions: GroupConstructorOptions{
			Namespace:       "codeintel",
			DescriptionRoot: "Precise code intelligence usage at a glance",
			Hidden:          true,

			ObservableConstructorOptions: ObservableConstructorOptions{
				MetricNameRoot:        "codeintel_resolvers",
				MetricDescriptionRoot: "graphql",
				By:                    []string{"op"},
			},
		},

		SharedObservationGroupOptions: SharedObservationGroupOptions{
			Total:     NoAlertsOption("none"),
			Duration:  NoAlertsOption("none"),
			Errors:    NoAlertsOption("none"),
			ErrorRate: NoAlertsOption("none"),
		},
		Aggregate: &SharedObservationGroupOptions{
			Total:     NoAlertsOption("none"),
			Duration:  NoAlertsOption("none"),
			Errors:    NoAlertsOption("none"),
			ErrorRate: NoAlertsOption("none"),
		},
	})
}

// src_codeintel_upload_total
// src_codeintel_upload_processor_total
func (codeIntelligence) NewUploadQueueGroup(containerName string) monitoring.Group {
	return Queue.NewGroup(containerName, monitoring.ObservableOwnerCodeIntel, QueueSizeGroupOptions{
		GroupConstructorOptions: GroupConstructorOptions{
			Namespace:       "codeintel",
			DescriptionRoot: "LSIF uploads",

			ObservableConstructorOptions: ObservableConstructorOptions{
				MetricNameRoot:        "codeintel_upload",
				MetricDescriptionRoot: "unprocessed upload record",
			},
		},

		QueueSize: NoAlertsOption("none"),
		QueueGrowthRate: NoAlertsOption(`
			This value compares the rate of enqueues against the rate of finished jobs.

				- A value < than 1 indicates that process rate > enqueue rate
				- A value = than 1 indicates that process rate = enqueue rate
				- A value > than 1 indicates that process rate < enqueue rate
		`),
	})
}

// src_codeintel_upload_processor_total
// src_codeintel_upload_processor_duration_seconds_bucket
// src_codeintel_upload_processor_errors_total
// src_codeintel_upload_processor_handlers
func (codeIntelligence) NewUploadProcessorGroup(containerName string) monitoring.Group {
	return Workerutil.NewGroup(containerName, monitoring.ObservableOwnerCodeIntel, WorkerutilGroupOptions{
		GroupConstructorOptions: GroupConstructorOptions{
			Namespace:       "codeintel",
			DescriptionRoot: "LSIF uploads",

			ObservableConstructorOptions: ObservableConstructorOptions{
				MetricNameRoot:        "codeintel_upload",
				MetricDescriptionRoot: "handler",
			},
		},

		SharedObservationGroupOptions: SharedObservationGroupOptions{
			Total:     NoAlertsOption("none"),
			Duration:  NoAlertsOption("none"),
			Errors:    NoAlertsOption("none"),
			ErrorRate: NoAlertsOption("none"),
		},
		Handlers: NoAlertsOption("none"),
	})
}

// src_codeintel_commit_graph_total
// src_codeintel_commit_graph_processor_total
func (codeIntelligence) NewCommitGraphQueueGroup(containerName string) monitoring.Group {
	return Queue.NewGroup(containerName, monitoring.ObservableOwnerCodeIntel, QueueSizeGroupOptions{
		GroupConstructorOptions: GroupConstructorOptions{
			Namespace:       "codeintel",
			DescriptionRoot: "Repository with stale commit graph",
			Hidden:          true,

			ObservableConstructorOptions: ObservableConstructorOptions{
				MetricNameRoot:        "codeintel_commit_graph",
				MetricDescriptionRoot: "repository",
			},
		},

		QueueSize: NoAlertsOption("none"),
		QueueGrowthRate: NoAlertsOption(`
			This value compares the rate of enqueues against the rate of finished jobs.

				- A value < than 1 indicates that process rate > enqueue rate
				- A value = than 1 indicates that process rate = enqueue rate
				- A value > than 1 indicates that process rate < enqueue rate
		`),
	})
}

// src_codeintel_commit_graph_processor_total
// src_codeintel_commit_graph_processor_duration_seconds_bucket
// src_codeintel_commit_graph_processor_errors_total
func (codeIntelligence) NewCommitGraphProcessorGroup(containerName string) monitoring.Group {
	return Observation.NewGroup(containerName, monitoring.ObservableOwnerCodeIntel, ObservationGroupOptions{
		GroupConstructorOptions: GroupConstructorOptions{
			Namespace:       "codeintel",
			DescriptionRoot: "Repository commit graph updates",
			Hidden:          true,

			ObservableConstructorOptions: ObservableConstructorOptions{
				MetricNameRoot:        "codeintel_commit_graph_processor",
				MetricDescriptionRoot: "update",
			},
		},

		SharedObservationGroupOptions: SharedObservationGroupOptions{
			Total:     NoAlertsOption("none"),
			Duration:  NoAlertsOption("none"),
			Errors:    NoAlertsOption("none"),
			ErrorRate: NoAlertsOption("none"),
		},
	})
}

// src_codeintel_index_scheduler_total
// src_codeintel_index_scheduler_duration_seconds_bucket
// src_codeintel_index_scheduler_errors_total
func (codeIntelligence) NewIndexSchedulerGroup(containerName string) monitoring.Group {
	return Observation.NewGroup(containerName, monitoring.ObservableOwnerCodeIntel, ObservationGroupOptions{
		GroupConstructorOptions: GroupConstructorOptions{
			Namespace:       "codeintel",
			DescriptionRoot: "Auto-index scheduler",
			Hidden:          true,

			ObservableConstructorOptions: ObservableConstructorOptions{
				MetricNameRoot:        "codeintel_index_scheduler",
				MetricDescriptionRoot: "scheduler",
				By:                    []string{"op"},
			},
		},

		SharedObservationGroupOptions: SharedObservationGroupOptions{
			Total:     NoAlertsOption("none"),
			Duration:  NoAlertsOption("none"),
			Errors:    NoAlertsOption("none"),
			ErrorRate: NoAlertsOption("none"),
		},
		Aggregate: &SharedObservationGroupOptions{
			Total:     NoAlertsOption("none"),
			Duration:  NoAlertsOption("none"),
			Errors:    NoAlertsOption("none"),
			ErrorRate: NoAlertsOption("none"),
		},
	})
}

// src_codeintel_dependency_index_total
// src_codeintel_dependency_index_processor_total
func (codeIntelligence) NewDependencyIndexQueueGroup(containerName string) monitoring.Group {
	return Queue.NewGroup(containerName, monitoring.ObservableOwnerCodeIntel, QueueSizeGroupOptions{
		GroupConstructorOptions: GroupConstructorOptions{
			Namespace:       "codeintel",
			DescriptionRoot: "Dependency index job",
			Hidden:          true,

			ObservableConstructorOptions: ObservableConstructorOptions{
				MetricNameRoot:        "codeintel_dependency_index",
				MetricDescriptionRoot: "dependency index job",
			},
		},

		QueueSize: NoAlertsOption("none"),
		QueueGrowthRate: NoAlertsOption(`
			This value compares the rate of enqueues against the rate of finished jobs.

				- A value < than 1 indicates that process rate > enqueue rate
				- A value = than 1 indicates that process rate = enqueue rate
				- A value > than 1 indicates that process rate < enqueue rate
		`),
	})
}

// src_codeintel_dependency_index_processor_total
// src_codeintel_dependency_index_processor_duration_seconds_bucket
// src_codeintel_dependency_index_processor_errors_total
// src_codeintel_dependency_index_processor_handlers
func (codeIntelligence) NewDependencyIndexProcessorGroup(containerName string) monitoring.Group {
	return Workerutil.NewGroup(containerName, monitoring.ObservableOwnerCodeIntel, WorkerutilGroupOptions{
		GroupConstructorOptions: GroupConstructorOptions{
			Namespace:       "codeintel",
			DescriptionRoot: "Dependency index jobs",
			Hidden:          true,

			ObservableConstructorOptions: ObservableConstructorOptions{
				MetricNameRoot:        "codeintel_dependency_index",
				MetricDescriptionRoot: "handler",
			},
		},

		SharedObservationGroupOptions: SharedObservationGroupOptions{
			Total:     NoAlertsOption("none"),
			Duration:  NoAlertsOption("none"),
			Errors:    NoAlertsOption("none"),
			ErrorRate: NoAlertsOption("none"),
		},
		Handlers: NoAlertsOption("none"),
	})
}

// src_executor_total
// src_executor_processor_total
func (codeIntelligence) NewExecutorQueueGroup(containerName string) monitoring.Group {
	return Queue.NewGroup(containerName, monitoring.ObservableOwnerCodeIntel, QueueSizeGroupOptions{
		GroupConstructorOptions: GroupConstructorOptions{
			Namespace:       "executor",
			DescriptionRoot: "Executor jobs",

			ObservableConstructorOptions: ObservableConstructorOptions{
				MetricNameRoot:        "executor",
				MetricDescriptionRoot: "unprocessed executor job",
				By:                    []string{"queue"},
			},
		},

		QueueSize: NoAlertsOption("none"),
		QueueGrowthRate: NoAlertsOption(`
			This value compares the rate of enqueues against the rate of finished jobs for the selected queue.

				- A value < than 1 indicates that process rate > enqueue rate
				- A value = than 1 indicates that process rate = enqueue rate
				- A value > than 1 indicates that process rate < enqueue rate
		`),
	})
}

// src_executor_processor_total
// src_executor_processor_duration_seconds_bucket
// src_executor_processor_errors_total
// src_executor_processor_handlers
func (codeIntelligence) NewExecutorProcessorGroup(containerName string) monitoring.Group {
	filters := []string{`queue=~"${queue:regex}"`}

	return Workerutil.NewGroup(containerName, monitoring.ObservableOwnerCodeIntel, WorkerutilGroupOptions{
		GroupConstructorOptions: GroupConstructorOptions{
			Namespace:       "executor",
			DescriptionRoot: "Executor jobs",

			ObservableConstructorOptions: ObservableConstructorOptions{
				MetricNameRoot:        "executor",
				MetricDescriptionRoot: "handler",
				Filters:               filters,
			},
		},

		SharedObservationGroupOptions: SharedObservationGroupOptions{
			Total:     NoAlertsOption("none"),
			Duration:  NoAlertsOption("none"),
			Errors:    NoAlertsOption("none"),
			ErrorRate: NoAlertsOption("none"),
		},
		Handlers: NoAlertsOption("none"),
	})
}

// src_executor_run_lock_wait_total
// src_executor_run_lock_held_total
func (codeIntelligence) NewExecutorExecutionRunLockContentionGroup(containerName string) monitoring.Group {
	return monitoring.Group{
		Title:  "Run lock contention",
		Hidden: true,
		Rows: []monitoring.Row{
			{
				Standard.Count("wait")(ObservableConstructorOptions{
					MetricNameRoot:        "executor_run_lock_wait",
					MetricDescriptionRoot: "milliseconds",
				})(containerName, monitoring.ObservableOwnerCodeIntel).WithNoAlerts(`
					Number of milliseconds spent waiting for the run lock every 5m
				`).Observable(),

				Standard.Count("held")(ObservableConstructorOptions{
					MetricNameRoot:        "executor_run_lock_held",
					MetricDescriptionRoot: "milliseconds",
				})(containerName, monitoring.ObservableOwnerCodeIntel).WithNoAlerts(`
					Number of milliseconds spent holding for the run lock every 5m
				`).Observable(),
			},
		},
	}
}

// src_apiworker_command_total
// src_apiworker_command_duration_seconds_bucket
// src_apiworker_command_errors_total
func (codeIntelligence) NewExecutorSetupCommandGroup(containerName string) monitoring.Group {
	return Observation.NewGroup(containerName, monitoring.ObservableOwnerCodeIntel, ObservationGroupOptions{
		GroupConstructorOptions: GroupConstructorOptions{
			Namespace:       "executor",
			DescriptionRoot: "Job setup",
			Hidden:          true,

			ObservableConstructorOptions: ObservableConstructorOptions{
				MetricNameRoot:        "apiworker_command",
				MetricDescriptionRoot: "command",
				Filters:               []string{`op=~"setup.*"`},
				By:                    []string{"op"},
			},
		},

		SharedObservationGroupOptions: SharedObservationGroupOptions{
			Total:     NoAlertsOption("none"),
			Duration:  NoAlertsOption("none"),
			Errors:    NoAlertsOption("none"),
			ErrorRate: NoAlertsOption("none"),
		},
		Aggregate: &SharedObservationGroupOptions{
			Total:     NoAlertsOption("none"),
			Duration:  NoAlertsOption("none"),
			Errors:    NoAlertsOption("none"),
			ErrorRate: NoAlertsOption("none"),
		},
	})
}

// src_apiworker_command_total
// src_apiworker_command_duration_seconds_bucket
// src_apiworker_command_errors_total
func (codeIntelligence) NewExecutorExecutionCommandGroup(containerName string) monitoring.Group {
	return Observation.NewGroup(containerName, monitoring.ObservableOwnerCodeIntel, ObservationGroupOptions{
		GroupConstructorOptions: GroupConstructorOptions{
			Namespace:       "executor",
			DescriptionRoot: "Job execution",
			Hidden:          true,

			ObservableConstructorOptions: ObservableConstructorOptions{
				MetricNameRoot:        "apiworker_command",
				MetricDescriptionRoot: "command",
				Filters:               []string{`op=~"exec.*"`},
				By:                    []string{"op"},
			},
		},

		SharedObservationGroupOptions: SharedObservationGroupOptions{
			Total:     NoAlertsOption("none"),
			Duration:  NoAlertsOption("none"),
			Errors:    NoAlertsOption("none"),
			ErrorRate: NoAlertsOption("none"),
		},
		Aggregate: &SharedObservationGroupOptions{
			Total:     NoAlertsOption("none"),
			Duration:  NoAlertsOption("none"),
			Errors:    NoAlertsOption("none"),
			ErrorRate: NoAlertsOption("none"),
		},
	})
}

// src_apiworker_command_total
// src_apiworker_command_duration_seconds_bucket
// src_apiworker_command_errors_total
func (codeIntelligence) NewExecutorTeardownCommandGroup(containerName string) monitoring.Group {
	return Observation.NewGroup(containerName, monitoring.ObservableOwnerCodeIntel, ObservationGroupOptions{
		GroupConstructorOptions: GroupConstructorOptions{
			Namespace:       "executor",
			DescriptionRoot: "Job teardown",
			Hidden:          true,

			ObservableConstructorOptions: ObservableConstructorOptions{
				MetricNameRoot:        "apiworker_command",
				MetricDescriptionRoot: "command",
				Filters:               []string{`op=~"teardown.*"`},
				By:                    []string{"op"},
			},
		},

		SharedObservationGroupOptions: SharedObservationGroupOptions{
			Total:     NoAlertsOption("none"),
			Duration:  NoAlertsOption("none"),
			Errors:    NoAlertsOption("none"),
			ErrorRate: NoAlertsOption("none"),
		},
		Aggregate: &SharedObservationGroupOptions{
			Total:     NoAlertsOption("none"),
			Duration:  NoAlertsOption("none"),
			Errors:    NoAlertsOption("none"),
			ErrorRate: NoAlertsOption("none"),
		},
	})
}

// src_apiworker_apiclient_total
// src_apiworker_apiclient_duration_seconds_bucket
// src_apiworker_apiclient_errors_total
func (codeIntelligence) NewExecutorAPIClientGroup(containerName string) monitoring.Group {
	return Observation.NewGroup(containerName, monitoring.ObservableOwnerCodeIntel, ObservationGroupOptions{
		GroupConstructorOptions: GroupConstructorOptions{
			Namespace:       "executor",
			DescriptionRoot: "Queue API client",
			Hidden:          true,

			ObservableConstructorOptions: ObservableConstructorOptions{
				MetricNameRoot:        "apiworker_apiclient",
				MetricDescriptionRoot: "client",
				Filters:               nil,
				By:                    []string{"op"},
			},
		},

		SharedObservationGroupOptions: SharedObservationGroupOptions{
			Total:     NoAlertsOption("none"),
			Duration:  NoAlertsOption("none"),
			Errors:    NoAlertsOption("none"),
			ErrorRate: NoAlertsOption("none"),
		},
		Aggregate: &SharedObservationGroupOptions{
			Total:     NoAlertsOption("none"),
			Duration:  NoAlertsOption("none"),
			Errors:    NoAlertsOption("none"),
			ErrorRate: NoAlertsOption("none"),
		},
	})
}

// src_codeintel_dbstore_total
// src_codeintel_dbstore_duration_seconds_bucket
// src_codeintel_dbstore_errors_total
func (codeIntelligence) NewDBStoreGroup(containerName string) monitoring.Group {
	return Observation.NewGroup(containerName, monitoring.ObservableOwnerCodeIntel, ObservationGroupOptions{
		GroupConstructorOptions: GroupConstructorOptions{
			Namespace:       "codeintel",
			DescriptionRoot: "dbstore stats",
			Hidden:          true,

			ObservableConstructorOptions: ObservableConstructorOptions{
				MetricNameRoot:        "codeintel_dbstore",
				MetricDescriptionRoot: "store",
				By:                    []string{"op"},
			},
		},

		SharedObservationGroupOptions: SharedObservationGroupOptions{
			Total:     NoAlertsOption("none"),
			Duration:  NoAlertsOption("none"),
			Errors:    NoAlertsOption("none"),
			ErrorRate: NoAlertsOption("none"),
		},
		Aggregate: &SharedObservationGroupOptions{
			Total:     NoAlertsOption("none"),
			Duration:  NoAlertsOption("none"),
			Errors:    NoAlertsOption("none"),
			ErrorRate: NoAlertsOption("none"),
		},
	})
}

// src_workerutil_dbworker_store_codeintel_upload_total
// src_workerutil_dbworker_store_codeintel_upload_duration_seconds_bucket
// src_workerutil_dbworker_store_codeintel_upload_errors_total
func (codeIntelligence) NewUploadDBWorkerStoreGroup(containerName string) monitoring.Group {
	return Observation.NewGroup(containerName, monitoring.ObservableOwnerCodeIntel, ObservationGroupOptions{
		GroupConstructorOptions: GroupConstructorOptions{
			Namespace:       "workerutil",
			DescriptionRoot: "lsif_uploads dbworker/store stats",
			Hidden:          true,

			ObservableConstructorOptions: ObservableConstructorOptions{
				MetricNameRoot:        "workerutil_dbworker_store_codeintel_upload",
				MetricDescriptionRoot: "store",
			},
		},

		SharedObservationGroupOptions: SharedObservationGroupOptions{
			Total:     NoAlertsOption("none"),
			Duration:  NoAlertsOption("none"),
			Errors:    NoAlertsOption("none"),
			ErrorRate: NoAlertsOption("none"),
		},
	})
}

// src_workerutil_dbworker_store_codeintel_index_total
// src_workerutil_dbworker_store_codeintel_index_duration_seconds_bucket
// src_workerutil_dbworker_store_codeintel_index_errors_total
func (codeIntelligence) NewIndexDBWorkerStoreGroup(containerName string) monitoring.Group {
	return Observation.NewGroup(containerName, monitoring.ObservableOwnerCodeIntel, ObservationGroupOptions{
		GroupConstructorOptions: GroupConstructorOptions{
			Namespace:       "workerutil",
			DescriptionRoot: "lsif_indexes dbworker/store stats",
			Hidden:          true,

			ObservableConstructorOptions: ObservableConstructorOptions{
				MetricNameRoot:        "workerutil_dbworker_store_codeintel_index",
				MetricDescriptionRoot: "store",
			},
		},

		SharedObservationGroupOptions: SharedObservationGroupOptions{
			Total:     NoAlertsOption("none"),
			Duration:  NoAlertsOption("none"),
			Errors:    NoAlertsOption("none"),
			ErrorRate: NoAlertsOption("none"),
		},
	})
}

// src_workerutil_dbworker_store_codeintel_dependency_index_total
// src_workerutil_dbworker_store_codeintel_dependency_index_duration_seconds_bucket
// src_workerutil_dbworker_store_codeintel_dependency_index_errors_total
func (codeIntelligence) NewDependencyIndexDBWorkerStoreGroup(containerName string) monitoring.Group {
	return Observation.NewGroup(containerName, monitoring.ObservableOwnerCodeIntel, ObservationGroupOptions{
		GroupConstructorOptions: GroupConstructorOptions{
			Namespace:       "workerutil",
			DescriptionRoot: "lsif_dependency_indexes dbworker/store stats",
			Hidden:          true,

			ObservableConstructorOptions: ObservableConstructorOptions{
				MetricNameRoot:        "workerutil_dbworker_store_codeintel_dependency_index",
				MetricDescriptionRoot: "store",
			},
		},

		SharedObservationGroupOptions: SharedObservationGroupOptions{
			Total:     NoAlertsOption("none"),
			Duration:  NoAlertsOption("none"),
			Errors:    NoAlertsOption("none"),
			ErrorRate: NoAlertsOption("none"),
		},
	})
}

// src_codeintel_lsifstore_total
// src_codeintel_lsifstore_duration_seconds_bucket
// src_codeintel_lsifstore_errors_total
func (codeIntelligence) NewLSIFStoreGroup(containerName string) monitoring.Group {
	return Observation.NewGroup(containerName, monitoring.ObservableOwnerCodeIntel, ObservationGroupOptions{
		GroupConstructorOptions: GroupConstructorOptions{
			Namespace:       "codeintel",
			DescriptionRoot: "lsifstore stats",
			Hidden:          true,

			ObservableConstructorOptions: ObservableConstructorOptions{
				MetricNameRoot:        "codeintel_lsifstore",
				MetricDescriptionRoot: "store",
				By:                    []string{"op"},
			},
		},

		SharedObservationGroupOptions: SharedObservationGroupOptions{
			Total:     NoAlertsOption("none"),
			Duration:  NoAlertsOption("none"),
			Errors:    NoAlertsOption("none"),
			ErrorRate: NoAlertsOption("none"),
		},
		Aggregate: &SharedObservationGroupOptions{
			Total:     NoAlertsOption("none"),
			Duration:  NoAlertsOption("none"),
			Errors:    NoAlertsOption("none"),
			ErrorRate: NoAlertsOption("none"),
		},
	})
}

// src_codeintel_gitserver_total
// src_codeintel_gitserver_duration_seconds_bucket
// src_codeintel_gitserver_errors_total
func (codeIntelligence) NewGitserverClientGroup(containerName string) monitoring.Group {
	return Observation.NewGroup(containerName, monitoring.ObservableOwnerCodeIntel, ObservationGroupOptions{
		GroupConstructorOptions: GroupConstructorOptions{
			Namespace:       "codeintel",
			DescriptionRoot: "gitserver client",
			Hidden:          true,

			ObservableConstructorOptions: ObservableConstructorOptions{
				MetricNameRoot:        "codeintel_gitserver",
				MetricDescriptionRoot: "client",
				By:                    []string{"op"},
			},
		},

		SharedObservationGroupOptions: SharedObservationGroupOptions{
			Total:     NoAlertsOption("none"),
			Duration:  NoAlertsOption("none"),
			Errors:    NoAlertsOption("none"),
			ErrorRate: NoAlertsOption("none"),
		},
		Aggregate: &SharedObservationGroupOptions{
			Total:     NoAlertsOption("none"),
			Duration:  NoAlertsOption("none"),
			Errors:    NoAlertsOption("none"),
			ErrorRate: NoAlertsOption("none"),
		},
	})
}

// src_codeintel_uploadstore_total
// src_codeintel_uploadstore_duration_seconds_bucket
// src_codeintel_uploadstore_errors_total
func (codeIntelligence) NewUploadStoreGroup(containerName string) monitoring.Group {
	return Observation.NewGroup(containerName, monitoring.ObservableOwnerCodeIntel, ObservationGroupOptions{
		GroupConstructorOptions: GroupConstructorOptions{
			Namespace:       "codeintel",
			DescriptionRoot: "uploadstore stats",
			Hidden:          true,

			ObservableConstructorOptions: ObservableConstructorOptions{
				MetricNameRoot:        "codeintel_uploadstore",
				MetricDescriptionRoot: "store",
				By:                    []string{"op"},
			},
		},

		SharedObservationGroupOptions: SharedObservationGroupOptions{
			Total:     NoAlertsOption("none"),
			Duration:  NoAlertsOption("none"),
			Errors:    NoAlertsOption("none"),
			ErrorRate: NoAlertsOption("none"),
		},
		Aggregate: &SharedObservationGroupOptions{
			Total:     NoAlertsOption("none"),
			Duration:  NoAlertsOption("none"),
			Errors:    NoAlertsOption("none"),
			ErrorRate: NoAlertsOption("none"),
		},
	})
}

// src_codeintel_autoindex_enqueuer_total
// src_codeintel_autoindex_enqueuer_duration_seconds_bucket
// src_codeintel_autoindex_enqueuer_errors_total
func (codeIntelligence) NewAutoIndexEnqueuerGroup(containerName string) monitoring.Group {
	return Observation.NewGroup(containerName, monitoring.ObservableOwnerCodeIntel, ObservationGroupOptions{
		GroupConstructorOptions: GroupConstructorOptions{
			Namespace:       "codeintel",
			DescriptionRoot: "Auto-index enqueuer",
			Hidden:          true,

			ObservableConstructorOptions: ObservableConstructorOptions{
				MetricNameRoot:        "codeintel_autoindex_enqueuer",
				MetricDescriptionRoot: "enqueuer",
				By:                    []string{"op"},
			},
		},

		SharedObservationGroupOptions: SharedObservationGroupOptions{
			Total:     NoAlertsOption("none"),
			Duration:  NoAlertsOption("none"),
			Errors:    NoAlertsOption("none"),
			ErrorRate: NoAlertsOption("none"),
		},
		Aggregate: &SharedObservationGroupOptions{
			Total:     NoAlertsOption("none"),
			Duration:  NoAlertsOption("none"),
			Errors:    NoAlertsOption("none"),
			ErrorRate: NoAlertsOption("none"),
		},
	})
}

// src_codeintel_background_upload_records_removed_total
// src_codeintel_background_index_records_removed_total
// src_codeintel_background_uploads_purged_total
// src_codeintel_background_errors_total
func (codeIntelligence) NewJanitorGroup(containerName string) monitoring.Group {
	return monitoring.Group{
		Title:  fmt.Sprintf("%s: %s", titlecase("codeintel"), "Janitor stats"),
		Hidden: true,
		Rows: []monitoring.Row{
			{
				Standard.Count("records scanned")(ObservableConstructorOptions{
					MetricNameRoot:        "codeintel_background_repositories_scanned",
					MetricDescriptionRoot: "repository",
				})(containerName, monitoring.ObservableOwnerCodeIntel).WithNoAlerts(`
					Number of repositories considered for data retention scanning every 5m
				`).Observable(),

				Standard.Count("records scanned")(ObservableConstructorOptions{
					MetricNameRoot:        "codeintel_background_upload_records_scanned",
					MetricDescriptionRoot: "lsif upload",
				})(containerName, monitoring.ObservableOwnerCodeIntel).WithNoAlerts(`
					Number of upload records considered for data retention scanning every 5m
				`).Observable(),

				Standard.Count("commits scanned")(ObservableConstructorOptions{
					MetricNameRoot:        "codeintel_background_commits_scanned",
					MetricDescriptionRoot: "lsif upload",
				})(containerName, monitoring.ObservableOwnerCodeIntel).WithNoAlerts(`
					Number of commits considered for data retention scanning every 5m
				`).Observable(),

				Standard.Count("records expired")(ObservableConstructorOptions{
					MetricNameRoot:        "codeintel_background_upload_records_expired",
					MetricDescriptionRoot: "lsif upload",
				})(containerName, monitoring.ObservableOwnerCodeIntel).WithNoAlerts(`
					Number of upload records found to be expired every 5m
				`).Observable(),
			},
			{
				Standard.Count("records deleted")(ObservableConstructorOptions{
					MetricNameRoot:        "codeintel_background_upload_records_removed",
					MetricDescriptionRoot: "lsif upload",
				})(containerName, monitoring.ObservableOwnerCodeIntel).WithNoAlerts(`
					Number of LSIF upload records deleted due to expiration or unreachability every 5m
				`).Observable(),

				Standard.Count("records deleted")(ObservableConstructorOptions{
					MetricNameRoot:        "codeintel_background_index_records_removed",
					MetricDescriptionRoot: "lsif index",
				})(containerName, monitoring.ObservableOwnerCodeIntel).WithNoAlerts(`
					Number of LSIF index records deleted due to expiration or unreachability every 5m
				`).Observable(),

				Standard.Count("data bundles deleted")(ObservableConstructorOptions{
					MetricNameRoot:        "codeintel_background_uploads_purged",
					MetricDescriptionRoot: "lsif upload",
				})(containerName, monitoring.ObservableOwnerCodeIntel).WithNoAlerts(`
					Number of LSIF upload data bundles purged from the codeintel-db database every 5m
				`).Observable(),

				Standard.Count("records deleted")(ObservableConstructorOptions{
					MetricNameRoot:        "codeintel_background_documentation_search_records_removed",
					MetricDescriptionRoot: "documentation search record",
				})(containerName, monitoring.ObservableOwnerCodeIntel).WithNoAlerts(`
					Number of documentation search records removed from the codeintel-db database every 5m
				`).Observable(),
			},
			{

				Observation.Errors(ObservableConstructorOptions{
					MetricNameRoot:        "codeintel_background",
					MetricDescriptionRoot: "janitor",
				})(containerName, monitoring.ObservableOwnerCodeIntel).WithNoAlerts(`
					Number of code intelligence janitor errors every 5m
				`).Observable(),
			},
		},
	}
}

func (codeIntelligence) NewCoursierGroup(containerName string) monitoring.Group {
	return Observation.NewGroup(containerName, monitoring.ObservableOwnerCodeIntel, ObservationGroupOptions{
		GroupConstructorOptions: GroupConstructorOptions{
			Namespace:       "codeintel",
			DescriptionRoot: "Coursier invocation stats",
			Hidden:          true,
			ObservableConstructorOptions: ObservableConstructorOptions{
				MetricNameRoot:        "codeintel_coursier",
				MetricDescriptionRoot: "invocations",
				Filters:               []string{`op!="RunCommand"`},
				By:                    []string{"op"},
			},
		},
		SharedObservationGroupOptions: SharedObservationGroupOptions{
			Total:     NoAlertsOption("none"),
			Duration:  NoAlertsOption("none"),
			Errors:    NoAlertsOption("none"),
			ErrorRate: NoAlertsOption("none"),
		},
		Aggregate: &SharedObservationGroupOptions{
			Total:     NoAlertsOption("none"),
			Duration:  NoAlertsOption("none"),
			Errors:    NoAlertsOption("none"),
			ErrorRate: NoAlertsOption("none"),
		},
	})
}

func (codeIntelligence) NewDependencyReposStoreGroup(containerName string) monitoring.Group {
	return Observation.NewGroup(containerName, monitoring.ObservableOwnerCodeIntel, ObservationGroupOptions{
		GroupConstructorOptions: GroupConstructorOptions{
			Namespace:       "codeintel",
			DescriptionRoot: "Dependency repository insert",
			Hidden:          true,
			ObservableConstructorOptions: ObservableConstructorOptions{
				MetricNameRoot:        "codeintel_dependency_repos",
				MetricDescriptionRoot: "insert",
				Filters:               []string{},
				By:                    []string{"scheme", "new"}, // TODO  add 'op' if more operations added
			},
		},
		SharedObservationGroupOptions: SharedObservationGroupOptions{
			Total:     NoAlertsOption("none"),
			Duration:  NoAlertsOption("none"),
			Errors:    NoAlertsOption("none"),
			ErrorRate: NoAlertsOption("none"),
		},
		Aggregate: &SharedObservationGroupOptions{
			Total:     NoAlertsOption("none"),
			Duration:  NoAlertsOption("none"),
			Errors:    NoAlertsOption("none"),
			ErrorRate: NoAlertsOption("none"),
		},
	})
}
