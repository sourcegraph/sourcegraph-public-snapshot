package main

func PreciseCodeIntelAPIServer() *Container {
	return &Container{
		Name:        "precise-code-intel-api-server",
		Title:       "Precise Code Intel API Server",
		Description: "Serves precise code intelligence requests.",
		Groups: []Group{
			{
				Title: "General",
				Rows: []Row{
					{
						{
							Name:        "99th_percentile_code_intel_api_duration",
							Description: "99th percentile successful code intel api query duration over 5m",
							// TODO(efritz) - ensure these exclude error durations
							Query:             `histogram_quantile(0.99, sum by (le,op)(rate(src_code_intel_api_duration_seconds_bucket[5m])))`,
							DataMayNotExist:   true,
							Warning:           Alert{GreaterOrEqual: 20},
							PanelOptions:      PanelOptions().LegendFormat("{{op}}").Unit(Seconds),
							PossibleSolutions: "none",
						},
						{
							Name:              "code_intel_api_errors",
							Description:       "code intel api errors every 5m",
							Query:             `sum by (op)(increase(src_code_intel_api_errors_total[5m]))`,
							DataMayNotExist:   true,
							Warning:           Alert{GreaterOrEqual: 20},
							PanelOptions:      PanelOptions().LegendFormat("{{op}}"),
							PossibleSolutions: "none",
						},
					},
					// TODO(efritz) - add bundle manager request meter
					// TODO(efritz) - add gitserver request meter
					{
						{
							Name:        "99th_percentile_db_duration",
							Description: "99th percentile successful database query duration over 5m",
							// TODO(efritz) - ensure these exclude error durations
							Query:             `histogram_quantile(0.99, sum by (le,op)(rate(src_code_intel_db_duration_seconds_bucket{job="precise-code-intel-api-server"}[5m])))`,
							DataMayNotExist:   true,
							Warning:           Alert{GreaterOrEqual: 20},
							PanelOptions:      PanelOptions().LegendFormat("{{op}}").Unit(Seconds),
							PossibleSolutions: "none",
						},
						{
							Name:              "db_errors",
							Description:       "database errors every 5m",
							Query:             `sum by (op)(increase(src_code_intel_db_errors_total{job="precise-code-intel-api-server"}[5m]))`,
							DataMayNotExist:   true,
							Warning:           Alert{GreaterOrEqual: 20},
							PanelOptions:      PanelOptions().LegendFormat("{{op}}"),
							PossibleSolutions: "none",
						},
					},
					{
						{
							Name:              "processing_uploads_reset",
							Description:       "jobs reset to queued state every 5m",
							Query:             `sum(increase(src_upload_queue_resets_total[5m]))`,
							DataMayNotExist:   true,
							Warning:           Alert{GreaterOrEqual: 20},
							PanelOptions:      PanelOptions().LegendFormat("jobs"),
							PossibleSolutions: "none",
						},
						{
							Name:              "upload_resetter_errors",
							Description:       "upload resetter errors every 5m",
							Query:             `sum(increase(src_upload_queue_reset_errors_total[5m]))`,
							DataMayNotExist:   true,
							Warning:           Alert{GreaterOrEqual: 20},
							PanelOptions:      PanelOptions().LegendFormat("errors"),
							PossibleSolutions: "none",
						},
					},
					{
						sharedFrontendInternalAPIErrorResponses("precise-code-intel-api-server"),
					},
				},
			},
			{
				Title:  "Container monitoring (not available on k8s or server)",
				Hidden: true,
				Rows: []Row{
					{
						sharedContainerRestarts("precise-code-intel-api-server"),
						sharedContainerMemoryUsage("precise-code-intel-api-server"),
						sharedContainerCPUUsage("precise-code-intel-api-server"),
					},
				},
			},
		},
	}
}
