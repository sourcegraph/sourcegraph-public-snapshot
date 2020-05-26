package main

func PreciseCodeIntelIndexer() *Container {
	return &Container{
		Name:        "precise-code-intel-indexer",
		Title:       "Precise Code Intel Indexer",
		Description: "Automatically indexes from popular, active Go repositories.",
		Groups: []Group{
			{
				Title:  "Internal service requests",
				Hidden: true,
				Rows: []Row{
					{
						sharedFrontendInternalAPIErrorResponses("precise-code-intel-indexer"),
					},
				},
			},
			{
				Title:  "Container monitoring (not available on k8s or server)",
				Hidden: true,
				Rows: []Row{
					{
						sharedContainerRestarts("precise-code-intel-indexer"),
						sharedContainerMemoryUsage("precise-code-intel-indexer"),
						sharedContainerCPUUsage("precise-code-intel-indexer"),
					},
				},
			},
		},
	}
}
