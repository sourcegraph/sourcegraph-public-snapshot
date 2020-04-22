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
				Title:  "Container monitoring (not available on k8s or server)",
				Hidden: true,
				Rows: []Row{
					{
						sharedContainerRestarts("query-runner"),
						sharedContainerMemoryUsage("query-runner"),
						sharedContainerCPUUsage("query-runner"),
					},
				},
			},
		},
	}
}
