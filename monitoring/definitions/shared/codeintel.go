package shared

import (
	"fmt"
	"strings"
	"time"

	"github.com/prometheus/common/model"

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
// src_codeintel_upload_queued_duration_seconds_total
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
		QueueMaxAge: CriticalOption(monitoring.Alert().GreaterOrEqual((time.Hour * 5).Seconds()), `
			An alert here could be indicative of a few things: an upload surfacing a pathological performance characteristic,
			precise-code-intel-worker being underprovisioned for the required upload processing throughput, or a higher replica
			count being required for the volume of uploads.
		`),
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
	group := Workerutil.NewGroup(containerName, monitoring.ObservableOwnerCodeIntel, WorkerutilGroupOptions{
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

	group.Rows[0] = append(group.Rows[0], monitoring.Observable{
		Name:           "codeintel_upload_processor_upload_size",
		Description:    "sum of upload sizes in bytes being processed by each precise code-intel worker instance",
		Owner:          monitoring.ObservableOwnerCodeIntel,
		Query:          `sum by(instance) (src_codeintel_upload_processor_upload_size{job="precise-code-intel-worker"})`,
		NoAlert:        true,
		Interpretation: "none",
		Panel:          monitoring.Panel().Unit(monitoring.Bytes).LegendFormat("{{instance}}"),
	})

	return group
}

// src_codeintel_commit_graph_total
// src_codeintel_commit_graph_processor_total
// src_codeintel_commit_graph_queued_duration_seconds_total
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
		QueueMaxAge: WarningOption(monitoring.Alert().GreaterOrEqual(time.Hour.Seconds()), `
			An alert here is generally indicative of either underprovisioned worker instance(s) and/or
			an underprovisioned main postgres instance.
		`),
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

// src_codeintel_autoindexing_total{op='HandleIndexSchedule'}
// src_codeintel_autoindexing_duration_seconds_bucket{op='HandleIndexSchedule'}
// src_codeintel_autoindexing_errors_total{op='HandleIndexSchedule'}
func (codeIntelligence) NewIndexSchedulerGroup(containerName string) monitoring.Group {
	return Observation.NewGroup(containerName, monitoring.ObservableOwnerCodeIntel, ObservationGroupOptions{
		GroupConstructorOptions: GroupConstructorOptions{
			Namespace:       "codeintel",
			DescriptionRoot: "Auto-index scheduler",
			Hidden:          true,

			ObservableConstructorOptions: ObservableConstructorOptions{
				MetricNameRoot:        "codeintel_autoindexing",
				Filters:               []string{"op='HandleIndexSchedule'"},
				MetricDescriptionRoot: "auto-indexing job scheduler",
				RangeWindow:           model.Duration(time.Minute) * 10,
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

// src_codeintel_dependency_index_total
// src_codeintel_dependency_index_processor_total
// src_codeintel_dependency_index_queued_duration_seconds_total
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

		QueueSize:   NoAlertsOption("none"),
		QueueMaxAge: NoAlertsOption("none"),
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
// src_executor_processor_duration_seconds_bucket
// src_executor_processor_errors_total
// src_executor_processor_handlers
func (codeIntelligence) NewExecutorProcessorGroup(containerName string) monitoring.Group {
	// TODO: pass in as variable like in NewExecutorQueueGroup?
	filters := []string{`queue=~"${queue:regex}"`}

	constructorOptions := ObservableConstructorOptions{
		MetricNameRoot:        "executor",
		JobLabel:              "sg_job",
		MetricDescriptionRoot: "executor",
		Filters:               filters,
	}

	queueConstructorOptions := ObservableConstructorOptions{
		MetricNameRoot:        "executor",
		MetricDescriptionRoot: "unprocessed executor job",
		By:                    []string{"queue"},
	}

	return Workerutil.NewGroup(containerName, monitoring.ObservableOwnerCodeIntel, WorkerutilGroupOptions{
		GroupConstructorOptions: GroupConstructorOptions{
			Namespace:       "executor",
			DescriptionRoot: "Executor jobs",

			ObservableConstructorOptions: constructorOptions,
		},

		SharedObservationGroupOptions: SharedObservationGroupOptions{
			Total:    NoAlertsOption("none"),
			Duration: NoAlertsOption("none"),
			Errors:   NoAlertsOption("none"),
			ErrorRate: WarningOption(
				monitoring.Alert().
					CustomQuery(Workerutil.LastOverTimeErrorRate(containerName, model.Duration(time.Hour*5), constructorOptions)).
					For(time.Hour).
					GreaterOrEqual(100),
				`
				- Determine the cause of failure from the auto-indexing job logs in the site-admin page.
				- This alert fires if all executor jobs have been failing for the past hour. The alert will continue for up
				to 5 hours until the error rate is no longer 100%, even if there are no running jobs in that time, as the
				problem is not know to be resolved until jobs start succeeding again.
			`),
		},
		Handlers: CriticalOption(
			monitoring.Alert().
				CustomQuery(Workerutil.QueueForwardProgress(containerName, constructorOptions, queueConstructorOptions)).
				CustomDescription("0 active executor handlers and > 0 queue size").
				LessOrEqual(0).
				// ~5min for scale-from-zero
				For(time.Minute*5),
			`
			- Check to see the state of any compute VMs, they may be taking longer than expected to boot.
			- Make sure the executors appear under Site Admin > Executors.
			- Check the Grafana dashboard section for APIClient, it should do frequent requests to Dequeue and Heartbeat and those must not fail.
		`),
	})
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
				JobLabel:              "sg_job",
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
				JobLabel:              "sg_job",
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
				JobLabel:              "sg_job",
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

// src_apiworker_apiclient_queue_total
// src_apiworker_apiclient_queue_duration_seconds_bucket
// src_apiworker_apiclient_queue_errors_total
func (codeIntelligence) NewExecutorAPIQueueClientGroup(containerName string) monitoring.Group {
	return Observation.NewGroup(containerName, monitoring.ObservableOwnerCodeIntel, ObservationGroupOptions{
		GroupConstructorOptions: GroupConstructorOptions{
			Namespace:       "executor",
			DescriptionRoot: "Queue API client",
			Hidden:          true,

			ObservableConstructorOptions: ObservableConstructorOptions{
				MetricNameRoot:        "apiworker_apiclient_queue",
				JobLabel:              "sg_job",
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

// src_apiworker_apiclient_files_total
// src_apiworker_apiclient_files_duration_seconds_bucket
// src_apiworker_apiclient_files_errors_total
func (codeIntelligence) NewExecutorAPIFilesClientGroup(containerName string) monitoring.Group {
	return Observation.NewGroup(containerName, monitoring.ObservableOwnerCodeIntel, ObservationGroupOptions{
		GroupConstructorOptions: GroupConstructorOptions{
			Namespace:       "executor",
			DescriptionRoot: "Files API client",
			Hidden:          true,

			ObservableConstructorOptions: ObservableConstructorOptions{
				MetricNameRoot:        "apiworker_apiclient_files",
				JobLabel:              "sg_job",
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
				MetricNameRoot:        "codeintel_uploads_store",
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
				MetricNameRoot:        "codeintel_uploads_lsifstore",
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

// src_codeintel_dependencies_total
// src_codeintel_dependencies_duration_seconds_bucket
// src_codeintel_dependencies_errors_total
func (codeIntelligence) NewDependencyServiceGroup(containerName string) monitoring.Group {
	return Observation.NewGroup(containerName, monitoring.ObservableOwnerCodeIntel, ObservationGroupOptions{
		GroupConstructorOptions: GroupConstructorOptions{
			Namespace:       "codeintel",
			DescriptionRoot: "dependencies service stats",
			Hidden:          true,

			ObservableConstructorOptions: ObservableConstructorOptions{
				MetricNameRoot:        "codeintel_dependencies",
				MetricDescriptionRoot: "service",
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

// src_codeintel_dependencies_store_total
// src_codeintel_dependencies_store_duration_seconds_bucket
// src_codeintel_dependencies_store_errors_total
func (codeIntelligence) NewDependencyStoreGroup(containerName string) monitoring.Group {
	return Observation.NewGroup(containerName, monitoring.ObservableOwnerCodeIntel, ObservationGroupOptions{
		GroupConstructorOptions: GroupConstructorOptions{
			Namespace:       "codeintel",
			DescriptionRoot: "dependencies service store stats",
			Hidden:          true,

			ObservableConstructorOptions: ObservableConstructorOptions{
				MetricNameRoot:        "codeintel_dependencies_background",
				MetricDescriptionRoot: "service",
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

// src_codeintel_dependencies_background_total
// src_codeintel_dependencies_background_duration_seconds_bucket
// src_codeintel_dependencies_background_errors_total
func (codeIntelligence) NewDependencyBackgroundJobGroup(containerName string) monitoring.Group {
	return Observation.NewGroup(containerName, monitoring.ObservableOwnerCodeIntel, ObservationGroupOptions{
		GroupConstructorOptions: GroupConstructorOptions{
			Namespace:       "codeintel",
			DescriptionRoot: "dependencies service background stats",
			Hidden:          true,

			ObservableConstructorOptions: ObservableConstructorOptions{
				MetricNameRoot:        "codeintel_dependencies_background",
				MetricDescriptionRoot: "service",
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

// src_codeintel_lockfiles_total
// src_codeintel_lockfiles_duration_seconds_bucket
// src_codeintel_lockfiles_errors_total
func (codeIntelligence) NewLockfilesGroup(containerName string) monitoring.Group {
	return Observation.NewGroup(containerName, monitoring.ObservableOwnerCodeIntel, ObservationGroupOptions{
		GroupConstructorOptions: GroupConstructorOptions{
			Namespace:       "codeintel",
			DescriptionRoot: "lockfiles service stats",
			Hidden:          true,

			ObservableConstructorOptions: ObservableConstructorOptions{
				MetricNameRoot:        "codeintel_lockfiles",
				MetricDescriptionRoot: "service",
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

func newPackageManagerGroup(packageManager string, containerName string) monitoring.Group {
	return Observation.NewGroup(containerName, monitoring.ObservableOwnerCodeIntel, ObservationGroupOptions{
		GroupConstructorOptions: GroupConstructorOptions{
			Namespace:       "codeintel",
			DescriptionRoot: fmt.Sprintf("%s invocation stats", packageManager),
			Hidden:          true,
			ObservableConstructorOptions: ObservableConstructorOptions{
				MetricNameRoot:        fmt.Sprintf("codeintel_%s", strings.ToLower(packageManager)),
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

func (codeIntelligence) NewCoursierGroup(containerName string) monitoring.Group {
	return newPackageManagerGroup("Coursier", containerName)
}

func (codeIntelligence) NewNpmGroup(containerName string) monitoring.Group {
	return newPackageManagerGroup("npm", containerName)
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

func (codeIntelligence) NewSymbolsAPIGroup(containerName string) monitoring.Group {
	return Observation.NewGroup(containerName, monitoring.ObservableOwnerCodeIntel, ObservationGroupOptions{
		GroupConstructorOptions: GroupConstructorOptions{
			Namespace:       "codeintel",
			DescriptionRoot: "Symbols API",
			Hidden:          false,
			ObservableConstructorOptions: ObservableConstructorOptions{
				MetricNameRoot:        "codeintel_symbols_api",
				MetricDescriptionRoot: "API",
				Filters:               []string{},
				By:                    []string{"op", "parseAmount"},
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

func (codeIntelligence) NewSymbolsParserGroup(containerName string) monitoring.Group {
	group := Observation.NewGroup(containerName, monitoring.ObservableOwnerCodeIntel, ObservationGroupOptions{
		GroupConstructorOptions: GroupConstructorOptions{
			Namespace:       "codeintel",
			DescriptionRoot: "Symbols parser",
			Hidden:          false,
			ObservableConstructorOptions: ObservableConstructorOptions{
				MetricNameRoot:        "codeintel_symbols_parser",
				MetricDescriptionRoot: "parser",
				Filters:               []string{},
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

	queueRow := monitoring.Row{
		{
			Name:           containerName,
			Description:    "in-flight parse jobs",
			Owner:          monitoring.ObservableOwnerCodeIntel,
			Query:          "max(src_codeintel_symbols_parsing{job=~\"^symbols.*\"})",
			NoAlert:        true,
			Interpretation: "none",
			Panel:          monitoring.Panel(),
		},
		{
			Name:           containerName,
			Description:    "parser queue size",
			Owner:          monitoring.ObservableOwnerCodeIntel,
			Query:          "max(src_codeintel_symbols_parse_queue_size{job=~\"^symbols.*\"})",
			NoAlert:        true,
			Interpretation: "none",
			Panel:          monitoring.Panel(),
		},
		{
			Name:           containerName,
			Description:    "parse queue timeouts",
			Owner:          monitoring.ObservableOwnerCodeIntel,
			Query:          "max(src_codeintel_symbols_parse_queue_timeouts_total{job=~\"^symbols.*\"})",
			NoAlert:        true,
			Interpretation: "none",
			Panel:          monitoring.Panel(),
		},
		{
			Name:           containerName,
			Description:    "parse failures every 5m",
			Owner:          monitoring.ObservableOwnerCodeIntel,
			Query:          "rate(src_codeintel_symbols_parse_failed_total{job=~\"^symbols.*\"}[5m])",
			NoAlert:        true,
			Interpretation: "none",
			Panel:          monitoring.Panel(),
		},
	}

	group.Rows = append([]monitoring.Row{queueRow}, group.Rows...)
	return group
}

func (codeIntelligence) NewSymbolsCacheJanitorGroup(containerName string) monitoring.Group {
	return monitoring.Group{
		Title:  fmt.Sprintf("%s: %s", "Codeintel", "Symbols cache janitor"),
		Hidden: true,
		Rows: []monitoring.Row{
			{
				{
					Name:           containerName,
					Description:    "size in bytes of the on-disk cache",
					Owner:          monitoring.ObservableOwnerCodeIntel,
					Query:          "src_codeintel_symbols_store_cache_size_bytes",
					NoAlert:        true,
					Interpretation: "no",
					Panel:          monitoring.Panel().Unit(monitoring.Bytes),
				},
				{
					Name:           containerName,
					Description:    "cache eviction operations every 5m",
					Owner:          monitoring.ObservableOwnerCodeIntel,
					Query:          "rate(src_codeintel_symbols_store_evictions_total[5m])",
					NoAlert:        true,
					Interpretation: "no",
					Panel:          monitoring.Panel(),
				},
				{
					Name:           containerName,
					Description:    "cache eviction operation errors every 5m",
					Owner:          monitoring.ObservableOwnerCodeIntel,
					Query:          "rate(src_codeintel_symbols_store_errors_total[5m])",
					NoAlert:        true,
					Interpretation: "no",
					Panel:          monitoring.Panel(),
				},
			},
		},
	}
}

func (codeIntelligence) NewSymbolsRepositoryFetcherGroup(containerName string) monitoring.Group {
	group := Observation.NewGroup(containerName, monitoring.ObservableOwnerCodeIntel, ObservationGroupOptions{
		GroupConstructorOptions: GroupConstructorOptions{
			Namespace:       "codeintel",
			DescriptionRoot: "Symbols repository fetcher",
			Hidden:          true,
			ObservableConstructorOptions: ObservableConstructorOptions{
				MetricNameRoot:        "codeintel_symbols_repository_fetcher",
				MetricDescriptionRoot: "fetcher",
				Filters:               []string{},
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

	queueRow := monitoring.Row{
		{
			Name:           containerName,
			Description:    "in-flight repository fetch operations",
			Owner:          monitoring.ObservableOwnerCodeIntel,
			Query:          "src_codeintel_symbols_fetching",
			NoAlert:        true,
			Interpretation: "none",
			Panel:          monitoring.Panel(),
		},
		{
			Name:           containerName,
			Description:    "repository fetch queue size",
			Owner:          monitoring.ObservableOwnerCodeIntel,
			Query:          "max(src_codeintel_symbols_fetch_queue_size{job=~\"^symbols.*\"})",
			NoAlert:        true,
			Interpretation: "none",
			Panel:          monitoring.Panel(),
		},
	}

	group.Rows = append([]monitoring.Row{queueRow}, group.Rows...)
	return group
}

func (codeIntelligence) NewSymbolsGitserverClientGroup(containerName string) monitoring.Group {
	return Observation.NewGroup(containerName, monitoring.ObservableOwnerCodeIntel, ObservationGroupOptions{
		GroupConstructorOptions: GroupConstructorOptions{
			Namespace:       "codeintel",
			DescriptionRoot: "Symbols gitserver client",
			Hidden:          true,
			ObservableConstructorOptions: ObservableConstructorOptions{
				MetricNameRoot:        "codeintel_symbols_gitserver",
				MetricDescriptionRoot: "gitserver client",
				Filters:               []string{},
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

// src_{task_name}_total
// src_{task_name}_duration_seconds_bucket
// src_{task_name}_errors_total
// src_{task_name}_records_processed_total
// src_{task_name}_records_altered_total
func (codeIntelligence) newPipelineGroups(
	titlePrefix string,
	containerName string,
	taskNames []string,
) []monitoring.Group {
	groups := make([]monitoring.Group, 0, len(taskNames))
	for _, taskName := range taskNames {
		groups = append(groups, CodeIntelligence.newPipelineGroup(titlePrefix, containerName, taskName))
	}

	return groups
}

// src_{task_name}_total
// src_{task_name}_duration_seconds_bucket
// src_{task_name}_errors_total
// src_{task_name}_records_processed_total
// src_{task_name}_records_altered_total
func (codeIntelligence) newPipelineGroup(
	titlePrefix string,
	containerName string,
	taskName string,
) monitoring.Group {
	group := Observation.NewGroup(containerName, monitoring.ObservableOwnerCodeIntel, ObservationGroupOptions{
		GroupConstructorOptions: GroupConstructorOptions{
			Namespace:       "codeintel",
			DescriptionRoot: fmt.Sprintf("%s > %s", titlePrefix, titlecase(strings.ReplaceAll(taskName, "_", " "))),
			Hidden:          true,
			ObservableConstructorOptions: ObservableConstructorOptions{
				MetricNameRoot:        taskName,
				MetricDescriptionRoot: "job invocation",
				Filters:               []string{},
				By:                    []string{"op"},
			},
		},
		SharedObservationGroupOptions: SharedObservationGroupOptions{
			Total:     NoAlertsOption("none"),
			Duration:  NoAlertsOption("none"),
			Errors:    NoAlertsOption("none"),
			ErrorRate: NoAlertsOption("none"),
		},
	})

	recordProgressRow := monitoring.Row{
		Standard.Count("processed")(ObservableConstructorOptions{
			MetricNameRoot:        fmt.Sprintf("%s_records_processed", taskName),
			MetricDescriptionRoot: "records",
		})(containerName, monitoring.ObservableOwnerCodeIntel).WithNoAlerts(`
			The number of candidate records considered for cleanup.
		`).Observable(),

		Standard.Count("altered")(ObservableConstructorOptions{
			MetricNameRoot:        fmt.Sprintf("%s_records_altered", taskName),
			MetricDescriptionRoot: "records",
		})(containerName, monitoring.ObservableOwnerCodeIntel).WithNoAlerts(`
			The number of candidate records altered as part of cleanup.
		`).Observable(),
	}

	group.Rows = append([]monitoring.Row{recordProgressRow}, group.Rows...)
	return group
}

// src_{task_name}_total
// src_{task_name}_duration_seconds_bucket
// src_{task_name}_errors_total
// src_{task_name}_records_scanned_total
// src_{task_name}_records_altered_total
func (codeIntelligence) newJanitorGroups(
	titlePrefix string,
	containerName string,
	taskNames []string,
) []monitoring.Group {
	groups := make([]monitoring.Group, 0, len(taskNames))
	for _, taskName := range taskNames {
		groups = append(groups, CodeIntelligence.newJanitorGroup(titlePrefix, containerName, taskName))
	}

	return groups
}

// src_{task_name}_total
// src_{task_name}_duration_seconds_bucket
// src_{task_name}_errors_total
// src_{task_name}_records_scanned_total
// src_{task_name}_records_altered_total
func (codeIntelligence) newJanitorGroup(
	titlePrefix string,
	containerName string,
	taskName string,
) monitoring.Group {
	group := Observation.NewGroup(containerName, monitoring.ObservableOwnerCodeIntel, ObservationGroupOptions{
		GroupConstructorOptions: GroupConstructorOptions{
			Namespace:       "codeintel",
			DescriptionRoot: fmt.Sprintf("%s > %s", titlePrefix, titlecase(strings.ReplaceAll(taskName, "_", " "))),
			Hidden:          true,
			ObservableConstructorOptions: ObservableConstructorOptions{
				MetricNameRoot:        taskName,
				MetricDescriptionRoot: "job invocation",
				Filters:               []string{},
				By:                    []string{"op"},
			},
		},
		SharedObservationGroupOptions: SharedObservationGroupOptions{
			Total:     NoAlertsOption("none"),
			Duration:  NoAlertsOption("none"),
			Errors:    NoAlertsOption("none"),
			ErrorRate: NoAlertsOption("none"),
		},
	})

	recordProgressRow := monitoring.Row{
		Standard.Count("scanned")(ObservableConstructorOptions{
			MetricNameRoot:        fmt.Sprintf("%s_records_scanned", taskName),
			MetricDescriptionRoot: "records",
		})(containerName, monitoring.ObservableOwnerCodeIntel).WithNoAlerts(`
			The number of candidate records considered for cleanup.
		`).Observable(),

		Standard.Count("altered")(ObservableConstructorOptions{
			MetricNameRoot:        fmt.Sprintf("%s_records_altered", taskName),
			MetricDescriptionRoot: "records",
		})(containerName, monitoring.ObservableOwnerCodeIntel).WithNoAlerts(`
			The number of candidate records altered as part of cleanup.
		`).Observable(),
	}

	group.Rows = append([]monitoring.Row{recordProgressRow}, group.Rows...)
	return group
}
