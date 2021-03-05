package definitions

import (
	"github.com/sourcegraph/sourcegraph/monitoring/definitions/shared"
	"github.com/sourcegraph/sourcegraph/monitoring/monitoring"
)

func ZoektIndexServer() *monitoring.Container {
	return &monitoring.Container{
		Name:        "zoekt-indexserver",
		Title:       "Zoekt Index Server",
		Description: "Indexes repositories and populates the search index.",
		Groups: []monitoring.Group{
			{
				Title: "General",
				Rows: []monitoring.Row{
					{
						{
							Name:              "average_resolve_revision_duration",
							Description:       "average resolve revision duration over 5m",
							Query:             `sum(rate(resolve_revision_seconds_sum[5m])) / sum(rate(resolve_revision_seconds_count[5m]))`,
							Warning:           monitoring.Alert().GreaterOrEqual(15, nil),
							Critical:          monitoring.Alert().GreaterOrEqual(30, nil),
							Panel:             monitoring.Panel().LegendFormat("{{duration}}").Unit(monitoring.Seconds),
							Owner:             monitoring.ObservableOwnerSearch,
							PossibleSolutions: "none",
						},
					},
				},
			},
			{
				Title:  shared.TitleContainerMonitoring,
				Hidden: true,
				Rows: []monitoring.Row{
					{
						shared.ContainerCPUUsage("zoekt-indexserver", monitoring.ObservableOwnerSearch).Observable(),
						shared.ContainerMemoryUsage("zoekt-indexserver", monitoring.ObservableOwnerSearch).Observable(),
					},
					{
						shared.ContainerMissing("zoekt-indexserver", monitoring.ObservableOwnerSearch).Observable(),
						shared.ContainerIOUsage("zoekt-indexserver", monitoring.ObservableOwnerSearch).Observable(),
					},
				},
			},
			{
				Title:  shared.TitleProvisioningIndicators,
				Hidden: true,
				Rows: []monitoring.Row{
					{
						shared.ProvisioningCPUUsageLongTerm("zoekt-indexserver", monitoring.ObservableOwnerSearch).Observable(),
						shared.ProvisioningMemoryUsageLongTerm("zoekt-indexserver", monitoring.ObservableOwnerSearch).Observable(),
					},
					{
						shared.ProvisioningCPUUsageShortTerm("zoekt-indexserver", monitoring.ObservableOwnerSearch).Observable(),
						shared.ProvisioningMemoryUsageShortTerm("zoekt-indexserver", monitoring.ObservableOwnerSearch).Observable(),
					},
				},
			},
			{
				Title:  shared.TitleKubernetesMonitoring,
				Hidden: true,
				Rows: []monitoring.Row{
					{
						// zoekt_index_server, zoekt_web_server are deployed together
						// as part of the indexed-search service, so only show pod
						// availability here.
						shared.KubernetesPodsAvailable("indexed-search", monitoring.ObservableOwnerSearch).Observable(),
					},
				},
			},
		},

		NoSourcegraphDebugServer: true,
	}
}
