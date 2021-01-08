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
							Warning:           monitoring.Alert().GreaterOrEqual(5),
							PanelOptions:      monitoring.PanelOptions().LegendFormat("failures"),
							Owner:             monitoring.ObservableOwnerCodeIntel,
							PossibleSolutions: "none",
						},
						{
							Name:              "current_fetch_queue_size",
							Description:       "current fetch queue size",
							Query:             `sum(symbols_store_fetch_queue_size)`,
							Warning:           monitoring.Alert().GreaterOrEqual(25),
							PanelOptions:      monitoring.PanelOptions().LegendFormat("size"),
							Owner:             monitoring.ObservableOwnerCodeIntel,
							PossibleSolutions: "none",
						},
					},
					{
						shared.FrontendInternalAPIErrorResponses("symbols", monitoring.ObservableOwnerCodeIntel).Observable(),
					},
				},
			},
			{
				Title:  "Container monitoring (not available on server)",
				Hidden: true,
				Rows: []monitoring.Row{
					{
						shared.ContainerCPUUsage("symbols", monitoring.ObservableOwnerCodeIntel).Observable(),
						shared.ContainerMemoryUsage("symbols", monitoring.ObservableOwnerCodeIntel).Observable(),
					},
					{
						shared.ContainerRestarts("symbols", monitoring.ObservableOwnerCodeIntel).Observable(),
						shared.ContainerFsInodes("symbols", monitoring.ObservableOwnerCodeIntel).Observable(),
					},
				},
			},
			{
				Title:  "Provisioning indicators (not available on server)",
				Hidden: true,
				Rows: []monitoring.Row{
					{
						shared.ProvisioningCPUUsageLongTerm("symbols", monitoring.ObservableOwnerCodeIntel).Observable(),
						shared.ProvisioningMemoryUsageLongTerm("symbols", monitoring.ObservableOwnerCodeIntel).Observable(),
					},
					{
						shared.ProvisioningCPUUsageShortTerm("symbols", monitoring.ObservableOwnerCodeIntel).Observable(),
						shared.ProvisioningMemoryUsageShortTerm("symbols", monitoring.ObservableOwnerCodeIntel).Observable(),
					},
				},
			},
			{
				Title:  "Golang runtime monitoring",
				Hidden: true,
				Rows: []monitoring.Row{
					{
						shared.GoGoroutines("symbols", monitoring.ObservableOwnerCodeIntel).Observable(),
						shared.GoGcDuration("symbols", monitoring.ObservableOwnerCodeIntel).Observable(),
					},
				},
			},
			{
				Title:  "Kubernetes monitoring (ignore if using Docker Compose or server)",
				Hidden: true,
				Rows: []monitoring.Row{
					{
						shared.KubernetesPodsAvailable("symbols", monitoring.ObservableOwnerCodeIntel).Observable(),
					},
				},
			},
		},
	}
}
