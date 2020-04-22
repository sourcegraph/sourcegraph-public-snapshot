package main

func Replacer() *Container {
	return &Container{
		Name:        "replacer",
		Title:       "Replacer",
		Description: "Backend for find-and-replace operations.",
		Groups: []Group{
			{
				Title: "General",
				Rows: []Row{
					{
						{
							Name:            "frontend_internal_api_error_responses",
							Description:     "frontend-internal API error responses every 5m by route",
							Query:           `sum by (category)(increase(src_frontend_internal_request_duration_seconds_count{job="replacer",code!~"2.."}[5m]))`,
							DataMayNotExist: true,
							Warning:         Alert{GreaterOrEqual: 5},
							PanelOptions:    PanelOptions().LegendFormat("{{category}}"),
						},
					},
				},
			},
			{
				Title:  "Container monitoring (not available on k8s or server)",
				Hidden: true,
				Rows: []Row{
					{
						sharedContainerRestarts("replacer"),
						sharedContainerMemoryUsage("replacer"),
						sharedContainerCPUUsage("replacer"),
					},
				},
			},
		},
	}
}
