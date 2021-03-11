package definitions

import (
	"github.com/sourcegraph/sourcegraph/monitoring/definitions/shared"
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
							Name:           "syntax_highlighting_errors",
							Description:    "syntax highlighting errors every 5m",
							Query:          `sum(increase(src_syntax_highlighting_requests{status="error"}[5m])) / sum(increase(src_syntax_highlighting_requests[5m])) * 100`,
							NoAlert:        true,
							Panel:          monitoring.Panel().LegendFormat("error").Unit(monitoring.Percentage),
							Owner:          monitoring.ObservableOwnerCoreApplication,
							Interpretation: "none",
						},
						{
							Name:           "syntax_highlighting_timeouts",
							Description:    "syntax highlighting timeouts every 5m",
							Query:          `sum(increase(src_syntax_highlighting_requests{status="timeout"}[5m])) / sum(increase(src_syntax_highlighting_requests[5m])) * 100`,
							NoAlert:        true,
							Panel:          monitoring.Panel().LegendFormat("timeout").Unit(monitoring.Percentage),
							Owner:          monitoring.ObservableOwnerCoreApplication,
							Interpretation: "none",
						},
					},
					{
						{
							Name:           "syntax_highlighting_panics",
							Description:    "syntax highlighting panics every 5m",
							Query:          `sum(increase(src_syntax_highlighting_requests{status="panic"}[5m]))`,
							NoAlert:        true,
							Panel:          monitoring.Panel().LegendFormat("panic"),
							Owner:          monitoring.ObservableOwnerCoreApplication,
							Interpretation: "none",
						},
						{
							Name:           "syntax_highlighting_worker_deaths",
							Description:    "syntax highlighter worker deaths every 5m",
							Query:          `sum(increase(src_syntax_highlighting_requests{status="hss_worker_timeout"}[5m]))`,
							NoAlert:        true,
							Panel:          monitoring.Panel().LegendFormat("worker death"),
							Owner:          monitoring.ObservableOwnerCoreApplication,
							Interpretation: "none",
						},
					},
				},
			},
			{
				Title:  shared.TitleContainerMonitoring,
				Hidden: true,
				Rows: []monitoring.Row{
					{
						shared.ContainerCPUUsage("syntect-server", monitoring.ObservableOwnerCoreApplication).Observable(),
						shared.ContainerMemoryUsage("syntect-server", monitoring.ObservableOwnerCoreApplication).Observable(),
					},
					{
						shared.ContainerMissing("syntect-server", monitoring.ObservableOwnerCoreApplication).Observable(),
					},
				},
			},
			{
				Title:  shared.TitleProvisioningIndicators,
				Hidden: true,
				Rows: []monitoring.Row{
					{
						shared.ProvisioningCPUUsageLongTerm("syntect-server", monitoring.ObservableOwnerCoreApplication).Observable(),
						shared.ProvisioningMemoryUsageLongTerm("syntect-server", monitoring.ObservableOwnerCoreApplication).Observable(),
					},
					{
						shared.ProvisioningCPUUsageShortTerm("syntect-server", monitoring.ObservableOwnerCoreApplication).Observable(),
						shared.ProvisioningMemoryUsageShortTerm("syntect-server", monitoring.ObservableOwnerCoreApplication).Observable(),
					},
				},
			},
			{
				Title:  shared.TitleKubernetesMonitoring,
				Hidden: true,
				Rows: []monitoring.Row{
					{
						shared.KubernetesPodsAvailable("syntect-server", monitoring.ObservableOwnerCoreApplication).Observable(),
					},
				},
			},
		},
		NoSourcegraphDebugServer: true,
	}
}
