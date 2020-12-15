package definitions

import (
	"time"

	"github.com/sourcegraph/sourcegraph/monitoring/definitions/shared"
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
						shared.FrontendInternalAPIErrorResponses("searcher", monitoring.ObservableOwnerSearch),
					},
				},
			},
			{
				Title:  "Container monitoring (not available on server)",
				Hidden: true,
				Rows: []monitoring.Row{
					{
						shared.ContainerCPUUsage("searcher", monitoring.ObservableOwnerSearch),
						shared.ContainerMemoryUsage("searcher", monitoring.ObservableOwnerSearch),
					},
					{
						shared.ContainerRestarts("searcher", monitoring.ObservableOwnerSearch),
						shared.ContainerFsInodes("searcher", monitoring.ObservableOwnerSearch),
					},
				},
			},
			{
				Title:  "Provisioning indicators (not available on server)",
				Hidden: true,
				Rows: []monitoring.Row{
					{
						shared.ProvisioningCPUUsageLongTerm("searcher", monitoring.ObservableOwnerSearch),
						shared.ProvisioningMemoryUsageLongTerm("searcher", monitoring.ObservableOwnerSearch),
					},
					{
						shared.ProvisioningCPUUsageShortTerm("searcher", monitoring.ObservableOwnerSearch),
						shared.ProvisioningMemoryUsageShortTerm("searcher", monitoring.ObservableOwnerSearch),
					},
				},
			},
			{
				Title:  "Golang runtime monitoring",
				Hidden: true,
				Rows: []monitoring.Row{
					{
						shared.GoGoroutines("searcher", monitoring.ObservableOwnerSearch),
						shared.GoGcDuration("searcher", monitoring.ObservableOwnerSearch),
					},
				},
			},
			{
				Title:  "Kubernetes monitoring (ignore if using Docker Compose or server)",
				Hidden: true,
				Rows: []monitoring.Row{
					{
						shared.KubernetesPodsAvailable("searcher", monitoring.ObservableOwnerSearch),
					},
				},
			},
		},
	}
}
