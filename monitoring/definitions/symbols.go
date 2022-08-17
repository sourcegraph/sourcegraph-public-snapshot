package definitions

import (
	"github.com/sourcegraph/sourcegraph/monitoring/definitions/shared"
	"github.com/sourcegraph/sourcegraph/monitoring/monitoring"
)

func Symbols() *monitoring.Dashboard {
	const containerName = "symbols"

	return &monitoring.Dashboard{
		Name:        "symbols",
		Title:       "Symbols",
		Description: "Handles symbol searches for unindexed branches.",
		Groups: []monitoring.Group{
			shared.CodeIntelligence.NewSymbolsAPIGroup(containerName),
			shared.CodeIntelligence.NewSymbolsParserGroup(containerName),
			shared.CodeIntelligence.NewSymbolsCacheJanitorGroup(containerName),
			shared.CodeIntelligence.NewSymbolsRepositoryFetcherGroup(containerName),
			shared.CodeIntelligence.NewSymbolsGitserverClientGroup(containerName),

			shared.NewDatabaseConnectionsMonitoringGroup(containerName),
			shared.NewFrontendInternalAPIErrorResponseMonitoringGroup(containerName, monitoring.ObservableOwnerCodeIntel, nil),
			shared.NewContainerMonitoringGroup(containerName, monitoring.ObservableOwnerCodeIntel, nil),
			shared.NewProvisioningIndicatorsGroup(containerName, monitoring.ObservableOwnerCodeIntel, nil),
			shared.NewGolangMonitoringGroup(containerName, monitoring.ObservableOwnerCodeIntel, nil),
			shared.NewKubernetesMonitoringGroup(containerName, monitoring.ObservableOwnerCodeIntel, nil),
		},
	}
}
