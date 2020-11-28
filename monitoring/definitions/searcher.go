package definitions

import (
	"time"

	"github.com/sourcegraph/sourcegraph/monitoring/monitoring"
)

func Searcher() *monitoring.Container {
	return &monitoring.Container{
		Name:        "searcher",
		Title:       "Searcher",
		Description: "Performs unindexed searches (diff and commit search, text search for unindexed branches).",
		Groups: []monitoring.Group{
			{
				Title: "General",
				Rows: []monitoring.Row{
					{
						{
							Name:              "unindexed_search_request_errors",
							Description:       "unindexed search request errors every 5m by code",
							Query:             `sum by (code)(increase(searcher_service_request_total{code!="200",code!="canceled"}[5m])) / ignoring(code) group_left sum(increase(searcher_service_request_total[5m])) * 100`,
							DataMayNotExist:   true,
							Warning:           monitoring.Alert().GreaterOrEqual(5).For(5 * time.Minute),
							PanelOptions:      monitoring.PanelOptions().LegendFormat("{{code}}").Unit(monitoring.Percentage),
							Owner:             monitoring.ObservableOwnerSearch,
							PossibleSolutions: "none",
						},
						{
							Name:              "replica_traffic",
							Description:       "requests per second over 10m",
							Query:             "sum by(instance) (rate(searcher_service_request_total[10m]))",
							Warning:           monitoring.Alert().GreaterOrEqual(5),
							PanelOptions:      monitoring.PanelOptions().LegendFormat("{{instance}}"),
							Owner:             monitoring.ObservableOwnerSearch,
							PossibleSolutions: "none",
						},
						sharedFrontendInternalAPIErrorResponses("searcher", monitoring.ObservableOwnerSearch),
					},
				},
			},
			{
				Title:  "Container monitoring (not available on server)",
				Hidden: true,
				Rows: []monitoring.Row{
					{
						sharedContainerCPUUsage("searcher", monitoring.ObservableOwnerSearch),
						sharedContainerMemoryUsage("searcher", monitoring.ObservableOwnerSearch),
					},
					{
						sharedContainerRestarts("searcher", monitoring.ObservableOwnerSearch),
						sharedContainerFsInodes("searcher", monitoring.ObservableOwnerSearch),
					},
				},
			},
			{
				Title:  "Provisioning indicators (not available on server)",
				Hidden: true,
				Rows: []monitoring.Row{
					{
						sharedProvisioningCPUUsageLongTerm("searcher", monitoring.ObservableOwnerSearch),
						sharedProvisioningMemoryUsageLongTerm("searcher", monitoring.ObservableOwnerSearch),
					},
					{
						sharedProvisioningCPUUsageShortTerm("searcher", monitoring.ObservableOwnerSearch),
						sharedProvisioningMemoryUsageShortTerm("searcher", monitoring.ObservableOwnerSearch),
					},
				},
			},
			{
				Title:  "Golang runtime monitoring",
				Hidden: true,
				Rows: []monitoring.Row{
					{
						sharedGoGoroutines("searcher", monitoring.ObservableOwnerSearch),
						sharedGoGcDuration("searcher", monitoring.ObservableOwnerSearch),
					},
				},
			},
			{
				Title:  "Kubernetes monitoring (ignore if using Docker Compose or server)",
				Hidden: true,
				Rows: []monitoring.Row{
					{
						sharedKubernetesPodsAvailable("searcher", monitoring.ObservableOwnerSearch),
					},
				},
			},
		},
	}
}
