package definitions

import (
	"github.com/sourcegraph/sourcegraph/monitoring/definitions/shared"
	"github.com/sourcegraph/sourcegraph/monitoring/monitoring"
)

func QueryRunner() *monitoring.Container {
	const containerName = "query-runner"

	return &monitoring.Container{
		Name:        "query-runner",
		Title:       "Query Runner",
		Description: "Periodically runs saved searches and instructs the frontend to send out notifications.",
		Groups: []monitoring.Group{
			shared.NewFrontendInternalAPIErrorResponseMonitoringGroup(containerName, monitoring.ObservableOwnerSearch, nil),
			shared.NewContainerMonitoringGroup(containerName, monitoring.ObservableOwnerSearch, nil),
			shared.NewProvisioningIndicatorsGroup(containerName, monitoring.ObservableOwnerSearch, nil),
			shared.NewGolangMonitoringGroup(containerName, monitoring.ObservableOwnerSearch, nil),
			shared.NewKubernetesMonitoringGroup(containerName, monitoring.ObservableOwnerSearch, nil),
		},
	}
}
