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
						{
							Name:            "frontend_internal_api_error_responses",
							Description:     "frontend-internal API error responses every 5m by route",
							Query:           `sum by (category)(increase(src_frontend_internal_request_duration_seconds_count{job="searcher",code!~"2.."}[5m]))`,
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
