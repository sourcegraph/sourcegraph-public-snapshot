package definitions

import (
	"time"

	"github.com/sourcegraph/sourcegraph/monitoring/definitions/shared"
	"github.com/sourcegraph/sourcegraph/monitoring/monitoring"
)

func GitHubProxy() *monitoring.Container {
	const (
		containerName = "github-proxy"
		primaryOwner  = monitoring.ObservableOwnerCoreApplication
	)

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
							Name:        "github_proxy_waiting_requests",
							Description: "number of requests waiting on the global mutex",
							Query:       `max(github_proxy_waiting_requests)`,
							Warning:     monitoring.Alert().GreaterOrEqual(100, nil).For(5 * time.Minute),
							Panel:       monitoring.Panel().LegendFormat("requests waiting"),
							Owner:       monitoring.ObservableOwnerCoreApplication,
							PossibleSolutions: `
								- **Check github-proxy logs for network connection issues.
								- **Check github status.`,
						},
					},
				},
			},
			shared.NewContainerMonitoringGroup(containerName, primaryOwner, nil),
			shared.NewProvisioningIndicatorsGroup(containerName, primaryOwner, nil),
			shared.NewGolangMonitoringGroup(containerName, primaryOwner, nil),
			shared.NewKubernetesMonitoringGroup(containerName, primaryOwner, nil),
		},
	}
}
