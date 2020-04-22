package main

func GitHubProxy() *Container {
	return &Container{
		Name:        "github-proxy",
		Title:       "GitHub Proxy",
		Description: "Proxies all requests to github.com, keeping track of and managing rate limits.",
		Groups: []Group{
			{
				Title:  "Container monitoring (not available on k8s or server)",
				Hidden: true,
				Rows: []Row{
					{
						sharedContainerRestarts("github-proxy"),
						sharedContainerMemoryUsage("github-proxy"),
						sharedContainerCPUUsage("github-proxy"),
					},
				},
			},
		},
	}
}
