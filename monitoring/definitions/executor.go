package definitions

import (
	"github.com/grafana-tools/sdk"

	"github.com/sourcegraph/sourcegraph/monitoring/definitions/shared"
	"github.com/sourcegraph/sourcegraph/monitoring/monitoring"
)

func Executor() *monitoring.Container {
	const containerName = "(executor|sourcegraph-code-intel-indexers|executor-batches)"

	// frontend is sometimes called sourcegraph-frontend in various contexts
	const queueContainerName = "(executor|sourcegraph-code-intel-indexers|executor-batches|frontend|sourcegraph-frontend|worker)"

	return &monitoring.Container{
		Name:        "executor",
		Title:       "Executor",
		Description: `Executes jobs in an isolated environment.`,
		Templates: []sdk.TemplateVar{
			{
				Label:      "Queue name",
				Name:       "queue",
				AllValue:   ".*",
				Current:    sdk.Current{Text: &sdk.StringSliceString{Value: []string{"all"}, Valid: true}, Value: "$__all"},
				IncludeAll: true,
				Options: []sdk.Option{
					{Text: "all", Value: "$__all", Selected: true},
					{Text: "batches", Value: "batches"},
					{Text: "codeintel", Value: "codeintel"},
				},
				Query: "batches,codeintel",
				Type:  "custom",
			},
		},
		Groups: []monitoring.Group{
			shared.CodeIntelligence.NewExecutorQueueGroup(queueContainerName),
			shared.CodeIntelligence.NewExecutorProcessorGroup(containerName),
			shared.CodeIntelligence.NewExecutorExecutionRunLockContentionGroup(containerName),
			shared.CodeIntelligence.NewExecutorAPIClientGroup(containerName),
			shared.CodeIntelligence.NewExecutorSetupCommandGroup(containerName),
			shared.CodeIntelligence.NewExecutorExecutionCommandGroup(containerName),
			shared.CodeIntelligence.NewExecutorTeardownCommandGroup(containerName),

			// Resource monitoring
			shared.NewContainerMonitoringGroup(containerName, monitoring.ObservableOwnerCodeIntel, nil),
			shared.NewProvisioningIndicatorsGroup(containerName, monitoring.ObservableOwnerCodeIntel, nil),
			shared.NewGolangMonitoringGroup(containerName, monitoring.ObservableOwnerCodeIntel, nil),
			shared.NewKubernetesMonitoringGroup(containerName, monitoring.ObservableOwnerCodeIntel, nil),
		},
	}
}
