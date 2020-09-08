package main

import (
	"time"
)

func RepoUpdater() *Container {
	return &Container{
		Name:        "repo-updater",
		Title:       "Repo Updater",
		Description: "Manages interaction with code hosts, instructs Gitserver to update repositories.",
		Groups: []Group{
			{
				Title: "General",
				Rows: []Row{
					{
						sharedFrontendInternalAPIErrorResponses("repo-updater", ObservableOwnerCloud),
					},
				},
			},
			{
				Title:  "Permissions",
				Hidden: true,
				Rows: []Row{
					{
						Observable{
							Name:              "perms_syncer_perms",
							Description:       "time gap between least and most up to date permissions",
							Query:             `src_repoupdater_perms_syncer_perms_gap_seconds`,
							DataMayNotExist:   true,
							Warning:           Alert{GreaterOrEqual: 3 * 24 * 60 * 60, For: 5 * time.Minute}, // 3 days
							PanelOptions:      PanelOptions().LegendFormat("{{type}}").Unit(Seconds),
							Owner:             ObservableOwnerCloud,
							PossibleSolutions: "Increase the API rate limit to the code host.",
						},
						Observable{
							Name:              "perms_syncer_stale_perms",
							Description:       "number of entities with stale permissions",
							Query:             `src_repoupdater_perms_syncer_stale_perms`,
							DataMayNotExist:   true,
							Warning:           Alert{GreaterOrEqual: 100, For: 5 * time.Minute},
							PanelOptions:      PanelOptions().LegendFormat("{{type}}").Unit(Number),
							Owner:             ObservableOwnerCloud,
							PossibleSolutions: "Increase the API rate limit to the code host.",
						},
						Observable{
							Name:            "perms_syncer_no_perms",
							Description:     "number of entities with no permissions",
							Query:           `src_repoupdater_perms_syncer_no_perms`,
							DataMayNotExist: true,
							Warning:         Alert{GreaterOrEqual: 100, For: 5 * time.Minute},
							PanelOptions:    PanelOptions().LegendFormat("{{type}}").Unit(Number),
							Owner:           ObservableOwnerCloud,
							PossibleSolutions: `
								- **Enabled permissions for the first time:** Wait for few minutes and see if the number goes down.
								- **Otherwise:** Increase the API rate limit to the code host.
							`,
						},
					},
					{
						Observable{
							Name:              "perms_syncer_sync_duration",
							Description:       "95th permissions sync duration",
							Query:             `histogram_quantile(0.95, rate(src_repoupdater_perms_syncer_sync_duration_seconds_bucket[1m]))`,
							DataMayNotExist:   true,
							Warning:           Alert{GreaterOrEqual: 30, For: 5 * time.Minute},
							PanelOptions:      PanelOptions().LegendFormat("{{type}}").Unit(Seconds),
							Owner:             ObservableOwnerCloud,
							PossibleSolutions: "Check the network latency is reasonable (<50ms) between the Sourcegraph and the code host.",
						},
						Observable{
							Name:            "perms_syncer_queue_size",
							Description:     "permissions sync queued items",
							Query:           `src_repoupdater_perms_syncer_queue_size`,
							DataMayNotExist: true,
							Warning:         Alert{GreaterOrEqual: 100, For: 5 * time.Minute},
							PanelOptions:    PanelOptions().Unit(Number),
							Owner:           ObservableOwnerCloud,
							PossibleSolutions: `
								- **Enabled permissions for the first time:** Wait for few minutes and see if the number goes down.
								- **Otherwise:** Increase the API rate limit to the code host.
							`,
						},
					},
					{
						Observable{
							Name:              "authz_filter_duration",
							Description:       "95th authorization duration",
							Query:             `histogram_quantile(0.95, rate(src_frontend_authz_filter_duration_seconds_bucket{success="true"}[1m]))`,
							DataMayNotExist:   true,
							Critical:          Alert{GreaterOrEqual: 1, For: time.Minute},
							PanelOptions:      PanelOptions().LegendFormat("seconds").Unit(Seconds),
							Owner:             ObservableOwnerCloud,
							PossibleSolutions: "Check if database is overloaded.",
						},
						Observable{
							Name:            "perms_syncer_sync_errors",
							Description:     "permissions sync error rate",
							Query:           `rate(src_repoupdater_perms_syncer_sync_errors_total[1m]) / rate(src_repoupdater_perms_syncer_sync_duration_seconds_count[1m])`,
							DataMayNotExist: true,
							Critical:        Alert{GreaterOrEqual: 1, For: time.Minute},
							PanelOptions:    PanelOptions().LegendFormat("{{type}}").Unit(Number),
							Owner:           ObservableOwnerCloud,
							PossibleSolutions: `
								- Check the network connectivity the Sourcegraph and the code host.
								- Check if API rate limit quota is exhausted.
							`,
						},
					},
				},
			},
			{
				Title:  "Container monitoring (not available on server)",
				Hidden: true,
				Rows: []Row{
					{
						sharedContainerCPUUsage("repo-updater", ObservableOwnerCloud),
						sharedContainerMemoryUsage("repo-updater", ObservableOwnerCloud),
					},
					{
						sharedContainerRestarts("repo-updater", ObservableOwnerCloud),
						sharedContainerFsInodes("repo-updater", ObservableOwnerCloud),
					},
				},
			},
			{
				Title:  "Provisioning indicators (not available on server)",
				Hidden: true,
				Rows: []Row{
					{
						sharedProvisioningCPUUsageLongTerm("repo-updater", ObservableOwnerCloud),
						sharedProvisioningMemoryUsageLongTerm("repo-updater", ObservableOwnerCloud),
					},
					{
						sharedProvisioningCPUUsageShortTerm("repo-updater", ObservableOwnerCloud),
						sharedProvisioningMemoryUsageShortTerm("repo-updater", ObservableOwnerCloud),
					},
				},
			},
			{
				Title:  "Golang runtime monitoring",
				Hidden: true,
				Rows: []Row{
					{
						sharedGoGoroutines("repo-updater", ObservableOwnerCloud),
						sharedGoGcDuration("repo-updater", ObservableOwnerCloud),
					},
				},
			},
			{
				Title:  "Kubernetes monitoring (ignore if using Docker Compose or server)",
				Hidden: true,
				Rows: []Row{
					{
						sharedKubernetesPodsAvailable("repo-updater", ObservableOwnerCloud),
					},
				},
			},
		},
	}
}
