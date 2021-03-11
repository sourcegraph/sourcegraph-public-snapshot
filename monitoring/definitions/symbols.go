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
							Warning:           monitoring.Alert().GreaterOrEqual(5, nil),
							Panel:             monitoring.Panel().LegendFormat("failures"),
							Owner:             monitoring.ObservableOwnerCodeIntel,
							PossibleSolutions: "none",
						},
						{
							Name:              "current_fetch_queue_size",
							Description:       "current fetch queue size",
							Query:             `sum(symbols_store_fetch_queue_size)`,
							Warning:           monitoring.Alert().GreaterOrEqual(25, nil),
							Panel:             monitoring.Panel().LegendFormat("size"),
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
				Title:  shared.TitleContainerMonitoring,
				Hidden: true,
				Rows: []monitoring.Row{
					{
						shared.ContainerCPUUsage("symbols", monitoring.ObservableOwnerCodeIntel).Observable(),
						shared.ContainerMemoryUsage("symbols", monitoring.ObservableOwnerCodeIntel).Observable(),
					},
					{
						shared.ContainerMissing("symbols", monitoring.ObservableOwnerCodeIntel).Observable(),
					},
				},
			},
			{
				Title:  shared.TitleProvisioningIndicators,
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
				Title:  shared.TitleGolangMonitoring,
				Hidden: true,
				Rows: []monitoring.Row{
					{
						shared.GoGoroutines("symbols", monitoring.ObservableOwnerCodeIntel).Observable(),
						shared.GoGcDuration("symbols", monitoring.ObservableOwnerCodeIntel).Observable(),
					},
				},
			},
			{
				Title:  shared.TitleKubernetesMonitoring,
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
