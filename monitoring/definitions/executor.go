package definitions

import (
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
		Variables: []monitoring.ContainerVariable{
			{
				Label:   "Queue name",
				Name:    "queue",
				Options: []string{"batches", "codeintel"},
			},
			{
				Label: "Compute instance",
				Name:  "instance",
				Query: "label_values(node_exporter_build_info{job=\"sourcegraph-code-intel-indexer-nodes\"}, instance)",
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
