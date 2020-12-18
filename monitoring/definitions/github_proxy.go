package definitions

import (
	"time"

	"github.com/sourcegraph/sourcegraph/monitoring/definitions/shared"
	"github.com/sourcegraph/sourcegraph/monitoring/monitoring"
)

func GitHubProxy() *monitoring.Container {
	return &monitoring.Container{
		Name:        "github-proxy",
		Title:       "GitHub Proxy",
		Description: "Proxies all requests to github.com, keeping track of and managing rate limits.",
		Groups: []monitoring.Group{
			{
				Title: "GitHub API monitoring",
				Rows: []monitoring.Row{
					{
						{
							Name:            "github_proxy_waiting_requests",
							Description:     "number of requests waiting on the global mutex",
							Query:           `max(github_proxy_waiting_requests)`,
							DataMayNotExist: true,
							Warning:         monitoring.Alert().GreaterOrEqual(100).For(5 * time.Minute),
							PanelOptions:    monitoring.PanelOptions().LegendFormat("requests waiting"),
							Owner:           monitoring.ObservableOwnerCloud,
							PossibleSolutions: `
								- **Check github-proxy logs for network connection issues.
								- **Check github status.`,
						},
					},
				},
			},
			{
				Title:  "Container monitoring (not available on server)",
				Hidden: true,
				Rows: []monitoring.Row{
					{
						shared.ContainerCPUUsage("github-proxy", monitoring.ObservableOwnerCloud),
						shared.ContainerMemoryUsage("github-proxy", monitoring.ObservableOwnerCloud),
					},
					{
						shared.ContainerRestarts("github-proxy", monitoring.ObservableOwnerCloud),
						shared.ContainerFsInodes("github-proxy", monitoring.ObservableOwnerCloud),
					},
				},
			},
			{
				Title:  "Provisioning indicators (not available on server)",
				Hidden: true,
				Rows: []monitoring.Row{
					{
						shared.ProvisioningCPUUsageLongTerm("github-proxy", monitoring.ObservableOwnerCloud),
						shared.ProvisioningMemoryUsageLongTerm("github-proxy", monitoring.ObservableOwnerCloud),
					},
					{
						shared.ProvisioningCPUUsageShortTerm("github-proxy", monitoring.ObservableOwnerCloud),
						shared.ProvisioningMemoryUsageShortTerm("github-proxy", monitoring.ObservableOwnerCloud),
					},
				},
			},
			{
				Title:  "Golang runtime monitoring",
				Hidden: true,
				Rows: []monitoring.Row{
					{
						shared.GoGoroutines("github-proxy", monitoring.ObservableOwnerCloud),
						shared.GoGcDuration("github-proxy", monitoring.ObservableOwnerCloud),
					},
				},
			},
			{
				Title:  "Kubernetes monitoring (ignore if using Docker Compose or server)",
				Hidden: true,
				Rows: []monitoring.Row{
					{
						shared.KubernetesPodsAvailable("github-proxy", monitoring.ObservableOwnerCloud),
					},
				},
			},
		},
	}
}
