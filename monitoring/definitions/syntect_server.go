package definitions

import (
	"time"

	"github.com/sourcegraph/sourcegraph/monitoring/monitoring"
)

func SyntectServer() *monitoring.Container {
	return &monitoring.Container{
		Name:        "syntect-server",
		Title:       "Syntect Server",
		Description: "Handles syntax highlighting for code files.",
		Groups: []monitoring.Group{
			{
				Title: "General",
				Rows: []monitoring.Row{
					{
						{
							Name:              "syntax_highlighting_errors",
							Description:       "syntax highlighting errors every 5m",
							Query:             `sum(increase(src_syntax_highlighting_requests{status="error"}[5m])) / sum(increase(src_syntax_highlighting_requests[5m])) * 100`,
							DataMayNotExist:   true,
							Warning:           monitoring.Alert().GreaterOrEqual(5).For(5 * time.Minute),
							PanelOptions:      monitoring.PanelOptions().LegendFormat("error").Unit(monitoring.Percentage),
							Owner:             monitoring.ObservableOwnerCloud,
							PossibleSolutions: "none",
						},
						{
							Name:              "syntax_highlighting_timeouts",
							Description:       "syntax highlighting timeouts every 5m",
							Query:             `sum(increase(src_syntax_highlighting_requests{status="timeout"}[5m])) / sum(increase(src_syntax_highlighting_requests[5m])) * 100`,
							DataMayNotExist:   true,
							Warning:           monitoring.Alert().GreaterOrEqual(5).For(5 * time.Minute),
							PanelOptions:      monitoring.PanelOptions().LegendFormat("timeout").Unit(monitoring.Percentage),
							Owner:             monitoring.ObservableOwnerCloud,
							PossibleSolutions: "none",
						},
					},
					{
						{
							Name:              "syntax_highlighting_panics",
							Description:       "syntax highlighting panics every 5m",
							Query:             `sum(increase(src_syntax_highlighting_requests{status="panic"}[5m]))`,
							DataMayNotExist:   true,
							Warning:           monitoring.Alert().GreaterOrEqual(5),
							PanelOptions:      monitoring.PanelOptions().LegendFormat("panic"),
							Owner:             monitoring.ObservableOwnerCloud,
							PossibleSolutions: "none",
						},
						{
							Name:              "syntax_highlighting_worker_deaths",
							Description:       "syntax highlighter worker deaths every 5m",
							Query:             `sum(increase(src_syntax_highlighting_requests{status="hss_worker_timeout"}[5m]))`,
							DataMayNotExist:   true,
							Warning:           monitoring.Alert().GreaterOrEqual(1),
							PanelOptions:      monitoring.PanelOptions().LegendFormat("worker death"),
							Owner:             monitoring.ObservableOwnerCloud,
							PossibleSolutions: "none",
						},
					},
				},
			},
			{
				Title:  "Container monitoring (not available on server)",
				Hidden: true,
				Rows: []monitoring.Row{
					{
						sharedContainerCPUUsage("syntect-server", monitoring.ObservableOwnerCloud),
						sharedContainerMemoryUsage("syntect-server", monitoring.ObservableOwnerCloud),
					},
					{
						sharedContainerRestarts("syntect-server", monitoring.ObservableOwnerCloud),
						sharedContainerFsInodes("syntect-server", monitoring.ObservableOwnerCloud),
					},
				},
			},
			{
				Title:  "Provisioning indicators (not available on server)",
				Hidden: true,
				Rows: []monitoring.Row{
					{
						sharedProvisioningCPUUsageLongTerm("syntect-server", monitoring.ObservableOwnerCloud),
						sharedProvisioningMemoryUsageLongTerm("syntect-server", monitoring.ObservableOwnerCloud),
					},
					{
						sharedProvisioningCPUUsageShortTerm("syntect-server", monitoring.ObservableOwnerCloud),
						sharedProvisioningMemoryUsageShortTerm("syntect-server", monitoring.ObservableOwnerCloud),
					},
				},
			},
			{
				Title:  "Kubernetes monitoring (ignore if using Docker Compose or server)",
				Hidden: true,
				Rows: []monitoring.Row{
					{
						sharedKubernetesPodsAvailable("syntect-server", monitoring.ObservableOwnerCloud),
					},
				},
			},
		},
	}
}
