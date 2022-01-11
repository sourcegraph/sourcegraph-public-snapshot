package definitions

import (
	"fmt"
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
		Variables: []monitoring.ContainerVariable{
			{
				Label: "Instance",
				Name:  "instance",
				Query: "label_values(index_num_assigned, instance)",
				Multi: true,
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
			{
				Title:  "Compound shards (experimental)",
				Hidden: true,
				Rows: []monitoring.Row{
					{
						{
							Name:        "compound_shards_aggregate",
							Description: "# of compound shards (aggregate)",
							Query:       "sum(index_number_compound_shards) by (app)",
							NoAlert:     true,
							Panel:       monitoring.Panel().LegendFormat("aggregate").Unit(monitoring.Number),
							Owner:       monitoring.ObservableOwnerSearchCore,
							Interpretation: `
								The total number of compound shards aggregated over all instances.

								This number should be consistent if the number of indexed repositories doesn't change.
							`,
						},
						{
							Name:        "compound_shards_per_instance",
							Description: "# of compound shards (per instance)",
							Query:       "sum(index_number_compound_shards{instance=~`${instance:regex}`}) by (instance)",
							NoAlert:     true,
							Panel:       monitoring.Panel().LegendFormat("{{instance}}").Unit(monitoring.Number),
							Owner:       monitoring.ObservableOwnerSearchCore,
							Interpretation: `
								The total number of compound shards per instance.

								This number should be consistent if the number of indexed repositories doesn't change.
							`,
						},
					},
					{
						{
							Name:        "average_shard_merging_duration_success",
							Description: "average successful shard merging duration over 1 hour",
							Query:       "sum(rate(index_shard_merging_duration_seconds_sum{error=\"false\"}[1h])) / sum(rate(index_shard_merging_duration_seconds_count{error=\"false\"}[1h]))",
							NoAlert:     true,
							Panel:       monitoring.Panel().LegendFormat("average").Unit(monitoring.Seconds),
							Owner:       monitoring.ObservableOwnerSearchCore,
							Interpretation: `
								Average duration of a successful merge over the last hour.

								The duration depends on the target compound shard size. The larger the compound shard the longer a merge will take.
								Since the target compound shard size is set on start of zoekt-indexserver, the average duration should be consistent.
							`,
						},
						{
							Name:        "average_shard_merging_duration_error",
							Description: "average failed shard merging duration over 1 hour",
							Query:       "sum(rate(index_shard_merging_duration_seconds_sum{error=\"true\"}[1h])) / sum(rate(index_shard_merging_duration_seconds_count{error=\"true\"}[1h]))",
							NoAlert:     true,
							Panel:       monitoring.Panel().LegendFormat("duration").Unit(monitoring.Seconds),
							Owner:       monitoring.ObservableOwnerSearchCore,
							Interpretation: `
								Average duration of a failed merge over the last hour.

								This curve should be flat. Any deviation should be investigated.
							`,
						},
					},
					{
						{
							Name:        "shard_merging_errors_aggregate",
							Description: "number of errors during shard merging (aggregate)",
							Query:       "sum(index_shard_merging_duration_seconds_count{error=\"true\"}) by (app)",
							NoAlert:     true,
							Panel:       monitoring.Panel().LegendFormat("aggregate").Unit(monitoring.Number),
							Owner:       monitoring.ObservableOwnerSearchCore,
							Interpretation: `
								Number of errors during shard merging aggregated over all instances.
							`,
						},
						{
							Name:        "shard_merging_errors_per_instance",
							Description: "number of errors during shard merging (per instance)",
							Query:       "sum(index_shard_merging_duration_seconds_count{instance=~`${instance:regex}`, error=\"true\"}) by (instance)",
							NoAlert:     true,
							Panel:       monitoring.Panel().LegendFormat("{{instance}}").Unit(monitoring.Number),
							Owner:       monitoring.ObservableOwnerSearchCore,
							Interpretation: `
								Number of errors during shard merging per instance.
							`,
						},
					},
					{
						{
							Name:        "shard_merging_merge_running_per_instance",
							Description: "if shard merging is running (per instance)",
							Query:       "max by (instance) (index_shard_merging_running{instance=~`${instance:regex}`})",
							NoAlert:     true,
							Panel:       monitoring.Panel().LegendFormat("{{instance}}").Unit(monitoring.Number),
							Owner:       monitoring.ObservableOwnerSearchCore,
							Interpretation: `
								Set to 1 if shard merging is running.
							`,
						},
						{
							Name:        "shard_merging_vacuum_running_per_instance",
							Description: "if vacuum is running (per instance)",
							Query:       "max by (instance) (index_vacuum_running{instance=~`${instance:regex}`})",
							NoAlert:     true,
							Panel:       monitoring.Panel().LegendFormat("{{instance}}").Unit(monitoring.Number),
							Owner:       monitoring.ObservableOwnerSearchCore,
							Interpretation: `
								Set to 1 if vacuum is running.
							`,
						},
					},
				},
			},
			{
				Title:  "Network I/O pod metrics (only available on Kubernetes)",
				Hidden: true,
				Rows: []monitoring.Row{
					{
						{
							Name:        "network_sent_bytes_aggregate",
							Description: "transmission rate over 5m (aggregate)",
							Query:       fmt.Sprintf("sum(rate(container_network_transmit_bytes_total{%s}[5m]))", shared.CadvisorPodNameMatcher(bundledContainerName)),
							NoAlert:     true,
							Panel: monitoring.Panel().LegendFormat(bundledContainerName).Unit(monitoring.BytesPerSecond).With(func(o monitoring.Observable, p *sdk.Panel) {
								p.GraphPanel.Legend.RightSide = true
							}),
							Owner:          monitoring.ObservableOwnerSearchCore,
							Interpretation: "The rate of bytes sent over the network across all Zoekt pods",
						},
						{
							Name:        "network_received_packets_per_instance",
							Description: "transmission rate over 5m (per instance)",
							Query:       "sum by (container_label_io_kubernetes_pod_name) (rate(container_network_transmit_bytes_total{container_label_io_kubernetes_pod_name=~`${instance:regex}`}[5m]))",
							NoAlert:     true,
							Panel: monitoring.Panel().LegendFormat("{{container_label_io_kubernetes_pod_name}}").Unit(monitoring.BytesPerSecond).With(func(o monitoring.Observable, p *sdk.Panel) {
								p.GraphPanel.Legend.RightSide = true
							}),
							Owner:          monitoring.ObservableOwnerSearchCore,
							Interpretation: "The amount of bytes sent over the network by individual Zoekt pods",
						},
					},
					{
						{
							Name:        "network_received_bytes_aggregate",
							Description: "receive rate over 5m (aggregate)",
							Query:       fmt.Sprintf("sum(rate(container_network_receive_bytes_total{%s}[5m]))", shared.CadvisorPodNameMatcher(bundledContainerName)),
							NoAlert:     true,
							Panel: monitoring.Panel().LegendFormat(bundledContainerName).Unit(monitoring.BytesPerSecond).With(func(o monitoring.Observable, p *sdk.Panel) {
								p.GraphPanel.Legend.RightSide = true
							}),
							Owner:          monitoring.ObservableOwnerSearchCore,
							Interpretation: "The amount of bytes received from the network across Zoekt pods",
						},
						{
							Name:        "network_received_bytes_per_instance",
							Description: "receive rate over 5m (per instance)",
							Query:       "sum by (container_label_io_kubernetes_pod_name) (rate(container_network_receive_bytes_total{container_label_io_kubernetes_pod_name=~`${instance:regex}`}[5m]))",
							NoAlert:     true,
							Panel: monitoring.Panel().LegendFormat("{{container_label_io_kubernetes_pod_name}}").Unit(monitoring.BytesPerSecond).With(func(o monitoring.Observable, p *sdk.Panel) {
								p.GraphPanel.Legend.RightSide = true
							}),
							Owner:          monitoring.ObservableOwnerSearchCore,
							Interpretation: "The amount of bytes received from the network by individual Zoekt pods",
						},
					},
					{
						{
							Name:        "network_transmitted_packets_dropped_by_instance",
							Description: "transmit packet drop rate over 5m (by instance)",
							Query:       "sum by (container_label_io_kubernetes_pod_name) (rate(container_network_transmit_packets_dropped_total{container_label_io_kubernetes_pod_name=~`${instance:regex}`}[5m]))",
							NoAlert:     true,
							Panel: monitoring.Panel().LegendFormat("{{container_label_io_kubernetes_pod_name}}").Unit(monitoring.PacketsPerSecond).With(func(o monitoring.Observable, p *sdk.Panel) {
								p.GraphPanel.Legend.RightSide = true
							}),
							Owner:          monitoring.ObservableOwnerSearchCore,
							Interpretation: "An increase in dropped packets could be a leading indicator of network saturation.",
						},
						{
							Name:        "network_transmitted_packets_errors_per_instance",
							Description: "errors encountered while transmitting over 5m (per instance)",
							Query:       "sum by (container_label_io_kubernetes_pod_name) (rate(container_network_transmit_errors_total{container_label_io_kubernetes_pod_name=~`${instance:regex}`}[5m]))",
							NoAlert:     true,
							Panel: monitoring.Panel().LegendFormat("{{container_label_io_kubernetes_pod_name}} errors").With(func(o monitoring.Observable, p *sdk.Panel) {
								p.GraphPanel.Legend.RightSide = true
							}),
							Owner:          monitoring.ObservableOwnerSearchCore,
							Interpretation: "An increase in transmission errors could indicate a networking issue",
						},
						{
							Name:        "network_received_packets_dropped_by_instance",
							Description: "receive packet drop rate over 5m (by instance)",
							Query:       "sum by (container_label_io_kubernetes_pod_name) (rate(container_network_receive_packets_dropped_total{container_label_io_kubernetes_pod_name=~`${instance:regex}`}[5m]))",
							NoAlert:     true,
							Panel: monitoring.Panel().LegendFormat("{{container_label_io_kubernetes_pod_name}}").Unit(monitoring.PacketsPerSecond).With(func(o monitoring.Observable, p *sdk.Panel) {
								p.GraphPanel.Legend.RightSide = true
							}),
							Owner:          monitoring.ObservableOwnerSearchCore,
							Interpretation: "An increase in dropped packets could be a leading indicator of network saturation.",
						},
						{
							Name:        "network_transmitted_packets_errors_by_instance",
							Description: "errors encountered while receiving over 5m (per instance)",
							Query:       "sum by (container_label_io_kubernetes_pod_name) (rate(container_network_receive_errors_total{container_label_io_kubernetes_pod_name=~`${instance:regex}`}[5m]))",
							NoAlert:     true,
							Panel: monitoring.Panel().LegendFormat("{{container_label_io_kubernetes_pod_name}} errors").With(func(o monitoring.Observable, p *sdk.Panel) {
								p.GraphPanel.Legend.RightSide = true
							}),
							Owner:          monitoring.ObservableOwnerSearchCore,
							Interpretation: "An increase in errors while receiving could indicate a networking issue.",
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
