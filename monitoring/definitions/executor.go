package definitions

import (
	"github.com/grafana-tools/sdk"

	"github.com/sourcegraph/sourcegraph/monitoring/definitions/shared"
	"github.com/sourcegraph/sourcegraph/monitoring/monitoring"
)

func Executor() *monitoring.Container {
	const containerName = "(executor|sourcegraph-code-intel-indexers|executor-batches|sourcegraph-executors)"

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
			{
				Label:      "Compute Instance",
				Name:       "instance",
				AllValue:   ".*",
				IncludeAll: true,
				Query:      "label_values(node_exporter_build_info{job=\"sourcegraph-code-intel-indexer-nodes\"}, instance)",
				Type:       "query",
				Datasource: monitoring.StringPtr("Prometheus"),
				Sort:       1,
				Refresh: sdk.BoolInt{
					Flag:  true,
					Value: monitoring.Int64Ptr(1),
				},
				Hide: 0,
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

			shared.NewNodeExporterGroup(containerName, "(sourcegraph-code-intel-indexer-nodes|sourcegraph-executor-nodes)", "Compute", "$instance"),
			shared.NewNodeExporterGroup(containerName, "(sourcegraph-code-intel-indexer-docker-registry-mirror-nodes|sourcegraph-executors-docker-registry-mirror-nodes)", "Docker Registry Mirror", ".*"),

			// Resource monitoring
			shared.NewGolangMonitoringGroup(containerName, monitoring.ObservableOwnerCodeIntel, nil),
		},
	}
}
