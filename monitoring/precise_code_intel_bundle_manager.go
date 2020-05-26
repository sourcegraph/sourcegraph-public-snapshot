package main

func PreciseCodeIntelBundleManager() *Container {
	return &Container{
		Name:        "precise-code-intel-bundle-manager",
		Title:       "Precise Code Intel Bundle Manager",
		Description: "Stores and manages precise code intelligence bundles.",
		Groups: []Group{
			{
				Title: "General",
				Rows: []Row{
					{
						{
							Name:        "99th_percentile_bundle_database_duration",
							Description: "99th percentile successful bundle database query duration over 5m",
							// TODO(efritz) - ensure these exclude error durations
							Query:             `histogram_quantile(0.99, sum by (le)(rate(src_bundle_database_duration_seconds_bucket[5m])))`,
							DataMayNotExist:   true,
							DataMayBeNaN:      true,
							Warning:           Alert{GreaterOrEqual: 20},
							PanelOptions:      PanelOptions().LegendFormat("database operation").Unit(Seconds),
							PossibleSolutions: "none",
						},
						{
							Name:              "bundle_database_errors",
							Description:       "bundle database errors every 5m",
							Query:             `increase(src_bundle_database_errors_total[5m])`,
							DataMayNotExist:   true,
							Warning:           Alert{GreaterOrEqual: 20},
							PanelOptions:      PanelOptions().LegendFormat("database operation"),
							PossibleSolutions: "none",
						},
					},
					{
						{
							Name:        "99th_percentile_bundle_reader_duration",
							Description: "99th percentile successful bundle reader query duration over 5m",
							// TODO(efritz) - ensure these exclude error durations
							Query:             `histogram_quantile(0.99, sum by (le)(rate(src_bundle_reader_duration_seconds_bucket[5m])))`,
							DataMayNotExist:   true,
							DataMayBeNaN:      true,
							Warning:           Alert{GreaterOrEqual: 20},
							PanelOptions:      PanelOptions().LegendFormat("reader operation").Unit(Seconds),
							PossibleSolutions: "none",
						},
						{
							Name:              "bundle_reader_errors",
							Description:       "bundle reader errors every 5m",
							Query:             `increase(src_bundle_reader_errors_total[5m])`,
							DataMayNotExist:   true,
							Warning:           Alert{GreaterOrEqual: 20},
							PanelOptions:      PanelOptions().LegendFormat("reader operation"),
							PossibleSolutions: "none",
						},
					},
					{
						{
							Name:            "disk_space_remaining",
							Description:     "disk space remaining by instance",
							Query:           `(src_disk_space_available_bytes{job="precise-code-intel-bundle-manager"} / src_disk_space_total_bytes{job="precise-code-intel-bundle-manager"}) * 100`,
							DataMayNotExist: true,
							Warning:         Alert{LessOrEqual: 25},
							Critical:        Alert{LessOrEqual: 15},
							PanelOptions:    PanelOptions().LegendFormat("{{instance}}").Unit(Percentage),
							PossibleSolutions: `
								- **Provision more disk space:** Sourcegraph will begin deleting the oldest uploaded bundle files at 10% disk space remaining.
							`,
						},
					},
				},
			},
			{
				Title:  "Janitor - synchronizes database and filesystem and keeps free space on disk",
				Hidden: true,
				Rows: []Row{
					{
						{
							Name:              "janitor_errors",
							Description:       "janitor errors every 5m",
							Query:             `sum(increase(src_bundle_manager_janitor_errors_total[5m]))`,
							DataMayNotExist:   true,
							Warning:           Alert{GreaterOrEqual: 20},
							PanelOptions:      PanelOptions().LegendFormat("errors"),
							PossibleSolutions: "none",
						},
						{
							Name:              "janitor_old_uploads",
							Description:       "upload files removed (due to age) every 5m",
							Query:             `sum(increase(src_bundle_manager_janitor_upload_files_removed_total[5m]))`,
							DataMayNotExist:   true,
							Warning:           Alert{GreaterOrEqual: 20},
							PanelOptions:      PanelOptions().LegendFormat("files removed"),
							PossibleSolutions: "none",
						},
						{
							Name:              "janitor_old_dumps",
							Description:       "bundle files removed (due to low disk space) every 5m",
							Query:             `sum(increase(src_bundle_manager_janitor_evicted_bundle_files_removed_total[5m]))`,
							DataMayNotExist:   true,
							Warning:           Alert{GreaterOrEqual: 20},
							PanelOptions:      PanelOptions().LegendFormat("files removed"),
							PossibleSolutions: "none",
						},
					},
					{
						{
							Name:              "janitor_orphaned_dumps",
							Description:       "bundle files removed (with no corresponding database entry) every 5m",
							Query:             `sum(increase(src_bundle_manager_janitor_orphaned_bundle_files_removed_total[5m]))`,
							DataMayNotExist:   true,
							Warning:           Alert{GreaterOrEqual: 20},
							PanelOptions:      PanelOptions().LegendFormat("files removed"),
							PossibleSolutions: "none",
						},
						{
							Name:              "janitor_uploads_without_bundle_files",
							Description:       "upload records removed (with no corresponding bundle file) every 5m",
							Query:             `sum(increase(src_bundle_manager_janitor_upload_records_removed_total[5m]))`,
							DataMayNotExist:   true,
							Warning:           Alert{GreaterOrEqual: 20},
							PanelOptions:      PanelOptions().LegendFormat("records removed"),
							PossibleSolutions: "none",
						},
					},
				},
			},
			{
				Title:  "Connection and data cache",
				Hidden: true,
				Rows: []Row{
					{
						{
							Name:              "cache_utilization",
							Description:       "cache utilization",
							Query:             `max by (cache)((src_cache_cost{job="precise-code-intel-bundle-manager"} / src_cache_capacity{job="precise-code-intel-bundle-manager"}) * 100)`,
							DataMayNotExist:   true,
							DataMayBeNaN:      true,
							Warning:           Alert{GreaterOrEqual: 110},
							PanelOptions:      PanelOptions().LegendFormat("{{cache}}").Unit(Percentage),
							PossibleSolutions: "none",
						},
						{
							Name:              "cache_miss_percentage",
							Description:       "percentage of cache misses over all cache activity every 5m",
							Query:             `max by (cache)((increase(src_cache_misses_total{job="precise-code-intel-bundle-manager"}[5m]) / (increase(src_cache_hits_total{job="precise-code-intel-bundle-manager"}[5m]) + increase(src_cache_misses_total{job="precise-code-intel-bundle-manager"}[5m]))) * 100)`,
							DataMayNotExist:   true,
							DataMayBeNaN:      true,
							Warning:           Alert{GreaterOrEqual: 110},
							PanelOptions:      PanelOptions().LegendFormat("{{cache}}").Unit(Percentage),
							PossibleSolutions: "none",
						},
					},
				},
			},
			{
				Title:  "Internal service requests",
				Hidden: true,
				Rows: []Row{
					{
						sharedFrontendInternalAPIErrorResponses("precise-code-intel-bundle-manager"),
					},
				},
			},
			{
				Title:  "Container monitoring (not available on k8s or server)",
				Hidden: true,
				Rows: []Row{
					{
						sharedContainerRestarts("precise-code-intel-bundle-manager"),
						sharedContainerMemoryUsage("precise-code-intel-bundle-manager"),
						sharedContainerCPUUsage("precise-code-intel-bundle-manager"),
					},
				},
			},
		},
	}
}
