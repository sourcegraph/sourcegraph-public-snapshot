package definitions

import (
	"github.com/sourcegraph/sourcegraph/monitoring/definitions/shared"
	"github.com/sourcegraph/sourcegraph/monitoring/monitoring"
)

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
						shared.FrontendInternalAPIErrorResponses("symbols", monitoring.ObservableOwnerCodeIntel),
					},
				},
			},
			{
				Title:  "Container monitoring (not available on server)",
				Hidden: true,
				Rows: []monitoring.Row{
					{
						shared.ContainerCPUUsage("symbols", monitoring.ObservableOwnerCodeIntel),
						shared.ContainerMemoryUsage("symbols", monitoring.ObservableOwnerCodeIntel),
					},
					{
						shared.ContainerRestarts("symbols", monitoring.ObservableOwnerCodeIntel),
						shared.ContainerFsInodes("symbols", monitoring.ObservableOwnerCodeIntel),
					},
				},
			},
			{
				Title:  "Provisioning indicators (not available on server)",
				Hidden: true,
				Rows: []monitoring.Row{
					{
						shared.ProvisioningCPUUsageLongTerm("symbols", monitoring.ObservableOwnerCodeIntel),
						shared.ProvisioningMemoryUsageLongTerm("symbols", monitoring.ObservableOwnerCodeIntel),
					},
					{
						shared.ProvisioningCPUUsageShortTerm("symbols", monitoring.ObservableOwnerCodeIntel),
						shared.ProvisioningMemoryUsageShortTerm("symbols", monitoring.ObservableOwnerCodeIntel),
					},
				},
			},
			{
				Title:  "Golang runtime monitoring",
				Hidden: true,
				Rows: []monitoring.Row{
					{
						shared.GoGoroutines("symbols", monitoring.ObservableOwnerCodeIntel),
						shared.GoGcDuration("symbols", monitoring.ObservableOwnerCodeIntel),
					},
				},
			},
			{
				Title:  "Kubernetes monitoring (ignore if using Docker Compose or server)",
				Hidden: true,
				Rows: []monitoring.Row{
					{
						shared.KubernetesPodsAvailable("symbols", monitoring.ObservableOwnerCodeIntel),
					},
				},
			},
		},
	}
}
