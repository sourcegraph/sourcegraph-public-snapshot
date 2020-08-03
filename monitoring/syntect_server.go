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
							Name:              "syntax_highlighting_errors",
							Description:       "syntax highlighting errors every 5m",
							Query:             `sum(increase(src_syntax_highlighting_requests{status="error"}[5m]))`,
							DataMayNotExist:   true,
							Warning:           Alert{GreaterOrEqual: 5},
							PanelOptions:      PanelOptions().LegendFormat("error"),
							Owner:             ObservableOwnerCodeIntel,
							PossibleSolutions: "none",
						},
						{
							Name:              "syntax_highlighting_panics",
							Description:       "syntax highlighting panics every 5m",
							Query:             `sum(increase(src_syntax_highlighting_requests{status="panic"}[5m]))`,
							DataMayNotExist:   true,
							Warning:           Alert{GreaterOrEqual: 5},
							PanelOptions:      PanelOptions().LegendFormat("panic"),
							Owner:             ObservableOwnerCodeIntel,
							PossibleSolutions: "none",
						},
					},
					{
						{
							Name:              "syntax_highlighting_timeouts",
							Description:       "syntax highlighting timeouts every 5m",
							Query:             `sum(increase(src_syntax_highlighting_requests{status="timeout"}[5m]))`,
							DataMayNotExist:   true,
							Warning:           Alert{GreaterOrEqual: 5},
							PanelOptions:      PanelOptions().LegendFormat("timeout"),
							Owner:             ObservableOwnerCodeIntel,
							PossibleSolutions: "none",
						},
						{
							Name:              "syntax_highlighting_worker_deaths",
							Description:       "syntax highlighter worker deaths every 5m",
							Query:             `sum(increase(src_syntax_highlighting_requests{status="hss_worker_timeout"}[5m]))`,
							DataMayNotExist:   true,
							Warning:           Alert{GreaterOrEqual: 1},
							PanelOptions:      PanelOptions().LegendFormat("worker death"),
							Owner:             ObservableOwnerCodeIntel,
							PossibleSolutions: "none",
						},
					},
				},
			},
			{
				Title:  "Container monitoring (not available on server)",
				Hidden: true,
				Rows: []Row{
					{
						sharedContainerCPUUsage("syntect-server"),
						sharedContainerMemoryUsage("syntect-server"),
					},
					{
						sharedContainerRestarts("syntect-server"),
						sharedContainerFsInodes("syntect-server"),
					},
				},
			},
			{
				Title:  "Provisioning indicators (not available on server)",
				Hidden: true,
				Rows: []Row{
					{
						sharedProvisioningCPUUsage7d("syntect-server"),
						sharedProvisioningMemoryUsage7d("syntect-server"),
					},
					{
						sharedProvisioningCPUUsage5m("syntect-server"),
						sharedProvisioningMemoryUsage5m("syntect-server"),
					},
				},
			},
			{
				Title:  "Kubernetes monitoring (ignore if using Docker Compose or server)",
				Hidden: true,
				Rows: []Row{
					{
						sharedKubernetesPodsAvailable("syntect-server"),
					},
				},
			},
		},
	}
}
