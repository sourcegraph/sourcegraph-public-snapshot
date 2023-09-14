package definitions

import (
	"github.com/sourcegraph/sourcegraph/monitoring/definitions/shared"
	"github.com/sourcegraph/sourcegraph/monitoring/monitoring"
)

func Executor() *monitoring.Dashboard {
	// sg_job value is hard-coded, see cmd/frontend/internal/executorqueue/handler/routes.go
	const executorsJobName = "sourcegraph-executors"
	const registryJobName = "sourcegraph-executors-registry"
	// sg_instance for registry is hard-coded, see cmd/executor/internal/metrics/metrics.go
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
				Label: "Compute instance",
				Name:  "instance",
				OptionsLabelValues: monitoring.ContainerVariableOptionsLabelValues{
					Query:         `node_exporter_build_info{sg_job="` + executorsJobName + `"}`,
					LabelName:     "sg_instance",
					ExampleOption: "codeintel-cloud-sourcegraph-executor-5rzf-ff9a035d-e34b-4bcf-b862-e78c69b99484",
				},

				// The options query can generate a massive result set that can cause issues.
				// shared.NewNodeExporterGroup filters by job as well so this is safe to use
				WildcardAllValue: true,
				Multi:            true,
			},
		},
		Groups: []monitoring.Group{
			shared.Executors.NewExecutorQueueGroup("executor", queueContainerName, "$queue"),
			shared.Executors.NewExecutorMultiqueueGroup("executor", queueContainerName, "$queue"),
			shared.CodeIntelligence.NewExecutorProcessorGroup(executorsJobName),
			shared.CodeIntelligence.NewExecutorAPIQueueClientGroup(executorsJobName),
			shared.CodeIntelligence.NewExecutorAPIFilesClientGroup(executorsJobName),
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
