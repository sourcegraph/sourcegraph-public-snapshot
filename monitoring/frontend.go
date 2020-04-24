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
							Description:     "99th percentile successful search request duration over 5m",
							Query:           `histogram_quantile(0.99, sum by (le)(rate(src_graphql_field_seconds_bucket{type="Search",field="results",error="false",source="browser",name!="CodeIntelSearch"}[5m])))`,
							DataMayNotExist: true,
							DataMayBeNaN:    true, // See https://github.com/sourcegraph/sourcegraph/issues/9834
							Warning:         Alert{GreaterOrEqual: 20},
							PanelOptions:    PanelOptions().LegendFormat("duration").Unit(Seconds),
							PossibleSolutions: `
								- **Get details on the exact queries that are slow** by configuring '"observability.logSlowSearches": 20,' in the site configuration and looking for 'frontend' warning logs prefixed with 'slow search request' for additional details.
								- **Check that most repositories are indexed** by visiting https://sourcegraph.example.com/site-admin/repositories?filter=needs-index (it should show few or no results.)
								- **Kubernetes:** Check CPU usage of zoekt-webserver in the indexed-search pod, consider increasing CPU limits in the 'indexed-search.Deployment.yaml' if regularly hitting max CPU utilization.
								- **Docker Compose:** Check CPU usage on the Zoekt Web Server dashboard, consider increasing 'cpus:' of the zoekt-webserver container in 'docker-compose.yml' if regularly hitting max CPU utilization.
							`,
						},
						{
							Name:            "90th_percentile_search_request_duration",
							Description:     "90th percentile successful search request duration over 5m",
							Query:           `histogram_quantile(0.90, sum by (le)(rate(src_graphql_field_seconds_bucket{type="Search",field="results",error="false",source="browser",name!="CodeIntelSearch"}[5m])))`,
							DataMayNotExist: true,
							DataMayBeNaN:    true, // See https://github.com/sourcegraph/sourcegraph/issues/9834
							Warning:         Alert{GreaterOrEqual: 15},
							PanelOptions:    PanelOptions().LegendFormat("duration").Unit(Seconds),
							PossibleSolutions: `
								- **Get details on the exact queries that are slow** by configuring '"observability.logSlowSearches": 15,' in the site configuration and looking for 'frontend' warning logs prefixed with 'slow search request' for additional details.
								- **Check that most repositories are indexed** by visiting https://sourcegraph.example.com/site-admin/repositories?filter=needs-index (it should show few or no results.)
								- **Kubernetes:** Check CPU usage of zoekt-webserver in the indexed-search pod, consider increasing CPU limits in the 'indexed-search.Deployment.yaml' if regularly hitting max CPU utilization.
								- **Docker Compose:** Check CPU usage on the Zoekt Web Server dashboard, consider increasing 'cpus:' of the zoekt-webserver container in 'docker-compose.yml' if regularly hitting max CPU utilization.
							`,
						},
					},
					{
						{
							Name:              "hard_timeout_search_responses",
							Description:       "hard timeout search responses every 5m",
							Query:             `sum(sum by (status)(increase(src_graphql_search_response{status="timeout",source="browser",name!="CodeIntelSearch"}[5m]))) + sum(sum by (status, alert_type)(increase(src_graphql_search_response{status="alert",alert_type="timed_out",source="browser",name!="CodeIntelSearch"}[5m])))`,
							DataMayNotExist:   true,
							Warning:           Alert{GreaterOrEqual: 5},
							Critical:          Alert{GreaterOrEqual: 20},
							PanelOptions:      PanelOptions().LegendFormat("hard timeout"),
							PossibleSolutions: "none",
						},
						{
							Name:              "hard_error_search_responses",
							Description:       "hard error search responses every 5m",
							Query:             `sum by (status)(increase(src_graphql_search_response{status=~"error",source="browser",name!="CodeIntelSearch"}[5m]))`,
							DataMayNotExist:   true,
							Warning:           Alert{GreaterOrEqual: 5},
							Critical:          Alert{GreaterOrEqual: 20},
							PanelOptions:      PanelOptions().LegendFormat("hard error"),
							PossibleSolutions: "none",
						},
						{
							Name:              "partial_timeout_search_responses",
							Description:       "partial timeout search responses every 5m",
							Query:             `sum by (status)(increase(src_graphql_search_response{status="partial_timeout",source="browser",name!="CodeIntelSearch"}[5m]))`,
							DataMayNotExist:   true,
							Warning:           Alert{GreaterOrEqual: 5},
							PanelOptions:      PanelOptions().LegendFormat("partial timeout"),
							PossibleSolutions: "none",
						},
						{
							Name:            "search_alert_user_suggestions",
							Description:     "search alert user suggestions shown every 5m",
							Query:           `sum by (alert_type)(increase(src_graphql_search_response{status="alert",alert_type!~"timed_out",source="browser",name!="CodeIntelSearch"}[5m]))`,
							DataMayNotExist: true,
							Warning:         Alert{GreaterOrEqual: 50},
							PanelOptions:    PanelOptions().LegendFormat("{{alert_type}}"),
							PossibleSolutions: `
								- This indicates your user's are making syntax errors or similar user errors.
							`,
						},
					},
				},
			},
			{
				Title:  "Search-based code intelligence at a glance",
				Hidden: true,
				Rows: []Row{
					{
						{
							Name:            "99th_percentile_search_codeintel_request_duration",
							Description:     "99th percentile code-intel successful search request duration over 5m",
							Query:           `histogram_quantile(0.99, sum by (le)(rate(src_graphql_field_seconds_bucket{type="Search",field="results",error="false",source="browser",request_name="CodeIntelSearch"}[5m])))`,
							DataMayNotExist: true,
							DataMayBeNaN:    true, // See https://github.com/sourcegraph/sourcegraph/issues/9834
							Warning:         Alert{GreaterOrEqual: 20},
							PanelOptions:    PanelOptions().LegendFormat("duration").Unit(Seconds),
							PossibleSolutions: `
								- **Get details on the exact queries that are slow** by configuring '"observability.logSlowSearches": 20,' in the site configuration and looking for 'frontend' warning logs prefixed with 'slow search request' for additional details.
								- **Check that most repositories are indexed** by visiting https://sourcegraph.example.com/site-admin/repositories?filter=needs-index (it should show few or no results.)
								- **Kubernetes:** Check CPU usage of zoekt-webserver in the indexed-search pod, consider increasing CPU limits in the 'indexed-search.Deployment.yaml' if regularly hitting max CPU utilization.
								- **Docker Compose:** Check CPU usage on the Zoekt Web Server dashboard, consider increasing 'cpus:' of the zoekt-webserver container in 'docker-compose.yml' if regularly hitting max CPU utilization.
							`,
						},
						{
							Name:            "90th_percentile_search_codeintel_request_duration",
							Description:     "90th percentile code-intel successful search request duration over 5m",
							Query:           `histogram_quantile(0.90, sum by (le)(rate(src_graphql_field_seconds_bucket{type="Search",field="results",error="false",source="browser",request_name="CodeIntelSearch"}[5m])))`,
							DataMayNotExist: true,
							DataMayBeNaN:    true, // See https://github.com/sourcegraph/sourcegraph/issues/9834
							Warning:         Alert{GreaterOrEqual: 15},
							PanelOptions:    PanelOptions().LegendFormat("duration").Unit(Seconds),
							PossibleSolutions: `
								- **Get details on the exact queries that are slow** by configuring '"observability.logSlowSearches": 15,' in the site configuration and looking for 'frontend' warning logs prefixed with 'slow search request' for additional details.
								- **Check that most repositories are indexed** by visiting https://sourcegraph.example.com/site-admin/repositories?filter=needs-index (it should show few or no results.)
								- **Kubernetes:** Check CPU usage of zoekt-webserver in the indexed-search pod, consider increasing CPU limits in the 'indexed-search.Deployment.yaml' if regularly hitting max CPU utilization.
								- **Docker Compose:** Check CPU usage on the Zoekt Web Server dashboard, consider increasing 'cpus:' of the zoekt-webserver container in 'docker-compose.yml' if regularly hitting max CPU utilization.
							`,
						},
					},
					{
						{
							Name:              "hard_timeout_search_codeintel_responses",
							Description:       "hard timeout search code-intel responses every 5m",
							Query:             `sum(sum by (status)(increase(src_graphql_search_response{status="timeout",source="browser",request_name="CodeIntelSearch"}[5m]))) + sum(sum by (status, alert_type)(increase(src_graphql_search_response{status="alert",alert_type="timed_out",source="browser",request_name="CodeIntelSearch"}[5m])))`,
							DataMayNotExist:   true,
							Warning:           Alert{GreaterOrEqual: 5},
							Critical:          Alert{GreaterOrEqual: 20},
							PanelOptions:      PanelOptions().LegendFormat("hard timeout"),
							PossibleSolutions: "none",
						},
						{
							Name:              "hard_error_search_codeintel_responses",
							Description:       "hard error search code-intel responses every 5m",
							Query:             `sum by (status)(increase(src_graphql_search_response{status=~"error",source="browser",request_name="CodeIntelSearch"}[5m]))`,
							DataMayNotExist:   true,
							Warning:           Alert{GreaterOrEqual: 5},
							Critical:          Alert{GreaterOrEqual: 20},
							PanelOptions:      PanelOptions().LegendFormat("hard error"),
							PossibleSolutions: "none",
						},
						{
							Name:              "partial_timeout_search_codeintel_responses",
							Description:       "partial timeout search code-intel responses every 5m",
							Query:             `sum by (status)(increase(src_graphql_search_response{status="partial_timeout",source="browser",request_name="CodeIntelSearch"}[5m]))`,
							DataMayNotExist:   true,
							Warning:           Alert{GreaterOrEqual: 5},
							PanelOptions:      PanelOptions().LegendFormat("partial timeout"),
							PossibleSolutions: "none",
						},
						{
							Name:            "search_codeintel_alert_user_suggestions",
							Description:     "search code-intel alert user suggestions shown every 5m",
							Query:           `sum by (alert_type)(increase(src_graphql_search_response{status="alert",alert_type!~"timed_out",source="browser",request_name="CodeIntelSearch"}[5m]))`,
							DataMayNotExist: true,
							Warning:         Alert{GreaterOrEqual: 50},
							PanelOptions:    PanelOptions().LegendFormat("{{alert_type}}"),
							PossibleSolutions: `
								- This indicates a bug in Sourcegraph, please [open an issue](https://github.com/sourcegraph/sourcegraph/issues/new/choose).
							`,
						},
					},
				},
			},
			{
				Title:  "Search API usage at a glance",
				Hidden: true,
				Rows: []Row{
					{
						{
							Name:            "99th_percentile_search_api_request_duration",
							Description:     "99th percentile successful search API request duration over 5m",
							Query:           `histogram_quantile(0.99, sum by (le)(rate(src_graphql_field_seconds_bucket{type="Search",field="results",error="false",source="other"}[5m])))`,
							DataMayNotExist: true,
							DataMayBeNaN:    true, // See https://github.com/sourcegraph/sourcegraph/issues/9834
							Warning:         Alert{GreaterOrEqual: 50},
							PanelOptions:    PanelOptions().LegendFormat("duration").Unit(Seconds),
							PossibleSolutions: `
								- **Get details on the exact queries that are slow** by configuring '"observability.logSlowSearches": 20,' in the site configuration and looking for 'frontend' warning logs prefixed with 'slow search request' for additional details.
								- **If your users are requesting many results** with a large 'count:' parameter, consider using our [search pagination API](../../api/graphql/search.md).
								- **Check that most repositories are indexed** by visiting https://sourcegraph.example.com/site-admin/repositories?filter=needs-index (it should show few or no results.)
								- **Kubernetes:** Check CPU usage of zoekt-webserver in the indexed-search pod, consider increasing CPU limits in the 'indexed-search.Deployment.yaml' if regularly hitting max CPU utilization.
								- **Docker Compose:** Check CPU usage on the Zoekt Web Server dashboard, consider increasing 'cpus:' of the zoekt-webserver container in 'docker-compose.yml' if regularly hitting max CPU utilization.
							`,
						},
						{
							Name:            "90th_percentile_search_api_request_duration",
							Description:     "90th percentile successful search API request duration over 5m",
							Query:           `histogram_quantile(0.90, sum by (le)(rate(src_graphql_field_seconds_bucket{type="Search",field="results",error="false",source="other"}[5m])))`,
							DataMayNotExist: true,
							DataMayBeNaN:    true, // See https://github.com/sourcegraph/sourcegraph/issues/9834
							Warning:         Alert{GreaterOrEqual: 40},
							PanelOptions:    PanelOptions().LegendFormat("duration").Unit(Seconds),
							PossibleSolutions: `
								- **Get details on the exact queries that are slow** by configuring '"observability.logSlowSearches": 15,' in the site configuration and looking for 'frontend' warning logs prefixed with 'slow search request' for additional details.
								- **If your users are requesting many results** with a large 'count:' parameter, consider using our [search pagination API](../../api/graphql/search.md).
								- **Check that most repositories are indexed** by visiting https://sourcegraph.example.com/site-admin/repositories?filter=needs-index (it should show few or no results.)
								- **Kubernetes:** Check CPU usage of zoekt-webserver in the indexed-search pod, consider increasing CPU limits in the 'indexed-search.Deployment.yaml' if regularly hitting max CPU utilization.
								- **Docker Compose:** Check CPU usage on the Zoekt Web Server dashboard, consider increasing 'cpus:' of the zoekt-webserver container in 'docker-compose.yml' if regularly hitting max CPU utilization.
							`,
						},
					},
					{
						{
							Name:              "hard_timeout_search_api_responses",
							Description:       "hard timeout search API responses every 5m",
							Query:             `sum(sum by (status)(increase(src_graphql_search_response{status="timeout",source="other"}[5m]))) + sum(sum by (status, alert_type)(increase(src_graphql_search_response{status="alert",alert_type="timed_out",source="other"}[5m])))`,
							DataMayNotExist:   true,
							Warning:           Alert{GreaterOrEqual: 5},
							Critical:          Alert{GreaterOrEqual: 20},
							PanelOptions:      PanelOptions().LegendFormat("hard timeout"),
							PossibleSolutions: "none",
						},
						{
							Name:              "hard_error_search_api_responses",
							Description:       "hard error search API responses every 5m",
							Query:             `sum by (status)(increase(src_graphql_search_response{status=~"error",source="other"}[5m]))`,
							DataMayNotExist:   true,
							Warning:           Alert{GreaterOrEqual: 5},
							Critical:          Alert{GreaterOrEqual: 20},
							PanelOptions:      PanelOptions().LegendFormat("hard error"),
							PossibleSolutions: "none",
						},
						{
							Name:              "partial_timeout_search_api_responses",
							Description:       "partial timeout search API responses every 5m",
							Query:             `sum by (status)(increase(src_graphql_search_response{status="partial_timeout",source="other"}[5m]))`,
							DataMayNotExist:   true,
							Warning:           Alert{GreaterOrEqual: 5},
							PanelOptions:      PanelOptions().LegendFormat("partial timeout"),
							PossibleSolutions: "none",
						},
						{
							Name:            "search_api_alert_user_suggestions",
							Description:     "search API alert user suggestions shown every 5m",
							Query:           `sum by (alert_type)(increase(src_graphql_search_response{status="alert",alert_type!~"timed_out",source="other"}[5m]))`,
							DataMayNotExist: true,
							Warning:         Alert{GreaterOrEqual: 50},
							PanelOptions:    PanelOptions().LegendFormat("{{alert_type}}"),
							PossibleSolutions: `
								- This indicates your user's search API requests have syntax errors or a similar user error. Check the responses the API sends back for an explanation.
							`,
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
							PossibleSolutions: `
								- Check the Zoekt Web Server dashboard for indications it might be unhealthy.
							`,
						},
						{
							Name:            "internal_unindexed_search_error_responses",
							Description:     "internal unindexed search error responses every 5m",
							Query:           `sum by (code)(increase(searcher_service_request_total{code!~"2.."}[5m]))`,
							DataMayNotExist: true,
							Warning:         Alert{GreaterOrEqual: 5},
							PanelOptions:    PanelOptions().LegendFormat("{{code}}"),
							PossibleSolutions: `
								- Check the Searcher dashboard for indications it might be unhealthy.
							`,
						},
						{
							Name:            "internal_api_error_responses",
							Description:     "internal API error responses every 5m by route",
							Query:           `sum by (category)(increase(src_frontend_internal_request_duration_seconds_count{code!~"2.."}[5m]))`,
							DataMayNotExist: true,
							Warning:         Alert{GreaterOrEqual: 25},
							PanelOptions:    PanelOptions().LegendFormat("{{category}}"),
							PossibleSolutions: `
								- May not be a substantial issue, check the 'frontend' logs for potential causes.
							`,
						},
					},
				},
			},
			{
				Title:  "Container monitoring (not available on k8s or server)",
				Hidden: true,
				Rows: []Row{
					{
						sharedContainerRestarts("frontend"),
						sharedContainerMemoryUsage("frontend"),
						sharedContainerCPUUsage("frontend"),
					},
				},
			},
		},
	}
}
