package main

func ExecutorAndExecutorQueue() *Container {
	return &Container{
		Name:        "executor-queue",
		Title:       "Executor Queue",
		Description: "Coordinates the and executes jobs from the executor work queue.",
		Groups: []Group{
			{
				Title: "Executor",
				Rows: []Row{
					{
						{
							Name:              "executor_queue_size",
							Description:       "executor queue size",
							Query:             `max(src_executor_queue_total)`,
							DataMayNotExist:   true,
							Warning:           Alert().GreaterOrEqual(100),
							PanelOptions:      PanelOptions().LegendFormat("uploads queued for processing"),
							Owner:             ObservableOwnerCodeIntel,
							PossibleSolutions: "none",
						},
						{
							Name:            "executor_queue_growth_rate",
							Description:     "executor queue growth rate every 5m",
							Query:           `sum(increase(src_executor_queue_total[30m])) / sum(increase(src_executor_queue_processor_total[30m]))`,
							DataMayNotExist: true,

							Warning:           Alert().GreaterOrEqual(5),
							PanelOptions:      PanelOptions().LegendFormat("executor queue growth rate"),
							Owner:             ObservableOwnerCodeIntel,
							PossibleSolutions: "none",
						},
						{
							Name:              "executor_process_errors",
							Description:       "executor process errors every 5m",
							Query:             `sum(increase(src_executor_queue_processor_errors_total[5m]))`,
							DataMayNotExist:   true,
							Warning:           Alert().GreaterOrEqual(20),
							PanelOptions:      PanelOptions().LegendFormat("errors"),
							Owner:             ObservableOwnerCodeIntel,
							PossibleSolutions: "none",
						},
					},
				},
			},
			{
				Title:  "Internal service requests",
				Hidden: true,
				Rows: []Row{
					{
						sharedFrontendInternalAPIErrorResponses("executor-queue", ObservableOwnerCodeIntel),
					},
				},
			},
			{
				Title:  "Container monitoring (not available on server)",
				Hidden: true,
				Rows: []Row{
					{
						sharedContainerCPUUsage("executor-queue", ObservableOwnerCodeIntel),
						sharedContainerMemoryUsage("executor-queue", ObservableOwnerCodeIntel),
					},
					{
						sharedContainerRestarts("executor-queue", ObservableOwnerCodeIntel),
						sharedContainerFsInodes("executor-queue", ObservableOwnerCodeIntel),
					},
				},
			},
			{
				Title:  "Provisioning indicators (not available on server)",
				Hidden: true,
				Rows: []Row{
					{
						sharedProvisioningCPUUsageLongTerm("executor-queue", ObservableOwnerCodeIntel),
						sharedProvisioningMemoryUsageLongTerm("executor-queue", ObservableOwnerCodeIntel),
					},
					{
						sharedProvisioningCPUUsageShortTerm("executor-queue", ObservableOwnerCodeIntel),
						sharedProvisioningMemoryUsageShortTerm("executor-queue", ObservableOwnerCodeIntel),
					},
				},
			},
			{
				Title:  "Golang runtime monitoring",
				Hidden: true,
				Rows: []Row{
					{
						sharedGoGoroutines("executor-queue", ObservableOwnerCodeIntel),
						sharedGoGcDuration("executor-queue", ObservableOwnerCodeIntel),
					},
				},
			},
			{
				Title:  "Kubernetes monitoring (ignore if using Docker Compose or server)",
				Hidden: true,
				Rows: []Row{
					{
						sharedKubernetesPodsAvailable("executor-queue", ObservableOwnerCodeIntel),
					},
				},
			},
		},
	}
}
