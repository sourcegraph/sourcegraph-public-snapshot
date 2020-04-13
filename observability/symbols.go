package main

func Symbols() *Container {
	return &Container{
		Name:        "symbols",
		Title:       "Symbols",
		Description: "Handles symbol searches for unindexed branches.",
		Groups: []Group{
			{
				Title: "General",
				Rows: []Row{
					{
						{
							Name:            "store_fetch_failures",
							Description:     "store fetch failures every 5m",
							Query:           `sum(increase(symbols_store_fetch_failed[5m]))`,
							DataMayNotExist: true,
							Warning:         Alert{GreaterOrEqual: 5},
							PanelOptions:    PanelOptions().LegendFormat("failures"),
						},
						{
							Name:            "current_fetch_queue_size",
							Description:     "current fetch queue size",
							Query:           `sum(symbols_store_fetch_queue_size)`,
							DataMayNotExist: true,
							Warning:         Alert{GreaterOrEqual: 25},
							PanelOptions:    PanelOptions().LegendFormat("size"),
						},
					},
					{
						{
							Name:            "frontend_internal_api_error_responses",
							Description:     "frontend-internal API error responses every 5m by route",
							Query:           `sum by (category)(increase(src_frontend_internal_request_duration_seconds_count{job="symbols",code!~"2.."}[5m]))`,
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
