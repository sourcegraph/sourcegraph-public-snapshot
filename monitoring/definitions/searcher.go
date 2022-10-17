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

			shared.NewDatabaseConnectionsMonitoringGroup(containerName),
			shared.NewFrontendInternalAPIErrorResponseMonitoringGroup(containerName, monitoring.ObservableOwnerSearchCore, nil),
			shared.NewContainerMonitoringGroup(containerName, monitoring.ObservableOwnerSearchCore, nil),
			shared.NewProvisioningIndicatorsGroup(containerName, monitoring.ObservableOwnerSearchCore, nil),
			shared.NewGolangMonitoringGroup(containerName, monitoring.ObservableOwnerSearchCore, nil),
			shared.NewKubernetesMonitoringGroup(containerName, monitoring.ObservableOwnerSearchCore, nil),
		},
	}
}
