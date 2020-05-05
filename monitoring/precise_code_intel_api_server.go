package main

func PreciseCodeIntelAPIServer() *Container {
	return &Container{
		Name:        "precise-code-intel-api-server",
		Title:       "Precise Code Intel API Server",
		Description: "Serves precise code intelligence requests.",
		Groups: []Group{
			{
				Title:  "Container monitoring (not available on k8s or server)",
				Hidden: true,
				Rows: []Row{
					{
						sharedContainerRestarts("precise-code-intel-api-server"),
						sharedContainerMemoryUsage("precise-code-intel-api-server"),
						sharedContainerCPUUsage("precise-code-intel-api-server"),
					},
				},
			},
		},
	}
}
