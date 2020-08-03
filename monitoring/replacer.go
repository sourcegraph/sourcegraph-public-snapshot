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
				Title:  "Container monitoring (not available on server)",
				Hidden: true,
				Rows: []Row{
					{
						sharedContainerCPUUsage("replacer"),
						sharedContainerMemoryUsage("replacer"),
					},
					{
						sharedContainerRestarts("replacer"),
						sharedContainerFsInodes("replacer"),
					},
				},
			},
			{
				Title:  "Provisioning indicators (not available on server)",
				Hidden: true,
				Rows: []Row{
					{
						sharedProvisioningCPUUsage7d("replacer"),
						sharedProvisioningMemoryUsage7d("replacer"),
					},
					{
						sharedProvisioningCPUUsage5m("replacer"),
						sharedProvisioningMemoryUsage5m("replacer"),
					},
				},
			},
			{
				Title:  "Golang runtime monitoring",
				Hidden: true,
				Rows: []Row{
					{
						sharedGoGoroutines("replacer"),
						sharedGoGcDuration("replacer"),
					},
				},
			},
			{
				Title:  "Kubernetes monitoring (ignore if using Docker Compose or server)",
				Hidden: true,
				Rows: []Row{
					{
						sharedKubernetesPodsAvailable("replacer"),
					},
				},
			},
		},
	}
}
