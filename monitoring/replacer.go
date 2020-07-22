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
				Title:  "Container monitoring (not available on server)",
				Hidden: true,
				Rows: []Row{
					{
						sharedContainerRestarts("replacer"),
						sharedContainerMemoryUsage("replacer"),
						sharedContainerCPUUsage("replacer"),
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
				Title:  "Kubernetes monitoring (only available on k8s)",
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
