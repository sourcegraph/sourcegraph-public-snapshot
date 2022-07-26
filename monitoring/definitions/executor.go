package definitions

import (
	"github.com/sourcegraph/sourcegraph/monitoring/definitions/shared"
	"github.com/sourcegraph/sourcegraph/monitoring/monitoring"
)

func Executor() *monitoring.Dashboard {
	// sg_job value is hard-coded, see enterprise/cmd/frontend/internal/executorqueue/handler/routes.go
	const containerName = "sourcegraph-executors"

	// frontend is sometimes called sourcegraph-frontend in various contexts
	const queueContainerName = "(executor|sourcegraph-code-intel-indexers|executor-batches|frontend|sourcegraph-frontend|worker|sourcegraph-executors)"

	return &monitoring.Dashboard{
		Name:        "executor",
		Title:       "Executor",
		Description: `Executes jobs in an isolated environment.`,
		Variables: []monitoring.ContainerVariable{
			{
				Label: "Queue name",
				Name:  "queue",
				Options: monitoring.ContainerVariableOptions{
					Options: []string{"batches", "codeintel"},
				},
			},
			{
				Label:        "Compute instance",
				Name:         "instance",
				OptionsQuery: "label_values(node_exporter_build_info{sg_job=\"sourcegraph-executor-nodes\"}, instance)",

				// The options query can generate a massive result set that can cause issues.
				// shared.NewNodeExporterGroup filters by job as well so this is safe to use
				WildcardAllValue: true,
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

			shared.NewNodeExporterGroup(containerName, "Compute", "$instance"),
			shared.NewNodeExporterGroup(containerName, "Docker Registry Mirror", ".*"),

			// Resource monitoring
			shared.NewGolangMonitoringGroup(containerName, monitoring.ObservableOwnerCodeIntel, &shared.GolangMonitoringOptions{
				InstanceLabelName: "sg_instance",
				JobLabelName:      "sg_job",
			}),
		},
	}
}
