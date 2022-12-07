package definitions

import (
	"time"

	"github.com/sourcegraph/sourcegraph/monitoring/definitions/shared"
	"github.com/sourcegraph/sourcegraph/monitoring/monitoring"
)

func Searcher() *monitoring.Dashboard {
	const containerName = "searcher"

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
		},
		Groups: []monitoring.Group{
			{
				Title: "General",
				Rows: []monitoring.Row{
					{
						{
							Name:        "unindexed_search_request_errors",
							Description: "unindexed search request errors every 5m by code",
							Query:       `sum by (code)(increase(searcher_service_request_total{code!="200",code!="canceled"}[5m])) / ignoring(code) group_left sum(increase(searcher_service_request_total[5m])) * 100`,
							Warning:     monitoring.Alert().GreaterOrEqual(5).For(5 * time.Minute),
							Panel:       monitoring.Panel().LegendFormat("{{code}}").Unit(monitoring.Percentage),
							Owner:       monitoring.ObservableOwnerSearchCore,
							NextSteps:   "none",
						},
						{
							Name:        "replica_traffic",
							Description: "requests per second over 10m",
							Query:       "sum by(instance) (rate(searcher_service_request_total[10m]))",
							Warning:     monitoring.Alert().GreaterOrEqual(5),
							Panel:       monitoring.Panel().LegendFormat("{{instance}}"),
							Owner:       monitoring.ObservableOwnerSearchCore,
							NextSteps:   "none",
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
							Query:   `sum by (state)(increase(searcher_hybrid_final_state_total[10m]))`,
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
							Query:   `sum by (reason)(increase(searcher_hybrid_retry_total[10m]))`,
							Panel:   monitoring.Panel(),
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

			shared.NewDatabaseConnectionsMonitoringGroup(containerName),
			shared.NewFrontendInternalAPIErrorResponseMonitoringGroup(containerName, monitoring.ObservableOwnerSearchCore, nil),
			shared.NewContainerMonitoringGroup(containerName, monitoring.ObservableOwnerSearchCore, nil),
			shared.NewProvisioningIndicatorsGroup(containerName, monitoring.ObservableOwnerSearchCore, nil),
			shared.NewGolangMonitoringGroup(containerName, monitoring.ObservableOwnerSearchCore, nil),
			shared.NewKubernetesMonitoringGroup(containerName, monitoring.ObservableOwnerSearchCore, nil),
		},
	}
}
