package definitions

import (
	"fmt"

	"github.com/grafana-tools/sdk"

	"github.com/sourcegraph/sourcegraph/monitoring/definitions/shared"
	"github.com/sourcegraph/sourcegraph/monitoring/monitoring"
)

func Symbols() *monitoring.Dashboard {
	const (
		containerName   = "symbols"
		grpcServiceName = "symbols.v1.SymbolsService"
	)

	scrapeJobRegex := fmt.Sprintf(".*%s", containerName)

	grpcMethodVariable := shared.GRPCMethodVariable("symbols", grpcServiceName)

	return &monitoring.Dashboard{
		Name:        "symbols",
		Title:       "Symbols",
		Description: "Handles symbol searches for unindexed branches.",
		Variables: []monitoring.ContainerVariable{
			{
				Label: "instance",
				Name:  "instance",
				OptionsLabelValues: monitoring.ContainerVariableOptionsLabelValues{
					Query:         "src_codeintel_symbols_fetching",
					LabelName:     "instance",
					ExampleOption: "symbols-0:3184",
				},
				Multi: true,
			},
			grpcMethodVariable,
		},
		Groups: []monitoring.Group{
			shared.CodeIntelligence.NewSymbolsAPIGroup(containerName),
			shared.CodeIntelligence.NewSymbolsParserGroup(containerName),
			shared.CodeIntelligence.NewSymbolsCacheJanitorGroup(containerName),
			shared.CodeIntelligence.NewSymbolsRepositoryFetcherGroup(containerName),
			shared.CodeIntelligence.NewSymbolsGitserverClientGroup(containerName),

			{
				Title:  "Rockskip",
				Hidden: true,
				Rows: []monitoring.Row{
					{
						{
							Name:        "p95_rockskip_search_request_duration",
							Description: "95th percentile search request duration over 5m",
							Query:       "histogram_quantile(0.95, sum(rate(src_rockskip_service_search_request_duration_seconds_bucket[5m])) by (le))",
							Panel:       monitoring.Panel().LegendFormat("duration").Unit(monitoring.Seconds),
							Owner:       monitoring.ObservableOwnerSearchCore,
							NoAlert:     true,
							Interpretation: `
								The 95th percentile duration of search requests to Rockskip in seconds. Lower is better.`,
						},
						{
							Name:        "rockskip_in_flight_search_requests",
							Description: "number of in-flight search requests",
							Query:       `sum(src_rockskip_service_in_flight_search_requests)`,
							Panel:       monitoring.Panel().LegendFormat("requests"),
							Owner:       monitoring.ObservableOwnerSearchCore,
							NoAlert:     true,
							Interpretation: `
								The number of search requests currently being processed by Rockskip.
								If there is not much traffic and the requests are served very fast relative to the polling window of Prometheus,
								it possible that that this number is 0 even if there are search requests being processed.`,
						},
						{
							Name:        "rockskip_search_request_errors",
							Description: "search request errors every 5m",
							Query:       `sum(increase(src_rockskip_service_search_request_errors[5m]))`,
							Panel:       monitoring.Panel().LegendFormat("errors"),
							Owner:       monitoring.ObservableOwnerSearchCore,
							NoAlert:     true,
							Interpretation: `
								The number of search requests that returned an error in the last 5 minutes.
								The errors tracked here are all application errors, grpc errors are not included.
								We generally want this to be 0.`,
						},
					},
					{
						{
							Name:        "p95_rockskip_index_job_duration",
							Description: "95th percentile index job duration over 5m",
							Query:       "histogram_quantile(0.95, sum(rate(src_rockskip_service_index_job_duration_seconds_bucket[5m])) by (le))",
							Panel: monitoring.Panel().LegendFormat("duration").Unit(monitoring.Seconds).With(
								func(o monitoring.Observable, p *sdk.Panel) {
									p.GraphPanel.Yaxes[0].LogBase = 2 // log to account for huge range of "new" vs "delta" index.
								}),
							Owner:   monitoring.ObservableOwnerSearchCore,
							NoAlert: true,
							Interpretation: `
								The 95th percentile duration of index jobs in seconds.
								The range of values is very large, because the metric measure quick delta updates as well as full index jobs.
								Lower is better.`,
						},
						{
							Name:        "rockskip_in_flight_index_jobs",
							Description: "number of in-flight index jobs",
							Query:       `sum(src_rockskip_service_in_flight_index_jobs)`,
							Panel:       monitoring.Panel().LegendFormat("jobs"),
							Owner:       monitoring.ObservableOwnerSearchCore,
							NoAlert:     true,
							Interpretation: `
								The number of index jobs currently being processed by Rockskip.
								This includes delta updates as well as full index jobs.`,
						},
						{
							Name:        "rockskip_index_job_errors",
							Description: "index job errors every 5m",
							Query:       `sum(increase(src_rockskip_service_index_job_errors[5m]))`,
							Panel:       monitoring.Panel().LegendFormat("errors"),
							Owner:       monitoring.ObservableOwnerSearchCore,
							NoAlert:     true,
							Interpretation: `
								The number of index jobs that returned an error in the last 5 minutes.
								If the errors are persistent, users will see alerts in the UI.
								The service logs will contain more detailed information about the kind of errors.
								We generally want this to be 0.`,
						},
					},
					{
						{
							Name:        "rockskip_number_of_repos_indexed",
							Description: "number of repositories indexed by Rockskip",
							Query:       `max(src_rockskip_service_repos_indexed)`, // "max" is used as hack to show only one value instead one per instance
							Panel:       monitoring.Panel().LegendFormat("num_indexed"),
							Owner:       monitoring.ObservableOwnerSearchCore,
							NoAlert:     true,
							Interpretation: `
								The number of repositories indexed by Rockskip.
								Apart from an initial transient phase in which many repos are being indexed,
								this number should be low and relatively stable and only increase by small increments.
								To verify if this number makes sense, compare ROCKSKIP_MIN_REPO_SIZE_MB with the repository sizes reported by gitserver_repos table.`,
						},
						{
							Name:        "p99.9_rockskip_index_queue_age",
							Description: "99.9th percentile index queue delay over 5m",
							Query:       "histogram_quantile(0.999, sum(rate(src_rockskip_service_index_queue_age_seconds_bucket[5m])) by (le))",
							Panel:       monitoring.Panel().LegendFormat("age").Unit(monitoring.Seconds),
							Owner:       monitoring.ObservableOwnerSearchCore,
							NoAlert:     true,
							Interpretation: `
								The 99.9th percentile age of index jobs in seconds
								This 99.9th percentile is useful to catch the long tail of queueing delays.`,
						},
						{
							Name:        "p95_rockskip_index_queue_age",
							Description: "95th percentile index queue delay over 5m",
							Query:       "histogram_quantile(0.95, sum(rate(src_rockskip_service_index_queue_age_seconds_bucket[5m])) by (le))",
							Panel:       monitoring.Panel().LegendFormat("age").Unit(monitoring.Seconds),
							Owner:       monitoring.ObservableOwnerSearchCore,
							NoAlert:     true,
							Interpretation: `
								The 95th percentile age of index jobs in seconds.
								A high delay might indicate a resource issue.
								Consider increasing indexing bandwidth by either increasing the number of queues or the number of symbol services.`,
						},
					},
				},
			},

			shared.NewGRPCServerMetricsGroup(
				shared.GRPCServerMetricsOptions{
					HumanServiceName:   containerName,
					RawGRPCServiceName: grpcServiceName,

					MethodFilterRegex:    fmt.Sprintf("${%s:regex}", grpcMethodVariable.Name),
					InstanceFilterRegex:  `${instance:regex}`,
					MessageSizeNamespace: "src",
				}, monitoring.ObservableOwnerCodeIntel),

			shared.NewGRPCInternalErrorMetricsGroup(
				shared.GRPCInternalErrorMetricsOptions{
					HumanServiceName:   containerName,
					RawGRPCServiceName: grpcServiceName,
					Namespace:          "src",

					MethodFilterRegex: fmt.Sprintf("${%s:regex}", grpcMethodVariable.Name),
				}, monitoring.ObservableOwnerCodeIntel),

			shared.NewGRPCRetryMetricsGroup(
				shared.GRPCRetryMetricsOptions{
					HumanServiceName:   containerName,
					RawGRPCServiceName: grpcServiceName,
					Namespace:          "src",

					MethodFilterRegex: fmt.Sprintf("${%s:regex}", grpcMethodVariable.Name),
				}, monitoring.ObservableOwnerCodeIntel),

			shared.NewSiteConfigurationClientMetricsGroup(shared.SiteConfigurationMetricsOptions{
				HumanServiceName:    "symbols",
				InstanceFilterRegex: `${instance:regex}`,
				JobFilterRegex:      scrapeJobRegex,
			}, monitoring.ObservableOwnerInfraOrg),
			shared.NewDatabaseConnectionsMonitoringGroup(containerName, monitoring.ObservableOwnerInfraOrg),
			shared.NewContainerMonitoringGroup(containerName, monitoring.ObservableOwnerCodeIntel, nil),
			shared.NewProvisioningIndicatorsGroup(containerName, monitoring.ObservableOwnerCodeIntel, nil),
			shared.NewGolangMonitoringGroup(containerName, monitoring.ObservableOwnerCodeIntel, nil),
			shared.NewKubernetesMonitoringGroup(containerName, monitoring.ObservableOwnerCodeIntel, nil),
		},
	}
}
