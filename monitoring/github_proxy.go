package main

func GitHubProxy() *Container {
	return &Container{
		Name:        "github-proxy",
		Title:       "GitHub Proxy",
		Description: "Proxies all requests to github.com, keeping track of and managing rate limits.",
		Groups: []Group{
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
