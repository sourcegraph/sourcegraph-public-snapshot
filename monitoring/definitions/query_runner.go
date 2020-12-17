package definitions

import (
	"github.com/sourcegraph/sourcegraph/monitoring/definitions/shared"
	"github.com/sourcegraph/sourcegraph/monitoring/monitoring"
)

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
						shared.FrontendInternalAPIErrorResponses("query-runner", monitoring.ObservableOwnerSearch),
					},
				},
			},
			{
				Title:  "Container monitoring (not available on server)",
				Hidden: true,
				Rows: []monitoring.Row{
					{
						shared.ContainerMemoryUsage("query-runner", monitoring.ObservableOwnerSearch),
						shared.ContainerCPUUsage("query-runner", monitoring.ObservableOwnerSearch),
					},
					{
						shared.ContainerRestarts("query-runner", monitoring.ObservableOwnerSearch),
						shared.ContainerFsInodes("query-runner", monitoring.ObservableOwnerSearch),
					},
				},
			},
			{
				Title:  "Provisioning indicators (not available on server)",
				Hidden: true,
				Rows: []monitoring.Row{
					{
						shared.ProvisioningCPUUsageLongTerm("query-runner", monitoring.ObservableOwnerSearch),
						shared.ProvisioningMemoryUsageLongTerm("query-runner", monitoring.ObservableOwnerSearch),
					},
					{
						shared.ProvisioningCPUUsageShortTerm("query-runner", monitoring.ObservableOwnerSearch),
						shared.ProvisioningMemoryUsageShortTerm("query-runner", monitoring.ObservableOwnerSearch),
					},
				},
			},
			{
				Title:  "Golang runtime monitoring",
				Hidden: true,
				Rows: []monitoring.Row{
					{
						shared.GoGoroutines("query-runner", monitoring.ObservableOwnerSearch),
						shared.GoGcDuration("query-runner", monitoring.ObservableOwnerSearch),
					},
				},
			},
			{
				Title:  "Kubernetes monitoring (ignore if using Docker Compose or server)",
				Hidden: true,
				Rows: []monitoring.Row{
					{
						shared.KubernetesPodsAvailable("query-runner", monitoring.ObservableOwnerSearch),
					},
				},
			},
		},
	}
}
