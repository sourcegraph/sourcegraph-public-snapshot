package definitions

import (
	"github.com/grafana-tools/sdk"

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
							Name:        "repos_assigned",
							Description: "total number of repos",
							Query:       `sum(index_num_assigned)`,
							NoAlert:     true,
							Panel: monitoring.Panel().With(func(o monitoring.Observable, p *sdk.Panel) {
								p.GraphPanel.Legend.Current = true
								p.GraphPanel.Targets = []sdk.Target{{
									Expr:         o.Query,
									LegendFormat: "assigned",
								}, {
									Expr:         `sum(index_num_indexed)`,
									LegendFormat: "indexed",
								}}
								p.GraphPanel.Tooltip.Shared = true
							}),
							Owner:          monitoring.ObservableOwnerSearch,
							Interpretation: "Sudden changes should be caused by indexing configuration changes.",
						},
						{
							Name:           "repos_priorities",
							Description:    "total number of repos with priorities for ranking",
							Query:          `sum(index_priorities_total)`,
							NoAlert:        true,
							Panel:          monitoring.Panel(),
							Owner:          monitoring.ObservableOwnerSearch,
							Interpretation: "Sudden changes should be caused by indexing configuration changes.",
						},
					},
					{
						{
							Name:        "repo_index_state",
							Description: "indexing results over 5m (noop=no changes, empty=no branches to index)",
							Query:       `sum by (state) (increase(index_repo_seconds_bucket[5m]))`,
							NoAlert:     true,
							Owner:       monitoring.ObservableOwnerSearch,
							Panel: monitoring.Panel().LegendFormat("{{state}}").With(func(o monitoring.Observable, p *sdk.Panel) {
								p.GraphPanel.Yaxes[0].LogBase = 2  // log to show the huge number of "noop" or "empty"
								p.GraphPanel.Tooltip.Shared = true // show multiple lines simultaneously
							}),
							Interpretation: "A persistent failing state indicates some repositories cannot be indexed, perhaps due to size and timeouts.",
						},
						{
							Name:        "repo_index_success_speed",
							Description: "successful indexing durations",
							Query:       `sum by (le, state) (increase(index_repo_seconds_bucket{state="success"}[$__rate_interval]))`,
							NoAlert:     true,
							Panel: monitoring.PanelHeatmap().With(func(o monitoring.Observable, p *sdk.Panel) {
								p.HeatmapPanel.YAxis.Format = string(monitoring.Seconds)
							}),
							Owner:          monitoring.ObservableOwnerSearch,
							Interpretation: "Latency increases can indicate bottlenecks in the indexserver.",
						},
						{
							Name:        "repo_index_fail_speed",
							Description: "failed indexing durations",
							Query:       `sum by (le, state) (increase(index_repo_seconds_bucket{state="fail"}[$__rate_interval]))`,
							NoAlert:     true,
							Panel: monitoring.PanelHeatmap().With(func(o monitoring.Observable, p *sdk.Panel) {
								p.HeatmapPanel.YAxis.Format = string(monitoring.Seconds)
							}),
							Owner:          monitoring.ObservableOwnerSearch,
							Interpretation: "Failures happening after a long time indicates timeouts.",
						},
					},
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
