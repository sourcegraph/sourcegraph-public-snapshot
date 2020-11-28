package definitions

import "github.com/sourcegraph/sourcegraph/monitoring/monitoring"

func Symbols() *monitoring.Container {
	return &monitoring.Container{
		Name:        "symbols",
		Title:       "Symbols",
		Description: "Handles symbol searches for unindexed branches.",
		Groups: []monitoring.Group{
			{
				Title: "General",
				Rows: []monitoring.Row{
					{
						{
							Name:              "store_fetch_failures",
							Description:       "store fetch failures every 5m",
							Query:             `sum(increase(symbols_store_fetch_failed[5m]))`,
							DataMayNotExist:   true,
							Warning:           monitoring.Alert().GreaterOrEqual(5),
							PanelOptions:      monitoring.PanelOptions().LegendFormat("failures"),
							Owner:             monitoring.ObservableOwnerCodeIntel,
							PossibleSolutions: "none",
						},
						{
							Name:              "current_fetch_queue_size",
							Description:       "current fetch queue size",
							Query:             `sum(symbols_store_fetch_queue_size)`,
							DataMayNotExist:   true,
							Warning:           monitoring.Alert().GreaterOrEqual(25),
							PanelOptions:      monitoring.PanelOptions().LegendFormat("size"),
							Owner:             monitoring.ObservableOwnerCodeIntel,
							PossibleSolutions: "none",
						},
					},
					{
						sharedFrontendInternalAPIErrorResponses("symbols", monitoring.ObservableOwnerCodeIntel),
					},
				},
			},
			{
				Title:  "Container monitoring (not available on server)",
				Hidden: true,
				Rows: []monitoring.Row{
					{
						sharedContainerCPUUsage("symbols", monitoring.ObservableOwnerCodeIntel),
						sharedContainerMemoryUsage("symbols", monitoring.ObservableOwnerCodeIntel),
					},
					{
						sharedContainerRestarts("symbols", monitoring.ObservableOwnerCodeIntel),
						sharedContainerFsInodes("symbols", monitoring.ObservableOwnerCodeIntel),
					},
				},
			},
			{
				Title:  "Provisioning indicators (not available on server)",
				Hidden: true,
				Rows: []monitoring.Row{
					{
						sharedProvisioningCPUUsageLongTerm("symbols", monitoring.ObservableOwnerCodeIntel),
						sharedProvisioningMemoryUsageLongTerm("symbols", monitoring.ObservableOwnerCodeIntel),
					},
					{
						sharedProvisioningCPUUsageShortTerm("symbols", monitoring.ObservableOwnerCodeIntel),
						sharedProvisioningMemoryUsageShortTerm("symbols", monitoring.ObservableOwnerCodeIntel),
					},
				},
			},
			{
				Title:  "Golang runtime monitoring",
				Hidden: true,
				Rows: []monitoring.Row{
					{
						sharedGoGoroutines("symbols", monitoring.ObservableOwnerCodeIntel),
						sharedGoGcDuration("symbols", monitoring.ObservableOwnerCodeIntel),
					},
				},
			},
			{
				Title:  "Kubernetes monitoring (ignore if using Docker Compose or server)",
				Hidden: true,
				Rows: []monitoring.Row{
					{
						sharedKubernetesPodsAvailable("symbols", monitoring.ObservableOwnerCodeIntel),
					},
				},
			},
		},
	}
}
