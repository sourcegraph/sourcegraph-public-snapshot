package definitions

import (
	"fmt"
	"time"

	"github.com/sourcegraph/sourcegraph/monitoring/monitoring"
)

// This is set a bit longer than maxSyncInterval in internal/repos/syncer.go
const syncDurationThreshold = 9 * time.Hour

func RepoUpdater() *monitoring.Container {
	return &monitoring.Container{
		Name:        "repo-updater",
		Title:       "Repo Updater",
		Description: "Manages interaction with code hosts, instructs Gitserver to update repositories.",
		Groups: []monitoring.Group{
			{
				Title: "General",
				Rows: []monitoring.Row{
					{
						sharedFrontendInternalAPIErrorResponses("repo-updater", monitoring.ObservableOwnerCloud),
					},
				},
			},
			{
				Title: "Repositories",
				Rows: []monitoring.Row{
					{
						{
							Name:              "syncer_sync_last_time",
							Description:       "time since last sync",
							Query:             `max(timestamp(vector(time()))) - max(src_repoupdater_syncer_sync_last_time)`,
							DataMayNotExist:   true,
							NoAlert:           true,
							PanelOptions:      monitoring.PanelOptions().Unit(monitoring.Seconds),
							Owner:             monitoring.ObservableOwnerCloud,
							PossibleSolutions: "Make sure there are external services added with valid tokens",
						},
						{
							Name:            "src_repoupdater_max_sync_backoff",
							Description:     "time since oldest sync",
							Query:           `max(src_repoupdater_max_sync_backoff)`,
							DataMayNotExist: true,
							Critical:        monitoring.Alert().GreaterOrEqual(syncDurationThreshold.Seconds()).For(10 * time.Minute),
							PanelOptions:    monitoring.PanelOptions().Unit(monitoring.Seconds),
							Owner:           monitoring.ObservableOwnerCloud,
							PossibleSolutions: fmt.Sprintf(`
								An alert here indicates that no code host connections have synced in at least %v. This indicates that there could be a configuration issue
								with your code hosts connections or networking issues affecting communication with your code hosts.
								- Check the code host status indicator (cloud icon in top right of Sourcegraph homepage) for errors.
								- Make sure external services do not have invalid tokens by navigating to them in the web UI and clicking save. If there are no errors, they are valid.
								- Check the repo-updater logs for errors about syncing.
								- Confirm that outbound network connections are allowed where repo-updater is deployed.
								- Check back in an hour to see if the issue has resolved itself.
							`, syncDurationThreshold),
						},
						{
							Name:            "src_repoupdater_syncer_sync_errors_total",
							Description:     "sync error rate",
							Query:           `sum by (family) (rate(src_repoupdater_syncer_sync_errors_total[5m]))`,
							DataMayNotExist: true,
							Critical:        monitoring.Alert().GreaterOrEqual(0.01).For(10 * time.Minute),
							PanelOptions:    monitoring.PanelOptions().Unit(monitoring.Number),
							Owner:           monitoring.ObservableOwnerCloud,
							PossibleSolutions: `
								An alert here indicates errors syncing repo metadata with code hosts. This indicates that there could be a configuration issue
								with your code hosts connections or networking issues affecting communication with your code hosts.
								- Check the code host status indicator (cloud icon in top right of Sourcegraph homepage) for errors.
								- Make sure external services do not have invalid tokens by navigating to them in the web UI and clicking save. If there are no errors, they are valid.
								- Check the repo-updater logs for errors about syncing.
								- Confirm that outbound network connections are allowed where repo-updater is deployed.
								- Check back in an hour to see if the issue has resolved itself.
							`,
						},
					},
					{
						{
							Name:              "syncer_sync_start",
							Description:       "sync was started",
							Query:             `sum by (family) (rate(src_repoupdater_syncer_start_sync[5m]))`,
							DataMayNotExist:   true,
							Warning:           monitoring.Alert().LessOrEqual(0).For(syncDurationThreshold),
							PanelOptions:      monitoring.PanelOptions().LegendFormat("{{family}}").Unit(monitoring.Number),
							Owner:             monitoring.ObservableOwnerCloud,
							PossibleSolutions: "Check repo-updater logs for errors.",
						},
						{
							Name:              "syncer_sync_duration",
							Description:       "95th repositories sync duration",
							Query:             `histogram_quantile(0.95, sum by (le, family, success) (rate(src_repoupdater_syncer_sync_duration_seconds_bucket[1m])))`,
							DataMayNotExist:   true,
							Warning:           monitoring.Alert().GreaterOrEqual(30).For(5 * time.Minute),
							PanelOptions:      monitoring.PanelOptions().LegendFormat("{{family}}-{{success}}").Unit(monitoring.Seconds),
							Owner:             monitoring.ObservableOwnerCloud,
							PossibleSolutions: "Check the network latency is reasonable (<50ms) between the Sourcegraph and the code host",
						},
						{
							Name:              "source_duration",
							Description:       "95th repositories source duration",
							Query:             `histogram_quantile(0.95, sum by (le) (rate(src_repoupdater_source_duration_seconds_bucket[1m])))`,
							DataMayNotExist:   true,
							Warning:           monitoring.Alert().GreaterOrEqual(30).For(5 * time.Minute),
							PanelOptions:      monitoring.PanelOptions().Unit(monitoring.Seconds),
							Owner:             monitoring.ObservableOwnerCloud,
							PossibleSolutions: "Check the network latency is reasonable (<50ms) between the Sourcegraph and the code host",
						},
					},
					{
						{
							Name:              "syncer_synced_repos",
							Description:       "repositories synced",
							Query:             `sum by (state) (rate(src_repoupdater_syncer_synced_repos_total[1m]))`,
							DataMayNotExist:   true,
							Warning:           monitoring.Alert().LessOrEqual(0).For(syncDurationThreshold),
							PanelOptions:      monitoring.PanelOptions().LegendFormat("{{state}}").Unit(monitoring.Number),
							Owner:             monitoring.ObservableOwnerCloud,
							PossibleSolutions: "Check network connectivity to code hosts",
						},
						{
							Name:              "sourced_repos",
							Description:       "repositories sourced",
							Query:             `sum(rate(src_repoupdater_source_repos_total[1m]))`,
							DataMayNotExist:   true,
							Warning:           monitoring.Alert().LessOrEqual(0).For(syncDurationThreshold),
							PanelOptions:      monitoring.PanelOptions().Unit(monitoring.Number),
							Owner:             monitoring.ObservableOwnerCloud,
							PossibleSolutions: "Check network connectivity to code hosts",
						},
						{
							Name:            "user_added_repos",
							Description:     "total number of user added repos",
							Query:           `sum(src_repoupdater_user_repos_total)`,
							DataMayNotExist: true,
							// 90% of our enforced limit
							Critical:          monitoring.Alert().GreaterOrEqual(200000 * 0.9).For(5 * time.Minute),
							PanelOptions:      monitoring.PanelOptions().Unit(monitoring.Number),
							Owner:             monitoring.ObservableOwnerCloud,
							PossibleSolutions: "Check for unusual spikes in user added repos. Each user is only allowed to add 2000",
						},
					},
					{
						{
							Name:              "purge_failed",
							Description:       "repositories purge failed",
							Query:             `sum(rate(src_repoupdater_purge_failed[1m]))`,
							DataMayNotExist:   true,
							Warning:           monitoring.Alert().GreaterOrEqual(0).For(5 * time.Minute),
							PanelOptions:      monitoring.PanelOptions().Unit(monitoring.Number),
							Owner:             monitoring.ObservableOwnerCloud,
							PossibleSolutions: "Check repo-updater's connectivity with gitserver and gitserver logs",
						},
					},
					{
						{
							Name:              "sched_auto_fetch",
							Description:       "repositories scheduled due to hitting a deadline",
							Query:             `sum(rate(src_repoupdater_sched_auto_fetch[1m]))`,
							DataMayNotExist:   true,
							Warning:           monitoring.Alert().LessOrEqual(0).For(syncDurationThreshold),
							PanelOptions:      monitoring.PanelOptions().Unit(monitoring.Number),
							Owner:             monitoring.ObservableOwnerCloud,
							PossibleSolutions: "Check repo-updater logs. This is expected to fire if there are no user added code hosts",
						},
						{
							Name:              "sched_manual_fetch",
							Description:       "repositories scheduled due to user traffic",
							Query:             `sum(rate(src_repoupdater_sched_manual_fetch[1m]))`,
							NoAlert:           true,
							DataMayNotExist:   true,
							PanelOptions:      monitoring.PanelOptions().Unit(monitoring.Number),
							Owner:             monitoring.ObservableOwnerCloud,
							PossibleSolutions: "Check repo-updater logs. This is expected to fire if there are no user added code hosts",
						},
					},
					{
						{
							Name:              "sched_known_repos",
							Description:       "repositories managed by the scheduler",
							Query:             `sum(src_repoupdater_sched_known_repos)`,
							DataMayNotExist:   true,
							Warning:           monitoring.Alert().LessOrEqual(0).For(10 * time.Minute),
							PanelOptions:      monitoring.PanelOptions().Unit(monitoring.Number),
							Owner:             monitoring.ObservableOwnerCloud,
							PossibleSolutions: "Check repo-updater logs. This is expected to fire if there are no user added code hosts",
						},
						{
							Name:            "sched_update_queue_length",
							Description:     "rate of growth of update queue length over 5 minutes",
							Query:           `max(deriv(src_repoupdater_sched_update_queue_length[5m]))`,
							DataMayNotExist: true,
							// Alert if the derivative is positive for longer than 30 minutes
							Critical:          monitoring.Alert().GreaterOrEqual(0).For(30 * time.Minute),
							PanelOptions:      monitoring.PanelOptions().Unit(monitoring.Number),
							Owner:             monitoring.ObservableOwnerCloud,
							PossibleSolutions: "Check repo-updater logs for indications that the queue is not being processed. The queue length should trend downwards over time as items are sent to GitServer",
						},
						{
							Name:              "sched_loops",
							Description:       "scheduler loops",
							Query:             `sum(rate(src_repoupdater_sched_loops[1m]))`,
							DataMayNotExist:   true,
							Warning:           monitoring.Alert().LessOrEqual(0).For(syncDurationThreshold),
							PanelOptions:      monitoring.PanelOptions().Unit(monitoring.Number),
							Owner:             monitoring.ObservableOwnerCloud,
							PossibleSolutions: "Check repo-updater logs for errors. This is expected to fire if there are no user added code hosts",
						},
					},
					{
						{
							Name:              "sched_error",
							Description:       "repositories schedule error rate",
							Query:             `sum(rate(src_repoupdater_sched_error[1m]))`,
							DataMayNotExist:   true,
							Critical:          monitoring.Alert().GreaterOrEqual(1).For(time.Minute),
							PanelOptions:      monitoring.PanelOptions().Unit(monitoring.Number),
							Owner:             monitoring.ObservableOwnerCloud,
							PossibleSolutions: "Check repo-updater logs for errors",
						},
					},
				},
			},
			{
				Title:  "Permissions",
				Hidden: true,
				Rows: []monitoring.Row{
					{
						{
							Name:              "perms_syncer_perms",
							Description:       "time gap between least and most up to date permissions",
							Query:             `sum by (type) (src_repoupdater_perms_syncer_perms_gap_seconds)`,
							DataMayNotExist:   true,
							Warning:           monitoring.Alert().GreaterOrEqual((3 * 24 * time.Hour).Seconds()).For(5 * time.Minute), // 3 days
							PanelOptions:      monitoring.PanelOptions().LegendFormat("{{type}}").Unit(monitoring.Seconds),
							Owner:             monitoring.ObservableOwnerCloud,
							PossibleSolutions: "Increase the API rate limit to [GitHub](https://docs.sourcegraph.com/admin/external_service/github#github-com-rate-limits), [GitLab](https://docs.sourcegraph.com/admin/external_service/gitlab#internal-rate-limits) or [Bitbucket Server](https://docs.sourcegraph.com/admin/external_service/bitbucket_server#internal-rate-limits).",
						},
						{
							Name:              "perms_syncer_stale_perms",
							Description:       "number of entities with stale permissions",
							Query:             `sum by (type) (src_repoupdater_perms_syncer_stale_perms)`,
							DataMayNotExist:   true,
							Warning:           monitoring.Alert().GreaterOrEqual(100).For(5 * time.Minute),
							PanelOptions:      monitoring.PanelOptions().LegendFormat("{{type}}").Unit(monitoring.Number),
							Owner:             monitoring.ObservableOwnerCloud,
							PossibleSolutions: "Increase the API rate limit to [GitHub](https://docs.sourcegraph.com/admin/external_service/github#github-com-rate-limits), [GitLab](https://docs.sourcegraph.com/admin/external_service/gitlab#internal-rate-limits) or [Bitbucket Server](https://docs.sourcegraph.com/admin/external_service/bitbucket_server#internal-rate-limits).",
						},
						{
							Name:            "perms_syncer_no_perms",
							Description:     "number of entities with no permissions",
							Query:           `sum by (type) (src_repoupdater_perms_syncer_no_perms)`,
							DataMayNotExist: true,
							Warning:         monitoring.Alert().GreaterOrEqual(100).For(5 * time.Minute),
							PanelOptions:    monitoring.PanelOptions().LegendFormat("{{type}}").Unit(monitoring.Number),
							Owner:           monitoring.ObservableOwnerCloud,
							PossibleSolutions: `
								- **Enabled permissions for the first time:** Wait for few minutes and see if the number goes down.
								- **Otherwise:** Increase the API rate limit to [GitHub](https://docs.sourcegraph.com/admin/external_service/github#github-com-rate-limits), [GitLab](https://docs.sourcegraph.com/admin/external_service/gitlab#internal-rate-limits) or [Bitbucket Server](https://docs.sourcegraph.com/admin/external_service/bitbucket_server#internal-rate-limits).
							`,
						},
					},
					{
						{
							Name:              "perms_syncer_sync_duration",
							Description:       "95th permissions sync duration",
							Query:             `histogram_quantile(0.95, sum by (le, type) (rate(src_repoupdater_perms_syncer_sync_duration_seconds_bucket[1m])))`,
							DataMayNotExist:   true,
							Warning:           monitoring.Alert().GreaterOrEqual(30).For(5 * time.Minute),
							PanelOptions:      monitoring.PanelOptions().LegendFormat("{{type}}").Unit(monitoring.Seconds),
							Owner:             monitoring.ObservableOwnerCloud,
							PossibleSolutions: "Check the network latency is reasonable (<50ms) between the Sourcegraph and the code host.",
						},
						{
							Name:            "perms_syncer_queue_size",
							Description:     "permissions sync queued items",
							Query:           `sum(src_repoupdater_perms_syncer_queue_size)`,
							DataMayNotExist: true,
							Warning:         monitoring.Alert().GreaterOrEqual(100).For(5 * time.Minute),
							PanelOptions:    monitoring.PanelOptions().Unit(monitoring.Number),
							Owner:           monitoring.ObservableOwnerCloud,
							PossibleSolutions: `
								- **Enabled permissions for the first time:** Wait for few minutes and see if the number goes down.
								- **Otherwise:** Increase the API rate limit to [GitHub](https://docs.sourcegraph.com/admin/external_service/github#github-com-rate-limits), [GitLab](https://docs.sourcegraph.com/admin/external_service/gitlab#internal-rate-limits) or [Bitbucket Server](https://docs.sourcegraph.com/admin/external_service/bitbucket_server#internal-rate-limits).
							`,
						},
					},
					{
						{
							Name:            "perms_syncer_sync_errors",
							Description:     "permissions sync error rate",
							Query:           `sum by (type) (rate(src_repoupdater_perms_syncer_sync_errors_total[1m])) / sum by (type) (rate(src_repoupdater_perms_syncer_sync_duration_seconds_count[1m]))`,
							DataMayNotExist: true,
							Critical:        monitoring.Alert().GreaterOrEqual(1).For(time.Minute),
							PanelOptions:    monitoring.PanelOptions().LegendFormat("{{type}}").Unit(monitoring.Number),
							Owner:           monitoring.ObservableOwnerCloud,
							PossibleSolutions: `
								- Check the network connectivity the Sourcegraph and the code host.
								- Check if API rate limit quota is exhausted on the code host.
							`,
						},
					},
				},
			},
			{
				Title:  "External services",
				Hidden: true,
				Rows: []monitoring.Row{
					{
						{
							Name:              "src_repoupdater_external_services_total",
							Description:       "the total number of external services",
							Query:             `sum(src_repoupdater_external_services_total)`,
							DataMayNotExist:   true,
							Critical:          monitoring.Alert().GreaterOrEqual(20000).For(1 * time.Hour),
							PanelOptions:      monitoring.PanelOptions().Unit(monitoring.Number),
							Owner:             monitoring.ObservableOwnerCloud,
							PossibleSolutions: "Check for spikes in external services, could be abuse",
						},
						{
							Name:              "src_repoupdater_user_external_services_total",
							Description:       "the total number of user added external services",
							Query:             `sum(src_repoupdater_user_external_services_total)`,
							DataMayNotExist:   true,
							Warning:           monitoring.Alert().GreaterOrEqual(20000).For(1 * time.Hour),
							PanelOptions:      monitoring.PanelOptions().Unit(monitoring.Number),
							Owner:             monitoring.ObservableOwnerCloud,
							PossibleSolutions: "Check for spikes in external services, could be abuse",
						},
					},
					{
						{
							Name:            "repoupdater_queued_sync_jobs_total",
							Description:     "the total number of queued sync jobs",
							Query:           `sum(src_repoupdater_queued_sync_jobs_total)`,
							DataMayNotExist: true,
							Warning:         monitoring.Alert().GreaterOrEqual(100).For(1 * time.Hour),
							PanelOptions:    monitoring.PanelOptions().Unit(monitoring.Number),
							Owner:           monitoring.ObservableOwnerCloud,
							PossibleSolutions: `
								- **Check if jobs are failing to sync:** "SELECT * FROM external_service_sync_jobs WHERE state = 'errored'";
								- **Increase the number of workers** using the 'repoConcurrentExternalServiceSyncers' site config.
							`,
						},
						{
							Name:              "repoupdater_completed_sync_jobs_total",
							Description:       "the total number of completed sync jobs",
							Query:             `sum(src_repoupdater_completed_sync_jobs_total)`,
							DataMayNotExist:   true,
							Warning:           monitoring.Alert().GreaterOrEqual(100000).For(1 * time.Hour),
							PanelOptions:      monitoring.PanelOptions().Unit(monitoring.Number),
							Owner:             monitoring.ObservableOwnerCloud,
							PossibleSolutions: "Check repo-updater logs. Jobs older than 1 day should have been removed.",
						},
						{
							Name:              "repoupdater_errored_sync_jobs_total",
							Description:       "the total number of errored sync jobs",
							Query:             `sum(src_repoupdater_errored_sync_jobs_total)`,
							DataMayNotExist:   true,
							Warning:           monitoring.Alert().GreaterOrEqual(100).For(1 * time.Hour),
							PanelOptions:      monitoring.PanelOptions().Unit(monitoring.Number),
							Owner:             monitoring.ObservableOwnerCloud,
							PossibleSolutions: "Check repo-updater logs. Check code host connectivity",
						},
					},
				},
			},
			{
				Title:  "Container monitoring (not available on server)",
				Hidden: true,
				Rows: []monitoring.Row{
					{
						sharedContainerCPUUsage("repo-updater", monitoring.ObservableOwnerCloud),
						sharedContainerMemoryUsage("repo-updater", monitoring.ObservableOwnerCloud),
					},
					{
						sharedContainerRestarts("repo-updater", monitoring.ObservableOwnerCloud),
						sharedContainerFsInodes("repo-updater", monitoring.ObservableOwnerCloud),
					},
				},
			},
			{
				Title:  "Provisioning indicators (not available on server)",
				Hidden: true,
				Rows: []monitoring.Row{
					{
						sharedProvisioningCPUUsageLongTerm("repo-updater", monitoring.ObservableOwnerCloud),
						sharedProvisioningMemoryUsageLongTerm("repo-updater", monitoring.ObservableOwnerCloud),
					},
					{
						sharedProvisioningCPUUsageShortTerm("repo-updater", monitoring.ObservableOwnerCloud),
						sharedProvisioningMemoryUsageShortTerm("repo-updater", monitoring.ObservableOwnerCloud),
					},
				},
			},
			{
				Title:  "Golang runtime monitoring",
				Hidden: true,
				Rows: []monitoring.Row{
					{
						sharedGoGoroutines("repo-updater", monitoring.ObservableOwnerCloud),
						sharedGoGcDuration("repo-updater", monitoring.ObservableOwnerCloud),
					},
				},
			},
			{
				Title:  "Kubernetes monitoring (ignore if using Docker Compose or server)",
				Hidden: true,
				Rows: []monitoring.Row{
					{
						sharedKubernetesPodsAvailable("repo-updater", monitoring.ObservableOwnerCloud),
					},
				},
			},
		},
	}
}
