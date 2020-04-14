package main

func QueryRunner() *Container {
	return &Container{
		Name:        "query-runner",
		Title:       "Query Runner",
		Description: "Periodically runs saved searches and instructs the frontend to send out notifications.",
		Groups: []Group{
			{
				Title: "General",
				Rows: []Row{
					{
						{
							Name:            "frontend_internal_api_error_responses",
							Description:     "frontend-internal API error responses every 5m by route",
							Query:           `sum by (category)(increase(src_frontend_internal_request_duration_seconds_count{job="query-runner",code!~"2.."}[5m]))`,
							DataMayNotExist: true,
							Warning:         Alert{GreaterOrEqual: 5},
							PanelOptions:    PanelOptions().LegendFormat("{{category}}"),
						},
					},
				},
			},
		},
	}
}
