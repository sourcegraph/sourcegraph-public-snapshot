package main

func PreciseCodeIntelBundleManager() *Container {
	return &Container{
		Name:        "precise-code-intel-bundle-manager",
		Title:       "Precise Code Intel Bundle Manager",
		Description: "Stores and manages precise code intelligence bundles.",
		Groups: []Group{
			{
				Title:  "Container monitoring (not available on k8s or server)",
				Hidden: true,
				Rows: []Row{
					{
						sharedContainerRestarts("precise-code-intel-bundle-manager"),
						sharedContainerMemoryUsage("precise-code-intel-bundle-manager"),
						sharedContainerCPUUsage("precise-code-intel-bundle-manager"),
					},
				},
			},
		},
	}
}
