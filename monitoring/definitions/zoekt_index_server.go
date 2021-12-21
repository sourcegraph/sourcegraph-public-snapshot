package definitions

import (
	"time"

	"github.com/grafana-tools/sdk"

	"github.com/sourcegraph/sourcegraph/monitoring/definitions/shared"
	"github.com/sourcegraph/sourcegraph/monitoring/monitoring"
)

func ZoektIndexServer() *monitoring.Container {
	const (
		containerName        = "zoekt-indexserver"
		bundledContainerName = "indexed-search"
	)

	return &monitoring.Container{
		Name: "zoekt-indexserver",

		Title:                    "Zoekt Index Server",
		Description:              "Indexes repositories and populates the search index.",
		NoSourcegraphDebugServer: true,
		Templates: []sdk.TemplateVar{
			{
				Label:      "Instance",
				Name:       "instance",
				Type:       "query",
				Datasource: monitoring.StringPtr("Prometheus"),
				Query:      "label_values(index_num_assigned, instance)",
				Multi:      true,
				Refresh:    sdk.BoolInt{Flag: true, Value: monitoring.Int64Ptr(2)}, // Refresh on time range change
				Sort:       3,
				IncludeAll: true,
				AllValue:   ".*",
				Current:    sdk.Current{Text: &sdk.StringSliceString{Value: []string{"all"}, Valid: true}, Value: "$__all"},
			},
		},
		Groups: []monitoring.Group{
			{
				Title: "General",
				Rows: []monitoring.Row{
					{
						{
							Name:        "total_repos_aggregate",
							Description: "total number of repos (aggregate)",
							Query:       `sum(index_num_assigned)`,
							NoAlert:     true,
							Panel: monitoring.Panel().With(func(o monitoring.Observable, p *sdk.Panel) {
								p.GraphPanel.Legend.Current = true
								p.GraphPanel.Legend.RightSide = true
								p.GraphPanel.Targets = []sdk.Target{{
									Expr:         o.Query,
									LegendFormat: "assigned",
								}, {
									Expr:         "sum(index_num_indexed)",
									LegendFormat: "indexed",
								}, {
									Expr:         "sum(index_queue_cap)",
									LegendFormat: "tracked",
								}}
								p.GraphPanel.Tooltip.Shared = true
							}),
							Owner: monitoring.ObservableOwnerSearchCore,
							Interpretation: `
								Sudden changes can be caused by indexing configuration changes.

								Additionally, a discrepancy between "assigned" and "tracked" could indicate a bug.

								Legend:
								- assigned: # of repos assigned to Zoekt
								- indexed: # of repos Zoekt has indexed
								- tracked: # of repos Zoekt is aware of, including those that it has finished indexing
							`,
						},
						{
							Name:        "total_repos_per_instance",
							Description: "total number of repos (per instance)",
							Query:       "sum by (instance) (index_num_assigned{instance=~`${instance:regex}`})",
							NoAlert:     true,
							Panel: monitoring.Panel().With(func(o monitoring.Observable, p *sdk.Panel) {
								p.GraphPanel.Legend.Current = true
								p.GraphPanel.Legend.RightSide = true
								p.GraphPanel.Targets = []sdk.Target{{
									Expr:         o.Query,
									LegendFormat: "{{instance}} assigned",
								}, {
									Expr:         "sum by (instance) (index_num_indexed{instance=~`${instance:regex}`})",
									LegendFormat: "{{instance}} indexed",
								}, {
									Expr:         "sum by (instance) (index_queue_cap{instance=~`${instance:regex}`})",
									LegendFormat: "{{instance}} tracked",
								}}
								p.GraphPanel.Tooltip.Shared = true
							}),
							Owner: monitoring.ObservableOwnerSearchCore,
							Interpretation: `
								Sudden changes can be caused by indexing configuration changes.

								Additionally, a discrepancy between "assigned" and "tracked" could indicate a bug.

								Legend:
								- assigned: # of repos assigned to Zoekt
								- indexed: # of repos Zoekt has indexed
								- tracked: # of repos Zoekt is aware of, including those that it has finished processing
							`,
						},
					},
					{
						{
							Name:        "repo_index_success_speed",
							Description: "successful indexing durations",
							Query:       `sum by (le, state) (increase(index_repo_seconds_bucket{state="success"}[$__rate_interval]))`,
							NoAlert:     true,
							Panel: monitoring.PanelHeatmap().With(func(o monitoring.Observable, p *sdk.Panel) {
								p.HeatmapPanel.YAxis.Format = string(monitoring.Seconds)
							}),
							Owner:          monitoring.ObservableOwnerSearchCore,
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
							Owner:          monitoring.ObservableOwnerSearchCore,
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
							Owner:             monitoring.ObservableOwnerSearchCore,
							PossibleSolutions: "none",
						},
						{
							Name:        "get_index_options_error_increase",
							Description: "the number of repositories we failed to get indexing options over 5m",
							Query:       `sum(increase(get_index_options_error_total[5m]))`,
							// This value can spike, so only if we have a
							// sustained error rate do we alert.
							Warning:  monitoring.Alert().GreaterOrEqual(100, nil).For(time.Minute),
							Critical: monitoring.Alert().GreaterOrEqual(100, nil).For(20 * time.Minute),
							Panel:    monitoring.Panel().Min(0),
							Owner:    monitoring.ObservableOwnerSearchCore,
							PossibleSolutions: `
								- View error rates on gitserver and frontend to identify root cause.
								- Rollback frontend/gitserver deployment if due to a bad code change.
								- View error logs for 'getIndexOptions' via net/trace debug interface. For example click on a 'indexed-search-indexer-' on https://sourcegraph.com/-/debug/. Then click on Traces. Replace sourcegraph.com with your instance address.
							`,
							Interpretation: `
								When considering indexing a repository we ask for the index configuration
								from frontend per repository. The most likely reason this would fail is
								failing to resolve branch names to git SHAs.

								This value can spike up during deployments/etc. Only if you encounter
								sustained periods of errors is there an underlying issue. When sustained
								this indicates repositories will not get updated indexes.
							`,
						},
					},
				},
			},
			{
				Title: "Indexing results",
				Rows: []monitoring.Row{
					{
						{
							Name:        "repo_index_state_aggregate",
							Description: "index results state count over 5m (aggregate)",
							Query:       "sum by (state) (increase(index_repo_seconds_count[5m]))",
							NoAlert:     true,
							Owner:       monitoring.ObservableOwnerSearchCore,
							Panel: monitoring.Panel().LegendFormat("{{state}}").With(func(o monitoring.Observable, p *sdk.Panel) {
								p.GraphPanel.Legend.RightSide = true
								p.GraphPanel.Yaxes[0].LogBase = 2  // log to show the huge number of "noop" or "empty"
								p.GraphPanel.Tooltip.Shared = true // show multiple lines simultaneously
							}),
							Interpretation: `
							This dashboard shows the outcomes of recently completed indexing jobs across all index-server instances.

							A persistent failing state indicates some repositories cannot be indexed, perhaps due to size and timeouts.

							Legend:
							- fail -> the indexing jobs failed
							- success -> the indexing job succeeded and the index was updated
							- success_meta -> the indexing job succeeded, but only metadata was updated
							- noop -> the indexing job succeed, but we didn't need to update anything
							- empty -> the indexing job succeeded, but the index was empty (i.e. the repository is empty)
						`,
						},
						{
							Name:        "repo_index_state_per_instance",
							Description: "index results state count over 5m (per instance)",
							Query:       "sum by (instance, state) (increase(index_repo_seconds_count{instance=~`${instance:regex}`}[5m]))",
							NoAlert:     true,
							Owner:       monitoring.ObservableOwnerSearchCore,
							Panel: monitoring.Panel().LegendFormat("{{instance}} {{state}}").With(func(o monitoring.Observable, p *sdk.Panel) {
								p.GraphPanel.Legend.RightSide = true
								p.GraphPanel.Yaxes[0].LogBase = 2  // log to show the huge number of "noop" or "empty"
								p.GraphPanel.Tooltip.Shared = true // show multiple lines simultaneously
							}),
							Interpretation: `
							This dashboard shows the outcomes of recently completed indexing jobs, split out across each index-server instance.

							(You can use the "instance" filter at the top of the page to select a particular instance.)

							A persistent failing state indicates some repositories cannot be indexed, perhaps due to size and timeouts.

							Legend:
							- fail -> the indexing jobs failed
							- success -> the indexing job succeeded and the index was updated
							- success_meta -> the indexing job succeeded, but only metadata was updated
							- noop -> the indexing job succeed, but we didn't need to update anything
							- empty -> the indexing job succeeded, but the index was empty (i.e. the repository is empty)
						`,
						},
					},
				},
			},
			{
				Title: "Indexing queue statistics",
				Rows: []monitoring.Row{
					{
						{
							Name:           "indexed_queue_size_aggregate",
							Description:    "# of outstanding index jobs (aggregate)",
							Query:          "sum(index_queue_len)", // total queue size amongst all index-server replicas
							NoAlert:        true,
							Panel:          monitoring.Panel().LegendFormat("jobs"),
							Owner:          monitoring.ObservableOwnerSearchCore,
							Interpretation: "A queue that is constantly growing could be a leading indicator of a bottleneck or under-provisioning",
						},
						{
							Name:           "indexed_queue_size_per_instance",
							Description:    "# of outstanding index jobs (per instance)",
							Query:          "index_queue_len{instance=~`${instance:regex}`}",
							NoAlert:        true,
							Panel:          monitoring.Panel().LegendFormat("{{instance}} jobs"),
							Owner:          monitoring.ObservableOwnerSearchCore,
							Interpretation: "A queue that is constantly growing could be a leading indicator of a bottleneck or under-provisioning",
						},
					},
				},
			},

			// Note:
			// zoekt_indexserver and zoekt_webserver are deployed together as part of the indexed-search service
			// We show pod availability here for both the webserver and indexserver as they are bundled together.

			shared.NewContainerMonitoringGroup(containerName, monitoring.ObservableOwnerSearchCore, nil),
			shared.NewProvisioningIndicatorsGroup(containerName, monitoring.ObservableOwnerSearchCore, nil),
			shared.NewKubernetesMonitoringGroup(bundledContainerName, monitoring.ObservableOwnerSearchCore, nil),
		},
	}
}
