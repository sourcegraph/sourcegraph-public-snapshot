package definitions

import (
	"github.com/sourcegraph/sourcegraph/monitoring/definitions/shared"
	"github.com/sourcegraph/sourcegraph/monitoring/monitoring"
)

func ExecutorAndExecutorQueue() *monitoring.Container {
	return &monitoring.Container{
		Name:        "executor-queue",
		Title:       "Executor Queue",
		Description: "Coordinates the and executes jobs from the executor work queue.",
		Groups: []monitoring.Group{
			{
				Title: "Executor",
				Rows: []monitoring.Row{
					{
						{
							Name:              "executor_queue_size",
							Description:       "executor queue size",
							Query:             `max(src_executor_queue_total)`,
							DataMayNotExist:   true,
							Warning:           monitoring.Alert().GreaterOrEqual(100),
							PanelOptions:      monitoring.PanelOptions().LegendFormat("uploads queued for processing"),
							Owner:             monitoring.ObservableOwnerCodeIntel,
							PossibleSolutions: "none",
						},
						{
							Name:            "executor_queue_growth_rate",
							Description:     "executor queue growth rate every 5m",
							Query:           `sum(increase(src_executor_queue_total[30m])) / sum(increase(src_executor_queue_processor_total[30m]))`,
							DataMayNotExist: true,

							Warning:           monitoring.Alert().GreaterOrEqual(5),
							PanelOptions:      monitoring.PanelOptions().LegendFormat("executor queue growth rate"),
							Owner:             monitoring.ObservableOwnerCodeIntel,
							PossibleSolutions: "none",
						},
						{
							Name:              "executor_process_errors",
							Description:       "executor process errors every 5m",
							Query:             `sum(increase(src_executor_queue_processor_errors_total[5m]))`,
							DataMayNotExist:   true,
							Warning:           monitoring.Alert().GreaterOrEqual(20),
							PanelOptions:      monitoring.PanelOptions().LegendFormat("errors"),
							Owner:             monitoring.ObservableOwnerCodeIntel,
							PossibleSolutions: "none",
						},
					},
				},
			},
			{
				Title:  "Internal service requests",
				Hidden: true,
				Rows: []monitoring.Row{
					{
						shared.FrontendInternalAPIErrorResponses("executor-queue", monitoring.ObservableOwnerCodeIntel),
					},
				},
			},
			{
				Title:  "Container monitoring (not available on server)",
				Hidden: true,
				Rows: []monitoring.Row{
					{
						shared.ContainerCPUUsage("executor-queue", monitoring.ObservableOwnerCodeIntel),
						shared.ContainerMemoryUsage("executor-queue", monitoring.ObservableOwnerCodeIntel),
					},
					{
						shared.ContainerRestarts("executor-queue", monitoring.ObservableOwnerCodeIntel),
						shared.ContainerFsInodes("executor-queue", monitoring.ObservableOwnerCodeIntel),
					},
				},
			},
			{
				Title:  "Provisioning indicators (not available on server)",
				Hidden: true,
				Rows: []monitoring.Row{
					{
						shared.ProvisioningCPUUsageLongTerm("executor-queue", monitoring.ObservableOwnerCodeIntel),
						shared.ProvisioningMemoryUsageLongTerm("executor-queue", monitoring.ObservableOwnerCodeIntel),
					},
					{
						shared.ProvisioningCPUUsageShortTerm("executor-queue", monitoring.ObservableOwnerCodeIntel),
						shared.ProvisioningMemoryUsageShortTerm("executor-queue", monitoring.ObservableOwnerCodeIntel),
					},
				},
			},
			{
				Title:  "Golang runtime monitoring",
				Hidden: true,
				Rows: []monitoring.Row{
					{
						shared.GoGoroutines("executor-queue", monitoring.ObservableOwnerCodeIntel),
						shared.GoGcDuration("executor-queue", monitoring.ObservableOwnerCodeIntel),
					},
				},
			},
			{
				Title:  "Kubernetes monitoring (ignore if using Docker Compose or server)",
				Hidden: true,
				Rows: []monitoring.Row{
					{
						shared.KubernetesPodsAvailable("executor-queue", monitoring.ObservableOwnerCodeIntel),
					},
				},
			},
		},
	}
}
