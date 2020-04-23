package main

func ZoektIndexServer() *Container {
	return &Container{
		Name:        "zoekt-indexserver",
		Title:       "Zoekt Index Server",
		Description: "Indexes repositories and populates the search index.",
		Groups: []Group{
			{
				Title: "General",
				Rows: []Row{
					{
						{
							Name:            "average_resolve_revision_duration",
							Description:     "average resolve revision duration over 5m",
							Query:           `sum(rate(resolve_revision_seconds_sum[5m]))`,
							DataMayNotExist: true,
							Warning:         Alert{GreaterOrEqual: 15},
							Critical:        Alert{GreaterOrEqual: 30},
							PanelOptions:    PanelOptions().LegendFormat("{{duration}}").Unit(Seconds),
						},
					},
				},
			},
			{
				Title:  "Container monitoring (not available on k8s or server)",
				Hidden: true,
				Rows: []Row{
					{
						sharedContainerRestarts("zoekt-indexserver"),
						sharedContainerMemoryUsage("zoekt-indexserver"),
						sharedContainerCPUUsage("zoekt-indexserver"),
					},
				},
			},
		},
	}
}
