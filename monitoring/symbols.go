package main

func Symbols() *Container {
	return &Container{
		Name:        "symbols",
		Title:       "Symbols",
		Description: "Handles symbol searches for unindexed branches.",
		Groups: []Group{
			{
				Title: "General",
				Rows: []Row{
					{
						{
							Name:              "store_fetch_failures",
							Description:       "store fetch failures every 5m",
							Query:             `sum(increase(symbols_store_fetch_failed[5m]))`,
							DataMayNotExist:   true,
							Warning:           Alert{GreaterOrEqual: 5},
							PanelOptions:      PanelOptions().LegendFormat("failures"),
							Owner:             ObservableOwnerSearch,
							PossibleSolutions: "none",
						},
						{
							Name:              "current_fetch_queue_size",
							Description:       "current fetch queue size",
							Query:             `sum(symbols_store_fetch_queue_size)`,
							DataMayNotExist:   true,
							Warning:           Alert{GreaterOrEqual: 25},
							PanelOptions:      PanelOptions().LegendFormat("size"),
							Owner:             ObservableOwnerSearch,
							PossibleSolutions: "none",
						},
					},
					{
						sharedFrontendInternalAPIErrorResponses("symbols"),
					},
				},
			},
			{
				Title:  "Container monitoring (not available on server)",
				Hidden: true,
				Rows: []Row{
					{
						sharedContainerCPUUsage("symbols"),
						sharedContainerMemoryUsage("symbols"),
					},
					{
						sharedContainerRestarts("symbols"),
						sharedContainerFsInodes("symbols"),
					},
				},
			},
			{
				Title:  "Provisioning indicators (not available on server)",
				Hidden: true,
				Rows: []Row{
					{
						sharedProvisioningCPUUsage7d("symbols"),
						sharedProvisioningMemoryUsage7d("symbols"),
					},
					{
						sharedProvisioningCPUUsage5m("symbols"),
						sharedProvisioningMemoryUsage5m("symbols"),
					},
				},
			},
			{
				Title:  "Golang runtime monitoring",
				Hidden: true,
				Rows: []Row{
					{
						sharedGoGoroutines("symbols"),
						sharedGoGcDuration("symbols"),
					},
				},
			},
			{
				Title:  "Kubernetes monitoring (ignore if using Docker Compose or server)",
				Hidden: true,
				Rows: []Row{
					{
						sharedKubernetesPodsAvailable("symbols"),
					},
				},
			},
		},
	}
}
