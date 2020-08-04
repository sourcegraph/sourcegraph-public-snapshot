package main

func GitHubProxy() *Container {
	return &Container{
		Name:        "github-proxy",
		Title:       "GitHub Proxy",
		Description: "Proxies all requests to github.com, keeping track of and managing rate limits.",
		Groups: []Group{
			{
				Title: "GitHub API monitoring",
				Rows: []Row{
					{
						{
							Name:              "github_core_rate_limit_remaining",
							Description:       "remaining calls to GitHub before hitting the rate limit",
							Query:             `src_github_rate_limit_remaining{resource="core"}`,
							DataMayNotExist:   true,
							Critical:          Alert{LessOrEqual: 1000},
							PanelOptions:      PanelOptions().LegendFormat("calls remaining"),
							Owner:             ObservableOwnerSearch,
							PossibleSolutions: `Try restarting the pod to get a different public IP.`,
						},
						{
							Name:              "github_search_rate_limit_remaining",
							Description:       "remaining calls to GitHub search before hitting the rate limit",
							Query:             `src_github_rate_limit_remaining{resource="search"}`,
							DataMayNotExist:   true,
							Warning:           Alert{LessOrEqual: 5},
							PanelOptions:      PanelOptions().LegendFormat("calls remaining"),
							Owner:             ObservableOwnerSearch,
							PossibleSolutions: `Try restarting the pod to get a different public IP.`,
						},
					},
				},
			},
			{
				Title:  "Container monitoring (not available on server)",
				Hidden: true,
				Rows: []Row{
					{
						sharedContainerCPUUsage("github-proxy"),
						sharedContainerMemoryUsage("github-proxy"),
					},
					{
						sharedContainerRestarts("github-proxy"),
						sharedContainerFsInodes("github-proxy"),
					},
				},
			},
			{
				Title:  "Provisioning indicators (not available on server)",
				Hidden: true,
				Rows: []Row{
					{
						sharedProvisioningCPUUsage7d("github-proxy"),
						sharedProvisioningMemoryUsage7d("github-proxy"),
					},
					{
						sharedProvisioningCPUUsage5m("github-proxy"),
						sharedProvisioningMemoryUsage5m("github-proxy"),
					},
				},
			},
			{
				Title:  "Golang runtime monitoring",
				Hidden: true,
				Rows: []Row{
					{
						sharedGoGoroutines("github-proxy"),
						sharedGoGcDuration("github-proxy"),
					},
				},
			},
			{
				Title:  "Kubernetes monitoring (ignore if using Docker Compose or server)",
				Hidden: true,
				Rows: []Row{
					{
						sharedKubernetesPodsAvailable("github-proxy"),
					},
				},
			},
		},
	}
}
