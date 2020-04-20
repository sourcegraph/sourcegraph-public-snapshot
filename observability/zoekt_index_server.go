package main

func ZoektIndexServer() *Container {
	return &Container{
		Name:        "zoekt-indexserver",
		Title:       "Zoekt Index Server",
		Description: "Indexes repositories and populates the search index.",
		Groups: []Group{
			{
				Title: "General",
				Rows: []Row{
					{
						{
							Name:            "average_resolve_revision_duration",
							Description:     "average resolve revision duration over 5m",
							Query:           `sum(rate(resolve_revision_seconds_sum[5m]))`,
							DataMayNotExist: true,
							Warning:         Alert{GreaterOrEqual: 15},
							Critical:        Alert{GreaterOrEqual: 30},
							PanelOptions:    PanelOptions().LegendFormat("{{duration}}").Unit(Seconds),
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
							Query:           `increase(cadvisor_container_restart_count{name=~".*zoekt-indexserver.*"}[5m])`,
							DataMayNotExist: true,
							Warning:         Alert{GreaterOrEqual: 1},
							PanelOptions:    PanelOptions().LegendFormat("{{name}}"),
						},
						{
							Name:            "container_memory_usage",
							Description:     "container memory usage by instance (not available on k8s or server)",
							Query:           `cadvisor_container_memory_usage_percentage_total{name=~".*zoekt-indexserver.*"}`,
							DataMayNotExist: true,
							Warning:         Alert{GreaterOrEqual: 90},
							PanelOptions:    PanelOptions().LegendFormat("{{name}}").Unit(Percentage),
						},
						{
							Name:            "container_cpu_usage",
							Description:     "container cpu usage total (5m average) across all cores by instance (not available on k8s or server)",
							Query:           `cadvisor_container_cpu_usage_percentage_total{name=~".*zoekt-indexserver.*"}`,
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
