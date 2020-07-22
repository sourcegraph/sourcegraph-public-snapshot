package main

func Searcher() *Container {
	return &Container{
		Name:        "searcher",
		Title:       "Searcher",
		Description: "Performs unindexed searches (diff and commit search, text search for unindexed branches).",
		Groups: []Group{
			{
				Title: "General",
				Rows: []Row{
					{
						{
							Name:              "unindexed_search_request_errors",
							Description:       "unindexed search request errors every 5m by code",
							Query:             `sum by (code)(increase(searcher_service_request_total{code!="200",code!="canceled"}[5m]))`,
							DataMayNotExist:   true,
							Warning:           Alert{GreaterOrEqual: 5},
							PanelOptions:      PanelOptions().LegendFormat("{{code}}"),
							PossibleSolutions: "none",
						},
						{
							Name:              "error_ratio",
							Description:       "error ratio over 10m",
							Query:             `searcher_errors:ratio10m`, // TODO: 20m
							Warning:           Alert{GreaterOrEqual: 0.1},
							PossibleSolutions: "none",
						},
						sharedFrontendInternalAPIErrorResponses("searcher"),
					},
				},
			},
			{
				Title:  "Golang runtime monitoring",
				Hidden: true,
				Rows: []Row{
					{
						sharedGoGoroutines("searcher"),
						sharedGoGcDuration("searcher"),
					},
				},
			},
			{
				Title:  "Container monitoring (not available on server)",
				Hidden: true,
				Rows: []Row{
					{
						sharedContainerRestarts("searcher"),
						sharedContainerMemoryUsage("searcher"),
						sharedContainerCPUUsage("searcher"),
					},
				},
			},
			{
				Title:  "Provisioning indicators (not available on server)",
				Hidden: true,
				Rows: []Row{
					{
						sharedProvisioningCPUUsage7d("searcher"),
						sharedProvisioningMemoryUsage7d("searcher"),
					},
					{
						sharedProvisioningCPUUsage5m("searcher"),
						sharedProvisioningMemoryUsage5m("searcher"),
					},
				},
			},
			{
				Title:  "Kubernetes monitoring (only available on k8s)",
				Hidden: true,
				Rows: []Row{
					{
						sharedKubernetesPodsAvailable("searcher"),
					},
				},
			},
		},
	}
}
