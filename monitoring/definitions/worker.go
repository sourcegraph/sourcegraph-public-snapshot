package definitions

import (
	"fmt"
	"time"

	"github.com/sourcegraph/sourcegraph/monitoring/definitions/shared"
	"github.com/sourcegraph/sourcegraph/monitoring/monitoring"
)

func Worker() *monitoring.Dashboard {
	const containerName = "worker"

	scrapeJobRegex := fmt.Sprintf("^%s.*", containerName)

	workerJobs := []struct {
		Name  string
		Owner monitoring.ObservableOwner
	}{
		{Name: "codeintel-upload-janitor", Owner: monitoring.ObservableOwnerCodeIntel},
		{Name: "codeintel-commitgraph-updater", Owner: monitoring.ObservableOwnerCodeIntel},
		{Name: "codeintel-autoindexing-scheduler", Owner: monitoring.ObservableOwnerCodeIntel},
	}

	var activeJobObservables []monitoring.Observable
	for _, job := range workerJobs {
		activeJobObservables = append(activeJobObservables, monitoring.Observable{
			Name:          fmt.Sprintf("worker_job_%s_count", job.Name),
			Description:   fmt.Sprintf("number of worker instances running the %s job", job.Name),
			Query:         fmt.Sprintf(`sum (src_worker_jobs{job=~%q, job_name="%s"})`, scrapeJobRegex, job.Name),
			Panel:         monitoring.Panel().LegendFormat(fmt.Sprintf("instances running %s", job.Name)),
			DataMustExist: true,
			Warning:       monitoring.Alert().Less(1).For(1 * time.Minute),
			Critical:      monitoring.Alert().Less(1).For(5 * time.Minute),
			Owner:         job.Owner,
			NextSteps: fmt.Sprintf(`
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
						Query:       fmt.Sprintf(`sum by (job_name) (src_worker_jobs{job=~%q})`, scrapeJobRegex),
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

	recordEncrypterGroup := monitoring.Group{
		Title:  "Database record encrypter",
		Hidden: true,
		Rows: []monitoring.Row{
			{
				func(owner monitoring.ObservableOwner) shared.Observable {
					return shared.Observable{
						Name:        "records_encrypted_at_rest_percentage",
						Description: "percentage of database records encrypted at rest",
						Query:       `(max(src_records_encrypted_at_rest_total) by (tableName)) / ((max(src_records_encrypted_at_rest_total) by (tableName)) + (max(src_records_unencrypted_at_rest_total) by (tableName))) * 100`,
						Panel:       monitoring.Panel().LegendFormat("{{tableName}}").Unit(monitoring.Percentage).Min(0).Max(100),
						Owner:       owner,
					}
				}(monitoring.ObservableOwnerSource).WithNoAlerts(`
					Percentage of encrypted database records
				`).Observable(),

				shared.Standard.Count("records encrypted")(shared.ObservableConstructorOptions{
					MetricNameRoot:        "records_encrypted",
					MetricDescriptionRoot: "database",
					By:                    []string{"tableName"},
				})(containerName, monitoring.ObservableOwnerSource).WithNoAlerts(`
					Number of encrypted database records every 5m
				`).Observable(),

				shared.Standard.Count("records decrypted")(shared.ObservableConstructorOptions{
					MetricNameRoot:        "records_decrypted",
					MetricDescriptionRoot: "database",
					By:                    []string{"tableName"},
				})(containerName, monitoring.ObservableOwnerSource).WithNoAlerts(`
					Number of encrypted database records every 5m
				`).Observable(),

				shared.Observation.Errors(shared.ObservableConstructorOptions{
					MetricNameRoot:        "record_encryption",
					MetricDescriptionRoot: "encryption",
				})(containerName, monitoring.ObservableOwnerSource).WithNoAlerts(`
					Number of database record encryption/decryption errors every 5m
				`).Observable(),
			},
		},
	}

	return &monitoring.Dashboard{
		Name:        "worker",
		Title:       "Worker",
		Description: "Manages background processes.",
		Variables: []monitoring.ContainerVariable{
			{
				Label: "Instance",
				Name:  "instance",
				OptionsLabelValues: monitoring.ContainerVariableOptionsLabelValues{
					Query:         "src_worker_jobs",
					LabelName:     "instance",
					ExampleOption: "worker:6089",
				},
				Multi: true,
			},
		},
		Groups: []monitoring.Group{
			// src_worker_jobs
			activeJobsGroup,

			// src_records_encrypted_at_rest_total
			// src_records_unencrypted_at_rest_total
			// src_records_encrypted_total
			// src_records_decrypted_total
			// src_record_encryption_errors_total
			recordEncrypterGroup,

			shared.CodeIntelligence.NewCommitGraphQueueGroup(containerName),
			shared.CodeIntelligence.NewCommitGraphProcessorGroup(containerName),
			shared.CodeIntelligence.NewDependencyIndexQueueGroup(containerName),
			shared.CodeIntelligence.NewDependencyIndexProcessorGroup(containerName),
			shared.CodeIntelligence.NewIndexSchedulerGroup(containerName),
			shared.CodeIntelligence.NewDBStoreGroup(containerName),
			shared.CodeIntelligence.NewLSIFStoreGroup(containerName),
			shared.CodeIntelligence.NewDependencyIndexDBWorkerStoreGroup(containerName),
			shared.CodeIntelligence.NewGitserverClientGroup(containerName),
			shared.CodeIntelligence.NewDependencyReposStoreGroup(containerName),

			repoPermsSyncerGroup(monitoring.ObservableOwnerSource),

			shared.GitServer.NewClientGroup(containerName),

			shared.Batches.NewDBStoreGroup(containerName),
			shared.Batches.NewServiceGroup(containerName),
			shared.Batches.NewBatchSpecResolutionDBWorkerStoreGroup(containerName),
			shared.Batches.NewBulkOperationDBWorkerStoreGroup(containerName),
			shared.Batches.NewReconcilerDBWorkerStoreGroup(containerName),
			// This is for the resetter only here, the queue is running in the frontend
			// through executorqueue.
			shared.Batches.NewWorkspaceExecutionDBWorkerStoreGroup(containerName),
			shared.Batches.NewExecutorQueueGroup(),

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
						MetricDescriptionRoot: "lsif upload",
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
						MetricDescriptionRoot: "lsif index",
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
						MetricDescriptionRoot: "lsif dependency index",
					},
				},

				RecordResets:        shared.NoAlertsOption("none"),
				RecordResetFailures: shared.NoAlertsOption("none"),
				Errors:              shared.NoAlertsOption("none"),
			}),
			shared.CodeInsights.NewInsightsQueryRunnerQueueGroup(containerName),
			shared.CodeInsights.NewInsightsQueryRunnerWorkerGroup(containerName),
			shared.CodeInsights.NewInsightsQueryRunnerResetterGroup(containerName),
			shared.CodeInsights.NewInsightsQueryRunnerStoreGroup(containerName),
			{
				Title:  "Code Insights queue utilization",
				Hidden: true,
				Rows: []monitoring.Row{{monitoring.Observable{
					Name:           "insights_queue_unutilized_size",
					Description:    "insights queue size that is not utilized (not processing)",
					Owner:          monitoring.ObservableOwnerCodeInsights,
					Query:          fmt.Sprintf("max(src_query_runner_worker_total{job=~%q}) > 0 and on(job) sum by (op)(increase(src_workerutil_dbworker_store_insights_query_runner_jobs_store_total{job=~%q,op=\"Dequeue\"}[5m])) < 1", scrapeJobRegex, scrapeJobRegex),
					DataMustExist:  false,
					Warning:        monitoring.Alert().Greater(0.0).For(time.Minute * 30),
					NextSteps:      "Verify code insights worker job has successfully started. Restart worker service and monitoring startup logs, looking for worker panics.",
					Interpretation: "Any value on this panel indicates code insights is not processing queries from its queue. This observable and alert only fire if there are records in the queue and there have been no dequeue attempts for 30 minutes.",
					Panel:          monitoring.Panel().LegendFormat("count"),
				}}},
			},

			// Resource monitoring
			shared.NewDatabaseConnectionsMonitoringGroup(containerName, monitoring.ObservableOwnerInfraOrg),
			shared.NewContainerMonitoringGroup(containerName, monitoring.ObservableOwnerCodeIntel, nil),
			shared.NewProvisioningIndicatorsGroup(containerName, monitoring.ObservableOwnerCodeIntel, nil),
			shared.NewGolangMonitoringGroup(containerName, monitoring.ObservableOwnerCodeIntel, nil),
			shared.NewKubernetesMonitoringGroup(containerName, monitoring.ObservableOwnerCodeIntel, nil),

			// Sourcegraph Own background jobs
			shared.SourcegraphOwn.NewOwnRepoIndexerStoreGroup(containerName),
			shared.SourcegraphOwn.NewOwnRepoIndexerWorkerGroup(containerName),
			shared.SourcegraphOwn.NewOwnRepoIndexerResetterGroup(containerName),
			shared.SourcegraphOwn.NewOwnRepoIndexerSchedulerGroup(containerName),

			shared.NewSiteConfigurationClientMetricsGroup(shared.SiteConfigurationMetricsOptions{
				HumanServiceName:    "worker",
				InstanceFilterRegex: `${instance:regex}`,
				JobFilterRegex:      scrapeJobRegex,
			}, monitoring.ObservableOwnerInfraOrg),
		},
	}
}

func repoPermsSyncerGroup(owner monitoring.ObservableOwner) monitoring.Group {
	return monitoring.Group{
		Title:  "Permissions",
		Hidden: true,
		Rows: []monitoring.Row{
			{
				{
					Name:           "user_success_syncs_total",
					Description:    "total number of user permissions syncs",
					Query:          `sum(src_repo_perms_syncer_success_syncs{type="user"})`,
					Panel:          monitoring.Panel().LegendFormat("{{type}}").Unit(monitoring.Number),
					Owner:          owner,
					NoAlert:        true,
					Interpretation: "Indicates the total number of user permissions sync completed.",
				},
				{
					Name:           "user_success_syncs",
					Description:    "number of user permissions syncs [5m]",
					Query:          `sum(increase(src_repo_perms_syncer_success_syncs{type="user"}[5m]))`,
					Panel:          monitoring.Panel().LegendFormat("{{type}}").Unit(monitoring.Number),
					Owner:          owner,
					NoAlert:        true,
					Interpretation: "Indicates the number of users permissions syncs completed.",
				},
				{
					Name:           "user_initial_syncs",
					Description:    "number of first user permissions syncs [5m]",
					Query:          `sum(increase(src_repo_perms_syncer_initial_syncs{type="user"}[5m]))`,
					Panel:          monitoring.Panel().LegendFormat("{{type}}").Unit(monitoring.Number),
					Owner:          owner,
					NoAlert:        true,
					Interpretation: "Indicates the number of permissions syncs done for the first time for the user.",
				},
			},
			{

				{
					Name:           "repo_success_syncs_total",
					Description:    "total number of repo permissions syncs",
					Query:          `sum(src_repo_perms_syncer_success_syncs{type="repo"})`,
					Panel:          monitoring.Panel().LegendFormat("{{type}}").Unit(monitoring.Number),
					Owner:          owner,
					NoAlert:        true,
					Interpretation: "Indicates the total number of repo permissions sync completed.",
				},
				{
					Name:           "repo_success_syncs",
					Description:    "number of repo permissions syncs over 5m",
					Query:          `sum(increase(src_repo_perms_syncer_success_syncs{type="repo"}[5m]))`,
					Panel:          monitoring.Panel().LegendFormat("{{type}}").Unit(monitoring.Number),
					Owner:          owner,
					NoAlert:        true,
					Interpretation: "Indicates the number of repos permissions syncs completed.",
				},
				{
					Name:           "repo_initial_syncs",
					Description:    "number of first repo permissions syncs over 5m",
					Query:          `sum(increase(src_repo_perms_syncer_initial_syncs{type="repo"}[5m]))`,
					Panel:          monitoring.Panel().LegendFormat("{{type}}").Unit(monitoring.Number),
					Owner:          owner,
					NoAlert:        true,
					Interpretation: "Indicates the number of permissions syncs done for the first time for the repo.",
				},
			},
			{
				{
					Name:           "users_consecutive_sync_delay",
					Description:    "max duration between two consecutive permissions sync for user",
					Query:          `max(max_over_time (src_repo_perms_syncer_perms_consecutive_sync_delay{type="user"} [1m]))`,
					Panel:          monitoring.Panel().LegendFormat("seconds").Unit(monitoring.Seconds),
					Owner:          owner,
					NoAlert:        true,
					Interpretation: "Indicates the max delay between two consecutive permissions sync for a user during the period.",
				},
				{
					Name:           "repos_consecutive_sync_delay",
					Description:    "max duration between two consecutive permissions sync for repo",
					Query:          `max(max_over_time (src_repo_perms_syncer_perms_consecutive_sync_delay{type="repo"} [1m]))`,
					Panel:          monitoring.Panel().LegendFormat("seconds").Unit(monitoring.Seconds),
					Owner:          owner,
					NoAlert:        true,
					Interpretation: "Indicates the max delay between two consecutive permissions sync for a repo during the period.",
				},
			},
			{
				{
					Name:           "users_first_sync_delay",
					Description:    "max duration between user creation and first permissions sync",
					Query:          `max(max_over_time(src_repo_perms_syncer_perms_first_sync_delay{type="user"}[1m]))`,
					Panel:          monitoring.Panel().LegendFormat("seconds").Unit(monitoring.Seconds),
					Owner:          owner,
					NoAlert:        true,
					Interpretation: "Indicates the max delay between user creation and their permissions sync",
				},
				{
					Name:           "repos_first_sync_delay",
					Description:    "max duration between repo creation and first permissions sync over 1m",
					Query:          `max(max_over_time(src_repo_perms_syncer_perms_first_sync_delay{type="repo"}[1m]))`,
					Panel:          monitoring.Panel().LegendFormat("seconds").Unit(monitoring.Seconds),
					Owner:          owner,
					NoAlert:        true,
					Interpretation: "Indicates the max delay between repo creation and their permissions sync",
				},
			},
			{
				{
					Name:           "permissions_found_count",
					Description:    "number of permissions found during user/repo permissions sync",
					Query:          `sum by (type) (src_repo_perms_syncer_perms_found)`,
					Panel:          monitoring.Panel().LegendFormat("{{type}}").Unit(monitoring.Number),
					Owner:          owner,
					NoAlert:        true,
					Interpretation: "Indicates the number permissions found during users/repos permissions sync.",
				},
				{
					Name:           "permissions_found_avg",
					Description:    "average number of permissions found during permissions sync per user/repo",
					Query:          `avg by (type) (src_repo_perms_syncer_perms_found)`,
					Panel:          monitoring.Panel().LegendFormat("{{type}}").Unit(monitoring.Number),
					Owner:          owner,
					NoAlert:        true,
					Interpretation: "Indicates the average number permissions found during permissions sync per user/repo.",
				},
			},
			{
				{
					Name:        "perms_syncer_outdated_perms",
					Description: "number of entities with outdated permissions",
					Query:       `max by (type) (src_repo_perms_syncer_outdated_perms)`,
					Warning:     monitoring.Alert().GreaterOrEqual(100).For(5 * time.Minute),
					Panel:       monitoring.Panel().LegendFormat("{{type}}").Unit(monitoring.Number),
					Owner:       owner,
					NextSteps: `
						- **Enabled permissions for the first time:** Wait for few minutes and see if the number goes down.
						- **Otherwise:** Increase the API rate limit to [GitHub](https://sourcegraph.com/docs/admin/code_hosts/github#github-com-rate-limits), [GitLab](https://sourcegraph.com/docs/admin/code_hosts/gitlab#internal-rate-limits) or [Bitbucket Server](https://sourcegraph.com/docs/admin/code_hosts/bitbucket_server#internal-rate-limits).
					`,
				},
			},
			{
				{
					Name:        "perms_syncer_sync_duration",
					Description: "95th permissions sync duration",
					Query:       `histogram_quantile(0.95, max by (le, type) (rate(src_repo_perms_syncer_sync_duration_seconds_bucket[1m])))`,
					Warning:     monitoring.Alert().GreaterOrEqual(30).For(5 * time.Minute),
					Panel:       monitoring.Panel().LegendFormat("{{type}}").Unit(monitoring.Seconds),
					Owner:       owner,
					NextSteps:   "Check the network latency is reasonable (<50ms) between the Sourcegraph and the code host.",
				},
			},
			{
				{
					Name:        "perms_syncer_sync_errors",
					Description: "permissions sync error rate",
					Query:       `max by (type) (ceil(rate(src_repo_perms_syncer_sync_errors_total[1m])))`,
					Critical:    monitoring.Alert().GreaterOrEqual(1).For(time.Minute),
					Panel:       monitoring.Panel().LegendFormat("{{type}}").Unit(monitoring.Number),
					Owner:       owner,
					NextSteps: `
						- Check the network connectivity the Sourcegraph and the code host.
						- Check if API rate limit quota is exhausted on the code host.
					`,
				},
				{
					Name:        "perms_syncer_scheduled_repos_total",
					Description: "total number of repos scheduled for permissions sync",
					Query:       `max(rate(src_repo_perms_syncer_schedule_repos_total[1m]))`,
					NoAlert:     true,
					Panel:       monitoring.Panel().Unit(monitoring.Number),
					Owner:       owner,
					Interpretation: `
						Indicates how many repositories have been scheduled for a permissions sync.
						More about repository permissions synchronization [here](https://sourcegraph.com/docs/admin/permissions/syncing#scheduling)
					`,
				},
			},
		},
	}
}
