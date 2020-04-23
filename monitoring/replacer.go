package main

func Replacer() *Container {
	return &Container{
		Name:        "replacer",
		Title:       "Replacer",
		Description: "Backend for find-and-replace operations.",
		Groups: []Group{
			{
				Title: "General",
				Rows: []Row{
					{
						sharedFrontendInternalAPIErrorResponses("replacer"),
					},
				},
			},
			{
				Title:  "Container monitoring (not available on k8s or server)",
				Hidden: true,
				Rows: []Row{
					{
						sharedContainerRestarts("replacer"),
						sharedContainerMemoryUsage("replacer"),
						sharedContainerCPUUsage("replacer"),
					},
				},
			},
		},
	}
}
