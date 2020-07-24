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
							Name:              "average_resolve_revision_duration",
							Description:       "average resolve revision duration over 5m",
							Query:             `sum(rate(resolve_revision_seconds_sum[5m])) / sum(rate(resolve_revision_seconds_count[5m]))`,
							DataMayNotExist:   true,
							Warning:           Alert{GreaterOrEqual: 15},
							Critical:          Alert{GreaterOrEqual: 30},
							PanelOptions:      PanelOptions().LegendFormat("{{duration}}").Unit(Seconds),
							Owner:             ObservableOwnerSearch,
							PossibleSolutions: "none",
						},
					},
				},
			},
			{
				Title:  "Container monitoring (not available on server)",
				Hidden: true,
				Rows: []Row{
					{
						sharedContainerRestarts("zoekt-indexserver"),
						sharedContainerMemoryUsage("zoekt-indexserver"),
						sharedContainerCPUUsage("zoekt-indexserver"),
					},
				},
			},
			{
				Title:  "Provisioning indicators (not available on server)",
				Hidden: true,
				Rows: []Row{
					{
						sharedProvisioningCPUUsage7d("zoekt-indexserver"),
						sharedProvisioningMemoryUsage7d("zoekt-indexserver"),
					},
					{
						sharedProvisioningCPUUsage5m("zoekt-indexserver"),
						sharedProvisioningMemoryUsage5m("zoekt-indexserver"),
					},
				},
			},
		},
	}
}
