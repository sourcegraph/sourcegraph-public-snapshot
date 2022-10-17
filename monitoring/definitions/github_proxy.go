package definitions

import (
	"time"

	"github.com/sourcegraph/sourcegraph/monitoring/definitions/shared"
	"github.com/sourcegraph/sourcegraph/monitoring/monitoring"
)

func GitHubProxy() *monitoring.Dashboard {
	const containerName = "github-proxy"

	return &monitoring.Dashboard{
		Name:        "github-proxy",
		Title:       "GitHub Proxy",
		Description: "Proxies all requests to github.com, keeping track of and managing rate limits.",
		Groups: []monitoring.Group{
			{
				Title: "GitHub API monitoring",
				Rows: []monitoring.Row{
					{
						{
							Name:        "github_proxy_waiting_requests",
							Description: "number of requests waiting on the global mutex",
							Query:       `max(github_proxy_waiting_requests)`,
							Warning:     monitoring.Alert().GreaterOrEqual(100).For(5 * time.Minute),
							Panel:       monitoring.Panel().LegendFormat("requests waiting"),
							Owner:       monitoring.ObservableOwnerRepoManagement,
							NextSteps: `
								- **Check github-proxy logs for network connection issues.
								- **Check github status.`,
						},
					},
				},
			},

			shared.NewContainerMonitoringGroup(containerName, monitoring.ObservableOwnerDevOps, nil),
			shared.NewProvisioningIndicatorsGroup(containerName, monitoring.ObservableOwnerDevOps, nil),
			shared.NewGolangMonitoringGroup(containerName, monitoring.ObservableOwnerDevOps, nil),
			shared.NewKubernetesMonitoringGroup(containerName, monitoring.ObservableOwnerDevOps, nil),
		},
	}
}
