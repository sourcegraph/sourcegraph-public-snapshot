package main

func QueryRunner() *Container {
	return &Container{
		Name:        "query-runner",
		Title:       "Query Runner",
		Description: "Periodically runs saved searches and instructs the frontend to send out notifications.",
		Groups: []Group{
			{
				Title: "General",
				Rows: []Row{
					{
						{
							Name:            "frontend_internal_api_error_responses",
							Description:     "frontend-internal API error responses every 5m by route",
							Query:           `sum by (category)(increase(src_frontend_internal_request_duration_seconds_count{job="query-runner",code!~"2.."}[5m]))`,
							DataMayNotExist: true,
							Warning:         Alert{GreaterOrEqual: 5},
							PanelOptions:    PanelOptions().LegendFormat("{{category}}"),
						},
					},
				},
			},
			{
				Title:  "Container monitoring (not available on k8s or server)",
				Hidden: true,
				Rows: []Row{
					{
						{
							Name:            "container_restarts",
							Description:     "container restarts every 5m by instance (not available on k8s or server)",
							Query:           `increase(cadvisor_container_restart_count{name=~".*query-runner.*"}[5m])`,
							DataMayNotExist: true,
							Warning:         Alert{GreaterOrEqual: 1},
							PanelOptions:    PanelOptions().LegendFormat("{{name}}"),
						},
						{
							Name:            "container_memory_usage",
							Description:     "container memory usage by instance (not available on k8s or server)",
							Query:           `cadvisor_container_memory_usage_percentage_total{name=~".*query-runner.*"}`,
							DataMayNotExist: true,
							Warning:         Alert{GreaterOrEqual: 90},
							PanelOptions:    PanelOptions().LegendFormat("{{name}}").Unit(Percentage),
						},
						{
							Name:            "container_cpu_usage",
							Description:     "container cpu usage total (5m average) across all cores by instance (not available on k8s or server)",
							Query:           `cadvisor_container_cpu_usage_percentage_total{name=~".*query-runner.*"}`,
							DataMayNotExist: true,
							Warning:         Alert{GreaterOrEqual: 90},
							PanelOptions:    PanelOptions().LegendFormat("{{name}}").Unit(Percentage),
						},
					},
				},
			},
		},
	}
}
