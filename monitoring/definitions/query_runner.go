package definitions

import (
	"github.com/sourcegraph/sourcegraph/monitoring/definitions/shared"
	"github.com/sourcegraph/sourcegraph/monitoring/monitoring"
)

func QueryRunner() *monitoring.Container {
	return &monitoring.Container{
		Name:        "query-runner",
		Title:       "Query Runner",
		Description: "Periodically runs saved searches and instructs the frontend to send out notifications.",
		Groups: []monitoring.Group{
			{
				Title: "General",
				Rows: []monitoring.Row{
					{
						shared.FrontendInternalAPIErrorResponses("query-runner", monitoring.ObservableOwnerSearch).Observable(),
					},
				},
			},
			shared.NewContainerMonitoringGroup("query-runner", monitoring.ObservableOwnerSearch, nil),
			shared.NewProvisioningIndicatorsGroup("query-runner", monitoring.ObservableOwnerSearch, nil),
			{
				Title:  shared.TitleGolangMonitoring,
				Hidden: true,
				Rows: []monitoring.Row{
					{
						shared.GoGoroutines("query-runner", monitoring.ObservableOwnerSearch).Observable(),
						shared.GoGcDuration("query-runner", monitoring.ObservableOwnerSearch).Observable(),
					},
				},
			},
			{
				Title:  shared.TitleKubernetesMonitoring,
				Hidden: true,
				Rows: []monitoring.Row{
					{
						shared.KubernetesPodsAvailable("query-runner", monitoring.ObservableOwnerSearch).Observable(),
					},
				},
			},
		},
	}
}
