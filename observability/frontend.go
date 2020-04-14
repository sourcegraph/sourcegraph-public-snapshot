package main

func Frontend() *Container {
	return &Container{
		Name:        "frontend",
		Title:       "Frontend",
		Description: "Serves all end-user browser and API requests.",
		Groups: []Group{
			{
				Title: "Search at a glance",
				Rows: []Row{
					{
						{
							Name:            "99th_percentile_search_request_duration",
							Description:     "99th percentile search request duration over 5m",
							Query:           `avg(histogram_quantile(0.99, rate(src_graphql_field_seconds_bucket{type="Search",field="results",error="false"}[5m])))`,
							DataMayNotExist: true,
							DataMayBeNaN:    true, // See https://github.com/sourcegraph/sourcegraph/issues/9834
							Warning:         Alert{GreaterOrEqual: 3},
							PanelOptions:    PanelOptions().LegendFormat("duration").Unit(Seconds),
						},
						{
							Name:            "90th_percentile_search_request_duration",
							Description:     "90th percentile search request duration over 5m",
							Query:           `avg(histogram_quantile(0.90, rate(src_graphql_field_seconds_bucket{type="Search",field="results",error="false"}[5m])))`,
							DataMayNotExist: true,
							DataMayBeNaN:    true, // See https://github.com/sourcegraph/sourcegraph/issues/9834
							Warning:         Alert{GreaterOrEqual: 3},
							PanelOptions:    PanelOptions().LegendFormat("duration").Unit(Seconds),
						},
					},
					{
						{
							Name:            "hard_timeout_search_responses",
							Description:     "hard timeout search responses every 5m",
							Query:           `sum(sum by (status)(increase(src_graphql_search_response{status="timeout"}[5m]))) + sum(sum by (status, alert_type)(increase(src_graphql_search_response{status="alert",alert_type="timed_out"}[5m])))`,
							DataMayNotExist: true,
							Warning:         Alert{GreaterOrEqual: 5},
							Critical:        Alert{GreaterOrEqual: 20},
							PanelOptions:    PanelOptions().LegendFormat("hard timeout"),
						},
						{
							Name:            "hard_error_search_responses",
							Description:     "hard error search responses every 5m",
							Query:           `sum by (status)(increase(src_graphql_search_response{status=~"error"}[5m]))`,
							DataMayNotExist: true,
							Warning:         Alert{GreaterOrEqual: 5},
							Critical:        Alert{GreaterOrEqual: 20},
							PanelOptions:    PanelOptions().LegendFormat("hard error"),
						},
						{
							Name:            "partial_timeout_search_responses",
							Description:     "partial timeout search responses every 5m",
							Query:           `sum by (status)(increase(src_graphql_search_response{status="partial_timeout"}[5m]))`,
							DataMayNotExist: true,
							Warning:         Alert{GreaterOrEqual: 5},
							PanelOptions:    PanelOptions().LegendFormat("partial timeout"),
						},
						{
							Name:            "search_alert_user_suggestions",
							Description:     "search alert user suggestions shown every 5m",
							Query:           `sum by (alert_type)(increase(src_graphql_search_response{status="alert",alert_type!~"timed_out"}[5m]))`,
							DataMayNotExist: true,
							Warning:         Alert{GreaterOrEqual: 50},
							PanelOptions:    PanelOptions().LegendFormat("{{alert_type}}"),
						},
					},
				},
			},
			{
				Title:  "Internal service requests",
				Hidden: true,
				Rows: []Row{
					{
						{
							Name:            "internal_indexed_search_error_responses",
							Description:     "internal indexed search error responses every 5m",
							Query:           `sum by (code)(increase(src_zoekt_request_duration_seconds_count{code!~"2.."}[5m]))`,
							DataMayNotExist: true,
							Warning:         Alert{GreaterOrEqual: 5},
							PanelOptions:    PanelOptions().LegendFormat("{{code}}"),
						},
						{
							Name:            "internal_unindexed_search_error_responses",
							Description:     "internal unindexed search error responses every 5m",
							Query:           `sum by (code)(increase(searcher_service_request_total{code!~"2.."}[5m]))`,
							DataMayNotExist: true,
							Warning:         Alert{GreaterOrEqual: 5},
							PanelOptions:    PanelOptions().LegendFormat("{{code}}"),
						},
						{
							Name:            "internal_api_error_responses",
							Description:     "internal API error responses every 5m by route",
							Query:           `sum by (category)(increase(src_frontend_internal_request_duration_seconds_count{code!~"2.."}[5m]))`,
							DataMayNotExist: true,
							Warning:         Alert{GreaterOrEqual: 25},
							PanelOptions:    PanelOptions().LegendFormat("{{category}}"),
						},
					},
				},
			},
		},
	}
}
