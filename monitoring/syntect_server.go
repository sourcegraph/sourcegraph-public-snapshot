package main

func SyntectServer() *Container {
	return &Container{
		Name:        "syntect-server",
		Title:       "Syntect Server",
		Description: "Handles syntax highlighting for code files.",
		Groups: []Group{
			{
				Title: "General",
				Rows: []Row{
					{
						{
							Name:            "syntax_highlighting_errors",
							Description:     "syntax highlighting errors every 5m",
							Query:           `sum(increase(src_syntax_highlighting_requests{status="error"}[5m]))`,
							DataMayNotExist: true,
							Warning:         Alert{GreaterOrEqual: 5},
							PanelOptions:    PanelOptions().LegendFormat("error"),
						},
						{
							Name:            "syntax_highlighting_panics",
							Description:     "syntax highlighting panics every 5m",
							Query:           `sum(increase(src_syntax_highlighting_requests{status="panic"}[5m]))`,
							DataMayNotExist: true,
							Warning:         Alert{GreaterOrEqual: 5},
							PanelOptions:    PanelOptions().LegendFormat("panic"),
						},
					},
					{
						{
							Name:            "syntax_highlighting_timeouts",
							Description:     "syntax highlighting timeouts every 5m",
							Query:           `sum(increase(src_syntax_highlighting_requests{status="timeout"}[5m]))`,
							DataMayNotExist: true,
							Warning:         Alert{GreaterOrEqual: 5},
							PanelOptions:    PanelOptions().LegendFormat("timeout"),
						},
						{
							Name:            "syntax_highlighting_worker_deaths",
							Description:     "syntax highlighter worker deaths every 5m",
							Query:           `sum(increase(src_syntax_highlighting_requests{status="hss_worker_timeout"}[5m]))`,
							DataMayNotExist: true,
							Warning:         Alert{GreaterOrEqual: 1},
							PanelOptions:    PanelOptions().LegendFormat("worker death"),
						},
					},
				},
			},
			{
				Title:  "Container monitoring (not available on k8s or server)",
				Hidden: true,
				Rows: []Row{
					{
						sharedContainerRestarts("syntect-server"),
						sharedContainerMemoryUsage("syntect-server"),
						sharedContainerCPUUsage("syntect-server"),
					},
				},
			},
		},
	}
}
