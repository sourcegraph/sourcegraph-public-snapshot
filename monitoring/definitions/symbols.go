package definitions

import (
	"fmt"

	"github.com/sourcegraph/sourcegraph/monitoring/definitions/shared"
	"github.com/sourcegraph/sourcegraph/monitoring/monitoring"
)

func Symbols() *monitoring.Dashboard {
	const containerName = "symbols"

	grpcMethodVariable := shared.GRPCMethodVariable(containerName)

	return &monitoring.Dashboard{
		Name:        "symbols",
		Title:       "Symbols",
		Description: "Handles symbol searches for unindexed branches.",
		Variables: []monitoring.ContainerVariable{
			{
				Label: "instance",
				Name:  "instance",
				OptionsLabelValues: monitoring.ContainerVariableOptionsLabelValues{
					Query:         "src_codeintel_symbols_fetching",
					LabelName:     "instance",
					ExampleOption: "symbols-0:3184",
				},
				Multi: true,
			},
			shared.GRPCMethodVariable(containerName),
		},
		Groups: []monitoring.Group{
			shared.CodeIntelligence.NewSymbolsAPIGroup(containerName),
			shared.CodeIntelligence.NewSymbolsParserGroup(containerName),
			shared.CodeIntelligence.NewSymbolsCacheJanitorGroup(containerName),
			shared.CodeIntelligence.NewSymbolsRepositoryFetcherGroup(containerName),
			shared.CodeIntelligence.NewSymbolsGitserverClientGroup(containerName),

			shared.NewGRPCServerMetricsGroup(
				shared.GRPCServerMetricsOptions{
					ServiceName:         containerName,
					MetricNamespace:     containerName,
					MethodFilterRegex:   fmt.Sprintf("${%s:regex}", grpcMethodVariable.Name),
					InstanceFilterRegex: `${instance:regex}`,
				}, monitoring.ObservableOwnerCodeIntel),

			shared.NewDatabaseConnectionsMonitoringGroup(containerName),
			shared.NewFrontendInternalAPIErrorResponseMonitoringGroup(containerName, monitoring.ObservableOwnerCodeIntel, nil),
			shared.NewContainerMonitoringGroup(containerName, monitoring.ObservableOwnerCodeIntel, nil),
			shared.NewProvisioningIndicatorsGroup(containerName, monitoring.ObservableOwnerCodeIntel, nil),
			shared.NewGolangMonitoringGroup(containerName, monitoring.ObservableOwnerCodeIntel, nil),
			shared.NewKubernetesMonitoringGroup(containerName, monitoring.ObservableOwnerCodeIntel, nil),
		},
	}
}
