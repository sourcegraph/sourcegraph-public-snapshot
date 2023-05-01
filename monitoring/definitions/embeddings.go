package definitions

import (
	"github.com/sourcegraph/sourcegraph/monitoring/definitions/shared"
	"github.com/sourcegraph/sourcegraph/monitoring/monitoring"
)

func Embeddings() *monitoring.Dashboard {
	const containerName = "embeddings"

	return &monitoring.Dashboard{
		Name:        "embeddings",
		Title:       "Embeddings",
		Description: "Handles embeddings searches.",
		Groups: []monitoring.Group{
			shared.NewDatabaseConnectionsMonitoringGroup(containerName),
			shared.NewFrontendInternalAPIErrorResponseMonitoringGroup(containerName, monitoring.ObservableOwnerCody, nil),
			shared.NewContainerMonitoringGroup(containerName, monitoring.ObservableOwnerCody, nil),
			shared.NewProvisioningIndicatorsGroup(containerName, monitoring.ObservableOwnerCody, nil),
			shared.NewGolangMonitoringGroup(containerName, monitoring.ObservableOwnerCody, nil),
			shared.NewKubernetesMonitoringGroup(containerName, monitoring.ObservableOwnerCody, nil),
		},
	}
}
