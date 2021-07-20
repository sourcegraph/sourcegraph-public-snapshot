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
			shared.NewGolangMonitoringGroup("query-runner", monitoring.ObservableOwnerSearch, nil),
			shared.NewKubernetesMonitoringGroup("query-runner", monitoring.ObservableOwnerSearch, nil),
		},
	}
}
