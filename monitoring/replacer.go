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
						sharedProvisioningCPUUsage1d("replacer"),
						sharedProvisioningMemoryUsage1d("replacer"),
					},
					{
						sharedProvisioningCPUUsage1h("replacer"),
						sharedProvisioningMemoryUsage1h("replacer"),
					},
					{
						sharedProvisioningCPUUsage5m("replacer"),
						sharedProvisioningMemoryUsage5m("replacer"),
					},
				},
			},
		},
	}
}
