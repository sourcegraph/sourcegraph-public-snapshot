package main

func GitServer() *Container {
	return &Container{
		Name:        "gitserver",
		Title:       "Git Server",
		Description: "Stores, manages, and operates Git repositories.",
		Groups: []Group{
			{
				Title: "General",
				Rows: []Row{
					{
						{
							Name:            "disk_space_remaining",
							Description:     "disk space remaining by instance",
							Query:           `(src_gitserver_disk_space_available / src_gitserver_disk_space_total) * 100`,
							DataMayNotExist: true,
							Warning:         Alert{LessOrEqual: 25},
							Critical:        Alert{LessOrEqual: 15},
							PanelOptions:    PanelOptions().LegendFormat("{{instance}}").Unit(Percentage),
						},
						{
							Name:            "running_git_commands",
							Description:     "running git commands (signals load)",
							Query:           "max(src_gitserver_exec_running)",
							DataMayNotExist: true,
							Warning:         Alert{GreaterOrEqual: 50},
							Critical:        Alert{GreaterOrEqual: 100},
							PanelOptions:    PanelOptions().LegendFormat("running commands"),
						},
					}, {
						{
							Name:            "repository_clone_queue_size",
							Description:     "repository clone queue size",
							Query:           "sum(src_gitserver_clone_queue)",
							DataMayNotExist: true,
							Warning:         Alert{GreaterOrEqual: 25},
							PanelOptions:    PanelOptions().LegendFormat("queue size"),
						},
						{
							Name:            "repository_existence_check_queue_size",
							Description:     "repository existence check queue size",
							Query:           "sum(src_gitserver_lsremote_queue)",
							DataMayNotExist: true,
							Warning:         Alert{GreaterOrEqual: 25},
							PanelOptions:    PanelOptions().LegendFormat("queue size"),
						},
					}, {
						{
							Name:            "echo_command_duration_test",
							Description:     "echo command duration test",
							Query:           "max(src_gitserver_echo_duration_seconds)",
							DataMayNotExist: true,
							Warning:         Alert{GreaterOrEqual: 1.0},
							Critical:        Alert{GreaterOrEqual: 2.0},
							PanelOptions:    PanelOptions().LegendFormat("running commands").Unit(Seconds),
						},
						{
							Name:            "frontend_internal_api_error_responses",
							Description:     "frontend-internal API error responses every 5m by route",
							Query:           `sum by (category)(increase(src_frontend_internal_request_duration_seconds_count{job="gitserver",code!~"2.."}[5m]))`,
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
							Query:           `increase(cadvisor_container_restart_count{name=~".*gitserver.*"}[5m])`,
							DataMayNotExist: true,
							Warning:         Alert{GreaterOrEqual: 1},
							PanelOptions:    PanelOptions().LegendFormat("{{name}}"),
						},
						{
							Name:            "container_memory_usage",
							Description:     "container memory usage by instance (not available on k8s or server)",
							Query:           `cadvisor_container_memory_usage_percentage_total{name=~".*gitserver.*"}`,
							DataMayNotExist: true,
							Warning:         Alert{GreaterOrEqual: 90},
							PanelOptions:    PanelOptions().LegendFormat("{{name}}").Unit(Percentage),
						},
						{
							Name:            "container_cpu_usage",
							Description:     "container cpu usage total (5m average) across all cores by instance (not available on k8s or server)",
							Query:           `cadvisor_container_cpu_usage_percentage_total{name=~".*gitserver.*"}`,
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
