package main

func PreciseCodeIntelWorker() *Container {
	return &Container{
		Name:        "precise-code-intel-worker",
		Title:       "Precise Code Intel Worker",
		Description: "Handles conversion of uploaded precise code intelligence bundles.",
		Groups: []Group{
			{
				Title:  "Container monitoring (not available on k8s or server)",
				Hidden: true,
				Rows: []Row{
					{
						{
							Name:            "container_restarts",
							Description:     "container restarts every 5m by instance (not available on k8s or server)",
							Query:           `increase(cadvisor_container_restart_count{name=~".*precise-code-intel-worker.*"}[5m])`,
							DataMayNotExist: true,
							Warning:         Alert{GreaterOrEqual: 1},
							PanelOptions:    PanelOptions().LegendFormat("{{name}}"),
						},
						{
							Name:            "container_memory_usage",
							Description:     "container memory usage by instance (not available on k8s or server)",
							Query:           `cadvisor_container_memory_usage_percentage_total{name=~".*precise-code-intel-worker.*"}`,
							DataMayNotExist: true,
							Warning:         Alert{GreaterOrEqual: 90},
							PanelOptions:    PanelOptions().LegendFormat("{{name}}").Unit(Percentage),
						},
						{
							Name:            "container_cpu_usage",
							Description:     "container cpu usage total (5m average) across all cores by instance (not available on k8s or server)",
							Query:           `cadvisor_container_cpu_usage_percentage_total{name=~".*precise-code-intel-worker.*"}`,
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
