package definitions

import (
	"fmt"
	"time"

	"github.com/sourcegraph/sourcegraph/monitoring/definitions/shared"
	"github.com/sourcegraph/sourcegraph/monitoring/monitoring"
)

func Worker() *monitoring.Container {
	const containerName = "worker"

	var workerJobs = []struct {
		Name  string
		Owner monitoring.ObservableOwner
	}{
		{Name: "codeintel-janitor", Owner: monitoring.ObservableOwnerCodeIntel},
		{Name: "codeintel-commitgraph", Owner: monitoring.ObservableOwnerCodeIntel},
		{Name: "codeintel-auto-indexing", Owner: monitoring.ObservableOwnerCodeIntel},
	}

	var activeJobObservables []monitoring.Observable
	for _, job := range workerJobs {
		activeJobObservables = append(activeJobObservables, monitoring.Observable{
			Name:          fmt.Sprintf("worker_job_%s_count", job.Name),
			Description:   fmt.Sprintf("number of worker instances running the %s job", job.Name),
			Query:         fmt.Sprintf(`sum (src_worker_jobs{job="worker", job_name="%s"})`, job.Name),
			Panel:         monitoring.Panel().LegendFormat(fmt.Sprintf("instances running %s", job.Name)),
			DataMustExist: true,
			Warning:       monitoring.Alert().Less(1, nil).For(1 * time.Minute),
			Critical:      monitoring.Alert().Less(1, nil).For(5 * time.Minute),
			Owner:         job.Owner,
			PossibleSolutions: fmt.Sprintf(`
				- Ensure your instance defines a worker container such that:
					- `+"`"+`WORKER_JOB_ALLOWLIST`+"`"+` contains "%[1]s" (or "all"), and
					- `+"`"+`WORKER_JOB_BLOCKLIST`+"`"+` does not contain "%[1]s"
				- Ensure that such a container is not failing to start or stay active
			`, job.Name),
		})
	}

	panelsPerRow := 4
	if rem := len(activeJobObservables) % panelsPerRow; rem == 1 || rem == 2 {
		// If we'd leave one or two panels on the only/last row, then reduce
		// the number of panels in previous rows so that we have less of a width
		// difference at the end
		panelsPerRow = 3
	}

	var activeJobRows []monitoring.Row
	for _, observable := range activeJobObservables {
		if n := len(activeJobRows); n == 0 || len(activeJobRows[n-1]) >= panelsPerRow {
			activeJobRows = append(activeJobRows, nil)
		}

		n := len(activeJobRows)
		activeJobRows[n-1] = append(activeJobRows[n-1], observable)
	}

	activeJobsGroup := monitoring.Group{
		Title: "Active jobs",
		Rows: append(
			[]monitoring.Row{
				{
					{
						Name:        "worker_job_count",
						Description: "number of worker instances running each job",
						Query:       `sum by (job_name) (src_worker_jobs{job="worker"})`,
						Panel:       monitoring.Panel().LegendFormat("instances running {{job_name}}"),
						NoAlert:     true,
						Interpretation: `
							The number of worker instances running each job type.
							It is necessary for each job type to be managed by at least one worker instance.
						`,
					},
				},
			},
			activeJobRows...,
		),
	}

	codeintelJanitorStatsGroup := monitoring.Group{
		Title:  "[codeintel] Janitor stats",
		Hidden: true,
		Rows: []monitoring.Row{
			{
				shared.Standard.Count("records deleted")(shared.ObservableConstructorOptions{
					MetricNameRoot:        "codeintel_background_upload_records_removed",
					MetricDescriptionRoot: "lsif_upload",
				})(containerName, monitoring.ObservableOwnerCodeIntel).WithNoAlerts(`
						Number of LSIF upload records deleted due to expiration or unreachability every 5m
					`).Observable(),

				shared.Standard.Count("records deleted")(shared.ObservableConstructorOptions{
					MetricNameRoot:        "codeintel_background_index_records_removed",
					MetricDescriptionRoot: "lsif_index",
				})(containerName, monitoring.ObservableOwnerCodeIntel).WithNoAlerts(`
						Number of LSIF index records deleted due to expiration or unreachability every 5m
					`).Observable(),

				shared.Standard.Count("data bundles deleted")(shared.ObservableConstructorOptions{
					MetricNameRoot:        "codeintel_background_uploads_purged",
					MetricDescriptionRoot: "lsif_upload",
				})(containerName, monitoring.ObservableOwnerCodeIntel).WithNoAlerts(`
						Number of LSIF upload data bundles purged from the codeintel-db database every 5m
					`).Observable(),

				shared.Observation.Errors(shared.ObservableConstructorOptions{
					MetricNameRoot:        "codeintel_background",
					MetricDescriptionRoot: "janitor",
				})(containerName, monitoring.ObservableOwnerCodeIntel).WithNoAlerts(`
						Number of code intelligence janitor errors every 5m
					`).Observable(),
			},
		},
	}

	return &monitoring.Container{
		Name:        "worker",
		Title:       "Worker",
		Description: "Manages background processes.",
		Groups: []monitoring.Group{
			// src_worker_jobs
			activeJobsGroup,

			// src_codeintel_commit_graph_total
			// src_codeintel_commit_graph_processor_total
			shared.Queue.NewGroup(containerName, monitoring.ObservableOwnerCodeIntel, shared.QueueSizeGroupOptions{
				GroupConstructorOptions: shared.GroupConstructorOptions{
					Namespace:       "codeintel",
					DescriptionRoot: "Repository with stale commit graph",
					Hidden:          true,

					ObservableConstructorOptions: shared.ObservableConstructorOptions{
						MetricNameRoot:        "codeintel_commit_graph",
						MetricDescriptionRoot: "repository",
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

			// src_codeintel_commit_graph_processor_total
			// src_codeintel_commit_graph_processor_duration_seconds_bucket
			// src_codeintel_commit_graph_processor_errors_total
			shared.Observation.NewGroup(containerName, monitoring.ObservableOwnerCodeIntel, shared.ObservationGroupOptions{
				GroupConstructorOptions: shared.GroupConstructorOptions{
					Namespace:       "codeintel",
					DescriptionRoot: "Repository commit graph updates",
					Hidden:          true,

					ObservableConstructorOptions: shared.ObservableConstructorOptions{
						MetricNameRoot:        "codeintel_commit_graph_processor",
						MetricDescriptionRoot: "update",
					},
				},

				Total:    shared.NoAlertsOption("none"),
				Duration: shared.NoAlertsOption("none"),
				Errors:   shared.NoAlertsOption("none"),
			}),

			// src_codeintel_dependency_index_total
			// src_codeintel_dependency_index_processor_total
			shared.Queue.NewGroup(containerName, monitoring.ObservableOwnerCodeIntel, shared.QueueSizeGroupOptions{
				GroupConstructorOptions: shared.GroupConstructorOptions{
					Namespace:       "codeintel",
					DescriptionRoot: "Dependency index job",
					Hidden:          true,

					ObservableConstructorOptions: shared.ObservableConstructorOptions{
						MetricNameRoot:        "codeintel_dependency_index",
						MetricDescriptionRoot: "dependency index job",
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

			// src_codeintel_dependency_index_processor_total
			// src_codeintel_dependency_index_processor_duration_seconds_bucket
			// src_codeintel_dependency_index_processor_errors_total
			// src_codeintel_dependency_index_processor_handlers
			shared.Workerutil.NewGroup(containerName, monitoring.ObservableOwnerCodeIntel, shared.WorkerutilGroupOptions{
				GroupConstructorOptions: shared.GroupConstructorOptions{
					Namespace:       "codeintel",
					DescriptionRoot: "Dependency index jobs",
					Hidden:          true,

					ObservableConstructorOptions: shared.ObservableConstructorOptions{
						MetricNameRoot:        "codeintel_dependency_index",
						MetricDescriptionRoot: "handler",
					},
				},

				Total:    shared.NoAlertsOption("none"),
				Duration: shared.NoAlertsOption("none"),
				Errors:   shared.NoAlertsOption("none"),
				Handlers: shared.NoAlertsOption("none"),
			}),

			// src_codeintel_background_upload_records_removed_total
			// src_codeintel_background_index_records_removed_total
			// src_codeintel_background_uploads_purged_total
			// src_codeintel_background_errors_total
			codeintelJanitorStatsGroup,

			// src_codeintel_index_scheduler_total
			// src_codeintel_index_scheduler_duration_seconds_bucket
			// src_codeintel_index_scheduler_errors_total
			shared.Observation.NewGroup(containerName, monitoring.ObservableOwnerCodeIntel, shared.ObservationGroupOptions{
				GroupConstructorOptions: shared.GroupConstructorOptions{
					Namespace:       "codeintel",
					DescriptionRoot: "Auto-index scheduler",
					Hidden:          true,

					ObservableConstructorOptions: shared.ObservableConstructorOptions{
						MetricNameRoot:        "codeintel_index_scheduler",
						MetricDescriptionRoot: "scheduler",
					},
				},

				Total:    shared.NoAlertsOption("none"),
				Duration: shared.NoAlertsOption("none"),
				Errors:   shared.NoAlertsOption("none"),
			}),

			// src_codeintel_autoindex_enqueuer_total
			// src_codeintel_autoindex_enqueuer_duration_seconds_bucket
			// src_codeintel_autoindex_enqueuer_errors_total
			shared.Observation.NewGroup(containerName, monitoring.ObservableOwnerCodeIntel, shared.ObservationGroupOptions{
				GroupConstructorOptions: shared.GroupConstructorOptions{
					Namespace:       "codeintel",
					DescriptionRoot: "Auto-index enqueuer",
					Hidden:          true,

					ObservableConstructorOptions: shared.ObservableConstructorOptions{
						MetricNameRoot:        "codeintel_autoindex_enqueuer",
						MetricDescriptionRoot: "enqueuer",
					},
				},

				Total:    shared.NoAlertsOption("none"),
				Duration: shared.NoAlertsOption("none"),
				Errors:   shared.NoAlertsOption("none"),
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

			// src_workerutil_dbworker_store_codeintel_dependency_index_total
			// src_workerutil_dbworker_store_codeintel_dependency_index_duration_seconds_bucket
			// src_workerutil_dbworker_store_codeintel_dependency_index_errors_total
			shared.Observation.NewGroup(containerName, monitoring.ObservableOwnerCodeIntel, shared.ObservationGroupOptions{
				GroupConstructorOptions: shared.GroupConstructorOptions{
					Namespace:       "workerutil",
					DescriptionRoot: "dbworker/store stats (db=frontend, table=lsif_dependency_indexes)",
					Hidden:          true,

					ObservableConstructorOptions: shared.ObservableConstructorOptions{
						MetricNameRoot:        "workerutil_dbworker_store_codeintel_dependency_index",
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

			// src_codeintel_background_upload_resets_total
			// src_codeintel_background_upload_reset_failures_total
			// src_codeintel_background_upload_reset_errors_total
			shared.WorkerutilResetter.NewGroup(containerName, monitoring.ObservableOwnerCodeIntel, shared.ResetterGroupOptions{
				GroupConstructorOptions: shared.GroupConstructorOptions{
					Namespace:       "codeintel",
					DescriptionRoot: "lsif_upload record resetter",
					Hidden:          true,

					ObservableConstructorOptions: shared.ObservableConstructorOptions{
						MetricNameRoot:        "codeintel_background_upload",
						MetricDescriptionRoot: "lsif_upload",
					},
				},

				RecordResets:        shared.NoAlertsOption("none"),
				RecordResetFailures: shared.NoAlertsOption("none"),
				Errors:              shared.NoAlertsOption("none"),
			}),

			// src_codeintel_background_index_resets_total
			// src_codeintel_background_index_reset_failures_total
			// src_codeintel_background_index_reset_errors_total
			shared.WorkerutilResetter.NewGroup(containerName, monitoring.ObservableOwnerCodeIntel, shared.ResetterGroupOptions{
				GroupConstructorOptions: shared.GroupConstructorOptions{
					Namespace:       "codeintel",
					DescriptionRoot: "lsif_index record resetter",
					Hidden:          true,

					ObservableConstructorOptions: shared.ObservableConstructorOptions{
						MetricNameRoot:        "codeintel_background_index",
						MetricDescriptionRoot: "lsif_index",
					},
				},

				RecordResets:        shared.NoAlertsOption("none"),
				RecordResetFailures: shared.NoAlertsOption("none"),
				Errors:              shared.NoAlertsOption("none"),
			}),

			// src_codeintel_background_dependency_index_resets_total
			// src_codeintel_background_dependency_index_reset_failures_total
			// src_codeintel_background_dependency_index_reset_errors_total
			shared.WorkerutilResetter.NewGroup(containerName, monitoring.ObservableOwnerCodeIntel, shared.ResetterGroupOptions{
				GroupConstructorOptions: shared.GroupConstructorOptions{
					Namespace:       "codeintel",
					DescriptionRoot: "lsif_dependency_index record resetter",
					Hidden:          true,

					ObservableConstructorOptions: shared.ObservableConstructorOptions{
						MetricNameRoot:        "codeintel_background_dependency_index",
						MetricDescriptionRoot: "lsif_dependency_index",
					},
				},

				RecordResets:        shared.NoAlertsOption("none"),
				RecordResetFailures: shared.NoAlertsOption("none"),
				Errors:              shared.NoAlertsOption("none"),
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
