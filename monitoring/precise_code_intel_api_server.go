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
						{
							Name:              "99th_percentile_bundle_manager_query_duration",
							Description:       "99th percentile successful bundle manager query duration over 5m",
							Query:             `histogram_quantile(0.99, sum by (le,category)(rate(src_precise_code_intel_bundle_manager_request_duration_seconds_bucket{job="precise-code-intel-api-server",category!="transfer"}[5m])))`,
							DataMayNotExist:   true,
							Warning:           Alert{GreaterOrEqual: 20},
							PanelOptions:      PanelOptions().LegendFormat("{{category}}").Unit(Seconds),
							PossibleSolutions: "none",
						},
						{
							Name:              "99th_percentile_bundle_manager_transfer_duration",
							Description:       "99th percentile successful bundle manager data transfer duration over 5m",
							Query:             `histogram_quantile(0.99, sum by (le,category)(rate(src_precise_code_intel_bundle_manager_request_duration_seconds_bucket{job="precise-code-intel-api-server",category="transfer"}[5m])))`,
							DataMayNotExist:   true,
							Warning:           Alert{GreaterOrEqual: 300},
							PanelOptions:      PanelOptions().LegendFormat("{{category}}").Unit(Seconds),
							PossibleSolutions: "none",
						},
						{
							Name:              "bundle_manager_error_responses",
							Description:       "bundle manager error responses every 5m",
							Query:             `sum by (category)(increase(src_precise_code_intel_bundle_manager_request_duration_seconds_count{job="precise-code-intel-api-server",code!~"2.."}[5m]))`,
							DataMayNotExist:   true,
							Warning:           Alert{GreaterOrEqual: 5},
							PanelOptions:      PanelOptions().LegendFormat("{{category}}"),
							PossibleSolutions: "none",
						},
					},
					{
						{
							Name:              "99th_percentile_gitserver_duration",
							Description:       "99th percentile successful gitserver query duration over 5m",
							Query:             `histogram_quantile(0.99, sum by (le,category)(rate(src_gitserver_request_duration_seconds_bucket{job="precise-code-intel-api-server"}[5m])))`,
							DataMayNotExist:   true,
							Warning:           Alert{GreaterOrEqual: 20},
							PanelOptions:      PanelOptions().LegendFormat("{{category}}").Unit(Seconds),
							PossibleSolutions: "none",
						},
						{
							Name:              "gitserver_error_responses",
							Description:       "gitserver error responses every 5m",
							Query:             `sum by (category)(increase(src_gitserver_request_duration_seconds_count{job="precise-code-intel-api-server",code!~"2.."}[5m]))`,
							DataMayNotExist:   true,
							Warning:           Alert{GreaterOrEqual: 5},
							PanelOptions:      PanelOptions().LegendFormat("{{category}}"),
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
