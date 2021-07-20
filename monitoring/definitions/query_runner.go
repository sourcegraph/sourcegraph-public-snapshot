package definitions

import (
	"github.com/sourcegraph/sourcegraph/monitoring/definitions/shared"
	"github.com/sourcegraph/sourcegraph/monitoring/monitoring"
)

func QueryRunner() *monitoring.Container {
	const (
		containerName = "query-runner"
		primaryOwner  = monitoring.ObservableOwnerSearch
	)

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
			shared.NewContainerMonitoringGroup(containerName, primaryOwner, nil),
			shared.NewProvisioningIndicatorsGroup(containerName, primaryOwner, nil),
			shared.NewGolangMonitoringGroup(containerName, primaryOwner, nil),
			shared.NewKubernetesMonitoringGroup(containerName, primaryOwner, nil),
		},
	}
}
