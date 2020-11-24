package main

func PreciseCodeIntelWorker() *Container {
	return &Container{
		Name:        "precise-code-intel-worker",
		Title:       "Precise Code Intel Worker",
		Description: "Handles conversion of uploaded precise code intelligence bundles.",
		Groups: []Group{
			{
				Title: "Upload processor",
				Rows: []Row{
					{
						{
							Name:              "upload_queue_size",
							Description:       "upload queue size",
							Query:             `max(src_upload_queue_uploads_total)`,
							DataMayNotExist:   true,
							Warning:           Alert().GreaterOrEqual(100),
							PanelOptions:      PanelOptions().LegendFormat("uploads queued for processing"),
							Owner:             ObservableOwnerCodeIntel,
							PossibleSolutions: "none",
						},
						{
							Name:            "upload_queue_growth_rate",
							Description:     "upload queue growth rate every 5m",
							Query:           `sum(increase(src_upload_queue_uploads_total[30m])) / sum(increase(src_upload_queue_processor_total[30m]))`,
							DataMayNotExist: true,

							Warning:           Alert().GreaterOrEqual(5),
							PanelOptions:      PanelOptions().LegendFormat("upload queue growth rate"),
							Owner:             ObservableOwnerCodeIntel,
							PossibleSolutions: "none",
						},
						{
							Name:              "upload_process_errors",
							Description:       "upload process errors every 5m",
							Query:             `sum(increase(src_upload_queue_processor_errors_total[5m]))`,
							DataMayNotExist:   true,
							Warning:           Alert().GreaterOrEqual(20),
							PanelOptions:      PanelOptions().LegendFormat("errors"),
							Owner:             ObservableOwnerCodeIntel,
							PossibleSolutions: "none",
						},
					},
				},
			},
			{
				Title: "Stores",
				Rows: []Row{
					{
						{
							Name:              "codeintel_dbstore_99th_percentile_duration",
							Description:       "99th percentile successful dbstore operation duration over 5m",
							Query:             `histogram_quantile(0.99, sum by (le)(rate(src_codeintel_dbstore_duration_seconds_bucket{job="precise-code-intel-worker"}[5m])))`,
							DataMayNotExist:   true,
							Warning:           Alert().GreaterOrEqual(20),
							PanelOptions:      PanelOptions().LegendFormat("store operation").Unit(Seconds),
							Owner:             ObservableOwnerCodeIntel,
							PossibleSolutions: "none",
						},
						{
							Name:              "codeintel_dbstore_errors",
							Description:       "dbstore errors every 5m",
							Query:             `sum(increase(src_codeintel_dbstore_errors_total{job="precise-code-intel-worker"}[5m]))`,
							DataMayNotExist:   true,
							Warning:           Alert().GreaterOrEqual(20),
							PanelOptions:      PanelOptions().LegendFormat("error"),
							Owner:             ObservableOwnerCodeIntel,
							PossibleSolutions: "none",
						},
					},
					{
						{
							Name:              "codeintel_lsifstore_99th_percentile_duration",
							Description:       "99th percentile successful lsifstore operation duration over 5m",
							Query:             `histogram_quantile(0.99, sum by (le)(rate(src_codeintel_lsifstore_duration_seconds_bucket{job="precise-code-intel-worker"}[5m])))`,
							DataMayNotExist:   true,
							Warning:           Alert().GreaterOrEqual(20),
							PanelOptions:      PanelOptions().LegendFormat("store operation").Unit(Seconds),
							Owner:             ObservableOwnerCodeIntel,
							PossibleSolutions: "none",
						},
						{
							Name:              "codeintel_lsifstore_errors",
							Description:       "lsifstore errors every 5m",
							Query:             `sum(increase(src_codeintel_lsifstore_errors_total{job="precise-code-intel-worker"}[5m]))`,
							DataMayNotExist:   true,
							Warning:           Alert().GreaterOrEqual(20),
							PanelOptions:      PanelOptions().LegendFormat("error"),
							Owner:             ObservableOwnerCodeIntel,
							PossibleSolutions: "none",
						},
					},
					{
						{
							Name:              "codeintel_uploadstore_99th_percentile_duration",
							Description:       "99th percentile successful uploadstore operation duration over 5m",
							Query:             `histogram_quantile(0.99, sum by (le)(rate(src_codeintel_uploadstore_duration_seconds_bucket{job="precise-code-intel-worker"}[5m])))`,
							DataMayNotExist:   true,
							Warning:           Alert().GreaterOrEqual(20),
							PanelOptions:      PanelOptions().LegendFormat("store operation").Unit(Seconds),
							Owner:             ObservableOwnerCodeIntel,
							PossibleSolutions: "none",
						},
						{
							Name:              "codeintel_uploadstore_errors",
							Description:       "uploadstore errors every 5m",
							Query:             `sum(increase(src_codeintel_uploadstore_errors_total{job="precise-code-intel-worker"}[5m]))`,
							DataMayNotExist:   true,
							Warning:           Alert().GreaterOrEqual(20),
							PanelOptions:      PanelOptions().LegendFormat("error"),
							Owner:             ObservableOwnerCodeIntel,
							PossibleSolutions: "none",
						},
					},
					{
						{
							Name:              "codeintel_gitserver_99th_percentile_duration",
							Description:       "99th percentile successful gitserver operation duration over 5m",
							Query:             `histogram_quantile(0.99, sum by (le)(rate(src_codeintel_gitserver_duration_seconds_bucket{job="precise-code-intel-worker"}[5m])))`,
							DataMayNotExist:   true,
							Warning:           Alert().GreaterOrEqual(20),
							PanelOptions:      PanelOptions().LegendFormat("store operation").Unit(Seconds),
							Owner:             ObservableOwnerCodeIntel,
							PossibleSolutions: "none",
						},
						{
							Name:              "codeintel_gitserver_errors",
							Description:       "gitserver errors every 5m",
							Query:             `sum(increase(src_codeintel_gitserver_errors_total{job="precise-code-intel-worker"}[5m]))`,
							DataMayNotExist:   true,
							Warning:           Alert().GreaterOrEqual(20),
							PanelOptions:      PanelOptions().LegendFormat("error"),
							Owner:             ObservableOwnerCodeIntel,
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
						sharedFrontendInternalAPIErrorResponses("precise-code-intel-worker", ObservableOwnerCodeIntel),
					},
				},
			},
			{
				Title:  "Container monitoring (not available on server)",
				Hidden: true,
				Rows: []Row{
					{
						sharedContainerCPUUsage("precise-code-intel-worker", ObservableOwnerCodeIntel),
						sharedContainerMemoryUsage("precise-code-intel-worker", ObservableOwnerCodeIntel),
					},
					{
						sharedContainerRestarts("precise-code-intel-worker", ObservableOwnerCodeIntel),
						sharedContainerFsInodes("precise-code-intel-worker", ObservableOwnerCodeIntel),
					},
				},
			},
			{
				Title:  "Provisioning indicators (not available on server)",
				Hidden: true,
				Rows: []Row{
					{
						sharedProvisioningCPUUsageLongTerm("precise-code-intel-worker", ObservableOwnerCodeIntel),
						sharedProvisioningMemoryUsageLongTerm("precise-code-intel-worker", ObservableOwnerCodeIntel),
					},
					{
						sharedProvisioningCPUUsageShortTerm("precise-code-intel-worker", ObservableOwnerCodeIntel),
						sharedProvisioningMemoryUsageShortTerm("precise-code-intel-worker", ObservableOwnerCodeIntel),
					},
				},
			},
			{
				Title:  "Golang runtime monitoring",
				Hidden: true,
				Rows: []Row{
					{
						sharedGoGoroutines("precise-code-intel-worker", ObservableOwnerCodeIntel),
						sharedGoGcDuration("precise-code-intel-worker", ObservableOwnerCodeIntel),
					},
				},
			},
			{
				Title:  "Kubernetes monitoring (ignore if using Docker Compose or server)",
				Hidden: true,
				Rows: []Row{
					{
						sharedKubernetesPodsAvailable("precise-code-intel-worker", ObservableOwnerCodeIntel),
					},
				},
			},
		},
	}
}
