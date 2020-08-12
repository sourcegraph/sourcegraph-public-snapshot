package main

func QueryRunner() *Container {
	return &Container{
		Name:        "query-runner",
		Title:       "Query Runner",
		Description: "Periodically runs saved searches and instructs the frontend to send out notifications.",
		Groups: []Group{
			{
				Title: "General",
				Rows: []Row{
					{
						sharedFrontendInternalAPIErrorResponses("query-runner"),
					},
				},
			},
			{
				Title:  "Container monitoring (not available on server)",
				Hidden: true,
				Rows: []Row{
					{
						sharedContainerMemoryUsage("query-runner"),
						sharedContainerCPUUsage("query-runner"),
					},
					{
						sharedContainerRestarts("query-runner"),
						sharedContainerFsInodes("query-runner"),
					},
				},
			},
			{
				Title:  "Provisioning indicators (not available on server)",
				Hidden: true,
				Rows: []Row{
					{
						sharedProvisioningCPUUsageLongTerm("query-runner"),
						sharedProvisioningMemoryUsageLongTerm("query-runner"),
					},
					{
						sharedProvisioningCPUUsageShortTerm("query-runner"),
						sharedProvisioningMemoryUsageShortTerm("query-runner"),
					},
				},
			},
			{
				Title:  "Golang runtime monitoring",
				Hidden: true,
				Rows: []Row{
					{
						sharedGoGoroutines("query-runner"),
						sharedGoGcDuration("query-runner"),
					},
				},
			},
			{
				Title:  "Kubernetes monitoring (ignore if using Docker Compose or server)",
				Hidden: true,
				Rows: []Row{
					{
						sharedKubernetesPodsAvailable("query-runner"),
					},
				},
			},
		},
	}
}
