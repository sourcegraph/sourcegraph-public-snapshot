package definitions

import (
	"time"

	"github.com/sourcegraph/sourcegraph/monitoring/monitoring"
)

func Frontend() *monitoring.Container {
	return &monitoring.Container{
		Name:        "frontend",
		Title:       "Frontend",
		Description: "Serves all end-user browser and API requests.",
		Groups: []monitoring.Group{
			{
				Title: "Search at a glance",
				Rows: []monitoring.Row{
					{
						{
							Name:            "99th_percentile_search_request_duration",
							Description:     "99th percentile successful search request duration over 5m",
							Query:           `histogram_quantile(0.99, sum by (le)(rate(src_graphql_field_seconds_bucket{type="Search",field="results",error="false",source="browser",request_name!="CodeIntelSearch"}[5m])))`,
							DataMayNotExist: true,

							Warning:      monitoring.Alert().GreaterOrEqual(20),
							PanelOptions: monitoring.PanelOptions().LegendFormat("duration").Unit(monitoring.Seconds),
							Owner:        monitoring.ObservableOwnerSearch,
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
							Query:           `histogram_quantile(0.90, sum by (le)(rate(src_graphql_field_seconds_bucket{type="Search",field="results",error="false",source="browser",request_name!="CodeIntelSearch"}[5m])))`,
							DataMayNotExist: true,

							Warning:      monitoring.Alert().GreaterOrEqual(15),
							PanelOptions: monitoring.PanelOptions().LegendFormat("duration").Unit(monitoring.Seconds),
							Owner:        monitoring.ObservableOwnerSearch,
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
							Name:            "hard_timeout_search_responses",
							Description:     "hard timeout search responses every 5m",
							Query:           `(sum(increase(src_graphql_search_response{status="timeout",source="browser",request_name!="CodeIntelSearch"}[5m])) + sum(increase(src_graphql_search_response{status="alert",alert_type="timed_out",source="browser",request_name!="CodeIntelSearch"}[5m]))) / sum(increase(src_graphql_search_response{source="browser",request_name!="CodeIntelSearch"}[5m])) * 100`,
							DataMayNotExist: true,

							Warning:           monitoring.Alert().GreaterOrEqual(2).For(15 * time.Minute),
							Critical:          monitoring.Alert().GreaterOrEqual(5).For(15 * time.Minute),
							PanelOptions:      monitoring.PanelOptions().LegendFormat("hard timeout").Unit(monitoring.Percentage),
							Owner:             monitoring.ObservableOwnerSearch,
							PossibleSolutions: "none",
						},
						{
							Name:            "hard_error_search_responses",
							Description:     "hard error search responses every 5m",
							Query:           `sum by (status)(increase(src_graphql_search_response{status=~"error",source="browser",request_name!="CodeIntelSearch"}[5m])) / ignoring(status) group_left sum(increase(src_graphql_search_response{source="browser",request_name!="CodeIntelSearch"}[5m])) * 100`,
							DataMayNotExist: true,

							Warning:           monitoring.Alert().GreaterOrEqual(2).For(15 * time.Minute),
							Critical:          monitoring.Alert().GreaterOrEqual(5).For(15 * time.Minute),
							PanelOptions:      monitoring.PanelOptions().LegendFormat("{{status}}").Unit(monitoring.Percentage),
							Owner:             monitoring.ObservableOwnerSearch,
							PossibleSolutions: "none",
						},
						{
							Name:            "partial_timeout_search_responses",
							Description:     "partial timeout search responses every 5m",
							Query:           `sum by (status)(increase(src_graphql_search_response{status="partial_timeout",source="browser",request_name!="CodeIntelSearch"}[5m])) / ignoring(status) group_left sum(increase(src_graphql_search_response{source="browser",request_name!="CodeIntelSearch"}[5m])) * 100`,
							DataMayNotExist: true,

							Warning:           monitoring.Alert().GreaterOrEqual(5).For(15 * time.Minute),
							PanelOptions:      monitoring.PanelOptions().LegendFormat("{{status}}").Unit(monitoring.Percentage),
							Owner:             monitoring.ObservableOwnerSearch,
							PossibleSolutions: "none",
						},
						{
							Name:            "search_alert_user_suggestions",
							Description:     "search alert user suggestions shown every 5m",
							Query:           `sum by (alert_type)(increase(src_graphql_search_response{status="alert",alert_type!~"timed_out|no_results__suggest_quotes",source="browser",request_name!="CodeIntelSearch"}[5m])) / ignoring(alert_type) group_left sum(increase(src_graphql_search_response{source="browser",request_name!="CodeIntelSearch"}[5m])) * 100`,
							DataMayNotExist: true,

							Warning:      monitoring.Alert().GreaterOrEqual(5).For(15 * time.Minute),
							PanelOptions: monitoring.PanelOptions().LegendFormat("{{alert_type}}").Unit(monitoring.Percentage),
							Owner:        monitoring.ObservableOwnerSearch,
							PossibleSolutions: `
								- This indicates your user's are making syntax errors or similar user errors.
							`,
						},
					},
					{
						{
							Name:            "page_load_latency",
							Description:     "90th percentile page load latency over all routes over 10m",
							Query:           `histogram_quantile(0.9, sum by(le) (rate(src_http_request_duration_seconds_bucket{route!="raw",route!="blob",route!~"graphql.*"}[10m])))`,
							DataMayNotExist: true,

							Critical:     monitoring.Alert().GreaterOrEqual(2),
							PanelOptions: monitoring.PanelOptions().LegendFormat("latency").Unit(monitoring.Seconds),
							Owner:        monitoring.ObservableOwnerCloud,
							PossibleSolutions: `
								- Confirm that the Sourcegraph frontend has enough CPU/memory using the provisioning panels.
								- Trace a request to see what the slowest part is: https://docs.sourcegraph.com/admin/observability/tracing
							`,
						},
						{
							Name:            "blob_load_latency",
							Description:     "90th percentile blob load latency over 10m",
							Query:           `histogram_quantile(0.9, sum by(le) (rate(src_http_request_duration_seconds_bucket{route="blob"}[10m])))`,
							DataMayNotExist: true,
							Critical:        monitoring.Alert().GreaterOrEqual(5),
							PanelOptions:    monitoring.PanelOptions().LegendFormat("latency").Unit(monitoring.Seconds),
							Owner:           monitoring.ObservableOwnerCloud,
							PossibleSolutions: `
								- Confirm that the Sourcegraph frontend has enough CPU/memory using the provisioning panels.
								- Trace a request to see what the slowest part is: https://docs.sourcegraph.com/admin/observability/tracing
							`,
						},
					},
				},
			},
			{
				Title:  "Search-based code intelligence at a glance",
				Hidden: true,
				Rows: []monitoring.Row{
					{
						{
							Name:            "99th_percentile_search_codeintel_request_duration",
							Description:     "99th percentile code-intel successful search request duration over 5m",
							Owner:           monitoring.ObservableOwnerCodeIntel,
							Query:           `histogram_quantile(0.99, sum by (le)(rate(src_graphql_field_seconds_bucket{type="Search",field="results",error="false",source="browser",request_name="CodeIntelSearch"}[5m])))`,
							DataMayNotExist: true,

							Warning:      monitoring.Alert().GreaterOrEqual(20),
							PanelOptions: monitoring.PanelOptions().LegendFormat("duration").Unit(monitoring.Seconds),
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

							Warning:      monitoring.Alert().GreaterOrEqual(15),
							PanelOptions: monitoring.PanelOptions().LegendFormat("duration").Unit(monitoring.Seconds),
							Owner:        monitoring.ObservableOwnerCodeIntel,
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
							Name:            "hard_timeout_search_codeintel_responses",
							Description:     "hard timeout search code-intel responses every 5m",
							Query:           `(sum(increase(src_graphql_search_response{status="timeout",source="browser",request_name="CodeIntelSearch"}[5m])) + sum(increase(src_graphql_search_response{status="alert",alert_type="timed_out",source="browser",request_name="CodeIntelSearch"}[5m]))) / sum(increase(src_graphql_search_response{source="browser",request_name="CodeIntelSearch"}[5m])) * 100`,
							DataMayNotExist: true,

							Warning:           monitoring.Alert().GreaterOrEqual(2).For(15 * time.Minute),
							Critical:          monitoring.Alert().GreaterOrEqual(5).For(15 * time.Minute),
							PanelOptions:      monitoring.PanelOptions().LegendFormat("hard timeout").Unit(monitoring.Percentage),
							Owner:             monitoring.ObservableOwnerCodeIntel,
							PossibleSolutions: "none",
						},
						{
							Name:            "hard_error_search_codeintel_responses",
							Description:     "hard error search code-intel responses every 5m",
							Query:           `sum by (status)(increase(src_graphql_search_response{status=~"error",source="browser",request_name="CodeIntelSearch"}[5m])) / ignoring(status) group_left sum(increase(src_graphql_search_response{source="browser",request_name="CodeIntelSearch"}[5m])) * 100`,
							DataMayNotExist: true,

							Warning:           monitoring.Alert().GreaterOrEqual(2).For(15 * time.Minute),
							Critical:          monitoring.Alert().GreaterOrEqual(5).For(15 * time.Minute),
							PanelOptions:      monitoring.PanelOptions().LegendFormat("hard error").Unit(monitoring.Percentage),
							Owner:             monitoring.ObservableOwnerCodeIntel,
							PossibleSolutions: "none",
						},
						{
							Name:            "partial_timeout_search_codeintel_responses",
							Description:     "partial timeout search code-intel responses every 5m",
							Query:           `sum by (status)(increase(src_graphql_search_response{status="partial_timeout",source="browser",request_name="CodeIntelSearch"}[5m])) / ignoring(status) group_left sum(increase(src_graphql_search_response{status="partial_timeout",source="browser",request_name="CodeIntelSearch"}[5m])) * 100`,
							DataMayNotExist: true,

							Warning:           monitoring.Alert().GreaterOrEqual(5).For(15 * time.Minute),
							PanelOptions:      monitoring.PanelOptions().LegendFormat("partial timeout").Unit(monitoring.Percentage),
							Owner:             monitoring.ObservableOwnerCodeIntel,
							PossibleSolutions: "none",
						},
						{
							Name:            "search_codeintel_alert_user_suggestions",
							Description:     "search code-intel alert user suggestions shown every 5m",
							Query:           `sum by (alert_type)(increase(src_graphql_search_response{status="alert",alert_type!~"timed_out",source="browser",request_name="CodeIntelSearch"}[5m])) / ignoring(alert_type) group_left sum(increase(src_graphql_search_response{source="browser",request_name="CodeIntelSearch"}[5m])) * 100`,
							DataMayNotExist: true,

							Warning:      monitoring.Alert().GreaterOrEqual(5).For(15 * time.Minute),
							PanelOptions: monitoring.PanelOptions().LegendFormat("{{alert_type}}").Unit(monitoring.Percentage),
							Owner:        monitoring.ObservableOwnerCodeIntel,
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
				Rows: []monitoring.Row{
					{
						{
							Name:            "99th_percentile_search_api_request_duration",
							Description:     "99th percentile successful search API request duration over 5m",
							Query:           `histogram_quantile(0.99, sum by (le)(rate(src_graphql_field_seconds_bucket{type="Search",field="results",error="false",source="other"}[5m])))`,
							DataMayNotExist: true,

							Warning:      monitoring.Alert().GreaterOrEqual(50),
							PanelOptions: monitoring.PanelOptions().LegendFormat("duration").Unit(monitoring.Seconds),
							Owner:        monitoring.ObservableOwnerSearch,
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

							Warning:      monitoring.Alert().GreaterOrEqual(40),
							PanelOptions: monitoring.PanelOptions().LegendFormat("duration").Unit(monitoring.Seconds),
							Owner:        monitoring.ObservableOwnerSearch,
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
							Name:            "hard_timeout_search_api_responses",
							Description:     "hard timeout search API responses every 5m",
							Query:           `(sum(increase(src_graphql_search_response{status="timeout",source="other"}[5m])) + sum(increase(src_graphql_search_response{status="alert",alert_type="timed_out",source="other"}[5m]))) / sum(increase(src_graphql_search_response{source="other"}[5m])) * 100`,
							DataMayNotExist: true,

							Warning:           monitoring.Alert().GreaterOrEqual(2).For(15 * time.Minute),
							Critical:          monitoring.Alert().GreaterOrEqual(5).For(15 * time.Minute),
							PanelOptions:      monitoring.PanelOptions().LegendFormat("hard timeout").Unit(monitoring.Percentage),
							Owner:             monitoring.ObservableOwnerSearch,
							PossibleSolutions: "none",
						},
						{
							Name:            "hard_error_search_api_responses",
							Description:     "hard error search API responses every 5m",
							Query:           `sum by (status)(increase(src_graphql_search_response{status=~"error",source="other"}[5m])) / ignoring(status) group_left sum(increase(src_graphql_search_response{source="other"}[5m]))`,
							DataMayNotExist: true,

							Warning:           monitoring.Alert().GreaterOrEqual(2).For(15 * time.Minute),
							Critical:          monitoring.Alert().GreaterOrEqual(5).For(15 * time.Minute),
							PanelOptions:      monitoring.PanelOptions().LegendFormat("{{status}}").Unit(monitoring.Percentage),
							Owner:             monitoring.ObservableOwnerSearch,
							PossibleSolutions: "none",
						},
						{
							Name:            "partial_timeout_search_api_responses",
							Description:     "partial timeout search API responses every 5m",
							Query:           `sum(increase(src_graphql_search_response{status="partial_timeout",source="other"}[5m])) / sum(increase(src_graphql_search_response{source="other"}[5m]))`,
							DataMayNotExist: true,

							Warning:           monitoring.Alert().GreaterOrEqual(5).For(15 * time.Minute),
							PanelOptions:      monitoring.PanelOptions().LegendFormat("partial timeout").Unit(monitoring.Percentage),
							Owner:             monitoring.ObservableOwnerSearch,
							PossibleSolutions: "none",
						},
						{
							Name:            "search_api_alert_user_suggestions",
							Description:     "search API alert user suggestions shown every 5m",
							Query:           `sum by (alert_type)(increase(src_graphql_search_response{status="alert",alert_type!~"timed_out|no_results__suggest_quotes",source="other"}[5m])) / ignoring(alert_type) group_left sum(increase(src_graphql_search_response{status="alert",source="other"}[5m]))`,
							DataMayNotExist: true,

							Warning:      monitoring.Alert().GreaterOrEqual(5),
							PanelOptions: monitoring.PanelOptions().LegendFormat("{{alert_type}}").Unit(monitoring.Percentage),
							Owner:        monitoring.ObservableOwnerSearch,
							PossibleSolutions: `
								- This indicates your user's search API requests have syntax errors or a similar user error. Check the responses the API sends back for an explanation.
							`,
						},
					},
				},
			},
			{
				Title:  "Precise code intel usage at a glance",
				Hidden: true,
				Rows: []monitoring.Row{
					{
						{
							Name:              "codeintel_resolvers_99th_percentile_duration",
							Description:       "99th percentile successful resolvers operation duration over 5m",
							Query:             `histogram_quantile(0.99, sum by (le)(rate(src_codeintel_resolvers_duration_seconds_bucket{job="sourcegraph-frontend"}[5m])))`,
							DataMayNotExist:   true,
							Warning:           monitoring.Alert().GreaterOrEqual(20),
							PanelOptions:      monitoring.PanelOptions().LegendFormat("resolvers operation").Unit(monitoring.Seconds),
							Owner:             monitoring.ObservableOwnerCodeIntel,
							PossibleSolutions: "none",
						},
						{
							Name:              "codeintel_resolvers_errors",
							Description:       "resolvers errors every 5m",
							Query:             `sum(increase(src_codeintel_resolvers_errors_total{job="sourcegraph-frontend"}[5m]))`,
							DataMayNotExist:   true,
							Warning:           monitoring.Alert().GreaterOrEqual(20),
							PanelOptions:      monitoring.PanelOptions().LegendFormat("error"),
							Owner:             monitoring.ObservableOwnerCodeIntel,
							PossibleSolutions: "none",
						},
					},
					{
						{
							Name:              "codeintel_api_99th_percentile_duration",
							Description:       "99th percentile successful api operation duration over 5m",
							Query:             `histogram_quantile(0.99, sum by (le)(rate(src_codeintel_api_duration_seconds_bucket{job="sourcegraph-frontend"}[5m])))`,
							DataMayNotExist:   true,
							Warning:           monitoring.Alert().GreaterOrEqual(20),
							PanelOptions:      monitoring.PanelOptions().LegendFormat("api operation").Unit(monitoring.Seconds),
							Owner:             monitoring.ObservableOwnerCodeIntel,
							PossibleSolutions: "none",
						},
						{
							Name:              "codeintel_api_errors",
							Description:       "api errors every 5m",
							Query:             `sum(increase(src_codeintel_api_errors_total{job="sourcegraph-frontend"}[5m]))`,
							DataMayNotExist:   true,
							Warning:           monitoring.Alert().GreaterOrEqual(20),
							PanelOptions:      monitoring.PanelOptions().LegendFormat("error"),
							Owner:             monitoring.ObservableOwnerCodeIntel,
							PossibleSolutions: "none",
						},
					},
					{
						{
							Name:              "codeintel_dbstore_99th_percentile_duration",
							Description:       "99th percentile successful dbstore operation duration over 5m",
							Query:             `histogram_quantile(0.99, sum by (le)(rate(src_codeintel_dbstore_duration_seconds_bucket{job="sourcegraph-frontend"}[5m])))`,
							DataMayNotExist:   true,
							Warning:           monitoring.Alert().GreaterOrEqual(20),
							PanelOptions:      monitoring.PanelOptions().LegendFormat("store operation").Unit(monitoring.Seconds),
							Owner:             monitoring.ObservableOwnerCodeIntel,
							PossibleSolutions: "none",
						},
						{
							Name:              "codeintel_dbstore_errors",
							Description:       "dbstore errors every 5m",
							Query:             `sum(increase(src_codeintel_dbstore_errors_total{job="sourcegraph-frontend"}[5m]))`,
							DataMayNotExist:   true,
							Warning:           monitoring.Alert().GreaterOrEqual(20),
							PanelOptions:      monitoring.PanelOptions().LegendFormat("error"),
							Owner:             monitoring.ObservableOwnerCodeIntel,
							PossibleSolutions: "none",
						},
					},
					{
						{
							Name:              "codeintel_lsifstore_99th_percentile_duration",
							Description:       "99th percentile successful lsifstore operation duration over 5m",
							Query:             `histogram_quantile(0.99, sum by (le)(rate(src_codeintel_lsifstore_duration_seconds_bucket{job="sourcegraph-frontend"}[5m])))`,
							DataMayNotExist:   true,
							Warning:           monitoring.Alert().GreaterOrEqual(20),
							PanelOptions:      monitoring.PanelOptions().LegendFormat("store operation").Unit(monitoring.Seconds),
							Owner:             monitoring.ObservableOwnerCodeIntel,
							PossibleSolutions: "none",
						},
						{
							Name:              "codeintel_lsifstore_errors",
							Description:       "lsifstore errors every 5m",
							Query:             `sum(increase(src_codeintel_lsifstore_errors_total{job="sourcegraph-frontend"}[5m]))`,
							DataMayNotExist:   true,
							Warning:           monitoring.Alert().GreaterOrEqual(20),
							PanelOptions:      monitoring.PanelOptions().LegendFormat("error"),
							Owner:             monitoring.ObservableOwnerCodeIntel,
							PossibleSolutions: "none",
						},
					},
					{
						{
							Name:              "codeintel_uploadstore_99th_percentile_duration",
							Description:       "99th percentile successful uploadstore operation duration over 5m",
							Query:             `histogram_quantile(0.99, sum by (le)(rate(src_codeintel_uploadstore_duration_seconds_bucket{job="sourcegraph-frontend"}[5m])))`,
							DataMayNotExist:   true,
							Warning:           monitoring.Alert().GreaterOrEqual(20),
							PanelOptions:      monitoring.PanelOptions().LegendFormat("store operation").Unit(monitoring.Seconds),
							Owner:             monitoring.ObservableOwnerCodeIntel,
							PossibleSolutions: "none",
						},
						{
							Name:              "codeintel_uploadstore_errors",
							Description:       "uploadstore errors every 5m",
							Query:             `sum(increase(src_codeintel_uploadstore_errors_total{job="sourcegraph-frontend"}[5m]))`,
							DataMayNotExist:   true,
							Warning:           monitoring.Alert().GreaterOrEqual(20),
							PanelOptions:      monitoring.PanelOptions().LegendFormat("error"),
							Owner:             monitoring.ObservableOwnerCodeIntel,
							PossibleSolutions: "none",
						},
					},
					{
						{
							Name:              "codeintel_gitserver_99th_percentile_duration",
							Description:       "99th percentile successful gitserver operation duration over 5m",
							Query:             `histogram_quantile(0.99, sum by (le)(rate(src_codeintel_gitserver_duration_seconds_bucket{job="sourcegraph-frontend"}[5m])))`,
							DataMayNotExist:   true,
							Warning:           monitoring.Alert().GreaterOrEqual(20),
							PanelOptions:      monitoring.PanelOptions().LegendFormat("store operation").Unit(monitoring.Seconds),
							Owner:             monitoring.ObservableOwnerCodeIntel,
							PossibleSolutions: "none",
						},
						{
							Name:              "codeintel_gitserver_errors",
							Description:       "gitserver errors every 5m",
							Query:             `sum(increase(src_codeintel_gitserver_errors_total{job="sourcegraph-frontend"}[5m]))`,
							DataMayNotExist:   true,
							Warning:           monitoring.Alert().GreaterOrEqual(20),
							PanelOptions:      monitoring.PanelOptions().LegendFormat("error"),
							Owner:             monitoring.ObservableOwnerCodeIntel,
							PossibleSolutions: "none",
						},
					},
				},
			},
			{
				Title:  "Internal service requests",
				Hidden: true,
				Rows: []monitoring.Row{
					{
						{
							Name:            "internal_indexed_search_error_responses",
							Description:     "internal indexed search error responses every 5m",
							Query:           `sum by(code) (increase(src_zoekt_request_duration_seconds_count{code!~"2.."}[5m])) / ignoring(code) group_left sum(increase(src_zoekt_request_duration_seconds_count[5m])) * 100`,
							DataMayNotExist: true,
							Warning:         monitoring.Alert().GreaterOrEqual(5).For(15 * time.Minute),
							PanelOptions:    monitoring.PanelOptions().LegendFormat("{{code}}").Unit(monitoring.Percentage),
							Owner:           monitoring.ObservableOwnerSearch,
							PossibleSolutions: `
								- Check the Zoekt Web Server dashboard for indications it might be unhealthy.
							`,
						},
						{
							Name:            "internal_unindexed_search_error_responses",
							Description:     "internal unindexed search error responses every 5m",
							Query:           `sum by(code) (increase(searcher_service_request_total{code!~"2.."}[5m])) / ignoring(code) group_left sum(increase(searcher_service_request_total[5m])) * 100`,
							DataMayNotExist: true,
							Warning:         monitoring.Alert().GreaterOrEqual(5).For(15 * time.Minute),
							PanelOptions:    monitoring.PanelOptions().LegendFormat("{{code}}").Unit(monitoring.Percentage),
							Owner:           monitoring.ObservableOwnerSearch,
							PossibleSolutions: `
								- Check the Searcher dashboard for indications it might be unhealthy.
							`,
						},
						{
							Name:            "internal_api_error_responses",
							Description:     "internal API error responses every 5m by route",
							Query:           `sum by(category) (increase(src_frontend_internal_request_duration_seconds_count{code!~"2.."}[5m])) / ignoring(code) group_left sum(increase(src_frontend_internal_request_duration_seconds_count[5m])) * 100`,
							DataMayNotExist: true,
							Warning:         monitoring.Alert().GreaterOrEqual(5).For(15 * time.Minute),
							PanelOptions:    monitoring.PanelOptions().LegendFormat("{{category}}").Unit(monitoring.Percentage),
							Owner:           monitoring.ObservableOwnerCloud,
							PossibleSolutions: `
								- May not be a substantial issue, check the 'frontend' logs for potential causes.
							`,
						},
					},
					{
						{
							Name:              "99th_percentile_gitserver_duration",
							Description:       "99th percentile successful gitserver query duration over 5m",
							Query:             `histogram_quantile(0.99, sum by (le,category)(rate(src_gitserver_request_duration_seconds_bucket{job="sourcegraph-frontend"}[5m])))`,
							DataMayNotExist:   true,
							Warning:           monitoring.Alert().GreaterOrEqual(20),
							PanelOptions:      monitoring.PanelOptions().LegendFormat("{{category}}").Unit(monitoring.Seconds),
							Owner:             monitoring.ObservableOwnerCloud,
							PossibleSolutions: "none",
						},
						{
							Name:              "gitserver_error_responses",
							Description:       "gitserver error responses every 5m",
							Query:             `sum by (category)(increase(src_gitserver_request_duration_seconds_count{job="sourcegraph-frontend",code!~"2.."}[5m])) / ignoring(code) group_left sum by (category)(increase(src_gitserver_request_duration_seconds_count{job="sourcegraph-frontend"}[5m])) * 100`,
							DataMayNotExist:   true,
							Warning:           monitoring.Alert().GreaterOrEqual(5).For(15 * time.Minute),
							PanelOptions:      monitoring.PanelOptions().LegendFormat("{{category}}").Unit(monitoring.Percentage),
							Owner:             monitoring.ObservableOwnerCloud,
							PossibleSolutions: "none",
						},
					},
					{
						{
							Name:              "observability_test_alert_warning",
							Description:       "warning test alert metric",
							Query:             `max by(owner) (observability_test_metric_warning)`,
							DataMayNotExist:   true,
							Warning:           monitoring.Alert().GreaterOrEqual(1),
							PanelOptions:      monitoring.PanelOptions().Max(1),
							Owner:             monitoring.ObservableOwnerDistribution,
							PossibleSolutions: "This alert is triggered via the `triggerObservabilityTestAlert` GraphQL endpoint, and will automatically resolve itself.",
						},
						{
							Name:              "observability_test_alert_critical",
							Description:       "critical test alert metric",
							Query:             `max by(owner) (observability_test_metric_critical)`,
							DataMayNotExist:   true,
							Critical:          monitoring.Alert().GreaterOrEqual(1),
							PanelOptions:      monitoring.PanelOptions().Max(1),
							Owner:             monitoring.ObservableOwnerDistribution,
							PossibleSolutions: "This alert is triggered via the `triggerObservabilityTestAlert` GraphQL endpoint, and will automatically resolve itself.",
						},
					},
				},
			},
			{
				Title:  "Container monitoring (not available on server)",
				Hidden: true,
				Rows: []monitoring.Row{
					{
						sharedContainerCPUUsage("frontend", monitoring.ObservableOwnerCloud),
						sharedContainerMemoryUsage("frontend", monitoring.ObservableOwnerCloud),
					},
					{
						sharedContainerRestarts("frontend", monitoring.ObservableOwnerCloud),
						sharedContainerFsInodes("frontend", monitoring.ObservableOwnerCloud),
					},
				},
			},
			{
				Title:  "Provisioning indicators (not available on server)",
				Hidden: true,
				Rows: []monitoring.Row{
					{
						sharedProvisioningCPUUsageLongTerm("frontend", monitoring.ObservableOwnerCloud),
						sharedProvisioningMemoryUsageLongTerm("frontend", monitoring.ObservableOwnerCloud),
					},
					{
						sharedProvisioningCPUUsageShortTerm("frontend", monitoring.ObservableOwnerCloud),
						sharedProvisioningMemoryUsageShortTerm("frontend", monitoring.ObservableOwnerCloud),
					},
				},
			},
			{
				Title:  "Golang runtime monitoring",
				Hidden: true,
				Rows: []monitoring.Row{
					{
						sharedGoGoroutines("frontend", monitoring.ObservableOwnerCloud),
						sharedGoGcDuration("frontend", monitoring.ObservableOwnerCloud),
					},
				},
			},
			{
				Title:  "Kubernetes monitoring (ignore if using Docker Compose or server)",
				Hidden: true,
				Rows: []monitoring.Row{
					{
						sharedKubernetesPodsAvailable("frontend", monitoring.ObservableOwnerCloud),
					},
				},
			},
		},
	}
}
