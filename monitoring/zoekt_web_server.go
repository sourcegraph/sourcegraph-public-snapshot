package main

func ZoektWebServer() *Container {
	return &Container{
		Name:        "zoekt-webserver",
		Title:       "Zoekt Web Server",
		Description: "Serves indexed search requests using the search index.",
		Groups: []Group{
			{
				Title: "General",
				Rows: []Row{
					{
						{
							Name:            "indexed_search_request_errors",
							Description:     "indexed search request errors every 5m by code",
							Query:           `sum by (code)(increase(src_zoekt_request_duration_seconds_count{code!~"2.."}[5m]))`,
							DataMayNotExist: true,
							Warning:         Alert{GreaterOrEqual: 50},
							PanelOptions:    PanelOptions().LegendFormat("{{code}}").Unit(Seconds),
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
							Query:           `increase(cadvisor_container_restart_count{name=~".*zoekt-webserver.*"}[5m])`,
							DataMayNotExist: true,
							Warning:         Alert{GreaterOrEqual: 1},
							PanelOptions:    PanelOptions().LegendFormat("{{name}}"),
						},
						{
							Name:            "container_memory_usage",
							Description:     "container memory usage by instance (not available on k8s or server)",
							Query:           `cadvisor_container_memory_usage_percentage_total{name=~".*zoekt-webserver.*"}`,
							DataMayNotExist: true,
							Warning:         Alert{GreaterOrEqual: 90},
							PanelOptions:    PanelOptions().LegendFormat("{{name}}").Unit(Percentage),
						},
						{
							Name:            "container_cpu_usage",
							Description:     "container cpu usage total (5m average) across all cores by instance (not available on k8s or server)",
							Query:           `cadvisor_container_cpu_usage_percentage_total{name=~".*zoekt-webserver.*"}`,
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
