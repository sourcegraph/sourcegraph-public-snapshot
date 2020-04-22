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
							Name:            "unindexed_search_request_errors",
							Description:     "unindexed search request errors every 5m by code",
							Query:           `sum by (code)(increase(searcher_service_request_total{code!="200",code!="canceled"}[5m]))`,
							DataMayNotExist: true,
							Warning:         Alert{GreaterOrEqual: 5},
							PanelOptions:    PanelOptions().LegendFormat("{{code}}"),
						},
						sharedFrontendInternalAPIErrorResponses("searcher"),
					},
				},
			},
			{
				Title:  "Container monitoring (not available on k8s or server)",
				Hidden: true,
				Rows: []Row{
					{
						sharedContainerRestarts("searcher"),
						sharedContainerMemoryUsage("searcher"),
						sharedContainerCPUUsage("searcher"),
					},
				},
			},
		},
	}
}
