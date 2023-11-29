package definitions

import (
	"fmt"
	"strings"
	"time"

	"github.com/sourcegraph/sourcegraph/monitoring/definitions/shared"
	"github.com/sourcegraph/sourcegraph/monitoring/monitoring"
)

func Searcher() *monitoring.Dashboard {
	const (
		containerName   = "searcher"
		grpcServiceName = "searcher.v1.SearcherService"
	)

	grpcMethodVariable := shared.GRPCMethodVariable("searcher", grpcServiceName)

	// instanceSelector is a helper for inserting the instance selector.
	// Should be used on strings created via `` since you can't escape in
	// those.
	instanceSelector := func(s string) string {
		return strings.ReplaceAll(s, "$$INSTANCE$$", "instance=~`${instance:regex}`")
	}

	return &monitoring.Dashboard{
		Name:        "searcher",
		Title:       "Searcher",
		Description: "Performs unindexed searches (diff and commit search, text search for unindexed branches).",
		Variables: []monitoring.ContainerVariable{
			{
				Label: "Instance",
				Name:  "instance",
				OptionsLabelValues: monitoring.ContainerVariableOptionsLabelValues{
					Query:         "searcher_service_request_total",
					LabelName:     "instance",
					ExampleOption: "searcher-7dd95df88c-5bjt9:3181",
				},
				Multi: true,
			},
			grpcMethodVariable,
		},
		Groups: []monitoring.Group{
			{
				Title: "General",
				Rows: []monitoring.Row{
					{
						{
							Name:        "traffic",
							Description: "requests per second by code over 10m",
							Query:       "sum by (code) (rate(searcher_service_request_total{instance=~`${instance:regex}`}[10m]))",
							Panel:       monitoring.Panel().LegendFormat("{{code}}"),
							Owner:       monitoring.ObservableOwnerSearchCore,
							NoAlert:     true,
							Interpretation: `
This graph is the average number of requests per second searcher is
experiencing over the last 10 minutes.

The code is the HTTP Status code. 200 is success. We have a special code
"canceled" which is common when doing a large search request and we find
enough results before searching all possible repos.

Note: A search query is translated into an unindexed search query per unique
(repo, commit). This means a single user query may result in thousands of
requests to searcher.`,
						},
						{
							Name:        "replica_traffic",
							Description: "requests per second per replica over 10m",
							Query:       "sum by (instance) (rate(searcher_service_request_total{instance=~`${instance:regex}`}[10m]))",
							Warning:     monitoring.Alert().GreaterOrEqual(5),
							Panel:       monitoring.Panel().LegendFormat("{{instance}}"),
							Owner:       monitoring.ObservableOwnerSearchCore,
							NextSteps:   "none",
							Interpretation: `
This graph is the average number of requests per second searcher is
experiencing over the last 10 minutes broken down per replica.

The code is the HTTP Status code. 200 is success. We have a special code
"canceled" which is common when doing a large search request and we find
enough results before searching all possible repos.

Note: A search query is translated into an unindexed search query per unique
(repo, commit). This means a single user query may result in thousands of
requests to searcher.`,
						},
					}, {
						{
							Name:        "concurrent_requests",
							Description: "amount of in-flight unindexed search requests (per instance)",
							Query:       "sum by (instance) (searcher_service_running{instance=~`${instance:regex}`})",
							Panel:       monitoring.Panel().LegendFormat("{{instance}}"),
							Owner:       monitoring.ObservableOwnerSearchCore,
							NoAlert:     true,
							Interpretation: `
This graph is the amount of in-flight unindexed search requests per instance.
Consistently high numbers here indicate you may need to scale out searcher.`,
						},
						{
							Name:        "unindexed_search_request_errors",
							Description: "unindexed search request errors every 5m by code",
							Query:       instanceSelector(`sum by (code)(increase(searcher_service_request_total{code!="200",code!="canceled",$$INSTANCE$$}[5m])) / ignoring(code) group_left sum(increase(searcher_service_request_total{$$INSTANCE$$}[5m])) * 100`),
							Warning:     monitoring.Alert().GreaterOrEqual(5).For(5 * time.Minute),
							Panel:       monitoring.Panel().LegendFormat("{{code}}").Unit(monitoring.Percentage),
							Owner:       monitoring.ObservableOwnerSearchCore,
							NextSteps:   "none",
						},
					},
				},
			},

			{
				Title:  "Cache store",
				Hidden: true,
				Rows: []monitoring.Row{
					{
						{
							Name:        "store_fetching",
							Description: "amount of in-flight unindexed search requests fetching code from gitserver (per instance)",
							Query:       "sum by (instance) (searcher_store_fetching{instance=~`${instance:regex}`})",
							Panel:       monitoring.Panel().LegendFormat("{{instance}}"),
							Owner:       monitoring.ObservableOwnerSearchCore,
							NoAlert:     true,
							Interpretation: `
Before we can search a commit we fetch the code from gitserver then cache it
for future search requests. This graph is the current number of search
requests which are in the state of fetching code from gitserver.

Generally this number should remain low since fetching code is fast, but
expect bursts. In the case of instances with a monorepo you would expect this
number to stay low for the duration of fetching the code (which in some cases
can take many minutes).`,
						},
						{
							Name:        "store_fetching_waiting",
							Description: "amount of in-flight unindexed search requests waiting to fetch code from gitserver (per instance)",
							Query:       "sum by (instance) (searcher_store_fetch_queue_size{instance=~`${instance:regex}`})",
							Panel:       monitoring.Panel().LegendFormat("{{instance}}"),
							Owner:       monitoring.ObservableOwnerSearchCore,
							NoAlert:     true,
							Interpretation: `
We limit the number of requests which can fetch code to prevent overwhelming
gitserver. This gauge is the number of requests waiting to be allowed to speak
to gitserver.`,
						},
						{
							Name:        "store_fetching_fail",
							Description: "amount of unindexed search requests that failed while fetching code from gitserver over 10m (per instance)",
							Query:       "sum by (instance) (rate(searcher_store_fetch_failed{instance=~`${instance:regex}`}[10m]))",
							Panel:       monitoring.Panel().LegendFormat("{{instance}}"),
							Owner:       monitoring.ObservableOwnerSearchCore,
							NoAlert:     true,
							Interpretation: `
This graph should be zero since fetching happens in the background and will
not be influenced by user timeouts/etc. Expected upticks in this graph are
during gitserver rollouts. If you regularly see this graph have non-zero
values please reach out to support.`,
						},
					},
				},
			},

			{
				Title:  "Index use",
				Hidden: true,
				Rows: []monitoring.Row{
					{
						{
							Name:        "searcher_hybrid_final_state_total",
							Description: "hybrid search final state over 10m",
							Interpretation: `
This graph is about our interactions with the search index (zoekt) to help
complete unindexed search requests. Searcher will use indexed search for the
files that have not changed between the unindexed commit and the index.

This graph should mostly be "success". The next most common state should be
"search-canceled" which happens when result limits are hit or the user starts
a new search. Finally the next most common should be "diff-too-large", which
happens if the commit is too far from the indexed commit. Otherwise other
state should be rare and likely are a sign for further investigation.

Note: On sourcegraph.com "zoekt-list-missing" is also common due to it
indexing a subset of repositories. Otherwise every other state should occur
rarely.

For a full list of possible state see
[recordHybridFinalState](https://sourcegraph.com/search?q=context:global+repo:%5Egithub%5C.com/sourcegraph/sourcegraph%24+f:cmd/searcher+recordHybridFinalState).`,
							Query:   "sum by (state)(increase(searcher_hybrid_final_state_total{instance=~`${instance:regex}`}[10m]))",
							Panel:   monitoring.Panel().LegendFormat("{{state}}"),
							Owner:   monitoring.ObservableOwnerSearchCore,
							NoAlert: true,
						},
						{
							Name:        "searcher_hybrid_retry_total",
							Description: "hybrid search retrying over 10m",
							Interpretation: `
Expectation is that this graph should mostly be 0. It will trigger if a user
manages to do a search and the underlying index changes while searching or
Zoekt goes down. So occasional bursts can be expected, but if this graph is
regularly above 0 it is a sign for further investigation.`,
							Query:   "sum by (reason)(increase(searcher_hybrid_retry_total{instance=~`${instance:regex}`}[10m]))",
							Panel:   monitoring.Panel().LegendFormat("{{reason}}"),
							Owner:   monitoring.ObservableOwnerSearchCore,
							NoAlert: true,
						},
					},
				},
			},

			shared.NewDiskMetricsGroup(
				shared.DiskMetricsGroupOptions{
					DiskTitle: "cache",

					MetricMountNameLabel: "cacheDir",
					MetricNamespace:      "searcher",

					ServiceName:         "searcher",
					InstanceFilterRegex: `${instance:regex}`,
				},
				monitoring.ObservableOwnerSearchCore,
			),

			shared.NewGRPCServerMetricsGroup(
				shared.GRPCServerMetricsOptions{
					HumanServiceName:   "searcher",
					RawGRPCServiceName: grpcServiceName,

					MethodFilterRegex: fmt.Sprintf("${%s:regex}", grpcMethodVariable.Name),

					InstanceFilterRegex:  `${instance:regex}`,
					MessageSizeNamespace: "src",
				}, monitoring.ObservableOwnerSearchCore),

			shared.NewGRPCInternalErrorMetricsGroup(
				shared.GRPCInternalErrorMetricsOptions{
					HumanServiceName:   "searcher",
					RawGRPCServiceName: grpcServiceName,
					Namespace:          "src",

					MethodFilterRegex: fmt.Sprintf("${%s:regex}", grpcMethodVariable.Name),
				}, monitoring.ObservableOwnerSearchCore),
			shared.NewSiteConfigurationClientMetricsGroup(shared.SiteConfigurationMetricsOptions{
				HumanServiceName:    "searcher",
				InstanceFilterRegex: `${instance:regex}`,
			}, monitoring.ObservableOwnerDevOps),
			shared.NewDatabaseConnectionsMonitoringGroup(containerName, monitoring.ObservableOwnerDevOps),
			shared.NewFrontendInternalAPIErrorResponseMonitoringGroup(containerName, monitoring.ObservableOwnerSearchCore, nil),
			shared.NewContainerMonitoringGroup(containerName, monitoring.ObservableOwnerSearchCore, nil),
			shared.NewProvisioningIndicatorsGroup(containerName, monitoring.ObservableOwnerSearchCore, nil),
			shared.NewGolangMonitoringGroup(containerName, monitoring.ObservableOwnerSearchCore, nil),
			shared.NewKubernetesMonitoringGroup(containerName, monitoring.ObservableOwnerSearchCore, nil),
		},
	}
}
