package definitions

import "github.com/sourcegraph/sourcegraph/monitoring/monitoring"

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
						sharedFrontendInternalAPIErrorResponses("executor-queue", monitoring.ObservableOwnerCodeIntel),
					},
				},
			},
			{
				Title:  "Container monitoring (not available on server)",
				Hidden: true,
				Rows: []monitoring.Row{
					{
						sharedContainerCPUUsage("executor-queue", monitoring.ObservableOwnerCodeIntel),
						sharedContainerMemoryUsage("executor-queue", monitoring.ObservableOwnerCodeIntel),
					},
					{
						sharedContainerRestarts("executor-queue", monitoring.ObservableOwnerCodeIntel),
						sharedContainerFsInodes("executor-queue", monitoring.ObservableOwnerCodeIntel),
					},
				},
			},
			{
				Title:  "Provisioning indicators (not available on server)",
				Hidden: true,
				Rows: []monitoring.Row{
					{
						sharedProvisioningCPUUsageLongTerm("executor-queue", monitoring.ObservableOwnerCodeIntel),
						sharedProvisioningMemoryUsageLongTerm("executor-queue", monitoring.ObservableOwnerCodeIntel),
					},
					{
						sharedProvisioningCPUUsageShortTerm("executor-queue", monitoring.ObservableOwnerCodeIntel),
						sharedProvisioningMemoryUsageShortTerm("executor-queue", monitoring.ObservableOwnerCodeIntel),
					},
				},
			},
			{
				Title:  "Golang runtime monitoring",
				Hidden: true,
				Rows: []monitoring.Row{
					{
						sharedGoGoroutines("executor-queue", monitoring.ObservableOwnerCodeIntel),
						sharedGoGcDuration("executor-queue", monitoring.ObservableOwnerCodeIntel),
					},
				},
			},
			{
				Title:  "Kubernetes monitoring (ignore if using Docker Compose or server)",
				Hidden: true,
				Rows: []monitoring.Row{
					{
						sharedKubernetesPodsAvailable("executor-queue", monitoring.ObservableOwnerCodeIntel),
					},
				},
			},
		},
	}
}
