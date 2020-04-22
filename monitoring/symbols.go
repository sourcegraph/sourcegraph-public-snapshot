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
							Name:            "store_fetch_failures",
							Description:     "store fetch failures every 5m",
							Query:           `sum(increase(symbols_store_fetch_failed[5m]))`,
							DataMayNotExist: true,
							Warning:         Alert{GreaterOrEqual: 5},
							PanelOptions:    PanelOptions().LegendFormat("failures"),
						},
						{
							Name:            "current_fetch_queue_size",
							Description:     "current fetch queue size",
							Query:           `sum(symbols_store_fetch_queue_size)`,
							DataMayNotExist: true,
							Warning:         Alert{GreaterOrEqual: 25},
							PanelOptions:    PanelOptions().LegendFormat("size"),
						},
					},
					{
						sharedFrontendInternalAPIErrorResponses("symbols"),
					},
				},
			},
			{
				Title:  "Container monitoring (not available on k8s or server)",
				Hidden: true,
				Rows: []Row{
					{
						sharedContainerRestarts("symbols"),
						sharedContainerMemoryUsage("symbols"),
						sharedContainerCPUUsage("symbols"),
					},
				},
			},
		},
	}
}
