package definitions

import (
	"github.com/sourcegraph/sourcegraph/monitoring/definitions/shared"
	"github.com/sourcegraph/sourcegraph/monitoring/monitoring"
)

func Worker() *monitoring.Container {
	return &monitoring.Container{
		Name:        "worker",
		Title:       "Worker",
		Description: "Manages background processes.",
		Groups: []monitoring.Group{
			{
				Title:  "Internal service requests",
				Hidden: true,
				Rows: []monitoring.Row{
					{
						shared.FrontendInternalAPIErrorResponses("worker", monitoring.ObservableOwnerCodeIntel).Observable(),
					},
				},
			},
			{
				Title:  shared.TitleContainerMonitoring,
				Hidden: true,
				Rows: []monitoring.Row{
					{
						shared.ContainerCPUUsage("worker", monitoring.ObservableOwnerCodeIntel).Observable(),
						shared.ContainerMemoryUsage("worker", monitoring.ObservableOwnerCodeIntel).Observable(),
					},
					{
						shared.ContainerMissing("worker", monitoring.ObservableOwnerCodeIntel).Observable(),
					},
				},
			},
			{
				Title:  shared.TitleProvisioningIndicators,
				Hidden: true,
				Rows: []monitoring.Row{
					{
						shared.ProvisioningCPUUsageLongTerm("worker", monitoring.ObservableOwnerCodeIntel).Observable(),
						shared.ProvisioningMemoryUsageLongTerm("worker", monitoring.ObservableOwnerCodeIntel).Observable(),
					},
					{
						shared.ProvisioningCPUUsageShortTerm("worker", monitoring.ObservableOwnerCodeIntel).Observable(),
						shared.ProvisioningMemoryUsageShortTerm("worker", monitoring.ObservableOwnerCodeIntel).Observable(),
					},
				},
			},
			{
				Title:  shared.TitleGolangMonitoring,
				Hidden: true,
				Rows: []monitoring.Row{
					{
						shared.GoGoroutines("worker", monitoring.ObservableOwnerCodeIntel).Observable(),
						shared.GoGcDuration("worker", monitoring.ObservableOwnerCodeIntel).Observable(),
					},
				},
			},
			{
				Title:  shared.TitleKubernetesMonitoring,
				Hidden: true,
				Rows: []monitoring.Row{
					{
						shared.KubernetesPodsAvailable("worker", monitoring.ObservableOwnerCodeIntel).Observable(),
					},
				},
			},
		},
	}
}
