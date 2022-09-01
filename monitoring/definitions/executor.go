package definitions

import (
	"github.com/sourcegraph/sourcegraph/monitoring/definitions/shared"
	"github.com/sourcegraph/sourcegraph/monitoring/monitoring"
)

func Executor() *monitoring.Dashboard {
	// sg_job value is hard-coded, see enterprise/cmd/frontend/internal/executorqueue/handler/routes.go
	const executorsJobName = "sourcegraph-executors"
	const registryJobName = "sourcegraph-executors-registry"
	// sg_instance for registry is hard-coded, see enterprise/cmd/executor/internal/metrics/metrics.go
	const registryInstanceName = "docker-registry"

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
				OptionsQuery: `label_values(node_exporter_build_info{sg_job="` + executorsJobName + `"}, sg_instance)`,

				// The options query can generate a massive result set that can cause issues.
				// shared.NewNodeExporterGroup filters by job as well so this is safe to use
				WildcardAllValue: true,
				Multi:            true,
			},
		},
		Groups: []monitoring.Group{
			shared.CodeIntelligence.NewExecutorQueueGroup(queueContainerName, "$queue"),
			shared.CodeIntelligence.NewExecutorProcessorGroup(executorsJobName),
			shared.CodeIntelligence.NewExecutorAPIClientGroup(executorsJobName),
			shared.CodeIntelligence.NewExecutorSetupCommandGroup(executorsJobName),
			shared.CodeIntelligence.NewExecutorExecutionCommandGroup(executorsJobName),
			shared.CodeIntelligence.NewExecutorTeardownCommandGroup(executorsJobName),

			shared.NewNodeExporterGroup(executorsJobName, "Compute", "$instance"),
			shared.NewNodeExporterGroup(registryJobName, "Docker Registry Mirror", registryInstanceName),

			// Resource monitoring
			shared.NewGolangMonitoringGroup(executorsJobName, monitoring.ObservableOwnerCodeIntel, &shared.GolangMonitoringOptions{
				InstanceLabelName: "sg_instance",
				JobLabelName:      "sg_job",
			}),
		},
	}
}
