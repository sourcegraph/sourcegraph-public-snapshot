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
							Name:              "github_core_rate_limit_remaining",
							Description:       "remaining calls to GitHub before hitting the rate limit",
							Query:             `src_github_rate_limit_remaining{resource="core"}`,
							DataMayNotExist:   true,
							Critical:          monitoring.Alert().LessOrEqual(500).For(5 * time.Minute),
							PanelOptions:      monitoring.PanelOptions().LegendFormat("calls remaining"),
							Owner:             monitoring.ObservableOwnerCloud,
							PossibleSolutions: `Try restarting the pod to get a different public IP.`,
						},
						{
							Name:              "github_search_rate_limit_remaining",
							Description:       "remaining calls to GitHub search before hitting the rate limit",
							Query:             `src_github_rate_limit_remaining{resource="search"}`,
							DataMayNotExist:   true,
							Warning:           monitoring.Alert().LessOrEqual(5),
							PanelOptions:      monitoring.PanelOptions().LegendFormat("calls remaining"),
							Owner:             monitoring.ObservableOwnerCloud,
							PossibleSolutions: `Try restarting the pod to get a different public IP.`,
						},
					},
					{
						{
							Name:            "github_proxy_waiting_requests",
							Description:     "number of requests waiting on the global mutex",
							Query:           `github_proxy_waiting_requests`,
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
