package definitions

import "github.com/sourcegraph/sourcegraph/monitoring/monitoring"

func QueryRunner() *monitoring.Container {
	return &monitoring.Container{
		Name:        "query-runner",
		Title:       "Query Runner",
		Description: "Periodically runs saved searches and instructs the frontend to send out notifications.",
		Groups: []monitoring.Group{
			{
				Title: "General",
				Rows: []monitoring.Row{
					{
						sharedFrontendInternalAPIErrorResponses("query-runner", monitoring.ObservableOwnerSearch),
					},
				},
			},
			{
				Title:  "Container monitoring (not available on server)",
				Hidden: true,
				Rows: []monitoring.Row{
					{
						sharedContainerMemoryUsage("query-runner", monitoring.ObservableOwnerSearch),
						sharedContainerCPUUsage("query-runner", monitoring.ObservableOwnerSearch),
					},
					{
						sharedContainerRestarts("query-runner", monitoring.ObservableOwnerSearch),
						sharedContainerFsInodes("query-runner", monitoring.ObservableOwnerSearch),
					},
				},
			},
			{
				Title:  "Provisioning indicators (not available on server)",
				Hidden: true,
				Rows: []monitoring.Row{
					{
						sharedProvisioningCPUUsageLongTerm("query-runner", monitoring.ObservableOwnerSearch),
						sharedProvisioningMemoryUsageLongTerm("query-runner", monitoring.ObservableOwnerSearch),
					},
					{
						sharedProvisioningCPUUsageShortTerm("query-runner", monitoring.ObservableOwnerSearch),
						sharedProvisioningMemoryUsageShortTerm("query-runner", monitoring.ObservableOwnerSearch),
					},
				},
			},
			{
				Title:  "Golang runtime monitoring",
				Hidden: true,
				Rows: []monitoring.Row{
					{
						sharedGoGoroutines("query-runner", monitoring.ObservableOwnerSearch),
						sharedGoGcDuration("query-runner", monitoring.ObservableOwnerSearch),
					},
				},
			},
			{
				Title:  "Kubernetes monitoring (ignore if using Docker Compose or server)",
				Hidden: true,
				Rows: []monitoring.Row{
					{
						sharedKubernetesPodsAvailable("query-runner", monitoring.ObservableOwnerSearch),
					},
				},
			},
		},
	}
}
