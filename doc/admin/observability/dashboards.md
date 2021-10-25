# Dashboards reference

<!-- DO NOT EDIT: generated via: go generate ./monitoring -->

This document contains a complete reference on Sourcegraph's available dashboards, as well as details on how to interpret the panels and metrics.

To learn more about Sourcegraph's metrics and how to view these dashboards, see [our metrics guide](https://docs.sourcegraph.com/admin/observability/metrics).

## Frontend

<p class="subtitle">Serves all end-user browser and API requests.</p>

To see this dashboard, visit `/-/debug/grafana/d/frontend/frontend` on your Sourcegraph instance.

### Frontend: Search at a glance

#### frontend: 99th_percentile_search_request_duration

<p class="subtitle">99th percentile successful search request duration over 5m</p>

Refer to the [alert solutions reference](./alert_solutions.md#frontend-99th-percentile-search-request-duration) for 1 alert related to this panel.

To see this panel, visit `/-/debug/grafana/d/frontend/frontend?viewPanel=100000` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Search team](https://about.sourcegraph.com/handbook/engineering/search).*</sub>

<details>
<summary>Technical details</summary>

Query: `histogram_quantile(0.99, sum by (le)(rate(src_graphql_field_seconds_bucket{type="Search",field="results",error="false",source="browser",request_name!="CodeIntelSearch"}[5m])))`

</details>

<br />

#### frontend: 90th_percentile_search_request_duration

<p class="subtitle">90th percentile successful search request duration over 5m</p>

Refer to the [alert solutions reference](./alert_solutions.md#frontend-90th-percentile-search-request-duration) for 1 alert related to this panel.

To see this panel, visit `/-/debug/grafana/d/frontend/frontend?viewPanel=100001` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Search team](https://about.sourcegraph.com/handbook/engineering/search).*</sub>

<details>
<summary>Technical details</summary>

Query: `histogram_quantile(0.90, sum by (le)(rate(src_graphql_field_seconds_bucket{type="Search",field="results",error="false",source="browser",request_name!="CodeIntelSearch"}[5m])))`

</details>

<br />

#### frontend: hard_timeout_search_responses

<p class="subtitle">Hard timeout search responses every 5m</p>

Refer to the [alert solutions reference](./alert_solutions.md#frontend-hard-timeout-search-responses) for 2 alerts related to this panel.

To see this panel, visit `/-/debug/grafana/d/frontend/frontend?viewPanel=100010` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Search team](https://about.sourcegraph.com/handbook/engineering/search).*</sub>

<details>
<summary>Technical details</summary>

Query: `(sum(increase(src_graphql_search_response{status="timeout",source="browser",request_name!="CodeIntelSearch"}[5m])) + sum(increase(src_graphql_search_response{status="alert",alert_type="timed_out",source="browser",request_name!="CodeIntelSearch"}[5m]))) / sum(increase(src_graphql_search_response{source="browser",request_name!="CodeIntelSearch"}[5m])) * 100`

</details>

<br />

#### frontend: hard_error_search_responses

<p class="subtitle">Hard error search responses every 5m</p>

Refer to the [alert solutions reference](./alert_solutions.md#frontend-hard-error-search-responses) for 2 alerts related to this panel.

To see this panel, visit `/-/debug/grafana/d/frontend/frontend?viewPanel=100011` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Search team](https://about.sourcegraph.com/handbook/engineering/search).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum by (status)(increase(src_graphql_search_response{status=~"error",source="browser",request_name!="CodeIntelSearch"}[5m])) / ignoring(status) group_left sum(increase(src_graphql_search_response{source="browser",request_name!="CodeIntelSearch"}[5m])) * 100`

</details>

<br />

#### frontend: partial_timeout_search_responses

<p class="subtitle">Partial timeout search responses every 5m</p>

Refer to the [alert solutions reference](./alert_solutions.md#frontend-partial-timeout-search-responses) for 1 alert related to this panel.

To see this panel, visit `/-/debug/grafana/d/frontend/frontend?viewPanel=100012` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Search team](https://about.sourcegraph.com/handbook/engineering/search).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum by (status)(increase(src_graphql_search_response{status="partial_timeout",source="browser",request_name!="CodeIntelSearch"}[5m])) / ignoring(status) group_left sum(increase(src_graphql_search_response{source="browser",request_name!="CodeIntelSearch"}[5m])) * 100`

</details>

<br />

#### frontend: search_alert_user_suggestions

<p class="subtitle">Search alert user suggestions shown every 5m</p>

Refer to the [alert solutions reference](./alert_solutions.md#frontend-search-alert-user-suggestions) for 1 alert related to this panel.

To see this panel, visit `/-/debug/grafana/d/frontend/frontend?viewPanel=100013` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Search team](https://about.sourcegraph.com/handbook/engineering/search).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum by (alert_type)(increase(src_graphql_search_response{status="alert",alert_type!~"timed_out|no_results__suggest_quotes",source="browser",request_name!="CodeIntelSearch"}[5m])) / ignoring(alert_type) group_left sum(increase(src_graphql_search_response{source="browser",request_name!="CodeIntelSearch"}[5m])) * 100`

</details>

<br />

#### frontend: page_load_latency

<p class="subtitle">90th percentile page load latency over all routes over 10m</p>

Refer to the [alert solutions reference](./alert_solutions.md#frontend-page-load-latency) for 1 alert related to this panel.

To see this panel, visit `/-/debug/grafana/d/frontend/frontend?viewPanel=100020` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<details>
<summary>Technical details</summary>

Query: `histogram_quantile(0.9, sum by(le) (rate(src_http_request_duration_seconds_bucket{route!="raw",route!="blob",route!~"graphql.*"}[10m])))`

</details>

<br />

#### frontend: blob_load_latency

<p class="subtitle">90th percentile blob load latency over 10m</p>

Refer to the [alert solutions reference](./alert_solutions.md#frontend-blob-load-latency) for 1 alert related to this panel.

To see this panel, visit `/-/debug/grafana/d/frontend/frontend?viewPanel=100021` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<details>
<summary>Technical details</summary>

Query: `histogram_quantile(0.9, sum by(le) (rate(src_http_request_duration_seconds_bucket{route="blob"}[10m])))`

</details>

<br />

### Frontend: Search-based code intelligence at a glance

#### frontend: 99th_percentile_search_codeintel_request_duration

<p class="subtitle">99th percentile code-intel successful search request duration over 5m</p>

Refer to the [alert solutions reference](./alert_solutions.md#frontend-99th-percentile-search-codeintel-request-duration) for 1 alert related to this panel.

To see this panel, visit `/-/debug/grafana/d/frontend/frontend?viewPanel=100100` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Search team](https://about.sourcegraph.com/handbook/engineering/search).*</sub>

<details>
<summary>Technical details</summary>

Query: `histogram_quantile(0.99, sum by (le)(rate(src_graphql_field_seconds_bucket{type="Search",field="results",error="false",source="browser",request_name="CodeIntelSearch"}[5m])))`

</details>

<br />

#### frontend: 90th_percentile_search_codeintel_request_duration

<p class="subtitle">90th percentile code-intel successful search request duration over 5m</p>

Refer to the [alert solutions reference](./alert_solutions.md#frontend-90th-percentile-search-codeintel-request-duration) for 1 alert related to this panel.

To see this panel, visit `/-/debug/grafana/d/frontend/frontend?viewPanel=100101` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Search team](https://about.sourcegraph.com/handbook/engineering/search).*</sub>

<details>
<summary>Technical details</summary>

Query: `histogram_quantile(0.90, sum by (le)(rate(src_graphql_field_seconds_bucket{type="Search",field="results",error="false",source="browser",request_name="CodeIntelSearch"}[5m])))`

</details>

<br />

#### frontend: hard_timeout_search_codeintel_responses

<p class="subtitle">Hard timeout search code-intel responses every 5m</p>

Refer to the [alert solutions reference](./alert_solutions.md#frontend-hard-timeout-search-codeintel-responses) for 2 alerts related to this panel.

To see this panel, visit `/-/debug/grafana/d/frontend/frontend?viewPanel=100110` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Search team](https://about.sourcegraph.com/handbook/engineering/search).*</sub>

<details>
<summary>Technical details</summary>

Query: `(sum(increase(src_graphql_search_response{status="timeout",source="browser",request_name="CodeIntelSearch"}[5m])) + sum(increase(src_graphql_search_response{status="alert",alert_type="timed_out",source="browser",request_name="CodeIntelSearch"}[5m]))) / sum(increase(src_graphql_search_response{source="browser",request_name="CodeIntelSearch"}[5m])) * 100`

</details>

<br />

#### frontend: hard_error_search_codeintel_responses

<p class="subtitle">Hard error search code-intel responses every 5m</p>

Refer to the [alert solutions reference](./alert_solutions.md#frontend-hard-error-search-codeintel-responses) for 2 alerts related to this panel.

To see this panel, visit `/-/debug/grafana/d/frontend/frontend?viewPanel=100111` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Search team](https://about.sourcegraph.com/handbook/engineering/search).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum by (status)(increase(src_graphql_search_response{status=~"error",source="browser",request_name="CodeIntelSearch"}[5m])) / ignoring(status) group_left sum(increase(src_graphql_search_response{source="browser",request_name="CodeIntelSearch"}[5m])) * 100`

</details>

<br />

#### frontend: partial_timeout_search_codeintel_responses

<p class="subtitle">Partial timeout search code-intel responses every 5m</p>

Refer to the [alert solutions reference](./alert_solutions.md#frontend-partial-timeout-search-codeintel-responses) for 1 alert related to this panel.

To see this panel, visit `/-/debug/grafana/d/frontend/frontend?viewPanel=100112` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Search team](https://about.sourcegraph.com/handbook/engineering/search).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum by (status)(increase(src_graphql_search_response{status="partial_timeout",source="browser",request_name="CodeIntelSearch"}[5m])) / ignoring(status) group_left sum(increase(src_graphql_search_response{status="partial_timeout",source="browser",request_name="CodeIntelSearch"}[5m])) * 100`

</details>

<br />

#### frontend: search_codeintel_alert_user_suggestions

<p class="subtitle">Search code-intel alert user suggestions shown every 5m</p>

Refer to the [alert solutions reference](./alert_solutions.md#frontend-search-codeintel-alert-user-suggestions) for 1 alert related to this panel.

To see this panel, visit `/-/debug/grafana/d/frontend/frontend?viewPanel=100113` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Search team](https://about.sourcegraph.com/handbook/engineering/search).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum by (alert_type)(increase(src_graphql_search_response{status="alert",alert_type!~"timed_out",source="browser",request_name="CodeIntelSearch"}[5m])) / ignoring(alert_type) group_left sum(increase(src_graphql_search_response{source="browser",request_name="CodeIntelSearch"}[5m])) * 100`

</details>

<br />

### Frontend: Search API usage at a glance

#### frontend: 99th_percentile_search_api_request_duration

<p class="subtitle">99th percentile successful search API request duration over 5m</p>

Refer to the [alert solutions reference](./alert_solutions.md#frontend-99th-percentile-search-api-request-duration) for 1 alert related to this panel.

To see this panel, visit `/-/debug/grafana/d/frontend/frontend?viewPanel=100200` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Search team](https://about.sourcegraph.com/handbook/engineering/search).*</sub>

<details>
<summary>Technical details</summary>

Query: `histogram_quantile(0.99, sum by (le)(rate(src_graphql_field_seconds_bucket{type="Search",field="results",error="false",source="other"}[5m])))`

</details>

<br />

#### frontend: 90th_percentile_search_api_request_duration

<p class="subtitle">90th percentile successful search API request duration over 5m</p>

Refer to the [alert solutions reference](./alert_solutions.md#frontend-90th-percentile-search-api-request-duration) for 1 alert related to this panel.

To see this panel, visit `/-/debug/grafana/d/frontend/frontend?viewPanel=100201` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Search team](https://about.sourcegraph.com/handbook/engineering/search).*</sub>

<details>
<summary>Technical details</summary>

Query: `histogram_quantile(0.90, sum by (le)(rate(src_graphql_field_seconds_bucket{type="Search",field="results",error="false",source="other"}[5m])))`

</details>

<br />

#### frontend: hard_error_search_api_responses

<p class="subtitle">Hard error search API responses every 5m</p>

Refer to the [alert solutions reference](./alert_solutions.md#frontend-hard-error-search-api-responses) for 2 alerts related to this panel.

To see this panel, visit `/-/debug/grafana/d/frontend/frontend?viewPanel=100210` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Search team](https://about.sourcegraph.com/handbook/engineering/search).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum by (status)(increase(src_graphql_search_response{status=~"error",source="other"}[5m])) / ignoring(status) group_left sum(increase(src_graphql_search_response{source="other"}[5m]))`

</details>

<br />

#### frontend: partial_timeout_search_api_responses

<p class="subtitle">Partial timeout search API responses every 5m</p>

Refer to the [alert solutions reference](./alert_solutions.md#frontend-partial-timeout-search-api-responses) for 1 alert related to this panel.

To see this panel, visit `/-/debug/grafana/d/frontend/frontend?viewPanel=100211` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Search team](https://about.sourcegraph.com/handbook/engineering/search).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum(increase(src_graphql_search_response{status="partial_timeout",source="other"}[5m])) / sum(increase(src_graphql_search_response{source="other"}[5m]))`

</details>

<br />

#### frontend: search_api_alert_user_suggestions

<p class="subtitle">Search API alert user suggestions shown every 5m</p>

Refer to the [alert solutions reference](./alert_solutions.md#frontend-search-api-alert-user-suggestions) for 1 alert related to this panel.

To see this panel, visit `/-/debug/grafana/d/frontend/frontend?viewPanel=100212` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Search team](https://about.sourcegraph.com/handbook/engineering/search).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum by (alert_type)(increase(src_graphql_search_response{status="alert",alert_type!~"timed_out|no_results__suggest_quotes",source="other"}[5m])) / ignoring(alert_type) group_left sum(increase(src_graphql_search_response{status="alert",source="other"}[5m]))`

</details>

<br />

### Frontend: Codeintel: Precise code intelligence usage at a glance

#### frontend: codeintel_resolvers_total

<p class="subtitle">Aggregate graphql operations every 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/frontend/frontend?viewPanel=100300` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum(increase(src_codeintel_resolvers_total{job=~"^(frontend|sourcegraph-frontend).*"}[5m]))`

</details>

<br />

#### frontend: codeintel_resolvers_99th_percentile_duration

<p class="subtitle">99th percentile successful aggregate graphql operation duration over 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/frontend/frontend?viewPanel=100301` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `histogram_quantile(0.99, sum  by (le)(rate(src_codeintel_resolvers_duration_seconds_bucket{job=~"^(frontend|sourcegraph-frontend).*"}[5m])))`

</details>

<br />

#### frontend: codeintel_resolvers_errors_total

<p class="subtitle">Aggregate graphql operation errors every 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/frontend/frontend?viewPanel=100302` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum(increase(src_codeintel_resolvers_errors_total{job=~"^(frontend|sourcegraph-frontend).*"}[5m]))`

</details>

<br />

#### frontend: codeintel_resolvers_error_rate

<p class="subtitle">Aggregate graphql operation error rate over 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/frontend/frontend?viewPanel=100303` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum(increase(src_codeintel_resolvers_errors_total{job=~"^(frontend|sourcegraph-frontend).*"}[5m])) / (sum(increase(src_codeintel_resolvers_total{job=~"^(frontend|sourcegraph-frontend).*"}[5m])) + sum(increase(src_codeintel_resolvers_errors_total{job=~"^(frontend|sourcegraph-frontend).*"}[5m]))) * 100`

</details>

<br />

#### frontend: codeintel_resolvers_total

<p class="subtitle">Graphql operations every 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/frontend/frontend?viewPanel=100310` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum by (op)(increase(src_codeintel_resolvers_total{job=~"^(frontend|sourcegraph-frontend).*"}[5m]))`

</details>

<br />

#### frontend: codeintel_resolvers_99th_percentile_duration

<p class="subtitle">99th percentile successful graphql operation duration over 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/frontend/frontend?viewPanel=100311` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `histogram_quantile(0.99, sum  by (le,op)(rate(src_codeintel_resolvers_duration_seconds_bucket{job=~"^(frontend|sourcegraph-frontend).*"}[5m])))`

</details>

<br />

#### frontend: codeintel_resolvers_errors_total

<p class="subtitle">Graphql operation errors every 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/frontend/frontend?viewPanel=100312` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum by (op)(increase(src_codeintel_resolvers_errors_total{job=~"^(frontend|sourcegraph-frontend).*"}[5m]))`

</details>

<br />

#### frontend: codeintel_resolvers_error_rate

<p class="subtitle">Graphql operation error rate over 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/frontend/frontend?viewPanel=100313` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum by (op)(increase(src_codeintel_resolvers_errors_total{job=~"^(frontend|sourcegraph-frontend).*"}[5m])) / (sum by (op)(increase(src_codeintel_resolvers_total{job=~"^(frontend|sourcegraph-frontend).*"}[5m])) + sum by (op)(increase(src_codeintel_resolvers_errors_total{job=~"^(frontend|sourcegraph-frontend).*"}[5m]))) * 100`

</details>

<br />

### Frontend: Codeintel: Auto-index enqueuer

#### frontend: codeintel_autoindex_enqueuer_total

<p class="subtitle">Aggregate enqueuer operations every 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/frontend/frontend?viewPanel=100400` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum(increase(src_codeintel_autoindex_enqueuer_total{job=~"^(frontend|sourcegraph-frontend).*"}[5m]))`

</details>

<br />

#### frontend: codeintel_autoindex_enqueuer_99th_percentile_duration

<p class="subtitle">99th percentile successful aggregate enqueuer operation duration over 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/frontend/frontend?viewPanel=100401` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `histogram_quantile(0.99, sum  by (le)(rate(src_codeintel_autoindex_enqueuer_duration_seconds_bucket{job=~"^(frontend|sourcegraph-frontend).*"}[5m])))`

</details>

<br />

#### frontend: codeintel_autoindex_enqueuer_errors_total

<p class="subtitle">Aggregate enqueuer operation errors every 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/frontend/frontend?viewPanel=100402` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum(increase(src_codeintel_autoindex_enqueuer_errors_total{job=~"^(frontend|sourcegraph-frontend).*"}[5m]))`

</details>

<br />

#### frontend: codeintel_autoindex_enqueuer_error_rate

<p class="subtitle">Aggregate enqueuer operation error rate over 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/frontend/frontend?viewPanel=100403` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum(increase(src_codeintel_autoindex_enqueuer_errors_total{job=~"^(frontend|sourcegraph-frontend).*"}[5m])) / (sum(increase(src_codeintel_autoindex_enqueuer_total{job=~"^(frontend|sourcegraph-frontend).*"}[5m])) + sum(increase(src_codeintel_autoindex_enqueuer_errors_total{job=~"^(frontend|sourcegraph-frontend).*"}[5m]))) * 100`

</details>

<br />

#### frontend: codeintel_autoindex_enqueuer_total

<p class="subtitle">Enqueuer operations every 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/frontend/frontend?viewPanel=100410` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum by (op)(increase(src_codeintel_autoindex_enqueuer_total{job=~"^(frontend|sourcegraph-frontend).*"}[5m]))`

</details>

<br />

#### frontend: codeintel_autoindex_enqueuer_99th_percentile_duration

<p class="subtitle">99th percentile successful enqueuer operation duration over 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/frontend/frontend?viewPanel=100411` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `histogram_quantile(0.99, sum  by (le,op)(rate(src_codeintel_autoindex_enqueuer_duration_seconds_bucket{job=~"^(frontend|sourcegraph-frontend).*"}[5m])))`

</details>

<br />

#### frontend: codeintel_autoindex_enqueuer_errors_total

<p class="subtitle">Enqueuer operation errors every 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/frontend/frontend?viewPanel=100412` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum by (op)(increase(src_codeintel_autoindex_enqueuer_errors_total{job=~"^(frontend|sourcegraph-frontend).*"}[5m]))`

</details>

<br />

#### frontend: codeintel_autoindex_enqueuer_error_rate

<p class="subtitle">Enqueuer operation error rate over 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/frontend/frontend?viewPanel=100413` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum by (op)(increase(src_codeintel_autoindex_enqueuer_errors_total{job=~"^(frontend|sourcegraph-frontend).*"}[5m])) / (sum by (op)(increase(src_codeintel_autoindex_enqueuer_total{job=~"^(frontend|sourcegraph-frontend).*"}[5m])) + sum by (op)(increase(src_codeintel_autoindex_enqueuer_errors_total{job=~"^(frontend|sourcegraph-frontend).*"}[5m]))) * 100`

</details>

<br />

### Frontend: Codeintel: dbstore stats

#### frontend: codeintel_dbstore_total

<p class="subtitle">Aggregate store operations every 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/frontend/frontend?viewPanel=100500` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum(increase(src_codeintel_dbstore_total{job=~"^(frontend|sourcegraph-frontend).*"}[5m]))`

</details>

<br />

#### frontend: codeintel_dbstore_99th_percentile_duration

<p class="subtitle">99th percentile successful aggregate store operation duration over 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/frontend/frontend?viewPanel=100501` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `histogram_quantile(0.99, sum  by (le)(rate(src_codeintel_dbstore_duration_seconds_bucket{job=~"^(frontend|sourcegraph-frontend).*"}[5m])))`

</details>

<br />

#### frontend: codeintel_dbstore_errors_total

<p class="subtitle">Aggregate store operation errors every 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/frontend/frontend?viewPanel=100502` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum(increase(src_codeintel_dbstore_errors_total{job=~"^(frontend|sourcegraph-frontend).*"}[5m]))`

</details>

<br />

#### frontend: codeintel_dbstore_error_rate

<p class="subtitle">Aggregate store operation error rate over 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/frontend/frontend?viewPanel=100503` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum(increase(src_codeintel_dbstore_errors_total{job=~"^(frontend|sourcegraph-frontend).*"}[5m])) / (sum(increase(src_codeintel_dbstore_total{job=~"^(frontend|sourcegraph-frontend).*"}[5m])) + sum(increase(src_codeintel_dbstore_errors_total{job=~"^(frontend|sourcegraph-frontend).*"}[5m]))) * 100`

</details>

<br />

#### frontend: codeintel_dbstore_total

<p class="subtitle">Store operations every 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/frontend/frontend?viewPanel=100510` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum by (op)(increase(src_codeintel_dbstore_total{job=~"^(frontend|sourcegraph-frontend).*"}[5m]))`

</details>

<br />

#### frontend: codeintel_dbstore_99th_percentile_duration

<p class="subtitle">99th percentile successful store operation duration over 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/frontend/frontend?viewPanel=100511` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `histogram_quantile(0.99, sum  by (le,op)(rate(src_codeintel_dbstore_duration_seconds_bucket{job=~"^(frontend|sourcegraph-frontend).*"}[5m])))`

</details>

<br />

#### frontend: codeintel_dbstore_errors_total

<p class="subtitle">Store operation errors every 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/frontend/frontend?viewPanel=100512` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum by (op)(increase(src_codeintel_dbstore_errors_total{job=~"^(frontend|sourcegraph-frontend).*"}[5m]))`

</details>

<br />

#### frontend: codeintel_dbstore_error_rate

<p class="subtitle">Store operation error rate over 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/frontend/frontend?viewPanel=100513` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum by (op)(increase(src_codeintel_dbstore_errors_total{job=~"^(frontend|sourcegraph-frontend).*"}[5m])) / (sum by (op)(increase(src_codeintel_dbstore_total{job=~"^(frontend|sourcegraph-frontend).*"}[5m])) + sum by (op)(increase(src_codeintel_dbstore_errors_total{job=~"^(frontend|sourcegraph-frontend).*"}[5m]))) * 100`

</details>

<br />

### Frontend: Workerutil: lsif_indexes dbworker/store stats

#### frontend: workerutil_dbworker_store_codeintel_index_total

<p class="subtitle">Store operations every 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/frontend/frontend?viewPanel=100600` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum(increase(src_workerutil_dbworker_store_codeintel_index_total{job=~"^(frontend|sourcegraph-frontend).*"}[5m]))`

</details>

<br />

#### frontend: workerutil_dbworker_store_codeintel_index_99th_percentile_duration

<p class="subtitle">99th percentile successful store operation duration over 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/frontend/frontend?viewPanel=100601` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `histogram_quantile(0.99, sum  by (le)(rate(src_workerutil_dbworker_store_codeintel_index_duration_seconds_bucket{job=~"^(frontend|sourcegraph-frontend).*"}[5m])))`

</details>

<br />

#### frontend: workerutil_dbworker_store_codeintel_index_errors_total

<p class="subtitle">Store operation errors every 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/frontend/frontend?viewPanel=100602` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum(increase(src_workerutil_dbworker_store_codeintel_index_errors_total{job=~"^(frontend|sourcegraph-frontend).*"}[5m]))`

</details>

<br />

#### frontend: workerutil_dbworker_store_codeintel_index_error_rate

<p class="subtitle">Store operation error rate over 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/frontend/frontend?viewPanel=100603` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum(increase(src_workerutil_dbworker_store_codeintel_index_errors_total{job=~"^(frontend|sourcegraph-frontend).*"}[5m])) / (sum(increase(src_workerutil_dbworker_store_codeintel_index_total{job=~"^(frontend|sourcegraph-frontend).*"}[5m])) + sum(increase(src_workerutil_dbworker_store_codeintel_index_errors_total{job=~"^(frontend|sourcegraph-frontend).*"}[5m]))) * 100`

</details>

<br />

### Frontend: Codeintel: lsifstore stats

#### frontend: codeintel_lsifstore_total

<p class="subtitle">Aggregate store operations every 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/frontend/frontend?viewPanel=100700` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum(increase(src_codeintel_lsifstore_total{job=~"^(frontend|sourcegraph-frontend).*"}[5m]))`

</details>

<br />

#### frontend: codeintel_lsifstore_99th_percentile_duration

<p class="subtitle">99th percentile successful aggregate store operation duration over 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/frontend/frontend?viewPanel=100701` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `histogram_quantile(0.99, sum  by (le)(rate(src_codeintel_lsifstore_duration_seconds_bucket{job=~"^(frontend|sourcegraph-frontend).*"}[5m])))`

</details>

<br />

#### frontend: codeintel_lsifstore_errors_total

<p class="subtitle">Aggregate store operation errors every 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/frontend/frontend?viewPanel=100702` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum(increase(src_codeintel_lsifstore_errors_total{job=~"^(frontend|sourcegraph-frontend).*"}[5m]))`

</details>

<br />

#### frontend: codeintel_lsifstore_error_rate

<p class="subtitle">Aggregate store operation error rate over 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/frontend/frontend?viewPanel=100703` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum(increase(src_codeintel_lsifstore_errors_total{job=~"^(frontend|sourcegraph-frontend).*"}[5m])) / (sum(increase(src_codeintel_lsifstore_total{job=~"^(frontend|sourcegraph-frontend).*"}[5m])) + sum(increase(src_codeintel_lsifstore_errors_total{job=~"^(frontend|sourcegraph-frontend).*"}[5m]))) * 100`

</details>

<br />

#### frontend: codeintel_lsifstore_total

<p class="subtitle">Store operations every 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/frontend/frontend?viewPanel=100710` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum by (op)(increase(src_codeintel_lsifstore_total{job=~"^(frontend|sourcegraph-frontend).*"}[5m]))`

</details>

<br />

#### frontend: codeintel_lsifstore_99th_percentile_duration

<p class="subtitle">99th percentile successful store operation duration over 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/frontend/frontend?viewPanel=100711` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `histogram_quantile(0.99, sum  by (le,op)(rate(src_codeintel_lsifstore_duration_seconds_bucket{job=~"^(frontend|sourcegraph-frontend).*"}[5m])))`

</details>

<br />

#### frontend: codeintel_lsifstore_errors_total

<p class="subtitle">Store operation errors every 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/frontend/frontend?viewPanel=100712` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum by (op)(increase(src_codeintel_lsifstore_errors_total{job=~"^(frontend|sourcegraph-frontend).*"}[5m]))`

</details>

<br />

#### frontend: codeintel_lsifstore_error_rate

<p class="subtitle">Store operation error rate over 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/frontend/frontend?viewPanel=100713` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum by (op)(increase(src_codeintel_lsifstore_errors_total{job=~"^(frontend|sourcegraph-frontend).*"}[5m])) / (sum by (op)(increase(src_codeintel_lsifstore_total{job=~"^(frontend|sourcegraph-frontend).*"}[5m])) + sum by (op)(increase(src_codeintel_lsifstore_errors_total{job=~"^(frontend|sourcegraph-frontend).*"}[5m]))) * 100`

</details>

<br />

### Frontend: Codeintel: gitserver client

#### frontend: codeintel_gitserver_total

<p class="subtitle">Aggregate client operations every 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/frontend/frontend?viewPanel=100800` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum(increase(src_codeintel_gitserver_total{job=~"^(frontend|sourcegraph-frontend).*"}[5m]))`

</details>

<br />

#### frontend: codeintel_gitserver_99th_percentile_duration

<p class="subtitle">99th percentile successful aggregate client operation duration over 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/frontend/frontend?viewPanel=100801` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `histogram_quantile(0.99, sum  by (le)(rate(src_codeintel_gitserver_duration_seconds_bucket{job=~"^(frontend|sourcegraph-frontend).*"}[5m])))`

</details>

<br />

#### frontend: codeintel_gitserver_errors_total

<p class="subtitle">Aggregate client operation errors every 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/frontend/frontend?viewPanel=100802` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum(increase(src_codeintel_gitserver_errors_total{job=~"^(frontend|sourcegraph-frontend).*"}[5m]))`

</details>

<br />

#### frontend: codeintel_gitserver_error_rate

<p class="subtitle">Aggregate client operation error rate over 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/frontend/frontend?viewPanel=100803` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum(increase(src_codeintel_gitserver_errors_total{job=~"^(frontend|sourcegraph-frontend).*"}[5m])) / (sum(increase(src_codeintel_gitserver_total{job=~"^(frontend|sourcegraph-frontend).*"}[5m])) + sum(increase(src_codeintel_gitserver_errors_total{job=~"^(frontend|sourcegraph-frontend).*"}[5m]))) * 100`

</details>

<br />

#### frontend: codeintel_gitserver_total

<p class="subtitle">Client operations every 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/frontend/frontend?viewPanel=100810` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum by (op)(increase(src_codeintel_gitserver_total{job=~"^(frontend|sourcegraph-frontend).*"}[5m]))`

</details>

<br />

#### frontend: codeintel_gitserver_99th_percentile_duration

<p class="subtitle">99th percentile successful client operation duration over 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/frontend/frontend?viewPanel=100811` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `histogram_quantile(0.99, sum  by (le,op)(rate(src_codeintel_gitserver_duration_seconds_bucket{job=~"^(frontend|sourcegraph-frontend).*"}[5m])))`

</details>

<br />

#### frontend: codeintel_gitserver_errors_total

<p class="subtitle">Client operation errors every 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/frontend/frontend?viewPanel=100812` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum by (op)(increase(src_codeintel_gitserver_errors_total{job=~"^(frontend|sourcegraph-frontend).*"}[5m]))`

</details>

<br />

#### frontend: codeintel_gitserver_error_rate

<p class="subtitle">Client operation error rate over 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/frontend/frontend?viewPanel=100813` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum by (op)(increase(src_codeintel_gitserver_errors_total{job=~"^(frontend|sourcegraph-frontend).*"}[5m])) / (sum by (op)(increase(src_codeintel_gitserver_total{job=~"^(frontend|sourcegraph-frontend).*"}[5m])) + sum by (op)(increase(src_codeintel_gitserver_errors_total{job=~"^(frontend|sourcegraph-frontend).*"}[5m]))) * 100`

</details>

<br />

### Frontend: Codeintel: uploadstore stats

#### frontend: codeintel_uploadstore_total

<p class="subtitle">Aggregate store operations every 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/frontend/frontend?viewPanel=100900` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum(increase(src_codeintel_uploadstore_total{job=~"^(frontend|sourcegraph-frontend).*"}[5m]))`

</details>

<br />

#### frontend: codeintel_uploadstore_99th_percentile_duration

<p class="subtitle">99th percentile successful aggregate store operation duration over 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/frontend/frontend?viewPanel=100901` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `histogram_quantile(0.99, sum  by (le)(rate(src_codeintel_uploadstore_duration_seconds_bucket{job=~"^(frontend|sourcegraph-frontend).*"}[5m])))`

</details>

<br />

#### frontend: codeintel_uploadstore_errors_total

<p class="subtitle">Aggregate store operation errors every 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/frontend/frontend?viewPanel=100902` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum(increase(src_codeintel_uploadstore_errors_total{job=~"^(frontend|sourcegraph-frontend).*"}[5m]))`

</details>

<br />

#### frontend: codeintel_uploadstore_error_rate

<p class="subtitle">Aggregate store operation error rate over 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/frontend/frontend?viewPanel=100903` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum(increase(src_codeintel_uploadstore_errors_total{job=~"^(frontend|sourcegraph-frontend).*"}[5m])) / (sum(increase(src_codeintel_uploadstore_total{job=~"^(frontend|sourcegraph-frontend).*"}[5m])) + sum(increase(src_codeintel_uploadstore_errors_total{job=~"^(frontend|sourcegraph-frontend).*"}[5m]))) * 100`

</details>

<br />

#### frontend: codeintel_uploadstore_total

<p class="subtitle">Store operations every 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/frontend/frontend?viewPanel=100910` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum by (op)(increase(src_codeintel_uploadstore_total{job=~"^(frontend|sourcegraph-frontend).*"}[5m]))`

</details>

<br />

#### frontend: codeintel_uploadstore_99th_percentile_duration

<p class="subtitle">99th percentile successful store operation duration over 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/frontend/frontend?viewPanel=100911` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `histogram_quantile(0.99, sum  by (le,op)(rate(src_codeintel_uploadstore_duration_seconds_bucket{job=~"^(frontend|sourcegraph-frontend).*"}[5m])))`

</details>

<br />

#### frontend: codeintel_uploadstore_errors_total

<p class="subtitle">Store operation errors every 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/frontend/frontend?viewPanel=100912` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum by (op)(increase(src_codeintel_uploadstore_errors_total{job=~"^(frontend|sourcegraph-frontend).*"}[5m]))`

</details>

<br />

#### frontend: codeintel_uploadstore_error_rate

<p class="subtitle">Store operation error rate over 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/frontend/frontend?viewPanel=100913` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum by (op)(increase(src_codeintel_uploadstore_errors_total{job=~"^(frontend|sourcegraph-frontend).*"}[5m])) / (sum by (op)(increase(src_codeintel_uploadstore_total{job=~"^(frontend|sourcegraph-frontend).*"}[5m])) + sum by (op)(increase(src_codeintel_uploadstore_errors_total{job=~"^(frontend|sourcegraph-frontend).*"}[5m]))) * 100`

</details>

<br />

### Frontend: Batches: dbstore stats

#### frontend: batches_dbstore_total

<p class="subtitle">Aggregate store operations every 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/frontend/frontend?viewPanel=101000` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Batches team](https://about.sourcegraph.com/handbook/engineering/batches).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum(increase(src_batches_dbstore_total{job=~"^(frontend|sourcegraph-frontend).*"}[5m]))`

</details>

<br />

#### frontend: batches_dbstore_99th_percentile_duration

<p class="subtitle">99th percentile successful aggregate store operation duration over 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/frontend/frontend?viewPanel=101001` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Batches team](https://about.sourcegraph.com/handbook/engineering/batches).*</sub>

<details>
<summary>Technical details</summary>

Query: `histogram_quantile(0.99, sum  by (le)(rate(src_batches_dbstore_duration_seconds_bucket{job=~"^(frontend|sourcegraph-frontend).*"}[5m])))`

</details>

<br />

#### frontend: batches_dbstore_errors_total

<p class="subtitle">Aggregate store operation errors every 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/frontend/frontend?viewPanel=101002` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Batches team](https://about.sourcegraph.com/handbook/engineering/batches).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum(increase(src_batches_dbstore_errors_total{job=~"^(frontend|sourcegraph-frontend).*"}[5m]))`

</details>

<br />

#### frontend: batches_dbstore_error_rate

<p class="subtitle">Aggregate store operation error rate over 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/frontend/frontend?viewPanel=101003` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Batches team](https://about.sourcegraph.com/handbook/engineering/batches).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum(increase(src_batches_dbstore_errors_total{job=~"^(frontend|sourcegraph-frontend).*"}[5m])) / (sum(increase(src_batches_dbstore_total{job=~"^(frontend|sourcegraph-frontend).*"}[5m])) + sum(increase(src_batches_dbstore_errors_total{job=~"^(frontend|sourcegraph-frontend).*"}[5m]))) * 100`

</details>

<br />

#### frontend: batches_dbstore_total

<p class="subtitle">Store operations every 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/frontend/frontend?viewPanel=101010` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Batches team](https://about.sourcegraph.com/handbook/engineering/batches).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum by (op)(increase(src_batches_dbstore_total{job=~"^(frontend|sourcegraph-frontend).*"}[5m]))`

</details>

<br />

#### frontend: batches_dbstore_99th_percentile_duration

<p class="subtitle">99th percentile successful store operation duration over 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/frontend/frontend?viewPanel=101011` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Batches team](https://about.sourcegraph.com/handbook/engineering/batches).*</sub>

<details>
<summary>Technical details</summary>

Query: `histogram_quantile(0.99, sum  by (le,op)(rate(src_batches_dbstore_duration_seconds_bucket{job=~"^(frontend|sourcegraph-frontend).*"}[5m])))`

</details>

<br />

#### frontend: batches_dbstore_errors_total

<p class="subtitle">Store operation errors every 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/frontend/frontend?viewPanel=101012` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Batches team](https://about.sourcegraph.com/handbook/engineering/batches).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum by (op)(increase(src_batches_dbstore_errors_total{job=~"^(frontend|sourcegraph-frontend).*"}[5m]))`

</details>

<br />

#### frontend: batches_dbstore_error_rate

<p class="subtitle">Store operation error rate over 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/frontend/frontend?viewPanel=101013` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Batches team](https://about.sourcegraph.com/handbook/engineering/batches).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum by (op)(increase(src_batches_dbstore_errors_total{job=~"^(frontend|sourcegraph-frontend).*"}[5m])) / (sum by (op)(increase(src_batches_dbstore_total{job=~"^(frontend|sourcegraph-frontend).*"}[5m])) + sum by (op)(increase(src_batches_dbstore_errors_total{job=~"^(frontend|sourcegraph-frontend).*"}[5m]))) * 100`

</details>

<br />

### Frontend: Batches: service stats

#### frontend: batches_service_total

<p class="subtitle">Aggregate service operations every 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/frontend/frontend?viewPanel=101100` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Batches team](https://about.sourcegraph.com/handbook/engineering/batches).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum(increase(src_batches_service_total{job=~"^(frontend|sourcegraph-frontend).*"}[5m]))`

</details>

<br />

#### frontend: batches_service_99th_percentile_duration

<p class="subtitle">99th percentile successful aggregate service operation duration over 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/frontend/frontend?viewPanel=101101` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Batches team](https://about.sourcegraph.com/handbook/engineering/batches).*</sub>

<details>
<summary>Technical details</summary>

Query: `histogram_quantile(0.99, sum  by (le)(rate(src_batches_service_duration_seconds_bucket{job=~"^(frontend|sourcegraph-frontend).*"}[5m])))`

</details>

<br />

#### frontend: batches_service_errors_total

<p class="subtitle">Aggregate service operation errors every 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/frontend/frontend?viewPanel=101102` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Batches team](https://about.sourcegraph.com/handbook/engineering/batches).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum(increase(src_batches_service_errors_total{job=~"^(frontend|sourcegraph-frontend).*"}[5m]))`

</details>

<br />

#### frontend: batches_service_error_rate

<p class="subtitle">Aggregate service operation error rate over 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/frontend/frontend?viewPanel=101103` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Batches team](https://about.sourcegraph.com/handbook/engineering/batches).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum(increase(src_batches_service_errors_total{job=~"^(frontend|sourcegraph-frontend).*"}[5m])) / (sum(increase(src_batches_service_total{job=~"^(frontend|sourcegraph-frontend).*"}[5m])) + sum(increase(src_batches_service_errors_total{job=~"^(frontend|sourcegraph-frontend).*"}[5m]))) * 100`

</details>

<br />

#### frontend: batches_service_total

<p class="subtitle">Service operations every 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/frontend/frontend?viewPanel=101110` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Batches team](https://about.sourcegraph.com/handbook/engineering/batches).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum by (op)(increase(src_batches_service_total{job=~"^(frontend|sourcegraph-frontend).*"}[5m]))`

</details>

<br />

#### frontend: batches_service_99th_percentile_duration

<p class="subtitle">99th percentile successful service operation duration over 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/frontend/frontend?viewPanel=101111` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Batches team](https://about.sourcegraph.com/handbook/engineering/batches).*</sub>

<details>
<summary>Technical details</summary>

Query: `histogram_quantile(0.99, sum  by (le,op)(rate(src_batches_service_duration_seconds_bucket{job=~"^(frontend|sourcegraph-frontend).*"}[5m])))`

</details>

<br />

#### frontend: batches_service_errors_total

<p class="subtitle">Service operation errors every 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/frontend/frontend?viewPanel=101112` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Batches team](https://about.sourcegraph.com/handbook/engineering/batches).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum by (op)(increase(src_batches_service_errors_total{job=~"^(frontend|sourcegraph-frontend).*"}[5m]))`

</details>

<br />

#### frontend: batches_service_error_rate

<p class="subtitle">Service operation error rate over 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/frontend/frontend?viewPanel=101113` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Batches team](https://about.sourcegraph.com/handbook/engineering/batches).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum by (op)(increase(src_batches_service_errors_total{job=~"^(frontend|sourcegraph-frontend).*"}[5m])) / (sum by (op)(increase(src_batches_service_total{job=~"^(frontend|sourcegraph-frontend).*"}[5m])) + sum by (op)(increase(src_batches_service_errors_total{job=~"^(frontend|sourcegraph-frontend).*"}[5m]))) * 100`

</details>

<br />

### Frontend: Out-of-band migrations: up migration invocation (one batch processed)

#### frontend: oobmigration_total

<p class="subtitle">Migration handler operations every 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/frontend/frontend?viewPanel=101200` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum(increase(src_oobmigration_total{op="up",job=~"^(frontend|sourcegraph-frontend).*"}[5m]))`

</details>

<br />

#### frontend: oobmigration_99th_percentile_duration

<p class="subtitle">99th percentile successful migration handler operation duration over 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/frontend/frontend?viewPanel=101201` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `histogram_quantile(0.99, sum  by (le)(rate(src_oobmigration_duration_seconds_bucket{op="up",job=~"^(frontend|sourcegraph-frontend).*"}[5m])))`

</details>

<br />

#### frontend: oobmigration_errors_total

<p class="subtitle">Migration handler operation errors every 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/frontend/frontend?viewPanel=101202` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum(increase(src_oobmigration_errors_total{op="up",job=~"^(frontend|sourcegraph-frontend).*"}[5m]))`

</details>

<br />

#### frontend: oobmigration_error_rate

<p class="subtitle">Migration handler operation error rate over 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/frontend/frontend?viewPanel=101203` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum(increase(src_oobmigration_errors_total{op="up",job=~"^(frontend|sourcegraph-frontend).*"}[5m])) / (sum(increase(src_oobmigration_total{op="up",job=~"^(frontend|sourcegraph-frontend).*"}[5m])) + sum(increase(src_oobmigration_errors_total{op="up",job=~"^(frontend|sourcegraph-frontend).*"}[5m]))) * 100`

</details>

<br />

### Frontend: Out-of-band migrations: down migration invocation (one batch processed)

#### frontend: oobmigration_total

<p class="subtitle">Migration handler operations every 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/frontend/frontend?viewPanel=101300` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum(increase(src_oobmigration_total{op="down",job=~"^(frontend|sourcegraph-frontend).*"}[5m]))`

</details>

<br />

#### frontend: oobmigration_99th_percentile_duration

<p class="subtitle">99th percentile successful migration handler operation duration over 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/frontend/frontend?viewPanel=101301` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `histogram_quantile(0.99, sum  by (le)(rate(src_oobmigration_duration_seconds_bucket{op="down",job=~"^(frontend|sourcegraph-frontend).*"}[5m])))`

</details>

<br />

#### frontend: oobmigration_errors_total

<p class="subtitle">Migration handler operation errors every 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/frontend/frontend?viewPanel=101302` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum(increase(src_oobmigration_errors_total{op="down",job=~"^(frontend|sourcegraph-frontend).*"}[5m]))`

</details>

<br />

#### frontend: oobmigration_error_rate

<p class="subtitle">Migration handler operation error rate over 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/frontend/frontend?viewPanel=101303` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum(increase(src_oobmigration_errors_total{op="down",job=~"^(frontend|sourcegraph-frontend).*"}[5m])) / (sum(increase(src_oobmigration_total{op="down",job=~"^(frontend|sourcegraph-frontend).*"}[5m])) + sum(increase(src_oobmigration_errors_total{op="down",job=~"^(frontend|sourcegraph-frontend).*"}[5m]))) * 100`

</details>

<br />

### Frontend: Internal service requests

#### frontend: internal_indexed_search_error_responses

<p class="subtitle">Internal indexed search error responses every 5m</p>

Refer to the [alert solutions reference](./alert_solutions.md#frontend-internal-indexed-search-error-responses) for 1 alert related to this panel.

To see this panel, visit `/-/debug/grafana/d/frontend/frontend?viewPanel=101400` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Search team](https://about.sourcegraph.com/handbook/engineering/search).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum by(code) (increase(src_zoekt_request_duration_seconds_count{code!~"2.."}[5m])) / ignoring(code) group_left sum(increase(src_zoekt_request_duration_seconds_count[5m])) * 100`

</details>

<br />

#### frontend: internal_unindexed_search_error_responses

<p class="subtitle">Internal unindexed search error responses every 5m</p>

Refer to the [alert solutions reference](./alert_solutions.md#frontend-internal-unindexed-search-error-responses) for 1 alert related to this panel.

To see this panel, visit `/-/debug/grafana/d/frontend/frontend?viewPanel=101401` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Search team](https://about.sourcegraph.com/handbook/engineering/search).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum by(code) (increase(searcher_service_request_total{code!~"2.."}[5m])) / ignoring(code) group_left sum(increase(searcher_service_request_total[5m])) * 100`

</details>

<br />

#### frontend: internal_api_error_responses

<p class="subtitle">Internal API error responses every 5m by route</p>

Refer to the [alert solutions reference](./alert_solutions.md#frontend-internal-api-error-responses) for 1 alert related to this panel.

To see this panel, visit `/-/debug/grafana/d/frontend/frontend?viewPanel=101402` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum by(category) (increase(src_frontend_internal_request_duration_seconds_count{code!~"2.."}[5m])) / ignoring(code) group_left sum(increase(src_frontend_internal_request_duration_seconds_count[5m])) * 100`

</details>

<br />

#### frontend: 99th_percentile_gitserver_duration

<p class="subtitle">99th percentile successful gitserver query duration over 5m</p>

Refer to the [alert solutions reference](./alert_solutions.md#frontend-99th-percentile-gitserver-duration) for 1 alert related to this panel.

To see this panel, visit `/-/debug/grafana/d/frontend/frontend?viewPanel=101410` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<details>
<summary>Technical details</summary>

Query: `histogram_quantile(0.99, sum by (le,category)(rate(src_gitserver_request_duration_seconds_bucket{job=~"(sourcegraph-)?frontend"}[5m])))`

</details>

<br />

#### frontend: gitserver_error_responses

<p class="subtitle">Gitserver error responses every 5m</p>

Refer to the [alert solutions reference](./alert_solutions.md#frontend-gitserver-error-responses) for 1 alert related to this panel.

To see this panel, visit `/-/debug/grafana/d/frontend/frontend?viewPanel=101411` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum by (category)(increase(src_gitserver_request_duration_seconds_count{job=~"(sourcegraph-)?frontend",code!~"2.."}[5m])) / ignoring(code) group_left sum by (category)(increase(src_gitserver_request_duration_seconds_count{job=~"(sourcegraph-)?frontend"}[5m])) * 100`

</details>

<br />

#### frontend: observability_test_alert_warning

<p class="subtitle">Warning test alert metric</p>

Refer to the [alert solutions reference](./alert_solutions.md#frontend-observability-test-alert-warning) for 1 alert related to this panel.

To see this panel, visit `/-/debug/grafana/d/frontend/frontend?viewPanel=101420` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Distribution team](https://about.sourcegraph.com/handbook/engineering/distribution).*</sub>

<details>
<summary>Technical details</summary>

Query: `max by(owner) (observability_test_metric_warning)`

</details>

<br />

#### frontend: observability_test_alert_critical

<p class="subtitle">Critical test alert metric</p>

Refer to the [alert solutions reference](./alert_solutions.md#frontend-observability-test-alert-critical) for 1 alert related to this panel.

To see this panel, visit `/-/debug/grafana/d/frontend/frontend?viewPanel=101421` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Distribution team](https://about.sourcegraph.com/handbook/engineering/distribution).*</sub>

<details>
<summary>Technical details</summary>

Query: `max by(owner) (observability_test_metric_critical)`

</details>

<br />

### Frontend: Database connections

#### frontend: max_open_conns

<p class="subtitle">Maximum open</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/frontend/frontend?viewPanel=101500` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum by (app_name, db_name) (src_pgsql_conns_max_open{app_name="frontend"})`

</details>

<br />

#### frontend: open_conns

<p class="subtitle">Established</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/frontend/frontend?viewPanel=101501` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum by (app_name, db_name) (src_pgsql_conns_open{app_name="frontend"})`

</details>

<br />

#### frontend: in_use

<p class="subtitle">Used</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/frontend/frontend?viewPanel=101510` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum by (app_name, db_name) (src_pgsql_conns_in_use{app_name="frontend"})`

</details>

<br />

#### frontend: idle

<p class="subtitle">Idle</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/frontend/frontend?viewPanel=101511` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum by (app_name, db_name) (src_pgsql_conns_idle{app_name="frontend"})`

</details>

<br />

#### frontend: mean_blocked_seconds_per_conn_request

<p class="subtitle">Mean blocked seconds per conn request</p>

Refer to the [alert solutions reference](./alert_solutions.md#frontend-mean-blocked-seconds-per-conn-request) for 2 alerts related to this panel.

To see this panel, visit `/-/debug/grafana/d/frontend/frontend?viewPanel=101520` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum by (app_name, db_name) (increase(src_pgsql_conns_blocked_seconds{app_name="frontend"}[5m])) / sum by (app_name, db_name) (increase(src_pgsql_conns_waited_for{app_name="frontend"}[5m]))`

</details>

<br />

#### frontend: closed_max_idle

<p class="subtitle">Closed by SetMaxIdleConns</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/frontend/frontend?viewPanel=101530` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum by (app_name, db_name) (increase(src_pgsql_conns_closed_max_idle{app_name="frontend"}[5m]))`

</details>

<br />

#### frontend: closed_max_lifetime

<p class="subtitle">Closed by SetConnMaxLifetime</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/frontend/frontend?viewPanel=101531` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum by (app_name, db_name) (increase(src_pgsql_conns_closed_max_lifetime{app_name="frontend"}[5m]))`

</details>

<br />

#### frontend: closed_max_idle_time

<p class="subtitle">Closed by SetConnMaxIdleTime</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/frontend/frontend?viewPanel=101532` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum by (app_name, db_name) (increase(src_pgsql_conns_closed_max_idle_time{app_name="frontend"}[5m]))`

</details>

<br />

### Frontend: Container monitoring (not available on server)

#### frontend: container_missing

<p class="subtitle">Container missing</p>

This value is the number of times a container has not been seen for more than one minute. If you observe this
value change independent of deployment events (such as an upgrade), it could indicate pods are being OOM killed or terminated for some other reasons.

- **Kubernetes:**
	- Determine if the pod was OOM killed using `kubectl describe pod (frontend|sourcegraph-frontend)` (look for `OOMKilled: true`) and, if so, consider increasing the memory limit in the relevant `Deployment.yaml`.
	- Check the logs before the container restarted to see if there are `panic:` messages or similar using `kubectl logs -p (frontend|sourcegraph-frontend)`.
- **Docker Compose:**
	- Determine if the pod was OOM killed using `docker inspect -f '{{json .State}}' (frontend|sourcegraph-frontend)` (look for `"OOMKilled":true`) and, if so, consider increasing the memory limit of the (frontend|sourcegraph-frontend) container in `docker-compose.yml`.
	- Check the logs before the container restarted to see if there are `panic:` messages or similar using `docker logs (frontend|sourcegraph-frontend)` (note this will include logs from the previous and currently running container).

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/frontend/frontend?viewPanel=101600` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<details>
<summary>Technical details</summary>

Query: `count by(name) ((time() - container_last_seen{name=~"^(frontend|sourcegraph-frontend).*"}) > 60)`

</details>

<br />

#### frontend: container_cpu_usage

<p class="subtitle">Container cpu usage total (1m average) across all cores by instance</p>

Refer to the [alert solutions reference](./alert_solutions.md#frontend-container-cpu-usage) for 1 alert related to this panel.

To see this panel, visit `/-/debug/grafana/d/frontend/frontend?viewPanel=101601` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<details>
<summary>Technical details</summary>

Query: `cadvisor_container_cpu_usage_percentage_total{name=~"^(frontend|sourcegraph-frontend).*"}`

</details>

<br />

#### frontend: container_memory_usage

<p class="subtitle">Container memory usage by instance</p>

Refer to the [alert solutions reference](./alert_solutions.md#frontend-container-memory-usage) for 1 alert related to this panel.

To see this panel, visit `/-/debug/grafana/d/frontend/frontend?viewPanel=101602` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<details>
<summary>Technical details</summary>

Query: `cadvisor_container_memory_usage_percentage_total{name=~"^(frontend|sourcegraph-frontend).*"}`

</details>

<br />

#### frontend: fs_io_operations

<p class="subtitle">Filesystem reads and writes rate by instance over 1h</p>

This value indicates the number of filesystem read and write operations by containers of this service.
When extremely high, this can indicate a resource usage problem, or can cause problems with the service itself, especially if high values or spikes correlate with {{CONTAINER_NAME}} issues.

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/frontend/frontend?viewPanel=101603` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum by(name) (rate(container_fs_reads_total{name=~"^(frontend|sourcegraph-frontend).*"}[1h]) + rate(container_fs_writes_total{name=~"^(frontend|sourcegraph-frontend).*"}[1h]))`

</details>

<br />

### Frontend: Provisioning indicators (not available on server)

#### frontend: provisioning_container_cpu_usage_long_term

<p class="subtitle">Container cpu usage total (90th percentile over 1d) across all cores by instance</p>

Refer to the [alert solutions reference](./alert_solutions.md#frontend-provisioning-container-cpu-usage-long-term) for 1 alert related to this panel.

To see this panel, visit `/-/debug/grafana/d/frontend/frontend?viewPanel=101700` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<details>
<summary>Technical details</summary>

Query: `quantile_over_time(0.9, cadvisor_container_cpu_usage_percentage_total{name=~"^(frontend|sourcegraph-frontend).*"}[1d])`

</details>

<br />

#### frontend: provisioning_container_memory_usage_long_term

<p class="subtitle">Container memory usage (1d maximum) by instance</p>

Refer to the [alert solutions reference](./alert_solutions.md#frontend-provisioning-container-memory-usage-long-term) for 1 alert related to this panel.

To see this panel, visit `/-/debug/grafana/d/frontend/frontend?viewPanel=101701` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<details>
<summary>Technical details</summary>

Query: `max_over_time(cadvisor_container_memory_usage_percentage_total{name=~"^(frontend|sourcegraph-frontend).*"}[1d])`

</details>

<br />

#### frontend: provisioning_container_cpu_usage_short_term

<p class="subtitle">Container cpu usage total (5m maximum) across all cores by instance</p>

Refer to the [alert solutions reference](./alert_solutions.md#frontend-provisioning-container-cpu-usage-short-term) for 1 alert related to this panel.

To see this panel, visit `/-/debug/grafana/d/frontend/frontend?viewPanel=101710` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<details>
<summary>Technical details</summary>

Query: `max_over_time(cadvisor_container_cpu_usage_percentage_total{name=~"^(frontend|sourcegraph-frontend).*"}[5m])`

</details>

<br />

#### frontend: provisioning_container_memory_usage_short_term

<p class="subtitle">Container memory usage (5m maximum) by instance</p>

Refer to the [alert solutions reference](./alert_solutions.md#frontend-provisioning-container-memory-usage-short-term) for 1 alert related to this panel.

To see this panel, visit `/-/debug/grafana/d/frontend/frontend?viewPanel=101711` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<details>
<summary>Technical details</summary>

Query: `max_over_time(cadvisor_container_memory_usage_percentage_total{name=~"^(frontend|sourcegraph-frontend).*"}[5m])`

</details>

<br />

### Frontend: Golang runtime monitoring

#### frontend: go_goroutines

<p class="subtitle">Maximum active goroutines</p>

A high value here indicates a possible goroutine leak.

Refer to the [alert solutions reference](./alert_solutions.md#frontend-go-goroutines) for 1 alert related to this panel.

To see this panel, visit `/-/debug/grafana/d/frontend/frontend?viewPanel=101800` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<details>
<summary>Technical details</summary>

Query: `max by(instance) (go_goroutines{job=~".*(frontend|sourcegraph-frontend)"})`

</details>

<br />

#### frontend: go_gc_duration_seconds

<p class="subtitle">Maximum go garbage collection duration</p>

Refer to the [alert solutions reference](./alert_solutions.md#frontend-go-gc-duration-seconds) for 1 alert related to this panel.

To see this panel, visit `/-/debug/grafana/d/frontend/frontend?viewPanel=101801` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<details>
<summary>Technical details</summary>

Query: `max by(instance) (go_gc_duration_seconds{job=~".*(frontend|sourcegraph-frontend)"})`

</details>

<br />

### Frontend: Kubernetes monitoring (only available on Kubernetes)

#### frontend: pods_available_percentage

<p class="subtitle">Percentage pods available</p>

Refer to the [alert solutions reference](./alert_solutions.md#frontend-pods-available-percentage) for 1 alert related to this panel.

To see this panel, visit `/-/debug/grafana/d/frontend/frontend?viewPanel=101900` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum by(app) (up{app=~".*(frontend|sourcegraph-frontend)"}) / count by (app) (up{app=~".*(frontend|sourcegraph-frontend)"}) * 100`

</details>

<br />

### Frontend: Sentinel queries (only on sourcegraph.com)

#### frontend: mean_successful_sentinel_duration_5m

<p class="subtitle">Mean successful sentinel search duration over 5m</p>

Refer to the [alert solutions reference](./alert_solutions.md#frontend-mean-successful-sentinel-duration-5m) for 2 alerts related to this panel.

To see this panel, visit `/-/debug/grafana/d/frontend/frontend?viewPanel=102000` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Search team](https://about.sourcegraph.com/handbook/engineering/search).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum(rate(src_search_response_latency_seconds_sum{source=~"searchblitz.*", status="success"}[5m])) / sum(rate(src_search_response_latency_seconds_count{source=~"searchblitz.*", status="success"}[5m]))`

</details>

<br />

#### frontend: mean_sentinel_stream_latency_5m

<p class="subtitle">Mean sentinel stream latency over 5m</p>

Refer to the [alert solutions reference](./alert_solutions.md#frontend-mean-sentinel-stream-latency-5m) for 2 alerts related to this panel.

To see this panel, visit `/-/debug/grafana/d/frontend/frontend?viewPanel=102001` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Search team](https://about.sourcegraph.com/handbook/engineering/search).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum(rate(src_search_streaming_latency_seconds_sum{source=~"searchblitz.*"}[5m])) / sum(rate(src_search_streaming_latency_seconds_count{source=~"searchblitz.*"}[5m]))`

</details>

<br />

#### frontend: 90th_percentile_successful_sentinel_duration_5m

<p class="subtitle">90th percentile successful sentinel search duration over 5m</p>

Refer to the [alert solutions reference](./alert_solutions.md#frontend-90th-percentile-successful-sentinel-duration-5m) for 2 alerts related to this panel.

To see this panel, visit `/-/debug/grafana/d/frontend/frontend?viewPanel=102010` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Search team](https://about.sourcegraph.com/handbook/engineering/search).*</sub>

<details>
<summary>Technical details</summary>

Query: `histogram_quantile(0.90, sum by (le)(label_replace(rate(src_search_response_latency_seconds_bucket{source=~"searchblitz.*", status="success"}[5m]), "source", "$1", "source", "searchblitz_(.*)")))`

</details>

<br />

#### frontend: 90th_percentile_sentinel_stream_latency_5m

<p class="subtitle">90th percentile sentinel stream latency over 5m</p>

Refer to the [alert solutions reference](./alert_solutions.md#frontend-90th-percentile-sentinel-stream-latency-5m) for 2 alerts related to this panel.

To see this panel, visit `/-/debug/grafana/d/frontend/frontend?viewPanel=102011` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Search team](https://about.sourcegraph.com/handbook/engineering/search).*</sub>

<details>
<summary>Technical details</summary>

Query: `histogram_quantile(0.90, sum by (le)(label_replace(rate(src_search_streaming_latency_seconds_bucket{source=~"searchblitz.*"}[5m]), "source", "$1", "source", "searchblitz_(.*)")))`

</details>

<br />

#### frontend: mean_successful_sentinel_duration_by_query_5m

<p class="subtitle">Mean successful sentinel search duration by query over 5m</p>

- The mean search duration for sentinel queries, broken down by query. Useful for debugging whether a slowdown is limited to a specific type of query.

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/frontend/frontend?viewPanel=102020` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Search team](https://about.sourcegraph.com/handbook/engineering/search).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum(rate(src_search_response_latency_seconds_sum{source=~"searchblitz.*", status="success"}[5m])) by (source) / sum(rate(src_search_response_latency_seconds_count{source=~"searchblitz.*", status="success"}[5m])) by (source)`

</details>

<br />

#### frontend: mean_sentinel_stream_latency_by_query_5m

<p class="subtitle">Mean sentinel stream latency by query over 5m</p>

- The mean streaming search latency for sentinel queries, broken down by query. Useful for debugging whether a slowdown is limited to a specific type of query.

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/frontend/frontend?viewPanel=102021` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Search team](https://about.sourcegraph.com/handbook/engineering/search).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum(rate(src_search_streaming_latency_seconds_sum{source=~"searchblitz.*"}[5m])) by (source) / sum(rate(src_search_streaming_latency_seconds_count{source=~"searchblitz.*"}[5m])) by (source)`

</details>

<br />

#### frontend: unsuccessful_status_rate_5m

<p class="subtitle">Unsuccessful status rate per 5m</p>

- The rate of unsuccessful sentinel query, broken down by failure type

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/frontend/frontend?viewPanel=102030` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Search team](https://about.sourcegraph.com/handbook/engineering/search).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum(rate(src_graphql_search_response{source=~"searchblitz.*", status!="success"}[5m])) by (status)`

</details>

<br />

## Git Server

<p class="subtitle">Stores, manages, and operates Git repositories.</p>

To see this dashboard, visit `/-/debug/grafana/d/gitserver/gitserver` on your Sourcegraph instance.

#### gitserver: memory_working_set

<p class="subtitle">Memory working set</p>



This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/gitserver/gitserver?viewPanel=100000` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum by (container_label_io_kubernetes_pod_name) (container_memory_working_set_bytes{container_label_io_kubernetes_container_name="gitserver", container_label_io_kubernetes_pod_name=~"${shard:regex}"})`

</details>

<br />

#### gitserver: go_routines

<p class="subtitle">Go routines</p>



This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/gitserver/gitserver?viewPanel=100001` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<details>
<summary>Technical details</summary>

Query: `go_goroutines{app="gitserver", instance=~"${shard:regex}"}`

</details>

<br />

#### gitserver: cpu_throttling_time

<p class="subtitle">Container CPU throttling time %</p>



This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/gitserver/gitserver?viewPanel=100010` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum by (container_label_io_kubernetes_pod_name) ((rate(container_cpu_cfs_throttled_periods_total{container_label_io_kubernetes_container_name="gitserver", container_label_io_kubernetes_pod_name=~"${shard:regex}"}[5m]) / rate(container_cpu_cfs_periods_total{container_label_io_kubernetes_container_name="gitserver", container_label_io_kubernetes_pod_name=~"${shard:regex}"}[5m])) * 100)`

</details>

<br />

#### gitserver: cpu_usage_seconds

<p class="subtitle">Cpu usage seconds</p>



This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/gitserver/gitserver?viewPanel=100011` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum by (container_label_io_kubernetes_pod_name) (rate(container_cpu_usage_seconds_total{container_label_io_kubernetes_container_name="gitserver", container_label_io_kubernetes_pod_name=~"${shard:regex}"}[5m]))`

</details>

<br />

#### gitserver: disk_space_remaining

<p class="subtitle">Disk space remaining by instance</p>

Refer to the [alert solutions reference](./alert_solutions.md#gitserver-disk-space-remaining) for 2 alerts related to this panel.

To see this panel, visit `/-/debug/grafana/d/gitserver/gitserver?viewPanel=100020` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<details>
<summary>Technical details</summary>

Query: `(src_gitserver_disk_space_available / src_gitserver_disk_space_total) * 100`

</details>

<br />

#### gitserver: io_reads_total

<p class="subtitle">I/o reads total</p>



This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/gitserver/gitserver?viewPanel=100030` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum by (container_label_io_kubernetes_container_name) (rate(container_fs_reads_total{container_label_io_kubernetes_container_name="gitserver"}[5m]))`

</details>

<br />

#### gitserver: io_writes_total

<p class="subtitle">I/o writes total</p>



This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/gitserver/gitserver?viewPanel=100031` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum by (container_label_io_kubernetes_container_name) (rate(container_fs_writes_total{container_label_io_kubernetes_container_name="gitserver"}[5m]))`

</details>

<br />

#### gitserver: io_reads

<p class="subtitle">I/o reads</p>



This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/gitserver/gitserver?viewPanel=100040` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum by (container_label_io_kubernetes_pod_name) (rate(container_fs_reads_total{container_label_io_kubernetes_container_name="gitserver", container_label_io_kubernetes_pod_name=~"${shard:regex}"}[5m]))`

</details>

<br />

#### gitserver: io_writes

<p class="subtitle">I/o writes</p>



This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/gitserver/gitserver?viewPanel=100041` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum by (container_label_io_kubernetes_pod_name) (rate(container_fs_writes_total{container_label_io_kubernetes_container_name="gitserver", container_label_io_kubernetes_pod_name=~"${shard:regex}"}[5m]))`

</details>

<br />

#### gitserver: io_read_througput

<p class="subtitle">I/o read throughput</p>



This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/gitserver/gitserver?viewPanel=100050` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum by (container_label_io_kubernetes_pod_name) (rate(container_fs_reads_bytes_total{container_label_io_kubernetes_container_name="gitserver", container_label_io_kubernetes_pod_name=~"${shard:regex}"}[5m]))`

</details>

<br />

#### gitserver: io_write_throughput

<p class="subtitle">I/o write throughput</p>



This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/gitserver/gitserver?viewPanel=100051` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum by (container_label_io_kubernetes_pod_name) (rate(container_fs_writes_bytes_total{container_label_io_kubernetes_container_name="gitserver", container_label_io_kubernetes_pod_name=~"${shard:regex}"}[5m]))`

</details>

<br />

#### gitserver: running_git_commands

<p class="subtitle">Git commands running on each gitserver instance</p>

A high value signals load.

Refer to the [alert solutions reference](./alert_solutions.md#gitserver-running-git-commands) for 2 alerts related to this panel.

To see this panel, visit `/-/debug/grafana/d/gitserver/gitserver?viewPanel=100060` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum by (instance, cmd) (src_gitserver_exec_running{instance=~"${shard:regex}"})`

</details>

<br />

#### gitserver: git_commands_received

<p class="subtitle">Rate of git commands received across all instances</p>

per second rate per command across all instances

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/gitserver/gitserver?viewPanel=100061` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum by (cmd) (rate(src_gitserver_exec_duration_seconds_count[5m]))`

</details>

<br />

#### gitserver: repository_clone_queue_size

<p class="subtitle">Repository clone queue size</p>

Refer to the [alert solutions reference](./alert_solutions.md#gitserver-repository-clone-queue-size) for 1 alert related to this panel.

To see this panel, visit `/-/debug/grafana/d/gitserver/gitserver?viewPanel=100070` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum(src_gitserver_clone_queue)`

</details>

<br />

#### gitserver: repository_existence_check_queue_size

<p class="subtitle">Repository existence check queue size</p>

Refer to the [alert solutions reference](./alert_solutions.md#gitserver-repository-existence-check-queue-size) for 1 alert related to this panel.

To see this panel, visit `/-/debug/grafana/d/gitserver/gitserver?viewPanel=100071` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum(src_gitserver_lsremote_queue)`

</details>

<br />

#### gitserver: echo_command_duration_test

<p class="subtitle">Echo test command duration</p>

A high value here likely indicates a problem, especially if consistently high.
You can query for individual commands using `sum by (cmd)(src_gitserver_exec_running)` in Grafana (`/-/debug/grafana`) to see if a specific Git Server command might be spiking in frequency.

If this value is consistently high, consider the following:

- **Single container deployments:** Upgrade to a [Docker Compose deployment](../install/docker-compose/migrate.md) which offers better scalability and resource isolation.
- **Kubernetes and Docker Compose:** Check that you are running a similar number of git server replicas and that their CPU/memory limits are allocated according to what is shown in the [Sourcegraph resource estimator](../install/resource_estimator.md).

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/gitserver/gitserver?viewPanel=100080` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<details>
<summary>Technical details</summary>

Query: `max(src_gitserver_echo_duration_seconds)`

</details>

<br />

#### gitserver: frontend_internal_api_error_responses

<p class="subtitle">Frontend-internal API error responses every 5m by route</p>

Refer to the [alert solutions reference](./alert_solutions.md#gitserver-frontend-internal-api-error-responses) for 1 alert related to this panel.

To see this panel, visit `/-/debug/grafana/d/gitserver/gitserver?viewPanel=100081` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum by (category)(increase(src_frontend_internal_request_duration_seconds_count{job="gitserver",code!~"2.."}[5m])) / ignoring(category) group_left sum(increase(src_frontend_internal_request_duration_seconds_count{job="gitserver"}[5m]))`

</details>

<br />

### Git Server: Gitserver cleanup jobs

#### gitserver: janitor_running

<p class="subtitle">If the janitor process is running</p>

1, if the janitor process is currently running

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/gitserver/gitserver?viewPanel=100100` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<details>
<summary>Technical details</summary>

Query: `max by (instance) (src_gitserver_janitor_running)`

</details>

<br />

#### gitserver: janitor_job_duration

<p class="subtitle">95th percentile job run duration</p>

95th percentile job run duration

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/gitserver/gitserver?viewPanel=100110` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<details>
<summary>Technical details</summary>

Query: `histogram_quantile(0.95, sum(rate(src_gitserver_janitor_job_duration_seconds_bucket[5m])) by (le, job_name))`

</details>

<br />

#### gitserver: repos_removed

<p class="subtitle">Repositories removed due to disk pressure</p>

Repositories removed due to disk pressure

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/gitserver/gitserver?viewPanel=100120` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum by (instance) (rate(src_gitserver_repos_removed_disk_pressure[5m]))`

</details>

<br />

### Git Server: Codeintel: Coursier invocation stats

#### gitserver: codeintel_coursier_total

<p class="subtitle">Aggregate invocations operations every 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/gitserver/gitserver?viewPanel=100200` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum(increase(src_codeintel_coursier_total{op!="RunCommand",job=~"^gitserver.*"}[5m]))`

</details>

<br />

#### gitserver: codeintel_coursier_99th_percentile_duration

<p class="subtitle">99th percentile successful aggregate invocations operation duration over 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/gitserver/gitserver?viewPanel=100201` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `histogram_quantile(0.99, sum  by (le)(rate(src_codeintel_coursier_duration_seconds_bucket{op!="RunCommand",job=~"^gitserver.*"}[5m])))`

</details>

<br />

#### gitserver: codeintel_coursier_errors_total

<p class="subtitle">Aggregate invocations operation errors every 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/gitserver/gitserver?viewPanel=100202` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum(increase(src_codeintel_coursier_errors_total{op!="RunCommand",job=~"^gitserver.*"}[5m]))`

</details>

<br />

#### gitserver: codeintel_coursier_error_rate

<p class="subtitle">Aggregate invocations operation error rate over 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/gitserver/gitserver?viewPanel=100203` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum(increase(src_codeintel_coursier_errors_total{op!="RunCommand",job=~"^gitserver.*"}[5m])) / (sum(increase(src_codeintel_coursier_total{op!="RunCommand",job=~"^gitserver.*"}[5m])) + sum(increase(src_codeintel_coursier_errors_total{op!="RunCommand",job=~"^gitserver.*"}[5m]))) * 100`

</details>

<br />

#### gitserver: codeintel_coursier_total

<p class="subtitle">Invocations operations every 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/gitserver/gitserver?viewPanel=100210` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum by (op)(increase(src_codeintel_coursier_total{op!="RunCommand",job=~"^gitserver.*"}[5m]))`

</details>

<br />

#### gitserver: codeintel_coursier_99th_percentile_duration

<p class="subtitle">99th percentile successful invocations operation duration over 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/gitserver/gitserver?viewPanel=100211` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `histogram_quantile(0.99, sum  by (le,op)(rate(src_codeintel_coursier_duration_seconds_bucket{op!="RunCommand",job=~"^gitserver.*"}[5m])))`

</details>

<br />

#### gitserver: codeintel_coursier_errors_total

<p class="subtitle">Invocations operation errors every 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/gitserver/gitserver?viewPanel=100212` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum by (op)(increase(src_codeintel_coursier_errors_total{op!="RunCommand",job=~"^gitserver.*"}[5m]))`

</details>

<br />

#### gitserver: codeintel_coursier_error_rate

<p class="subtitle">Invocations operation error rate over 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/gitserver/gitserver?viewPanel=100213` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum by (op)(increase(src_codeintel_coursier_errors_total{op!="RunCommand",job=~"^gitserver.*"}[5m])) / (sum by (op)(increase(src_codeintel_coursier_total{op!="RunCommand",job=~"^gitserver.*"}[5m])) + sum by (op)(increase(src_codeintel_coursier_errors_total{op!="RunCommand",job=~"^gitserver.*"}[5m]))) * 100`

</details>

<br />

### Git Server: Database connections

#### gitserver: max_open_conns

<p class="subtitle">Maximum open</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/gitserver/gitserver?viewPanel=100300` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum by (app_name, db_name) (src_pgsql_conns_max_open{app_name="gitserver"})`

</details>

<br />

#### gitserver: open_conns

<p class="subtitle">Established</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/gitserver/gitserver?viewPanel=100301` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum by (app_name, db_name) (src_pgsql_conns_open{app_name="gitserver"})`

</details>

<br />

#### gitserver: in_use

<p class="subtitle">Used</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/gitserver/gitserver?viewPanel=100310` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum by (app_name, db_name) (src_pgsql_conns_in_use{app_name="gitserver"})`

</details>

<br />

#### gitserver: idle

<p class="subtitle">Idle</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/gitserver/gitserver?viewPanel=100311` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum by (app_name, db_name) (src_pgsql_conns_idle{app_name="gitserver"})`

</details>

<br />

#### gitserver: mean_blocked_seconds_per_conn_request

<p class="subtitle">Mean blocked seconds per conn request</p>

Refer to the [alert solutions reference](./alert_solutions.md#gitserver-mean-blocked-seconds-per-conn-request) for 2 alerts related to this panel.

To see this panel, visit `/-/debug/grafana/d/gitserver/gitserver?viewPanel=100320` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum by (app_name, db_name) (increase(src_pgsql_conns_blocked_seconds{app_name="gitserver"}[5m])) / sum by (app_name, db_name) (increase(src_pgsql_conns_waited_for{app_name="gitserver"}[5m]))`

</details>

<br />

#### gitserver: closed_max_idle

<p class="subtitle">Closed by SetMaxIdleConns</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/gitserver/gitserver?viewPanel=100330` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum by (app_name, db_name) (increase(src_pgsql_conns_closed_max_idle{app_name="gitserver"}[5m]))`

</details>

<br />

#### gitserver: closed_max_lifetime

<p class="subtitle">Closed by SetConnMaxLifetime</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/gitserver/gitserver?viewPanel=100331` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum by (app_name, db_name) (increase(src_pgsql_conns_closed_max_lifetime{app_name="gitserver"}[5m]))`

</details>

<br />

#### gitserver: closed_max_idle_time

<p class="subtitle">Closed by SetConnMaxIdleTime</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/gitserver/gitserver?viewPanel=100332` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum by (app_name, db_name) (increase(src_pgsql_conns_closed_max_idle_time{app_name="gitserver"}[5m]))`

</details>

<br />

### Git Server: Container monitoring (not available on server)

#### gitserver: container_missing

<p class="subtitle">Container missing</p>

This value is the number of times a container has not been seen for more than one minute. If you observe this
value change independent of deployment events (such as an upgrade), it could indicate pods are being OOM killed or terminated for some other reasons.

- **Kubernetes:**
	- Determine if the pod was OOM killed using `kubectl describe pod gitserver` (look for `OOMKilled: true`) and, if so, consider increasing the memory limit in the relevant `Deployment.yaml`.
	- Check the logs before the container restarted to see if there are `panic:` messages or similar using `kubectl logs -p gitserver`.
- **Docker Compose:**
	- Determine if the pod was OOM killed using `docker inspect -f '{{json .State}}' gitserver` (look for `"OOMKilled":true`) and, if so, consider increasing the memory limit of the gitserver container in `docker-compose.yml`.
	- Check the logs before the container restarted to see if there are `panic:` messages or similar using `docker logs gitserver` (note this will include logs from the previous and currently running container).

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/gitserver/gitserver?viewPanel=100400` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<details>
<summary>Technical details</summary>

Query: `count by(name) ((time() - container_last_seen{name=~"^gitserver.*"}) > 60)`

</details>

<br />

#### gitserver: container_cpu_usage

<p class="subtitle">Container cpu usage total (1m average) across all cores by instance</p>

Refer to the [alert solutions reference](./alert_solutions.md#gitserver-container-cpu-usage) for 1 alert related to this panel.

To see this panel, visit `/-/debug/grafana/d/gitserver/gitserver?viewPanel=100401` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<details>
<summary>Technical details</summary>

Query: `cadvisor_container_cpu_usage_percentage_total{name=~"^gitserver.*"}`

</details>

<br />

#### gitserver: container_memory_usage

<p class="subtitle">Container memory usage by instance</p>

Refer to the [alert solutions reference](./alert_solutions.md#gitserver-container-memory-usage) for 1 alert related to this panel.

To see this panel, visit `/-/debug/grafana/d/gitserver/gitserver?viewPanel=100402` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<details>
<summary>Technical details</summary>

Query: `cadvisor_container_memory_usage_percentage_total{name=~"^gitserver.*"}`

</details>

<br />

#### gitserver: fs_io_operations

<p class="subtitle">Filesystem reads and writes rate by instance over 1h</p>

This value indicates the number of filesystem read and write operations by containers of this service.
When extremely high, this can indicate a resource usage problem, or can cause problems with the service itself, especially if high values or spikes correlate with {{CONTAINER_NAME}} issues.

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/gitserver/gitserver?viewPanel=100403` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum by(name) (rate(container_fs_reads_total{name=~"^gitserver.*"}[1h]) + rate(container_fs_writes_total{name=~"^gitserver.*"}[1h]))`

</details>

<br />

### Git Server: Provisioning indicators (not available on server)

#### gitserver: provisioning_container_cpu_usage_long_term

<p class="subtitle">Container cpu usage total (90th percentile over 1d) across all cores by instance</p>

Refer to the [alert solutions reference](./alert_solutions.md#gitserver-provisioning-container-cpu-usage-long-term) for 1 alert related to this panel.

To see this panel, visit `/-/debug/grafana/d/gitserver/gitserver?viewPanel=100500` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<details>
<summary>Technical details</summary>

Query: `quantile_over_time(0.9, cadvisor_container_cpu_usage_percentage_total{name=~"^gitserver.*"}[1d])`

</details>

<br />

#### gitserver: provisioning_container_memory_usage_long_term

<p class="subtitle">Container memory usage (1d maximum) by instance</p>

Git Server is expected to use up all the memory it is provided.

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/gitserver/gitserver?viewPanel=100501` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<details>
<summary>Technical details</summary>

Query: `max_over_time(cadvisor_container_memory_usage_percentage_total{name=~"^gitserver.*"}[1d])`

</details>

<br />

#### gitserver: provisioning_container_cpu_usage_short_term

<p class="subtitle">Container cpu usage total (5m maximum) across all cores by instance</p>

Refer to the [alert solutions reference](./alert_solutions.md#gitserver-provisioning-container-cpu-usage-short-term) for 1 alert related to this panel.

To see this panel, visit `/-/debug/grafana/d/gitserver/gitserver?viewPanel=100510` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<details>
<summary>Technical details</summary>

Query: `max_over_time(cadvisor_container_cpu_usage_percentage_total{name=~"^gitserver.*"}[5m])`

</details>

<br />

#### gitserver: provisioning_container_memory_usage_short_term

<p class="subtitle">Container memory usage (5m maximum) by instance</p>

Git Server is expected to use up all the memory it is provided.

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/gitserver/gitserver?viewPanel=100511` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<details>
<summary>Technical details</summary>

Query: `max_over_time(cadvisor_container_memory_usage_percentage_total{name=~"^gitserver.*"}[5m])`

</details>

<br />

### Git Server: Golang runtime monitoring

#### gitserver: go_goroutines

<p class="subtitle">Maximum active goroutines</p>

A high value here indicates a possible goroutine leak.

Refer to the [alert solutions reference](./alert_solutions.md#gitserver-go-goroutines) for 1 alert related to this panel.

To see this panel, visit `/-/debug/grafana/d/gitserver/gitserver?viewPanel=100600` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<details>
<summary>Technical details</summary>

Query: `max by(instance) (go_goroutines{job=~".*gitserver"})`

</details>

<br />

#### gitserver: go_gc_duration_seconds

<p class="subtitle">Maximum go garbage collection duration</p>

Refer to the [alert solutions reference](./alert_solutions.md#gitserver-go-gc-duration-seconds) for 1 alert related to this panel.

To see this panel, visit `/-/debug/grafana/d/gitserver/gitserver?viewPanel=100601` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<details>
<summary>Technical details</summary>

Query: `max by(instance) (go_gc_duration_seconds{job=~".*gitserver"})`

</details>

<br />

### Git Server: Kubernetes monitoring (only available on Kubernetes)

#### gitserver: pods_available_percentage

<p class="subtitle">Percentage pods available</p>

Refer to the [alert solutions reference](./alert_solutions.md#gitserver-pods-available-percentage) for 1 alert related to this panel.

To see this panel, visit `/-/debug/grafana/d/gitserver/gitserver?viewPanel=100700` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum by(app) (up{app=~".*gitserver"}) / count by (app) (up{app=~".*gitserver"}) * 100`

</details>

<br />

## GitHub Proxy

<p class="subtitle">Proxies all requests to github.com, keeping track of and managing rate limits.</p>

To see this dashboard, visit `/-/debug/grafana/d/github-proxy/github-proxy` on your Sourcegraph instance.

### GitHub Proxy: GitHub API monitoring

#### github-proxy: github_proxy_waiting_requests

<p class="subtitle">Number of requests waiting on the global mutex</p>

Refer to the [alert solutions reference](./alert_solutions.md#github-proxy-github-proxy-waiting-requests) for 1 alert related to this panel.

To see this panel, visit `/-/debug/grafana/d/github-proxy/github-proxy?viewPanel=100000` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<details>
<summary>Technical details</summary>

Query: `max(github_proxy_waiting_requests)`

</details>

<br />

### GitHub Proxy: Container monitoring (not available on server)

#### github-proxy: container_missing

<p class="subtitle">Container missing</p>

This value is the number of times a container has not been seen for more than one minute. If you observe this
value change independent of deployment events (such as an upgrade), it could indicate pods are being OOM killed or terminated for some other reasons.

- **Kubernetes:**
	- Determine if the pod was OOM killed using `kubectl describe pod github-proxy` (look for `OOMKilled: true`) and, if so, consider increasing the memory limit in the relevant `Deployment.yaml`.
	- Check the logs before the container restarted to see if there are `panic:` messages or similar using `kubectl logs -p github-proxy`.
- **Docker Compose:**
	- Determine if the pod was OOM killed using `docker inspect -f '{{json .State}}' github-proxy` (look for `"OOMKilled":true`) and, if so, consider increasing the memory limit of the github-proxy container in `docker-compose.yml`.
	- Check the logs before the container restarted to see if there are `panic:` messages or similar using `docker logs github-proxy` (note this will include logs from the previous and currently running container).

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/github-proxy/github-proxy?viewPanel=100100` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<details>
<summary>Technical details</summary>

Query: `count by(name) ((time() - container_last_seen{name=~"^github-proxy.*"}) > 60)`

</details>

<br />

#### github-proxy: container_cpu_usage

<p class="subtitle">Container cpu usage total (1m average) across all cores by instance</p>

Refer to the [alert solutions reference](./alert_solutions.md#github-proxy-container-cpu-usage) for 1 alert related to this panel.

To see this panel, visit `/-/debug/grafana/d/github-proxy/github-proxy?viewPanel=100101` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<details>
<summary>Technical details</summary>

Query: `cadvisor_container_cpu_usage_percentage_total{name=~"^github-proxy.*"}`

</details>

<br />

#### github-proxy: container_memory_usage

<p class="subtitle">Container memory usage by instance</p>

Refer to the [alert solutions reference](./alert_solutions.md#github-proxy-container-memory-usage) for 1 alert related to this panel.

To see this panel, visit `/-/debug/grafana/d/github-proxy/github-proxy?viewPanel=100102` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<details>
<summary>Technical details</summary>

Query: `cadvisor_container_memory_usage_percentage_total{name=~"^github-proxy.*"}`

</details>

<br />

#### github-proxy: fs_io_operations

<p class="subtitle">Filesystem reads and writes rate by instance over 1h</p>

This value indicates the number of filesystem read and write operations by containers of this service.
When extremely high, this can indicate a resource usage problem, or can cause problems with the service itself, especially if high values or spikes correlate with {{CONTAINER_NAME}} issues.

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/github-proxy/github-proxy?viewPanel=100103` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum by(name) (rate(container_fs_reads_total{name=~"^github-proxy.*"}[1h]) + rate(container_fs_writes_total{name=~"^github-proxy.*"}[1h]))`

</details>

<br />

### GitHub Proxy: Provisioning indicators (not available on server)

#### github-proxy: provisioning_container_cpu_usage_long_term

<p class="subtitle">Container cpu usage total (90th percentile over 1d) across all cores by instance</p>

Refer to the [alert solutions reference](./alert_solutions.md#github-proxy-provisioning-container-cpu-usage-long-term) for 1 alert related to this panel.

To see this panel, visit `/-/debug/grafana/d/github-proxy/github-proxy?viewPanel=100200` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<details>
<summary>Technical details</summary>

Query: `quantile_over_time(0.9, cadvisor_container_cpu_usage_percentage_total{name=~"^github-proxy.*"}[1d])`

</details>

<br />

#### github-proxy: provisioning_container_memory_usage_long_term

<p class="subtitle">Container memory usage (1d maximum) by instance</p>

Refer to the [alert solutions reference](./alert_solutions.md#github-proxy-provisioning-container-memory-usage-long-term) for 1 alert related to this panel.

To see this panel, visit `/-/debug/grafana/d/github-proxy/github-proxy?viewPanel=100201` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<details>
<summary>Technical details</summary>

Query: `max_over_time(cadvisor_container_memory_usage_percentage_total{name=~"^github-proxy.*"}[1d])`

</details>

<br />

#### github-proxy: provisioning_container_cpu_usage_short_term

<p class="subtitle">Container cpu usage total (5m maximum) across all cores by instance</p>

Refer to the [alert solutions reference](./alert_solutions.md#github-proxy-provisioning-container-cpu-usage-short-term) for 1 alert related to this panel.

To see this panel, visit `/-/debug/grafana/d/github-proxy/github-proxy?viewPanel=100210` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<details>
<summary>Technical details</summary>

Query: `max_over_time(cadvisor_container_cpu_usage_percentage_total{name=~"^github-proxy.*"}[5m])`

</details>

<br />

#### github-proxy: provisioning_container_memory_usage_short_term

<p class="subtitle">Container memory usage (5m maximum) by instance</p>

Refer to the [alert solutions reference](./alert_solutions.md#github-proxy-provisioning-container-memory-usage-short-term) for 1 alert related to this panel.

To see this panel, visit `/-/debug/grafana/d/github-proxy/github-proxy?viewPanel=100211` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<details>
<summary>Technical details</summary>

Query: `max_over_time(cadvisor_container_memory_usage_percentage_total{name=~"^github-proxy.*"}[5m])`

</details>

<br />

### GitHub Proxy: Golang runtime monitoring

#### github-proxy: go_goroutines

<p class="subtitle">Maximum active goroutines</p>

A high value here indicates a possible goroutine leak.

Refer to the [alert solutions reference](./alert_solutions.md#github-proxy-go-goroutines) for 1 alert related to this panel.

To see this panel, visit `/-/debug/grafana/d/github-proxy/github-proxy?viewPanel=100300` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<details>
<summary>Technical details</summary>

Query: `max by(instance) (go_goroutines{job=~".*github-proxy"})`

</details>

<br />

#### github-proxy: go_gc_duration_seconds

<p class="subtitle">Maximum go garbage collection duration</p>

Refer to the [alert solutions reference](./alert_solutions.md#github-proxy-go-gc-duration-seconds) for 1 alert related to this panel.

To see this panel, visit `/-/debug/grafana/d/github-proxy/github-proxy?viewPanel=100301` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<details>
<summary>Technical details</summary>

Query: `max by(instance) (go_gc_duration_seconds{job=~".*github-proxy"})`

</details>

<br />

### GitHub Proxy: Kubernetes monitoring (only available on Kubernetes)

#### github-proxy: pods_available_percentage

<p class="subtitle">Percentage pods available</p>

Refer to the [alert solutions reference](./alert_solutions.md#github-proxy-pods-available-percentage) for 1 alert related to this panel.

To see this panel, visit `/-/debug/grafana/d/github-proxy/github-proxy?viewPanel=100400` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum by(app) (up{app=~".*github-proxy"}) / count by (app) (up{app=~".*github-proxy"}) * 100`

</details>

<br />

## Postgres

<p class="subtitle">Postgres metrics, exported from postgres_exporter (only available on Kubernetes).</p>

To see this dashboard, visit `/-/debug/grafana/d/postgres/postgres` on your Sourcegraph instance.

#### postgres: connections

<p class="subtitle">Active connections</p>

Refer to the [alert solutions reference](./alert_solutions.md#postgres-connections) for 1 alert related to this panel.

To see this panel, visit `/-/debug/grafana/d/postgres/postgres?viewPanel=100000` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum by (job) (pg_stat_activity_count{datname!~"template.*|postgres|cloudsqladmin"})`

</details>

<br />

#### postgres: transaction_durations

<p class="subtitle">Maximum transaction durations</p>

Refer to the [alert solutions reference](./alert_solutions.md#postgres-transaction-durations) for 2 alerts related to this panel.

To see this panel, visit `/-/debug/grafana/d/postgres/postgres?viewPanel=100001` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum by (datname) (pg_stat_activity_max_tx_duration{datname!~"template.*|postgres|cloudsqladmin"})`

</details>

<br />

### Postgres: Database and collector status

#### postgres: postgres_up

<p class="subtitle">Database availability</p>

A non-zero value indicates the database is online.

Refer to the [alert solutions reference](./alert_solutions.md#postgres-postgres-up) for 1 alert related to this panel.

To see this panel, visit `/-/debug/grafana/d/postgres/postgres?viewPanel=100100` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<details>
<summary>Technical details</summary>

Query: `pg_up`

</details>

<br />

#### postgres: invalid_indexes

<p class="subtitle">Invalid indexes (unusable by the query planner)</p>

A non-zero value indicates the that Postgres failed to build an index. Expect degraded performance until the index is manually rebuilt.

Refer to the [alert solutions reference](./alert_solutions.md#postgres-invalid-indexes) for 1 alert related to this panel.

To see this panel, visit `/-/debug/grafana/d/postgres/postgres?viewPanel=100101` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<details>
<summary>Technical details</summary>

Query: `max by (relname)(pg_invalid_index_count)`

</details>

<br />

#### postgres: pg_exporter_err

<p class="subtitle">Errors scraping postgres exporter</p>

This value indicates issues retrieving metrics from postgres_exporter.

Refer to the [alert solutions reference](./alert_solutions.md#postgres-pg-exporter-err) for 1 alert related to this panel.

To see this panel, visit `/-/debug/grafana/d/postgres/postgres?viewPanel=100110` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<details>
<summary>Technical details</summary>

Query: `pg_exporter_last_scrape_error`

</details>

<br />

#### postgres: migration_in_progress

<p class="subtitle">Active schema migration</p>

A 0 value indicates that no migration is in progress.

Refer to the [alert solutions reference](./alert_solutions.md#postgres-migration-in-progress) for 1 alert related to this panel.

To see this panel, visit `/-/debug/grafana/d/postgres/postgres?viewPanel=100111` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<details>
<summary>Technical details</summary>

Query: `pg_sg_migration_status`

</details>

<br />

### Postgres: Object size and bloat

#### postgres: pg_table_size

<p class="subtitle">Table size</p>

Total size of this table

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/postgres/postgres?viewPanel=100200` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<details>
<summary>Technical details</summary>

Query: `max by (relname)(pg_table_bloat_size)`

</details>

<br />

#### postgres: pg_table_bloat_ratio

<p class="subtitle">Table bloat ratio</p>

Estimated bloat ratio of this table (high bloat = high overhead)

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/postgres/postgres?viewPanel=100201` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<details>
<summary>Technical details</summary>

Query: `max by (relname)(pg_table_bloat_ratio) * 100`

</details>

<br />

#### postgres: pg_index_size

<p class="subtitle">Index size</p>

Total size of this index

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/postgres/postgres?viewPanel=100210` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<details>
<summary>Technical details</summary>

Query: `max by (relname)(pg_index_bloat_size)`

</details>

<br />

#### postgres: pg_index_bloat_ratio

<p class="subtitle">Index bloat ratio</p>

Estimated bloat ratio of this index (high bloat = high overhead)

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/postgres/postgres?viewPanel=100211` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<details>
<summary>Technical details</summary>

Query: `max by (relname)(pg_index_bloat_ratio) * 100`

</details>

<br />

### Postgres: Provisioning indicators (not available on server)

#### postgres: provisioning_container_cpu_usage_long_term

<p class="subtitle">Container cpu usage total (90th percentile over 1d) across all cores by instance</p>

Refer to the [alert solutions reference](./alert_solutions.md#postgres-provisioning-container-cpu-usage-long-term) for 1 alert related to this panel.

To see this panel, visit `/-/debug/grafana/d/postgres/postgres?viewPanel=100300` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<details>
<summary>Technical details</summary>

Query: `quantile_over_time(0.9, cadvisor_container_cpu_usage_percentage_total{name=~"^(pgsql|codeintel-db).*"}[1d])`

</details>

<br />

#### postgres: provisioning_container_memory_usage_long_term

<p class="subtitle">Container memory usage (1d maximum) by instance</p>

Refer to the [alert solutions reference](./alert_solutions.md#postgres-provisioning-container-memory-usage-long-term) for 1 alert related to this panel.

To see this panel, visit `/-/debug/grafana/d/postgres/postgres?viewPanel=100301` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<details>
<summary>Technical details</summary>

Query: `max_over_time(cadvisor_container_memory_usage_percentage_total{name=~"^(pgsql|codeintel-db).*"}[1d])`

</details>

<br />

#### postgres: provisioning_container_cpu_usage_short_term

<p class="subtitle">Container cpu usage total (5m maximum) across all cores by instance</p>

Refer to the [alert solutions reference](./alert_solutions.md#postgres-provisioning-container-cpu-usage-short-term) for 1 alert related to this panel.

To see this panel, visit `/-/debug/grafana/d/postgres/postgres?viewPanel=100310` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<details>
<summary>Technical details</summary>

Query: `max_over_time(cadvisor_container_cpu_usage_percentage_total{name=~"^(pgsql|codeintel-db).*"}[5m])`

</details>

<br />

#### postgres: provisioning_container_memory_usage_short_term

<p class="subtitle">Container memory usage (5m maximum) by instance</p>

Refer to the [alert solutions reference](./alert_solutions.md#postgres-provisioning-container-memory-usage-short-term) for 1 alert related to this panel.

To see this panel, visit `/-/debug/grafana/d/postgres/postgres?viewPanel=100311` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<details>
<summary>Technical details</summary>

Query: `max_over_time(cadvisor_container_memory_usage_percentage_total{name=~"^(pgsql|codeintel-db).*"}[5m])`

</details>

<br />

### Postgres: Kubernetes monitoring (only available on Kubernetes)

#### postgres: pods_available_percentage

<p class="subtitle">Percentage pods available</p>

Refer to the [alert solutions reference](./alert_solutions.md#postgres-pods-available-percentage) for 1 alert related to this panel.

To see this panel, visit `/-/debug/grafana/d/postgres/postgres?viewPanel=100400` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum by(app) (up{app=~".*(pgsql|codeintel-db)"}) / count by (app) (up{app=~".*(pgsql|codeintel-db)"}) * 100`

</details>

<br />

## Precise Code Intel Worker

<p class="subtitle">Handles conversion of uploaded precise code intelligence bundles.</p>

To see this dashboard, visit `/-/debug/grafana/d/precise-code-intel-worker/precise-code-intel-worker` on your Sourcegraph instance.

### Precise Code Intel Worker: Codeintel: LSIF uploads

#### precise-code-intel-worker: codeintel_upload_queue_size

<p class="subtitle">Unprocessed upload record queue size</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/precise-code-intel-worker/precise-code-intel-worker?viewPanel=100000` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `max(src_codeintel_upload_total{job=~"^precise-code-intel-worker.*"})`

</details>

<br />

#### precise-code-intel-worker: codeintel_upload_queue_growth_rate

<p class="subtitle">Unprocessed upload record queue growth rate over 30m</p>

This value compares the rate of enqueues against the rate of finished jobs.

	- A value < than 1 indicates that process rate > enqueue rate
	- A value = than 1 indicates that process rate = enqueue rate
	- A value > than 1 indicates that process rate < enqueue rate

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/precise-code-intel-worker/precise-code-intel-worker?viewPanel=100001` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum(increase(src_codeintel_upload_total{job=~"^precise-code-intel-worker.*"}[30m])) / sum(increase(src_codeintel_upload_processor_total{job=~"^precise-code-intel-worker.*"}[30m]))`

</details>

<br />

### Precise Code Intel Worker: Codeintel: LSIF uploads

#### precise-code-intel-worker: codeintel_upload_handlers

<p class="subtitle">Handler active handlers</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/precise-code-intel-worker/precise-code-intel-worker?viewPanel=100100` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum(src_codeintel_upload_processor_handlers{job=~"^precise-code-intel-worker.*"})`

</details>

<br />

#### precise-code-intel-worker: codeintel_upload_processor_total

<p class="subtitle">Handler operations every 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/precise-code-intel-worker/precise-code-intel-worker?viewPanel=100110` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum(increase(src_codeintel_upload_processor_total{job=~"^precise-code-intel-worker.*"}[5m]))`

</details>

<br />

#### precise-code-intel-worker: codeintel_upload_processor_99th_percentile_duration

<p class="subtitle">99th percentile successful handler operation duration over 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/precise-code-intel-worker/precise-code-intel-worker?viewPanel=100111` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `histogram_quantile(0.99, sum  by (le)(rate(src_codeintel_upload_processor_duration_seconds_bucket{job=~"^precise-code-intel-worker.*"}[5m])))`

</details>

<br />

#### precise-code-intel-worker: codeintel_upload_processor_errors_total

<p class="subtitle">Handler operation errors every 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/precise-code-intel-worker/precise-code-intel-worker?viewPanel=100112` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum(increase(src_codeintel_upload_processor_errors_total{job=~"^precise-code-intel-worker.*"}[5m]))`

</details>

<br />

#### precise-code-intel-worker: codeintel_upload_processor_error_rate

<p class="subtitle">Handler operation error rate over 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/precise-code-intel-worker/precise-code-intel-worker?viewPanel=100113` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum(increase(src_codeintel_upload_processor_errors_total{job=~"^precise-code-intel-worker.*"}[5m])) / (sum(increase(src_codeintel_upload_processor_total{job=~"^precise-code-intel-worker.*"}[5m])) + sum(increase(src_codeintel_upload_processor_errors_total{job=~"^precise-code-intel-worker.*"}[5m]))) * 100`

</details>

<br />

### Precise Code Intel Worker: Codeintel: dbstore stats

#### precise-code-intel-worker: codeintel_dbstore_total

<p class="subtitle">Aggregate store operations every 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/precise-code-intel-worker/precise-code-intel-worker?viewPanel=100200` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum(increase(src_codeintel_dbstore_total{job=~"^precise-code-intel-worker.*"}[5m]))`

</details>

<br />

#### precise-code-intel-worker: codeintel_dbstore_99th_percentile_duration

<p class="subtitle">99th percentile successful aggregate store operation duration over 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/precise-code-intel-worker/precise-code-intel-worker?viewPanel=100201` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `histogram_quantile(0.99, sum  by (le)(rate(src_codeintel_dbstore_duration_seconds_bucket{job=~"^precise-code-intel-worker.*"}[5m])))`

</details>

<br />

#### precise-code-intel-worker: codeintel_dbstore_errors_total

<p class="subtitle">Aggregate store operation errors every 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/precise-code-intel-worker/precise-code-intel-worker?viewPanel=100202` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum(increase(src_codeintel_dbstore_errors_total{job=~"^precise-code-intel-worker.*"}[5m]))`

</details>

<br />

#### precise-code-intel-worker: codeintel_dbstore_error_rate

<p class="subtitle">Aggregate store operation error rate over 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/precise-code-intel-worker/precise-code-intel-worker?viewPanel=100203` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum(increase(src_codeintel_dbstore_errors_total{job=~"^precise-code-intel-worker.*"}[5m])) / (sum(increase(src_codeintel_dbstore_total{job=~"^precise-code-intel-worker.*"}[5m])) + sum(increase(src_codeintel_dbstore_errors_total{job=~"^precise-code-intel-worker.*"}[5m]))) * 100`

</details>

<br />

#### precise-code-intel-worker: codeintel_dbstore_total

<p class="subtitle">Store operations every 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/precise-code-intel-worker/precise-code-intel-worker?viewPanel=100210` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum by (op)(increase(src_codeintel_dbstore_total{job=~"^precise-code-intel-worker.*"}[5m]))`

</details>

<br />

#### precise-code-intel-worker: codeintel_dbstore_99th_percentile_duration

<p class="subtitle">99th percentile successful store operation duration over 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/precise-code-intel-worker/precise-code-intel-worker?viewPanel=100211` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `histogram_quantile(0.99, sum  by (le,op)(rate(src_codeintel_dbstore_duration_seconds_bucket{job=~"^precise-code-intel-worker.*"}[5m])))`

</details>

<br />

#### precise-code-intel-worker: codeintel_dbstore_errors_total

<p class="subtitle">Store operation errors every 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/precise-code-intel-worker/precise-code-intel-worker?viewPanel=100212` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum by (op)(increase(src_codeintel_dbstore_errors_total{job=~"^precise-code-intel-worker.*"}[5m]))`

</details>

<br />

#### precise-code-intel-worker: codeintel_dbstore_error_rate

<p class="subtitle">Store operation error rate over 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/precise-code-intel-worker/precise-code-intel-worker?viewPanel=100213` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum by (op)(increase(src_codeintel_dbstore_errors_total{job=~"^precise-code-intel-worker.*"}[5m])) / (sum by (op)(increase(src_codeintel_dbstore_total{job=~"^precise-code-intel-worker.*"}[5m])) + sum by (op)(increase(src_codeintel_dbstore_errors_total{job=~"^precise-code-intel-worker.*"}[5m]))) * 100`

</details>

<br />

### Precise Code Intel Worker: Codeintel: lsifstore stats

#### precise-code-intel-worker: codeintel_lsifstore_total

<p class="subtitle">Aggregate store operations every 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/precise-code-intel-worker/precise-code-intel-worker?viewPanel=100300` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum(increase(src_codeintel_lsifstore_total{job=~"^precise-code-intel-worker.*"}[5m]))`

</details>

<br />

#### precise-code-intel-worker: codeintel_lsifstore_99th_percentile_duration

<p class="subtitle">99th percentile successful aggregate store operation duration over 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/precise-code-intel-worker/precise-code-intel-worker?viewPanel=100301` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `histogram_quantile(0.99, sum  by (le)(rate(src_codeintel_lsifstore_duration_seconds_bucket{job=~"^precise-code-intel-worker.*"}[5m])))`

</details>

<br />

#### precise-code-intel-worker: codeintel_lsifstore_errors_total

<p class="subtitle">Aggregate store operation errors every 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/precise-code-intel-worker/precise-code-intel-worker?viewPanel=100302` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum(increase(src_codeintel_lsifstore_errors_total{job=~"^precise-code-intel-worker.*"}[5m]))`

</details>

<br />

#### precise-code-intel-worker: codeintel_lsifstore_error_rate

<p class="subtitle">Aggregate store operation error rate over 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/precise-code-intel-worker/precise-code-intel-worker?viewPanel=100303` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum(increase(src_codeintel_lsifstore_errors_total{job=~"^precise-code-intel-worker.*"}[5m])) / (sum(increase(src_codeintel_lsifstore_total{job=~"^precise-code-intel-worker.*"}[5m])) + sum(increase(src_codeintel_lsifstore_errors_total{job=~"^precise-code-intel-worker.*"}[5m]))) * 100`

</details>

<br />

#### precise-code-intel-worker: codeintel_lsifstore_total

<p class="subtitle">Store operations every 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/precise-code-intel-worker/precise-code-intel-worker?viewPanel=100310` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum by (op)(increase(src_codeintel_lsifstore_total{job=~"^precise-code-intel-worker.*"}[5m]))`

</details>

<br />

#### precise-code-intel-worker: codeintel_lsifstore_99th_percentile_duration

<p class="subtitle">99th percentile successful store operation duration over 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/precise-code-intel-worker/precise-code-intel-worker?viewPanel=100311` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `histogram_quantile(0.99, sum  by (le,op)(rate(src_codeintel_lsifstore_duration_seconds_bucket{job=~"^precise-code-intel-worker.*"}[5m])))`

</details>

<br />

#### precise-code-intel-worker: codeintel_lsifstore_errors_total

<p class="subtitle">Store operation errors every 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/precise-code-intel-worker/precise-code-intel-worker?viewPanel=100312` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum by (op)(increase(src_codeintel_lsifstore_errors_total{job=~"^precise-code-intel-worker.*"}[5m]))`

</details>

<br />

#### precise-code-intel-worker: codeintel_lsifstore_error_rate

<p class="subtitle">Store operation error rate over 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/precise-code-intel-worker/precise-code-intel-worker?viewPanel=100313` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum by (op)(increase(src_codeintel_lsifstore_errors_total{job=~"^precise-code-intel-worker.*"}[5m])) / (sum by (op)(increase(src_codeintel_lsifstore_total{job=~"^precise-code-intel-worker.*"}[5m])) + sum by (op)(increase(src_codeintel_lsifstore_errors_total{job=~"^precise-code-intel-worker.*"}[5m]))) * 100`

</details>

<br />

### Precise Code Intel Worker: Workerutil: lsif_uploads dbworker/store stats

#### precise-code-intel-worker: workerutil_dbworker_store_codeintel_upload_total

<p class="subtitle">Store operations every 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/precise-code-intel-worker/precise-code-intel-worker?viewPanel=100400` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum(increase(src_workerutil_dbworker_store_codeintel_upload_total{job=~"^precise-code-intel-worker.*"}[5m]))`

</details>

<br />

#### precise-code-intel-worker: workerutil_dbworker_store_codeintel_upload_99th_percentile_duration

<p class="subtitle">99th percentile successful store operation duration over 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/precise-code-intel-worker/precise-code-intel-worker?viewPanel=100401` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `histogram_quantile(0.99, sum  by (le)(rate(src_workerutil_dbworker_store_codeintel_upload_duration_seconds_bucket{job=~"^precise-code-intel-worker.*"}[5m])))`

</details>

<br />

#### precise-code-intel-worker: workerutil_dbworker_store_codeintel_upload_errors_total

<p class="subtitle">Store operation errors every 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/precise-code-intel-worker/precise-code-intel-worker?viewPanel=100402` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum(increase(src_workerutil_dbworker_store_codeintel_upload_errors_total{job=~"^precise-code-intel-worker.*"}[5m]))`

</details>

<br />

#### precise-code-intel-worker: workerutil_dbworker_store_codeintel_upload_error_rate

<p class="subtitle">Store operation error rate over 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/precise-code-intel-worker/precise-code-intel-worker?viewPanel=100403` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum(increase(src_workerutil_dbworker_store_codeintel_upload_errors_total{job=~"^precise-code-intel-worker.*"}[5m])) / (sum(increase(src_workerutil_dbworker_store_codeintel_upload_total{job=~"^precise-code-intel-worker.*"}[5m])) + sum(increase(src_workerutil_dbworker_store_codeintel_upload_errors_total{job=~"^precise-code-intel-worker.*"}[5m]))) * 100`

</details>

<br />

### Precise Code Intel Worker: Codeintel: gitserver client

#### precise-code-intel-worker: codeintel_gitserver_total

<p class="subtitle">Aggregate client operations every 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/precise-code-intel-worker/precise-code-intel-worker?viewPanel=100500` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum(increase(src_codeintel_gitserver_total{job=~"^precise-code-intel-worker.*"}[5m]))`

</details>

<br />

#### precise-code-intel-worker: codeintel_gitserver_99th_percentile_duration

<p class="subtitle">99th percentile successful aggregate client operation duration over 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/precise-code-intel-worker/precise-code-intel-worker?viewPanel=100501` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `histogram_quantile(0.99, sum  by (le)(rate(src_codeintel_gitserver_duration_seconds_bucket{job=~"^precise-code-intel-worker.*"}[5m])))`

</details>

<br />

#### precise-code-intel-worker: codeintel_gitserver_errors_total

<p class="subtitle">Aggregate client operation errors every 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/precise-code-intel-worker/precise-code-intel-worker?viewPanel=100502` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum(increase(src_codeintel_gitserver_errors_total{job=~"^precise-code-intel-worker.*"}[5m]))`

</details>

<br />

#### precise-code-intel-worker: codeintel_gitserver_error_rate

<p class="subtitle">Aggregate client operation error rate over 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/precise-code-intel-worker/precise-code-intel-worker?viewPanel=100503` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum(increase(src_codeintel_gitserver_errors_total{job=~"^precise-code-intel-worker.*"}[5m])) / (sum(increase(src_codeintel_gitserver_total{job=~"^precise-code-intel-worker.*"}[5m])) + sum(increase(src_codeintel_gitserver_errors_total{job=~"^precise-code-intel-worker.*"}[5m]))) * 100`

</details>

<br />

#### precise-code-intel-worker: codeintel_gitserver_total

<p class="subtitle">Client operations every 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/precise-code-intel-worker/precise-code-intel-worker?viewPanel=100510` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum by (op)(increase(src_codeintel_gitserver_total{job=~"^precise-code-intel-worker.*"}[5m]))`

</details>

<br />

#### precise-code-intel-worker: codeintel_gitserver_99th_percentile_duration

<p class="subtitle">99th percentile successful client operation duration over 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/precise-code-intel-worker/precise-code-intel-worker?viewPanel=100511` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `histogram_quantile(0.99, sum  by (le,op)(rate(src_codeintel_gitserver_duration_seconds_bucket{job=~"^precise-code-intel-worker.*"}[5m])))`

</details>

<br />

#### precise-code-intel-worker: codeintel_gitserver_errors_total

<p class="subtitle">Client operation errors every 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/precise-code-intel-worker/precise-code-intel-worker?viewPanel=100512` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum by (op)(increase(src_codeintel_gitserver_errors_total{job=~"^precise-code-intel-worker.*"}[5m]))`

</details>

<br />

#### precise-code-intel-worker: codeintel_gitserver_error_rate

<p class="subtitle">Client operation error rate over 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/precise-code-intel-worker/precise-code-intel-worker?viewPanel=100513` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum by (op)(increase(src_codeintel_gitserver_errors_total{job=~"^precise-code-intel-worker.*"}[5m])) / (sum by (op)(increase(src_codeintel_gitserver_total{job=~"^precise-code-intel-worker.*"}[5m])) + sum by (op)(increase(src_codeintel_gitserver_errors_total{job=~"^precise-code-intel-worker.*"}[5m]))) * 100`

</details>

<br />

### Precise Code Intel Worker: Codeintel: uploadstore stats

#### precise-code-intel-worker: codeintel_uploadstore_total

<p class="subtitle">Aggregate store operations every 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/precise-code-intel-worker/precise-code-intel-worker?viewPanel=100600` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum(increase(src_codeintel_uploadstore_total{job=~"^precise-code-intel-worker.*"}[5m]))`

</details>

<br />

#### precise-code-intel-worker: codeintel_uploadstore_99th_percentile_duration

<p class="subtitle">99th percentile successful aggregate store operation duration over 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/precise-code-intel-worker/precise-code-intel-worker?viewPanel=100601` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `histogram_quantile(0.99, sum  by (le)(rate(src_codeintel_uploadstore_duration_seconds_bucket{job=~"^precise-code-intel-worker.*"}[5m])))`

</details>

<br />

#### precise-code-intel-worker: codeintel_uploadstore_errors_total

<p class="subtitle">Aggregate store operation errors every 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/precise-code-intel-worker/precise-code-intel-worker?viewPanel=100602` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum(increase(src_codeintel_uploadstore_errors_total{job=~"^precise-code-intel-worker.*"}[5m]))`

</details>

<br />

#### precise-code-intel-worker: codeintel_uploadstore_error_rate

<p class="subtitle">Aggregate store operation error rate over 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/precise-code-intel-worker/precise-code-intel-worker?viewPanel=100603` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum(increase(src_codeintel_uploadstore_errors_total{job=~"^precise-code-intel-worker.*"}[5m])) / (sum(increase(src_codeintel_uploadstore_total{job=~"^precise-code-intel-worker.*"}[5m])) + sum(increase(src_codeintel_uploadstore_errors_total{job=~"^precise-code-intel-worker.*"}[5m]))) * 100`

</details>

<br />

#### precise-code-intel-worker: codeintel_uploadstore_total

<p class="subtitle">Store operations every 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/precise-code-intel-worker/precise-code-intel-worker?viewPanel=100610` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum by (op)(increase(src_codeintel_uploadstore_total{job=~"^precise-code-intel-worker.*"}[5m]))`

</details>

<br />

#### precise-code-intel-worker: codeintel_uploadstore_99th_percentile_duration

<p class="subtitle">99th percentile successful store operation duration over 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/precise-code-intel-worker/precise-code-intel-worker?viewPanel=100611` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `histogram_quantile(0.99, sum  by (le,op)(rate(src_codeintel_uploadstore_duration_seconds_bucket{job=~"^precise-code-intel-worker.*"}[5m])))`

</details>

<br />

#### precise-code-intel-worker: codeintel_uploadstore_errors_total

<p class="subtitle">Store operation errors every 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/precise-code-intel-worker/precise-code-intel-worker?viewPanel=100612` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum by (op)(increase(src_codeintel_uploadstore_errors_total{job=~"^precise-code-intel-worker.*"}[5m]))`

</details>

<br />

#### precise-code-intel-worker: codeintel_uploadstore_error_rate

<p class="subtitle">Store operation error rate over 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/precise-code-intel-worker/precise-code-intel-worker?viewPanel=100613` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum by (op)(increase(src_codeintel_uploadstore_errors_total{job=~"^precise-code-intel-worker.*"}[5m])) / (sum by (op)(increase(src_codeintel_uploadstore_total{job=~"^precise-code-intel-worker.*"}[5m])) + sum by (op)(increase(src_codeintel_uploadstore_errors_total{job=~"^precise-code-intel-worker.*"}[5m]))) * 100`

</details>

<br />

### Precise Code Intel Worker: Internal service requests

#### precise-code-intel-worker: frontend_internal_api_error_responses

<p class="subtitle">Frontend-internal API error responses every 5m by route</p>

Refer to the [alert solutions reference](./alert_solutions.md#precise-code-intel-worker-frontend-internal-api-error-responses) for 1 alert related to this panel.

To see this panel, visit `/-/debug/grafana/d/precise-code-intel-worker/precise-code-intel-worker?viewPanel=100700` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum by (category)(increase(src_frontend_internal_request_duration_seconds_count{job="precise-code-intel-worker",code!~"2.."}[5m])) / ignoring(category) group_left sum(increase(src_frontend_internal_request_duration_seconds_count{job="precise-code-intel-worker"}[5m]))`

</details>

<br />

### Precise Code Intel Worker: Database connections

#### precise-code-intel-worker: max_open_conns

<p class="subtitle">Maximum open</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/precise-code-intel-worker/precise-code-intel-worker?viewPanel=100800` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum by (app_name, db_name) (src_pgsql_conns_max_open{app_name="precise-code-intel-worker"})`

</details>

<br />

#### precise-code-intel-worker: open_conns

<p class="subtitle">Established</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/precise-code-intel-worker/precise-code-intel-worker?viewPanel=100801` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum by (app_name, db_name) (src_pgsql_conns_open{app_name="precise-code-intel-worker"})`

</details>

<br />

#### precise-code-intel-worker: in_use

<p class="subtitle">Used</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/precise-code-intel-worker/precise-code-intel-worker?viewPanel=100810` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum by (app_name, db_name) (src_pgsql_conns_in_use{app_name="precise-code-intel-worker"})`

</details>

<br />

#### precise-code-intel-worker: idle

<p class="subtitle">Idle</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/precise-code-intel-worker/precise-code-intel-worker?viewPanel=100811` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum by (app_name, db_name) (src_pgsql_conns_idle{app_name="precise-code-intel-worker"})`

</details>

<br />

#### precise-code-intel-worker: mean_blocked_seconds_per_conn_request

<p class="subtitle">Mean blocked seconds per conn request</p>

Refer to the [alert solutions reference](./alert_solutions.md#precise-code-intel-worker-mean-blocked-seconds-per-conn-request) for 2 alerts related to this panel.

To see this panel, visit `/-/debug/grafana/d/precise-code-intel-worker/precise-code-intel-worker?viewPanel=100820` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum by (app_name, db_name) (increase(src_pgsql_conns_blocked_seconds{app_name="precise-code-intel-worker"}[5m])) / sum by (app_name, db_name) (increase(src_pgsql_conns_waited_for{app_name="precise-code-intel-worker"}[5m]))`

</details>

<br />

#### precise-code-intel-worker: closed_max_idle

<p class="subtitle">Closed by SetMaxIdleConns</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/precise-code-intel-worker/precise-code-intel-worker?viewPanel=100830` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum by (app_name, db_name) (increase(src_pgsql_conns_closed_max_idle{app_name="precise-code-intel-worker"}[5m]))`

</details>

<br />

#### precise-code-intel-worker: closed_max_lifetime

<p class="subtitle">Closed by SetConnMaxLifetime</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/precise-code-intel-worker/precise-code-intel-worker?viewPanel=100831` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum by (app_name, db_name) (increase(src_pgsql_conns_closed_max_lifetime{app_name="precise-code-intel-worker"}[5m]))`

</details>

<br />

#### precise-code-intel-worker: closed_max_idle_time

<p class="subtitle">Closed by SetConnMaxIdleTime</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/precise-code-intel-worker/precise-code-intel-worker?viewPanel=100832` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum by (app_name, db_name) (increase(src_pgsql_conns_closed_max_idle_time{app_name="precise-code-intel-worker"}[5m]))`

</details>

<br />

### Precise Code Intel Worker: Container monitoring (not available on server)

#### precise-code-intel-worker: container_missing

<p class="subtitle">Container missing</p>

This value is the number of times a container has not been seen for more than one minute. If you observe this
value change independent of deployment events (such as an upgrade), it could indicate pods are being OOM killed or terminated for some other reasons.

- **Kubernetes:**
	- Determine if the pod was OOM killed using `kubectl describe pod precise-code-intel-worker` (look for `OOMKilled: true`) and, if so, consider increasing the memory limit in the relevant `Deployment.yaml`.
	- Check the logs before the container restarted to see if there are `panic:` messages or similar using `kubectl logs -p precise-code-intel-worker`.
- **Docker Compose:**
	- Determine if the pod was OOM killed using `docker inspect -f '{{json .State}}' precise-code-intel-worker` (look for `"OOMKilled":true`) and, if so, consider increasing the memory limit of the precise-code-intel-worker container in `docker-compose.yml`.
	- Check the logs before the container restarted to see if there are `panic:` messages or similar using `docker logs precise-code-intel-worker` (note this will include logs from the previous and currently running container).

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/precise-code-intel-worker/precise-code-intel-worker?viewPanel=100900` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `count by(name) ((time() - container_last_seen{name=~"^precise-code-intel-worker.*"}) > 60)`

</details>

<br />

#### precise-code-intel-worker: container_cpu_usage

<p class="subtitle">Container cpu usage total (1m average) across all cores by instance</p>

Refer to the [alert solutions reference](./alert_solutions.md#precise-code-intel-worker-container-cpu-usage) for 1 alert related to this panel.

To see this panel, visit `/-/debug/grafana/d/precise-code-intel-worker/precise-code-intel-worker?viewPanel=100901` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `cadvisor_container_cpu_usage_percentage_total{name=~"^precise-code-intel-worker.*"}`

</details>

<br />

#### precise-code-intel-worker: container_memory_usage

<p class="subtitle">Container memory usage by instance</p>

Refer to the [alert solutions reference](./alert_solutions.md#precise-code-intel-worker-container-memory-usage) for 1 alert related to this panel.

To see this panel, visit `/-/debug/grafana/d/precise-code-intel-worker/precise-code-intel-worker?viewPanel=100902` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `cadvisor_container_memory_usage_percentage_total{name=~"^precise-code-intel-worker.*"}`

</details>

<br />

#### precise-code-intel-worker: fs_io_operations

<p class="subtitle">Filesystem reads and writes rate by instance over 1h</p>

This value indicates the number of filesystem read and write operations by containers of this service.
When extremely high, this can indicate a resource usage problem, or can cause problems with the service itself, especially if high values or spikes correlate with {{CONTAINER_NAME}} issues.

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/precise-code-intel-worker/precise-code-intel-worker?viewPanel=100903` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum by(name) (rate(container_fs_reads_total{name=~"^precise-code-intel-worker.*"}[1h]) + rate(container_fs_writes_total{name=~"^precise-code-intel-worker.*"}[1h]))`

</details>

<br />

### Precise Code Intel Worker: Provisioning indicators (not available on server)

#### precise-code-intel-worker: provisioning_container_cpu_usage_long_term

<p class="subtitle">Container cpu usage total (90th percentile over 1d) across all cores by instance</p>

Refer to the [alert solutions reference](./alert_solutions.md#precise-code-intel-worker-provisioning-container-cpu-usage-long-term) for 1 alert related to this panel.

To see this panel, visit `/-/debug/grafana/d/precise-code-intel-worker/precise-code-intel-worker?viewPanel=101000` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `quantile_over_time(0.9, cadvisor_container_cpu_usage_percentage_total{name=~"^precise-code-intel-worker.*"}[1d])`

</details>

<br />

#### precise-code-intel-worker: provisioning_container_memory_usage_long_term

<p class="subtitle">Container memory usage (1d maximum) by instance</p>

Refer to the [alert solutions reference](./alert_solutions.md#precise-code-intel-worker-provisioning-container-memory-usage-long-term) for 1 alert related to this panel.

To see this panel, visit `/-/debug/grafana/d/precise-code-intel-worker/precise-code-intel-worker?viewPanel=101001` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `max_over_time(cadvisor_container_memory_usage_percentage_total{name=~"^precise-code-intel-worker.*"}[1d])`

</details>

<br />

#### precise-code-intel-worker: provisioning_container_cpu_usage_short_term

<p class="subtitle">Container cpu usage total (5m maximum) across all cores by instance</p>

Refer to the [alert solutions reference](./alert_solutions.md#precise-code-intel-worker-provisioning-container-cpu-usage-short-term) for 1 alert related to this panel.

To see this panel, visit `/-/debug/grafana/d/precise-code-intel-worker/precise-code-intel-worker?viewPanel=101010` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `max_over_time(cadvisor_container_cpu_usage_percentage_total{name=~"^precise-code-intel-worker.*"}[5m])`

</details>

<br />

#### precise-code-intel-worker: provisioning_container_memory_usage_short_term

<p class="subtitle">Container memory usage (5m maximum) by instance</p>

Refer to the [alert solutions reference](./alert_solutions.md#precise-code-intel-worker-provisioning-container-memory-usage-short-term) for 1 alert related to this panel.

To see this panel, visit `/-/debug/grafana/d/precise-code-intel-worker/precise-code-intel-worker?viewPanel=101011` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `max_over_time(cadvisor_container_memory_usage_percentage_total{name=~"^precise-code-intel-worker.*"}[5m])`

</details>

<br />

### Precise Code Intel Worker: Golang runtime monitoring

#### precise-code-intel-worker: go_goroutines

<p class="subtitle">Maximum active goroutines</p>

A high value here indicates a possible goroutine leak.

Refer to the [alert solutions reference](./alert_solutions.md#precise-code-intel-worker-go-goroutines) for 1 alert related to this panel.

To see this panel, visit `/-/debug/grafana/d/precise-code-intel-worker/precise-code-intel-worker?viewPanel=101100` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `max by(instance) (go_goroutines{job=~".*precise-code-intel-worker"})`

</details>

<br />

#### precise-code-intel-worker: go_gc_duration_seconds

<p class="subtitle">Maximum go garbage collection duration</p>

Refer to the [alert solutions reference](./alert_solutions.md#precise-code-intel-worker-go-gc-duration-seconds) for 1 alert related to this panel.

To see this panel, visit `/-/debug/grafana/d/precise-code-intel-worker/precise-code-intel-worker?viewPanel=101101` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `max by(instance) (go_gc_duration_seconds{job=~".*precise-code-intel-worker"})`

</details>

<br />

### Precise Code Intel Worker: Kubernetes monitoring (only available on Kubernetes)

#### precise-code-intel-worker: pods_available_percentage

<p class="subtitle">Percentage pods available</p>

Refer to the [alert solutions reference](./alert_solutions.md#precise-code-intel-worker-pods-available-percentage) for 1 alert related to this panel.

To see this panel, visit `/-/debug/grafana/d/precise-code-intel-worker/precise-code-intel-worker?viewPanel=101200` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum by(app) (up{app=~".*precise-code-intel-worker"}) / count by (app) (up{app=~".*precise-code-intel-worker"}) * 100`

</details>

<br />

## Query Runner

<p class="subtitle">Periodically runs saved searches and instructs the frontend to send out notifications.</p>

To see this dashboard, visit `/-/debug/grafana/d/query-runner/query-runner` on your Sourcegraph instance.

### Query Runner: Internal service requests

#### query-runner: frontend_internal_api_error_responses

<p class="subtitle">Frontend-internal API error responses every 5m by route</p>

Refer to the [alert solutions reference](./alert_solutions.md#query-runner-frontend-internal-api-error-responses) for 1 alert related to this panel.

To see this panel, visit `/-/debug/grafana/d/query-runner/query-runner?viewPanel=100000` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Search team](https://about.sourcegraph.com/handbook/engineering/search).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum by (category)(increase(src_frontend_internal_request_duration_seconds_count{job="query-runner",code!~"2.."}[5m])) / ignoring(category) group_left sum(increase(src_frontend_internal_request_duration_seconds_count{job="query-runner"}[5m]))`

</details>

<br />

### Query Runner: Container monitoring (not available on server)

#### query-runner: container_missing

<p class="subtitle">Container missing</p>

This value is the number of times a container has not been seen for more than one minute. If you observe this
value change independent of deployment events (such as an upgrade), it could indicate pods are being OOM killed or terminated for some other reasons.

- **Kubernetes:**
	- Determine if the pod was OOM killed using `kubectl describe pod query-runner` (look for `OOMKilled: true`) and, if so, consider increasing the memory limit in the relevant `Deployment.yaml`.
	- Check the logs before the container restarted to see if there are `panic:` messages or similar using `kubectl logs -p query-runner`.
- **Docker Compose:**
	- Determine if the pod was OOM killed using `docker inspect -f '{{json .State}}' query-runner` (look for `"OOMKilled":true`) and, if so, consider increasing the memory limit of the query-runner container in `docker-compose.yml`.
	- Check the logs before the container restarted to see if there are `panic:` messages or similar using `docker logs query-runner` (note this will include logs from the previous and currently running container).

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/query-runner/query-runner?viewPanel=100100` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Search team](https://about.sourcegraph.com/handbook/engineering/search).*</sub>

<details>
<summary>Technical details</summary>

Query: `count by(name) ((time() - container_last_seen{name=~"^query-runner.*"}) > 60)`

</details>

<br />

#### query-runner: container_cpu_usage

<p class="subtitle">Container cpu usage total (1m average) across all cores by instance</p>

Refer to the [alert solutions reference](./alert_solutions.md#query-runner-container-cpu-usage) for 1 alert related to this panel.

To see this panel, visit `/-/debug/grafana/d/query-runner/query-runner?viewPanel=100101` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Search team](https://about.sourcegraph.com/handbook/engineering/search).*</sub>

<details>
<summary>Technical details</summary>

Query: `cadvisor_container_cpu_usage_percentage_total{name=~"^query-runner.*"}`

</details>

<br />

#### query-runner: container_memory_usage

<p class="subtitle">Container memory usage by instance</p>

Refer to the [alert solutions reference](./alert_solutions.md#query-runner-container-memory-usage) for 1 alert related to this panel.

To see this panel, visit `/-/debug/grafana/d/query-runner/query-runner?viewPanel=100102` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Search team](https://about.sourcegraph.com/handbook/engineering/search).*</sub>

<details>
<summary>Technical details</summary>

Query: `cadvisor_container_memory_usage_percentage_total{name=~"^query-runner.*"}`

</details>

<br />

#### query-runner: fs_io_operations

<p class="subtitle">Filesystem reads and writes rate by instance over 1h</p>

This value indicates the number of filesystem read and write operations by containers of this service.
When extremely high, this can indicate a resource usage problem, or can cause problems with the service itself, especially if high values or spikes correlate with {{CONTAINER_NAME}} issues.

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/query-runner/query-runner?viewPanel=100103` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum by(name) (rate(container_fs_reads_total{name=~"^query-runner.*"}[1h]) + rate(container_fs_writes_total{name=~"^query-runner.*"}[1h]))`

</details>

<br />

### Query Runner: Provisioning indicators (not available on server)

#### query-runner: provisioning_container_cpu_usage_long_term

<p class="subtitle">Container cpu usage total (90th percentile over 1d) across all cores by instance</p>

Refer to the [alert solutions reference](./alert_solutions.md#query-runner-provisioning-container-cpu-usage-long-term) for 1 alert related to this panel.

To see this panel, visit `/-/debug/grafana/d/query-runner/query-runner?viewPanel=100200` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Search team](https://about.sourcegraph.com/handbook/engineering/search).*</sub>

<details>
<summary>Technical details</summary>

Query: `quantile_over_time(0.9, cadvisor_container_cpu_usage_percentage_total{name=~"^query-runner.*"}[1d])`

</details>

<br />

#### query-runner: provisioning_container_memory_usage_long_term

<p class="subtitle">Container memory usage (1d maximum) by instance</p>

Refer to the [alert solutions reference](./alert_solutions.md#query-runner-provisioning-container-memory-usage-long-term) for 1 alert related to this panel.

To see this panel, visit `/-/debug/grafana/d/query-runner/query-runner?viewPanel=100201` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Search team](https://about.sourcegraph.com/handbook/engineering/search).*</sub>

<details>
<summary>Technical details</summary>

Query: `max_over_time(cadvisor_container_memory_usage_percentage_total{name=~"^query-runner.*"}[1d])`

</details>

<br />

#### query-runner: provisioning_container_cpu_usage_short_term

<p class="subtitle">Container cpu usage total (5m maximum) across all cores by instance</p>

Refer to the [alert solutions reference](./alert_solutions.md#query-runner-provisioning-container-cpu-usage-short-term) for 1 alert related to this panel.

To see this panel, visit `/-/debug/grafana/d/query-runner/query-runner?viewPanel=100210` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Search team](https://about.sourcegraph.com/handbook/engineering/search).*</sub>

<details>
<summary>Technical details</summary>

Query: `max_over_time(cadvisor_container_cpu_usage_percentage_total{name=~"^query-runner.*"}[5m])`

</details>

<br />

#### query-runner: provisioning_container_memory_usage_short_term

<p class="subtitle">Container memory usage (5m maximum) by instance</p>

Refer to the [alert solutions reference](./alert_solutions.md#query-runner-provisioning-container-memory-usage-short-term) for 1 alert related to this panel.

To see this panel, visit `/-/debug/grafana/d/query-runner/query-runner?viewPanel=100211` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Search team](https://about.sourcegraph.com/handbook/engineering/search).*</sub>

<details>
<summary>Technical details</summary>

Query: `max_over_time(cadvisor_container_memory_usage_percentage_total{name=~"^query-runner.*"}[5m])`

</details>

<br />

### Query Runner: Golang runtime monitoring

#### query-runner: go_goroutines

<p class="subtitle">Maximum active goroutines</p>

A high value here indicates a possible goroutine leak.

Refer to the [alert solutions reference](./alert_solutions.md#query-runner-go-goroutines) for 1 alert related to this panel.

To see this panel, visit `/-/debug/grafana/d/query-runner/query-runner?viewPanel=100300` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Search team](https://about.sourcegraph.com/handbook/engineering/search).*</sub>

<details>
<summary>Technical details</summary>

Query: `max by(instance) (go_goroutines{job=~".*query-runner"})`

</details>

<br />

#### query-runner: go_gc_duration_seconds

<p class="subtitle">Maximum go garbage collection duration</p>

Refer to the [alert solutions reference](./alert_solutions.md#query-runner-go-gc-duration-seconds) for 1 alert related to this panel.

To see this panel, visit `/-/debug/grafana/d/query-runner/query-runner?viewPanel=100301` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Search team](https://about.sourcegraph.com/handbook/engineering/search).*</sub>

<details>
<summary>Technical details</summary>

Query: `max by(instance) (go_gc_duration_seconds{job=~".*query-runner"})`

</details>

<br />

### Query Runner: Kubernetes monitoring (only available on Kubernetes)

#### query-runner: pods_available_percentage

<p class="subtitle">Percentage pods available</p>

Refer to the [alert solutions reference](./alert_solutions.md#query-runner-pods-available-percentage) for 1 alert related to this panel.

To see this panel, visit `/-/debug/grafana/d/query-runner/query-runner?viewPanel=100400` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Search team](https://about.sourcegraph.com/handbook/engineering/search).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum by(app) (up{app=~".*query-runner"}) / count by (app) (up{app=~".*query-runner"}) * 100`

</details>

<br />

## Worker

<p class="subtitle">Manages background processes.</p>

To see this dashboard, visit `/-/debug/grafana/d/worker/worker` on your Sourcegraph instance.

### Worker: Active jobs

#### worker: worker_job_count

<p class="subtitle">Number of worker instances running each job</p>

The number of worker instances running each job type.
It is necessary for each job type to be managed by at least one worker instance.

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/worker/worker?viewPanel=100000` on your Sourcegraph instance.


<details>
<summary>Technical details</summary>

Query: `sum by (job_name) (src_worker_jobs{job="worker"})`

</details>

<br />

#### worker: worker_job_codeintel-janitor_count

<p class="subtitle">Number of worker instances running the codeintel-janitor job</p>

Refer to the [alert solutions reference](./alert_solutions.md#worker-worker-job-codeintel-janitor-count) for 2 alerts related to this panel.

To see this panel, visit `/-/debug/grafana/d/worker/worker?viewPanel=100010` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum (src_worker_jobs{job="worker", job_name="codeintel-janitor"})`

</details>

<br />

#### worker: worker_job_codeintel-commitgraph_count

<p class="subtitle">Number of worker instances running the codeintel-commitgraph job</p>

Refer to the [alert solutions reference](./alert_solutions.md#worker-worker-job-codeintel-commitgraph-count) for 2 alerts related to this panel.

To see this panel, visit `/-/debug/grafana/d/worker/worker?viewPanel=100011` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum (src_worker_jobs{job="worker", job_name="codeintel-commitgraph"})`

</details>

<br />

#### worker: worker_job_codeintel-auto-indexing_count

<p class="subtitle">Number of worker instances running the codeintel-auto-indexing job</p>

Refer to the [alert solutions reference](./alert_solutions.md#worker-worker-job-codeintel-auto-indexing-count) for 2 alerts related to this panel.

To see this panel, visit `/-/debug/grafana/d/worker/worker?viewPanel=100012` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum (src_worker_jobs{job="worker", job_name="codeintel-auto-indexing"})`

</details>

<br />

### Worker: Codeintel: Repository with stale commit graph

#### worker: codeintel_commit_graph_queue_size

<p class="subtitle">Repository queue size</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/worker/worker?viewPanel=100100` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `max(src_codeintel_commit_graph_total{job=~"^worker.*"})`

</details>

<br />

#### worker: codeintel_commit_graph_queue_growth_rate

<p class="subtitle">Repository queue growth rate over 30m</p>

This value compares the rate of enqueues against the rate of finished jobs.

	- A value < than 1 indicates that process rate > enqueue rate
	- A value = than 1 indicates that process rate = enqueue rate
	- A value > than 1 indicates that process rate < enqueue rate

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/worker/worker?viewPanel=100101` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum(increase(src_codeintel_commit_graph_total{job=~"^worker.*"}[30m])) / sum(increase(src_codeintel_commit_graph_processor_total{job=~"^worker.*"}[30m]))`

</details>

<br />

### Worker: Codeintel: Repository commit graph updates

#### worker: codeintel_commit_graph_processor_total

<p class="subtitle">Update operations every 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/worker/worker?viewPanel=100200` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum(increase(src_codeintel_commit_graph_processor_total{job=~"^worker.*"}[5m]))`

</details>

<br />

#### worker: codeintel_commit_graph_processor_99th_percentile_duration

<p class="subtitle">99th percentile successful update operation duration over 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/worker/worker?viewPanel=100201` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `histogram_quantile(0.99, sum  by (le)(rate(src_codeintel_commit_graph_processor_duration_seconds_bucket{job=~"^worker.*"}[5m])))`

</details>

<br />

#### worker: codeintel_commit_graph_processor_errors_total

<p class="subtitle">Update operation errors every 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/worker/worker?viewPanel=100202` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum(increase(src_codeintel_commit_graph_processor_errors_total{job=~"^worker.*"}[5m]))`

</details>

<br />

#### worker: codeintel_commit_graph_processor_error_rate

<p class="subtitle">Update operation error rate over 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/worker/worker?viewPanel=100203` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum(increase(src_codeintel_commit_graph_processor_errors_total{job=~"^worker.*"}[5m])) / (sum(increase(src_codeintel_commit_graph_processor_total{job=~"^worker.*"}[5m])) + sum(increase(src_codeintel_commit_graph_processor_errors_total{job=~"^worker.*"}[5m]))) * 100`

</details>

<br />

### Worker: Codeintel: Dependency index job

#### worker: codeintel_dependency_index_queue_size

<p class="subtitle">Dependency index job queue size</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/worker/worker?viewPanel=100300` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `max(src_codeintel_dependency_index_total{job=~"^worker.*"})`

</details>

<br />

#### worker: codeintel_dependency_index_queue_growth_rate

<p class="subtitle">Dependency index job queue growth rate over 30m</p>

This value compares the rate of enqueues against the rate of finished jobs.

	- A value < than 1 indicates that process rate > enqueue rate
	- A value = than 1 indicates that process rate = enqueue rate
	- A value > than 1 indicates that process rate < enqueue rate

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/worker/worker?viewPanel=100301` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum(increase(src_codeintel_dependency_index_total{job=~"^worker.*"}[30m])) / sum(increase(src_codeintel_dependency_index_processor_total{job=~"^worker.*"}[30m]))`

</details>

<br />

### Worker: Codeintel: Dependency index jobs

#### worker: codeintel_dependency_index_handlers

<p class="subtitle">Handler active handlers</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/worker/worker?viewPanel=100400` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum(src_codeintel_dependency_index_processor_handlers{job=~"^worker.*"})`

</details>

<br />

#### worker: codeintel_dependency_index_processor_total

<p class="subtitle">Handler operations every 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/worker/worker?viewPanel=100410` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum(increase(src_codeintel_dependency_index_processor_total{job=~"^worker.*"}[5m]))`

</details>

<br />

#### worker: codeintel_dependency_index_processor_99th_percentile_duration

<p class="subtitle">99th percentile successful handler operation duration over 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/worker/worker?viewPanel=100411` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `histogram_quantile(0.99, sum  by (le)(rate(src_codeintel_dependency_index_processor_duration_seconds_bucket{job=~"^worker.*"}[5m])))`

</details>

<br />

#### worker: codeintel_dependency_index_processor_errors_total

<p class="subtitle">Handler operation errors every 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/worker/worker?viewPanel=100412` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum(increase(src_codeintel_dependency_index_processor_errors_total{job=~"^worker.*"}[5m]))`

</details>

<br />

#### worker: codeintel_dependency_index_processor_error_rate

<p class="subtitle">Handler operation error rate over 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/worker/worker?viewPanel=100413` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum(increase(src_codeintel_dependency_index_processor_errors_total{job=~"^worker.*"}[5m])) / (sum(increase(src_codeintel_dependency_index_processor_total{job=~"^worker.*"}[5m])) + sum(increase(src_codeintel_dependency_index_processor_errors_total{job=~"^worker.*"}[5m]))) * 100`

</details>

<br />

### Worker: Codeintel: Janitor stats

#### worker: codeintel_background_repositories_scanned_total

<p class="subtitle">Repository records scanned every 5m</p>

Number of repositories considered for data retention scanning every 5m

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/worker/worker?viewPanel=100500` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum(increase(src_codeintel_background_repositories_scanned_total{job=~"^worker.*"}[5m]))`

</details>

<br />

#### worker: codeintel_background_upload_records_scanned_total

<p class="subtitle">Lsif upload records scanned every 5m</p>

Number of upload records considered for data retention scanning every 5m

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/worker/worker?viewPanel=100501` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum(increase(src_codeintel_background_upload_records_scanned_total{job=~"^worker.*"}[5m]))`

</details>

<br />

#### worker: codeintel_background_commits_scanned_total

<p class="subtitle">Lsif upload commits scanned every 5m</p>

Number of commits considered for data retention scanning every 5m

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/worker/worker?viewPanel=100502` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum(increase(src_codeintel_background_commits_scanned_total{job=~"^worker.*"}[5m]))`

</details>

<br />

#### worker: codeintel_background_upload_records_expired_total

<p class="subtitle">Lsif upload records expired every 5m</p>

Number of upload records found to be expired every 5m

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/worker/worker?viewPanel=100503` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum(increase(src_codeintel_background_upload_records_expired_total{job=~"^worker.*"}[5m]))`

</details>

<br />

#### worker: codeintel_background_upload_records_removed_total

<p class="subtitle">Lsif upload records deleted every 5m</p>

Number of LSIF upload records deleted due to expiration or unreachability every 5m

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/worker/worker?viewPanel=100510` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum(increase(src_codeintel_background_upload_records_removed_total{job=~"^worker.*"}[5m]))`

</details>

<br />

#### worker: codeintel_background_index_records_removed_total

<p class="subtitle">Lsif index records deleted every 5m</p>

Number of LSIF index records deleted due to expiration or unreachability every 5m

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/worker/worker?viewPanel=100511` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum(increase(src_codeintel_background_index_records_removed_total{job=~"^worker.*"}[5m]))`

</details>

<br />

#### worker: codeintel_background_uploads_purged_total

<p class="subtitle">Lsif upload data bundles deleted every 5m</p>

Number of LSIF upload data bundles purged from the codeintel-db database every 5m

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/worker/worker?viewPanel=100512` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum(increase(src_codeintel_background_uploads_purged_total{job=~"^worker.*"}[5m]))`

</details>

<br />

#### worker: codeintel_background_documentation_search_records_removed_total

<p class="subtitle">Documentation search record records deleted every 5m</p>

Number of documentation search records removed from the codeintel-db database every 5m

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/worker/worker?viewPanel=100513` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum(increase(src_codeintel_background_documentation_search_records_removed_total{job=~"^worker.*"}[5m]))`

</details>

<br />

#### worker: codeintel_background_errors_total

<p class="subtitle">Janitor operation errors every 5m</p>

Number of code intelligence janitor errors every 5m

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/worker/worker?viewPanel=100520` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum(increase(src_codeintel_background_errors_total{job=~"^worker.*"}[5m]))`

</details>

<br />

### Worker: Codeintel: Auto-index scheduler

#### worker: codeintel_index_scheduler_total

<p class="subtitle">Aggregate scheduler operations every 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/worker/worker?viewPanel=100600` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum(increase(src_codeintel_index_scheduler_total{job=~"^worker.*"}[5m]))`

</details>

<br />

#### worker: codeintel_index_scheduler_99th_percentile_duration

<p class="subtitle">99th percentile successful aggregate scheduler operation duration over 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/worker/worker?viewPanel=100601` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `histogram_quantile(0.99, sum  by (le)(rate(src_codeintel_index_scheduler_duration_seconds_bucket{job=~"^worker.*"}[5m])))`

</details>

<br />

#### worker: codeintel_index_scheduler_errors_total

<p class="subtitle">Aggregate scheduler operation errors every 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/worker/worker?viewPanel=100602` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum(increase(src_codeintel_index_scheduler_errors_total{job=~"^worker.*"}[5m]))`

</details>

<br />

#### worker: codeintel_index_scheduler_error_rate

<p class="subtitle">Aggregate scheduler operation error rate over 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/worker/worker?viewPanel=100603` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum(increase(src_codeintel_index_scheduler_errors_total{job=~"^worker.*"}[5m])) / (sum(increase(src_codeintel_index_scheduler_total{job=~"^worker.*"}[5m])) + sum(increase(src_codeintel_index_scheduler_errors_total{job=~"^worker.*"}[5m]))) * 100`

</details>

<br />

#### worker: codeintel_index_scheduler_total

<p class="subtitle">Scheduler operations every 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/worker/worker?viewPanel=100610` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum by (op)(increase(src_codeintel_index_scheduler_total{job=~"^worker.*"}[5m]))`

</details>

<br />

#### worker: codeintel_index_scheduler_99th_percentile_duration

<p class="subtitle">99th percentile successful scheduler operation duration over 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/worker/worker?viewPanel=100611` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `histogram_quantile(0.99, sum  by (le,op)(rate(src_codeintel_index_scheduler_duration_seconds_bucket{job=~"^worker.*"}[5m])))`

</details>

<br />

#### worker: codeintel_index_scheduler_errors_total

<p class="subtitle">Scheduler operation errors every 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/worker/worker?viewPanel=100612` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum by (op)(increase(src_codeintel_index_scheduler_errors_total{job=~"^worker.*"}[5m]))`

</details>

<br />

#### worker: codeintel_index_scheduler_error_rate

<p class="subtitle">Scheduler operation error rate over 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/worker/worker?viewPanel=100613` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum by (op)(increase(src_codeintel_index_scheduler_errors_total{job=~"^worker.*"}[5m])) / (sum by (op)(increase(src_codeintel_index_scheduler_total{job=~"^worker.*"}[5m])) + sum by (op)(increase(src_codeintel_index_scheduler_errors_total{job=~"^worker.*"}[5m]))) * 100`

</details>

<br />

### Worker: Codeintel: Auto-index enqueuer

#### worker: codeintel_autoindex_enqueuer_total

<p class="subtitle">Aggregate enqueuer operations every 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/worker/worker?viewPanel=100700` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum(increase(src_codeintel_autoindex_enqueuer_total{job=~"^worker.*"}[5m]))`

</details>

<br />

#### worker: codeintel_autoindex_enqueuer_99th_percentile_duration

<p class="subtitle">99th percentile successful aggregate enqueuer operation duration over 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/worker/worker?viewPanel=100701` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `histogram_quantile(0.99, sum  by (le)(rate(src_codeintel_autoindex_enqueuer_duration_seconds_bucket{job=~"^worker.*"}[5m])))`

</details>

<br />

#### worker: codeintel_autoindex_enqueuer_errors_total

<p class="subtitle">Aggregate enqueuer operation errors every 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/worker/worker?viewPanel=100702` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum(increase(src_codeintel_autoindex_enqueuer_errors_total{job=~"^worker.*"}[5m]))`

</details>

<br />

#### worker: codeintel_autoindex_enqueuer_error_rate

<p class="subtitle">Aggregate enqueuer operation error rate over 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/worker/worker?viewPanel=100703` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum(increase(src_codeintel_autoindex_enqueuer_errors_total{job=~"^worker.*"}[5m])) / (sum(increase(src_codeintel_autoindex_enqueuer_total{job=~"^worker.*"}[5m])) + sum(increase(src_codeintel_autoindex_enqueuer_errors_total{job=~"^worker.*"}[5m]))) * 100`

</details>

<br />

#### worker: codeintel_autoindex_enqueuer_total

<p class="subtitle">Enqueuer operations every 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/worker/worker?viewPanel=100710` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum by (op)(increase(src_codeintel_autoindex_enqueuer_total{job=~"^worker.*"}[5m]))`

</details>

<br />

#### worker: codeintel_autoindex_enqueuer_99th_percentile_duration

<p class="subtitle">99th percentile successful enqueuer operation duration over 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/worker/worker?viewPanel=100711` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `histogram_quantile(0.99, sum  by (le,op)(rate(src_codeintel_autoindex_enqueuer_duration_seconds_bucket{job=~"^worker.*"}[5m])))`

</details>

<br />

#### worker: codeintel_autoindex_enqueuer_errors_total

<p class="subtitle">Enqueuer operation errors every 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/worker/worker?viewPanel=100712` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum by (op)(increase(src_codeintel_autoindex_enqueuer_errors_total{job=~"^worker.*"}[5m]))`

</details>

<br />

#### worker: codeintel_autoindex_enqueuer_error_rate

<p class="subtitle">Enqueuer operation error rate over 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/worker/worker?viewPanel=100713` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum by (op)(increase(src_codeintel_autoindex_enqueuer_errors_total{job=~"^worker.*"}[5m])) / (sum by (op)(increase(src_codeintel_autoindex_enqueuer_total{job=~"^worker.*"}[5m])) + sum by (op)(increase(src_codeintel_autoindex_enqueuer_errors_total{job=~"^worker.*"}[5m]))) * 100`

</details>

<br />

### Worker: Codeintel: dbstore stats

#### worker: codeintel_dbstore_total

<p class="subtitle">Aggregate store operations every 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/worker/worker?viewPanel=100800` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum(increase(src_codeintel_dbstore_total{job=~"^worker.*"}[5m]))`

</details>

<br />

#### worker: codeintel_dbstore_99th_percentile_duration

<p class="subtitle">99th percentile successful aggregate store operation duration over 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/worker/worker?viewPanel=100801` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `histogram_quantile(0.99, sum  by (le)(rate(src_codeintel_dbstore_duration_seconds_bucket{job=~"^worker.*"}[5m])))`

</details>

<br />

#### worker: codeintel_dbstore_errors_total

<p class="subtitle">Aggregate store operation errors every 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/worker/worker?viewPanel=100802` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum(increase(src_codeintel_dbstore_errors_total{job=~"^worker.*"}[5m]))`

</details>

<br />

#### worker: codeintel_dbstore_error_rate

<p class="subtitle">Aggregate store operation error rate over 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/worker/worker?viewPanel=100803` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum(increase(src_codeintel_dbstore_errors_total{job=~"^worker.*"}[5m])) / (sum(increase(src_codeintel_dbstore_total{job=~"^worker.*"}[5m])) + sum(increase(src_codeintel_dbstore_errors_total{job=~"^worker.*"}[5m]))) * 100`

</details>

<br />

#### worker: codeintel_dbstore_total

<p class="subtitle">Store operations every 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/worker/worker?viewPanel=100810` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum by (op)(increase(src_codeintel_dbstore_total{job=~"^worker.*"}[5m]))`

</details>

<br />

#### worker: codeintel_dbstore_99th_percentile_duration

<p class="subtitle">99th percentile successful store operation duration over 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/worker/worker?viewPanel=100811` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `histogram_quantile(0.99, sum  by (le,op)(rate(src_codeintel_dbstore_duration_seconds_bucket{job=~"^worker.*"}[5m])))`

</details>

<br />

#### worker: codeintel_dbstore_errors_total

<p class="subtitle">Store operation errors every 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/worker/worker?viewPanel=100812` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum by (op)(increase(src_codeintel_dbstore_errors_total{job=~"^worker.*"}[5m]))`

</details>

<br />

#### worker: codeintel_dbstore_error_rate

<p class="subtitle">Store operation error rate over 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/worker/worker?viewPanel=100813` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum by (op)(increase(src_codeintel_dbstore_errors_total{job=~"^worker.*"}[5m])) / (sum by (op)(increase(src_codeintel_dbstore_total{job=~"^worker.*"}[5m])) + sum by (op)(increase(src_codeintel_dbstore_errors_total{job=~"^worker.*"}[5m]))) * 100`

</details>

<br />

### Worker: Codeintel: lsifstore stats

#### worker: codeintel_lsifstore_total

<p class="subtitle">Aggregate store operations every 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/worker/worker?viewPanel=100900` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum(increase(src_codeintel_lsifstore_total{job=~"^worker.*"}[5m]))`

</details>

<br />

#### worker: codeintel_lsifstore_99th_percentile_duration

<p class="subtitle">99th percentile successful aggregate store operation duration over 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/worker/worker?viewPanel=100901` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `histogram_quantile(0.99, sum  by (le)(rate(src_codeintel_lsifstore_duration_seconds_bucket{job=~"^worker.*"}[5m])))`

</details>

<br />

#### worker: codeintel_lsifstore_errors_total

<p class="subtitle">Aggregate store operation errors every 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/worker/worker?viewPanel=100902` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum(increase(src_codeintel_lsifstore_errors_total{job=~"^worker.*"}[5m]))`

</details>

<br />

#### worker: codeintel_lsifstore_error_rate

<p class="subtitle">Aggregate store operation error rate over 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/worker/worker?viewPanel=100903` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum(increase(src_codeintel_lsifstore_errors_total{job=~"^worker.*"}[5m])) / (sum(increase(src_codeintel_lsifstore_total{job=~"^worker.*"}[5m])) + sum(increase(src_codeintel_lsifstore_errors_total{job=~"^worker.*"}[5m]))) * 100`

</details>

<br />

#### worker: codeintel_lsifstore_total

<p class="subtitle">Store operations every 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/worker/worker?viewPanel=100910` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum by (op)(increase(src_codeintel_lsifstore_total{job=~"^worker.*"}[5m]))`

</details>

<br />

#### worker: codeintel_lsifstore_99th_percentile_duration

<p class="subtitle">99th percentile successful store operation duration over 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/worker/worker?viewPanel=100911` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `histogram_quantile(0.99, sum  by (le,op)(rate(src_codeintel_lsifstore_duration_seconds_bucket{job=~"^worker.*"}[5m])))`

</details>

<br />

#### worker: codeintel_lsifstore_errors_total

<p class="subtitle">Store operation errors every 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/worker/worker?viewPanel=100912` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum by (op)(increase(src_codeintel_lsifstore_errors_total{job=~"^worker.*"}[5m]))`

</details>

<br />

#### worker: codeintel_lsifstore_error_rate

<p class="subtitle">Store operation error rate over 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/worker/worker?viewPanel=100913` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum by (op)(increase(src_codeintel_lsifstore_errors_total{job=~"^worker.*"}[5m])) / (sum by (op)(increase(src_codeintel_lsifstore_total{job=~"^worker.*"}[5m])) + sum by (op)(increase(src_codeintel_lsifstore_errors_total{job=~"^worker.*"}[5m]))) * 100`

</details>

<br />

### Worker: Workerutil: lsif_dependency_indexes dbworker/store stats

#### worker: workerutil_dbworker_store_codeintel_dependency_index_total

<p class="subtitle">Store operations every 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/worker/worker?viewPanel=101000` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum(increase(src_workerutil_dbworker_store_codeintel_dependency_index_total{job=~"^worker.*"}[5m]))`

</details>

<br />

#### worker: workerutil_dbworker_store_codeintel_dependency_index_99th_percentile_duration

<p class="subtitle">99th percentile successful store operation duration over 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/worker/worker?viewPanel=101001` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `histogram_quantile(0.99, sum  by (le)(rate(src_workerutil_dbworker_store_codeintel_dependency_index_duration_seconds_bucket{job=~"^worker.*"}[5m])))`

</details>

<br />

#### worker: workerutil_dbworker_store_codeintel_dependency_index_errors_total

<p class="subtitle">Store operation errors every 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/worker/worker?viewPanel=101002` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum(increase(src_workerutil_dbworker_store_codeintel_dependency_index_errors_total{job=~"^worker.*"}[5m]))`

</details>

<br />

#### worker: workerutil_dbworker_store_codeintel_dependency_index_error_rate

<p class="subtitle">Store operation error rate over 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/worker/worker?viewPanel=101003` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum(increase(src_workerutil_dbworker_store_codeintel_dependency_index_errors_total{job=~"^worker.*"}[5m])) / (sum(increase(src_workerutil_dbworker_store_codeintel_dependency_index_total{job=~"^worker.*"}[5m])) + sum(increase(src_workerutil_dbworker_store_codeintel_dependency_index_errors_total{job=~"^worker.*"}[5m]))) * 100`

</details>

<br />

### Worker: Codeintel: gitserver client

#### worker: codeintel_gitserver_total

<p class="subtitle">Aggregate client operations every 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/worker/worker?viewPanel=101100` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum(increase(src_codeintel_gitserver_total{job=~"^worker.*"}[5m]))`

</details>

<br />

#### worker: codeintel_gitserver_99th_percentile_duration

<p class="subtitle">99th percentile successful aggregate client operation duration over 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/worker/worker?viewPanel=101101` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `histogram_quantile(0.99, sum  by (le)(rate(src_codeintel_gitserver_duration_seconds_bucket{job=~"^worker.*"}[5m])))`

</details>

<br />

#### worker: codeintel_gitserver_errors_total

<p class="subtitle">Aggregate client operation errors every 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/worker/worker?viewPanel=101102` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum(increase(src_codeintel_gitserver_errors_total{job=~"^worker.*"}[5m]))`

</details>

<br />

#### worker: codeintel_gitserver_error_rate

<p class="subtitle">Aggregate client operation error rate over 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/worker/worker?viewPanel=101103` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum(increase(src_codeintel_gitserver_errors_total{job=~"^worker.*"}[5m])) / (sum(increase(src_codeintel_gitserver_total{job=~"^worker.*"}[5m])) + sum(increase(src_codeintel_gitserver_errors_total{job=~"^worker.*"}[5m]))) * 100`

</details>

<br />

#### worker: codeintel_gitserver_total

<p class="subtitle">Client operations every 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/worker/worker?viewPanel=101110` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum by (op)(increase(src_codeintel_gitserver_total{job=~"^worker.*"}[5m]))`

</details>

<br />

#### worker: codeintel_gitserver_99th_percentile_duration

<p class="subtitle">99th percentile successful client operation duration over 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/worker/worker?viewPanel=101111` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `histogram_quantile(0.99, sum  by (le,op)(rate(src_codeintel_gitserver_duration_seconds_bucket{job=~"^worker.*"}[5m])))`

</details>

<br />

#### worker: codeintel_gitserver_errors_total

<p class="subtitle">Client operation errors every 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/worker/worker?viewPanel=101112` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum by (op)(increase(src_codeintel_gitserver_errors_total{job=~"^worker.*"}[5m]))`

</details>

<br />

#### worker: codeintel_gitserver_error_rate

<p class="subtitle">Client operation error rate over 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/worker/worker?viewPanel=101113` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum by (op)(increase(src_codeintel_gitserver_errors_total{job=~"^worker.*"}[5m])) / (sum by (op)(increase(src_codeintel_gitserver_total{job=~"^worker.*"}[5m])) + sum by (op)(increase(src_codeintel_gitserver_errors_total{job=~"^worker.*"}[5m]))) * 100`

</details>

<br />

### Worker: Codeintel: Dependency repository insert

#### worker: codeintel_dependency_repos_total

<p class="subtitle">Aggregate insert operations every 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/worker/worker?viewPanel=101200` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum(increase(src_codeintel_dependency_repos_total{job=~"^worker.*"}[5m]))`

</details>

<br />

#### worker: codeintel_dependency_repos_99th_percentile_duration

<p class="subtitle">99th percentile successful aggregate insert operation duration over 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/worker/worker?viewPanel=101201` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `histogram_quantile(0.99, sum  by (le)(rate(src_codeintel_dependency_repos_duration_seconds_bucket{job=~"^worker.*"}[5m])))`

</details>

<br />

#### worker: codeintel_dependency_repos_errors_total

<p class="subtitle">Aggregate insert operation errors every 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/worker/worker?viewPanel=101202` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum(increase(src_codeintel_dependency_repos_errors_total{job=~"^worker.*"}[5m]))`

</details>

<br />

#### worker: codeintel_dependency_repos_error_rate

<p class="subtitle">Aggregate insert operation error rate over 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/worker/worker?viewPanel=101203` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum(increase(src_codeintel_dependency_repos_errors_total{job=~"^worker.*"}[5m])) / (sum(increase(src_codeintel_dependency_repos_total{job=~"^worker.*"}[5m])) + sum(increase(src_codeintel_dependency_repos_errors_total{job=~"^worker.*"}[5m]))) * 100`

</details>

<br />

#### worker: codeintel_dependency_repos_total

<p class="subtitle">Insert operations every 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/worker/worker?viewPanel=101210` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum by (scheme,new)(increase(src_codeintel_dependency_repos_total{job=~"^worker.*"}[5m]))`

</details>

<br />

#### worker: codeintel_dependency_repos_99th_percentile_duration

<p class="subtitle">99th percentile successful insert operation duration over 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/worker/worker?viewPanel=101211` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `histogram_quantile(0.99, sum  by (le,scheme,new)(rate(src_codeintel_dependency_repos_duration_seconds_bucket{job=~"^worker.*"}[5m])))`

</details>

<br />

#### worker: codeintel_dependency_repos_errors_total

<p class="subtitle">Insert operation errors every 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/worker/worker?viewPanel=101212` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum by (scheme,new)(increase(src_codeintel_dependency_repos_errors_total{job=~"^worker.*"}[5m]))`

</details>

<br />

#### worker: codeintel_dependency_repos_error_rate

<p class="subtitle">Insert operation error rate over 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/worker/worker?viewPanel=101213` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum by (scheme,new)(increase(src_codeintel_dependency_repos_errors_total{job=~"^worker.*"}[5m])) / (sum by (scheme,new)(increase(src_codeintel_dependency_repos_total{job=~"^worker.*"}[5m])) + sum by (scheme,new)(increase(src_codeintel_dependency_repos_errors_total{job=~"^worker.*"}[5m]))) * 100`

</details>

<br />

### Worker: Batches: dbstore stats

#### worker: batches_dbstore_total

<p class="subtitle">Aggregate store operations every 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/worker/worker?viewPanel=101300` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Batches team](https://about.sourcegraph.com/handbook/engineering/batches).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum(increase(src_batches_dbstore_total{job=~"^worker.*"}[5m]))`

</details>

<br />

#### worker: batches_dbstore_99th_percentile_duration

<p class="subtitle">99th percentile successful aggregate store operation duration over 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/worker/worker?viewPanel=101301` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Batches team](https://about.sourcegraph.com/handbook/engineering/batches).*</sub>

<details>
<summary>Technical details</summary>

Query: `histogram_quantile(0.99, sum  by (le)(rate(src_batches_dbstore_duration_seconds_bucket{job=~"^worker.*"}[5m])))`

</details>

<br />

#### worker: batches_dbstore_errors_total

<p class="subtitle">Aggregate store operation errors every 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/worker/worker?viewPanel=101302` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Batches team](https://about.sourcegraph.com/handbook/engineering/batches).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum(increase(src_batches_dbstore_errors_total{job=~"^worker.*"}[5m]))`

</details>

<br />

#### worker: batches_dbstore_error_rate

<p class="subtitle">Aggregate store operation error rate over 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/worker/worker?viewPanel=101303` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Batches team](https://about.sourcegraph.com/handbook/engineering/batches).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum(increase(src_batches_dbstore_errors_total{job=~"^worker.*"}[5m])) / (sum(increase(src_batches_dbstore_total{job=~"^worker.*"}[5m])) + sum(increase(src_batches_dbstore_errors_total{job=~"^worker.*"}[5m]))) * 100`

</details>

<br />

#### worker: batches_dbstore_total

<p class="subtitle">Store operations every 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/worker/worker?viewPanel=101310` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Batches team](https://about.sourcegraph.com/handbook/engineering/batches).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum by (op)(increase(src_batches_dbstore_total{job=~"^worker.*"}[5m]))`

</details>

<br />

#### worker: batches_dbstore_99th_percentile_duration

<p class="subtitle">99th percentile successful store operation duration over 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/worker/worker?viewPanel=101311` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Batches team](https://about.sourcegraph.com/handbook/engineering/batches).*</sub>

<details>
<summary>Technical details</summary>

Query: `histogram_quantile(0.99, sum  by (le,op)(rate(src_batches_dbstore_duration_seconds_bucket{job=~"^worker.*"}[5m])))`

</details>

<br />

#### worker: batches_dbstore_errors_total

<p class="subtitle">Store operation errors every 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/worker/worker?viewPanel=101312` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Batches team](https://about.sourcegraph.com/handbook/engineering/batches).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum by (op)(increase(src_batches_dbstore_errors_total{job=~"^worker.*"}[5m]))`

</details>

<br />

#### worker: batches_dbstore_error_rate

<p class="subtitle">Store operation error rate over 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/worker/worker?viewPanel=101313` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Batches team](https://about.sourcegraph.com/handbook/engineering/batches).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum by (op)(increase(src_batches_dbstore_errors_total{job=~"^worker.*"}[5m])) / (sum by (op)(increase(src_batches_dbstore_total{job=~"^worker.*"}[5m])) + sum by (op)(increase(src_batches_dbstore_errors_total{job=~"^worker.*"}[5m]))) * 100`

</details>

<br />

### Worker: Batches: service stats

#### worker: batches_service_total

<p class="subtitle">Aggregate service operations every 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/worker/worker?viewPanel=101400` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Batches team](https://about.sourcegraph.com/handbook/engineering/batches).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum(increase(src_batches_service_total{job=~"^worker.*"}[5m]))`

</details>

<br />

#### worker: batches_service_99th_percentile_duration

<p class="subtitle">99th percentile successful aggregate service operation duration over 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/worker/worker?viewPanel=101401` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Batches team](https://about.sourcegraph.com/handbook/engineering/batches).*</sub>

<details>
<summary>Technical details</summary>

Query: `histogram_quantile(0.99, sum  by (le)(rate(src_batches_service_duration_seconds_bucket{job=~"^worker.*"}[5m])))`

</details>

<br />

#### worker: batches_service_errors_total

<p class="subtitle">Aggregate service operation errors every 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/worker/worker?viewPanel=101402` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Batches team](https://about.sourcegraph.com/handbook/engineering/batches).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum(increase(src_batches_service_errors_total{job=~"^worker.*"}[5m]))`

</details>

<br />

#### worker: batches_service_error_rate

<p class="subtitle">Aggregate service operation error rate over 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/worker/worker?viewPanel=101403` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Batches team](https://about.sourcegraph.com/handbook/engineering/batches).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum(increase(src_batches_service_errors_total{job=~"^worker.*"}[5m])) / (sum(increase(src_batches_service_total{job=~"^worker.*"}[5m])) + sum(increase(src_batches_service_errors_total{job=~"^worker.*"}[5m]))) * 100`

</details>

<br />

#### worker: batches_service_total

<p class="subtitle">Service operations every 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/worker/worker?viewPanel=101410` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Batches team](https://about.sourcegraph.com/handbook/engineering/batches).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum by (op)(increase(src_batches_service_total{job=~"^worker.*"}[5m]))`

</details>

<br />

#### worker: batches_service_99th_percentile_duration

<p class="subtitle">99th percentile successful service operation duration over 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/worker/worker?viewPanel=101411` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Batches team](https://about.sourcegraph.com/handbook/engineering/batches).*</sub>

<details>
<summary>Technical details</summary>

Query: `histogram_quantile(0.99, sum  by (le,op)(rate(src_batches_service_duration_seconds_bucket{job=~"^worker.*"}[5m])))`

</details>

<br />

#### worker: batches_service_errors_total

<p class="subtitle">Service operation errors every 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/worker/worker?viewPanel=101412` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Batches team](https://about.sourcegraph.com/handbook/engineering/batches).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum by (op)(increase(src_batches_service_errors_total{job=~"^worker.*"}[5m]))`

</details>

<br />

#### worker: batches_service_error_rate

<p class="subtitle">Service operation error rate over 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/worker/worker?viewPanel=101413` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Batches team](https://about.sourcegraph.com/handbook/engineering/batches).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum by (op)(increase(src_batches_service_errors_total{job=~"^worker.*"}[5m])) / (sum by (op)(increase(src_batches_service_total{job=~"^worker.*"}[5m])) + sum by (op)(increase(src_batches_service_errors_total{job=~"^worker.*"}[5m]))) * 100`

</details>

<br />

### Worker: Codeintel: lsif_upload record resetter

#### worker: codeintel_background_upload_record_resets_total

<p class="subtitle">Lsif upload records reset to queued state every 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/worker/worker?viewPanel=101500` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum(increase(src_codeintel_background_upload_record_resets_total{job=~"^worker.*"}[5m]))`

</details>

<br />

#### worker: codeintel_background_upload_record_reset_failures_total

<p class="subtitle">Lsif upload records reset to errored state every 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/worker/worker?viewPanel=101501` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum(increase(src_codeintel_background_upload_record_reset_failures_total{job=~"^worker.*"}[5m]))`

</details>

<br />

#### worker: codeintel_background_upload_record_reset_errors_total

<p class="subtitle">Lsif upload operation errors every 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/worker/worker?viewPanel=101502` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum(increase(src_codeintel_background_upload_record_reset_errors_total{job=~"^worker.*"}[5m]))`

</details>

<br />

### Worker: Codeintel: lsif_index record resetter

#### worker: codeintel_background_index_record_resets_total

<p class="subtitle">Lsif index records reset to queued state every 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/worker/worker?viewPanel=101600` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum(increase(src_codeintel_background_index_record_resets_total{job=~"^worker.*"}[5m]))`

</details>

<br />

#### worker: codeintel_background_index_record_reset_failures_total

<p class="subtitle">Lsif index records reset to errored state every 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/worker/worker?viewPanel=101601` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum(increase(src_codeintel_background_index_record_reset_failures_total{job=~"^worker.*"}[5m]))`

</details>

<br />

#### worker: codeintel_background_index_record_reset_errors_total

<p class="subtitle">Lsif index operation errors every 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/worker/worker?viewPanel=101602` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum(increase(src_codeintel_background_index_record_reset_errors_total{job=~"^worker.*"}[5m]))`

</details>

<br />

### Worker: Codeintel: lsif_dependency_index record resetter

#### worker: codeintel_background_dependency_index_record_resets_total

<p class="subtitle">Lsif dependency index records reset to queued state every 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/worker/worker?viewPanel=101700` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum(increase(src_codeintel_background_dependency_index_record_resets_total{job=~"^worker.*"}[5m]))`

</details>

<br />

#### worker: codeintel_background_dependency_index_record_reset_failures_total

<p class="subtitle">Lsif dependency index records reset to errored state every 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/worker/worker?viewPanel=101701` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum(increase(src_codeintel_background_dependency_index_record_reset_failures_total{job=~"^worker.*"}[5m]))`

</details>

<br />

#### worker: codeintel_background_dependency_index_record_reset_errors_total

<p class="subtitle">Lsif dependency index operation errors every 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/worker/worker?viewPanel=101702` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum(increase(src_codeintel_background_dependency_index_record_reset_errors_total{job=~"^worker.*"}[5m]))`

</details>

<br />

### Worker: Codeinsights: Query Runner Queue

#### worker: insights_search_queue_queue_size

<p class="subtitle">Code insights search queue queue size</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/worker/worker?viewPanel=101800` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-insights team](https://about.sourcegraph.com/handbook/engineering/developer-insights/code-insights).*</sub>

<details>
<summary>Technical details</summary>

Query: `max(src_insights_search_queue_total{job=~"^worker.*"})`

</details>

<br />

#### worker: insights_search_queue_queue_growth_rate

<p class="subtitle">Code insights search queue queue growth rate over 30m</p>

This value compares the rate of enqueues against the rate of finished jobs.

	- A value < than 1 indicates that process rate > enqueue rate
	- A value = than 1 indicates that process rate = enqueue rate
	- A value > than 1 indicates that process rate < enqueue rate

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/worker/worker?viewPanel=101801` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-insights team](https://about.sourcegraph.com/handbook/engineering/developer-insights/code-insights).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum(increase(src_insights_search_queue_total{job=~"^worker.*"}[30m])) / sum(increase(src_insights_search_queue_processor_total{job=~"^worker.*"}[30m]))`

</details>

<br />

### Worker: Codeinsights: insights queue processor

#### worker: insights_search_queue_handlers

<p class="subtitle">Handler active handlers</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/worker/worker?viewPanel=101900` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-insights team](https://about.sourcegraph.com/handbook/engineering/developer-insights/code-insights).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum(src_insights_search_queue_processor_handlers{job=~"^worker.*"})`

</details>

<br />

#### worker: insights_search_queue_processor_total

<p class="subtitle">Handler operations every 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/worker/worker?viewPanel=101910` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-insights team](https://about.sourcegraph.com/handbook/engineering/developer-insights/code-insights).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum(increase(src_insights_search_queue_processor_total{job=~"^worker.*"}[5m]))`

</details>

<br />

#### worker: insights_search_queue_processor_99th_percentile_duration

<p class="subtitle">99th percentile successful handler operation duration over 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/worker/worker?viewPanel=101911` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-insights team](https://about.sourcegraph.com/handbook/engineering/developer-insights/code-insights).*</sub>

<details>
<summary>Technical details</summary>

Query: `histogram_quantile(0.99, sum  by (le)(rate(src_insights_search_queue_processor_duration_seconds_bucket{job=~"^worker.*"}[5m])))`

</details>

<br />

#### worker: insights_search_queue_processor_errors_total

<p class="subtitle">Handler operation errors every 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/worker/worker?viewPanel=101912` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-insights team](https://about.sourcegraph.com/handbook/engineering/developer-insights/code-insights).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum(increase(src_insights_search_queue_processor_errors_total{job=~"^worker.*"}[5m]))`

</details>

<br />

#### worker: insights_search_queue_processor_error_rate

<p class="subtitle">Handler operation error rate over 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/worker/worker?viewPanel=101913` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-insights team](https://about.sourcegraph.com/handbook/engineering/developer-insights/code-insights).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum(increase(src_insights_search_queue_processor_errors_total{job=~"^worker.*"}[5m])) / (sum(increase(src_insights_search_queue_processor_total{job=~"^worker.*"}[5m])) + sum(increase(src_insights_search_queue_processor_errors_total{job=~"^worker.*"}[5m]))) * 100`

</details>

<br />

### Worker: Codeinsights: code insights search queue record resetter

#### worker: insights_search_queue_record_resets_total

<p class="subtitle">Insights search queue records reset to queued state every 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/worker/worker?viewPanel=102000` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-insights team](https://about.sourcegraph.com/handbook/engineering/developer-insights/code-insights).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum(increase(src_insights_search_queue_record_resets_total{job=~"^worker.*"}[5m]))`

</details>

<br />

#### worker: insights_search_queue_record_reset_failures_total

<p class="subtitle">Insights search queue records reset to errored state every 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/worker/worker?viewPanel=102001` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-insights team](https://about.sourcegraph.com/handbook/engineering/developer-insights/code-insights).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum(increase(src_insights_search_queue_record_reset_failures_total{job=~"^worker.*"}[5m]))`

</details>

<br />

#### worker: insights_search_queue_record_reset_errors_total

<p class="subtitle">Insights search queue operation errors every 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/worker/worker?viewPanel=102002` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-insights team](https://about.sourcegraph.com/handbook/engineering/developer-insights/code-insights).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum(increase(src_insights_search_queue_record_reset_errors_total{job=~"^worker.*"}[5m]))`

</details>

<br />

### Worker: Codeinsights: dbstore stats

#### worker: workerutil_dbworker_store_insights_query_runner_jobs_store_total

<p class="subtitle">Aggregate store operations every 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/worker/worker?viewPanel=102100` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-insights team](https://about.sourcegraph.com/handbook/engineering/developer-insights/code-insights).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum(increase(src_workerutil_dbworker_store_insights_query_runner_jobs_store_total{job=~"^worker.*"}[5m]))`

</details>

<br />

#### worker: workerutil_dbworker_store_insights_query_runner_jobs_store_99th_percentile_duration

<p class="subtitle">99th percentile successful aggregate store operation duration over 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/worker/worker?viewPanel=102101` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-insights team](https://about.sourcegraph.com/handbook/engineering/developer-insights/code-insights).*</sub>

<details>
<summary>Technical details</summary>

Query: `histogram_quantile(0.99, sum  by (le)(rate(src_workerutil_dbworker_store_insights_query_runner_jobs_store_duration_seconds_bucket{job=~"^worker.*"}[5m])))`

</details>

<br />

#### worker: workerutil_dbworker_store_insights_query_runner_jobs_store_errors_total

<p class="subtitle">Aggregate store operation errors every 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/worker/worker?viewPanel=102102` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-insights team](https://about.sourcegraph.com/handbook/engineering/developer-insights/code-insights).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum(increase(src_workerutil_dbworker_store_insights_query_runner_jobs_store_errors_total{job=~"^worker.*"}[5m]))`

</details>

<br />

#### worker: workerutil_dbworker_store_insights_query_runner_jobs_store_error_rate

<p class="subtitle">Aggregate store operation error rate over 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/worker/worker?viewPanel=102103` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-insights team](https://about.sourcegraph.com/handbook/engineering/developer-insights/code-insights).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum(increase(src_workerutil_dbworker_store_insights_query_runner_jobs_store_errors_total{job=~"^worker.*"}[5m])) / (sum(increase(src_workerutil_dbworker_store_insights_query_runner_jobs_store_total{job=~"^worker.*"}[5m])) + sum(increase(src_workerutil_dbworker_store_insights_query_runner_jobs_store_errors_total{job=~"^worker.*"}[5m]))) * 100`

</details>

<br />

#### worker: workerutil_dbworker_store_insights_query_runner_jobs_store_total

<p class="subtitle">Store operations every 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/worker/worker?viewPanel=102110` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-insights team](https://about.sourcegraph.com/handbook/engineering/developer-insights/code-insights).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum by (op)(increase(src_workerutil_dbworker_store_insights_query_runner_jobs_store_total{job=~"^worker.*"}[5m]))`

</details>

<br />

#### worker: workerutil_dbworker_store_insights_query_runner_jobs_store_99th_percentile_duration

<p class="subtitle">99th percentile successful store operation duration over 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/worker/worker?viewPanel=102111` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-insights team](https://about.sourcegraph.com/handbook/engineering/developer-insights/code-insights).*</sub>

<details>
<summary>Technical details</summary>

Query: `histogram_quantile(0.99, sum  by (le,op)(rate(src_workerutil_dbworker_store_insights_query_runner_jobs_store_duration_seconds_bucket{job=~"^worker.*"}[5m])))`

</details>

<br />

#### worker: workerutil_dbworker_store_insights_query_runner_jobs_store_errors_total

<p class="subtitle">Store operation errors every 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/worker/worker?viewPanel=102112` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-insights team](https://about.sourcegraph.com/handbook/engineering/developer-insights/code-insights).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum by (op)(increase(src_workerutil_dbworker_store_insights_query_runner_jobs_store_errors_total{job=~"^worker.*"}[5m]))`

</details>

<br />

#### worker: workerutil_dbworker_store_insights_query_runner_jobs_store_error_rate

<p class="subtitle">Store operation error rate over 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/worker/worker?viewPanel=102113` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-insights team](https://about.sourcegraph.com/handbook/engineering/developer-insights/code-insights).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum by (op)(increase(src_workerutil_dbworker_store_insights_query_runner_jobs_store_errors_total{job=~"^worker.*"}[5m])) / (sum by (op)(increase(src_workerutil_dbworker_store_insights_query_runner_jobs_store_total{job=~"^worker.*"}[5m])) + sum by (op)(increase(src_workerutil_dbworker_store_insights_query_runner_jobs_store_errors_total{job=~"^worker.*"}[5m]))) * 100`

</details>

<br />

### Worker: Code Insights queue utilization

#### worker: insights_queue_unutilized_size

<p class="subtitle">Insights queue size that is not utilized (not processing)</p>

Any value on this panel indicates code insights is not processing queries from its queue. This observable and alert only fire if there are records in the queue and there have been no dequeue attempts for 30 minutes.

Refer to the [alert solutions reference](./alert_solutions.md#worker-insights-queue-unutilized-size) for 1 alert related to this panel.

To see this panel, visit `/-/debug/grafana/d/worker/worker?viewPanel=102200` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-insights team](https://about.sourcegraph.com/handbook/engineering/developer-insights/code-insights).*</sub>

<details>
<summary>Technical details</summary>

Query: `max(src_insights_search_queue_total{job=~"^worker.*"}) > 0 and on(job) sum by (op)(increase(src_workerutil_dbworker_store_insights_query_runner_jobs_store_total{job=~"^worker.*",op="Dequeue"}[5m])) < 1`

</details>

<br />

### Worker: Internal service requests

#### worker: frontend_internal_api_error_responses

<p class="subtitle">Frontend-internal API error responses every 5m by route</p>

Refer to the [alert solutions reference](./alert_solutions.md#worker-frontend-internal-api-error-responses) for 1 alert related to this panel.

To see this panel, visit `/-/debug/grafana/d/worker/worker?viewPanel=102300` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum by (category)(increase(src_frontend_internal_request_duration_seconds_count{job="worker",code!~"2.."}[5m])) / ignoring(category) group_left sum(increase(src_frontend_internal_request_duration_seconds_count{job="worker"}[5m]))`

</details>

<br />

### Worker: Database connections

#### worker: max_open_conns

<p class="subtitle">Maximum open</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/worker/worker?viewPanel=102400` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum by (app_name, db_name) (src_pgsql_conns_max_open{app_name="worker"})`

</details>

<br />

#### worker: open_conns

<p class="subtitle">Established</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/worker/worker?viewPanel=102401` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum by (app_name, db_name) (src_pgsql_conns_open{app_name="worker"})`

</details>

<br />

#### worker: in_use

<p class="subtitle">Used</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/worker/worker?viewPanel=102410` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum by (app_name, db_name) (src_pgsql_conns_in_use{app_name="worker"})`

</details>

<br />

#### worker: idle

<p class="subtitle">Idle</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/worker/worker?viewPanel=102411` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum by (app_name, db_name) (src_pgsql_conns_idle{app_name="worker"})`

</details>

<br />

#### worker: mean_blocked_seconds_per_conn_request

<p class="subtitle">Mean blocked seconds per conn request</p>

Refer to the [alert solutions reference](./alert_solutions.md#worker-mean-blocked-seconds-per-conn-request) for 2 alerts related to this panel.

To see this panel, visit `/-/debug/grafana/d/worker/worker?viewPanel=102420` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum by (app_name, db_name) (increase(src_pgsql_conns_blocked_seconds{app_name="worker"}[5m])) / sum by (app_name, db_name) (increase(src_pgsql_conns_waited_for{app_name="worker"}[5m]))`

</details>

<br />

#### worker: closed_max_idle

<p class="subtitle">Closed by SetMaxIdleConns</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/worker/worker?viewPanel=102430` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum by (app_name, db_name) (increase(src_pgsql_conns_closed_max_idle{app_name="worker"}[5m]))`

</details>

<br />

#### worker: closed_max_lifetime

<p class="subtitle">Closed by SetConnMaxLifetime</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/worker/worker?viewPanel=102431` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum by (app_name, db_name) (increase(src_pgsql_conns_closed_max_lifetime{app_name="worker"}[5m]))`

</details>

<br />

#### worker: closed_max_idle_time

<p class="subtitle">Closed by SetConnMaxIdleTime</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/worker/worker?viewPanel=102432` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum by (app_name, db_name) (increase(src_pgsql_conns_closed_max_idle_time{app_name="worker"}[5m]))`

</details>

<br />

### Worker: Container monitoring (not available on server)

#### worker: container_missing

<p class="subtitle">Container missing</p>

This value is the number of times a container has not been seen for more than one minute. If you observe this
value change independent of deployment events (such as an upgrade), it could indicate pods are being OOM killed or terminated for some other reasons.

- **Kubernetes:**
	- Determine if the pod was OOM killed using `kubectl describe pod worker` (look for `OOMKilled: true`) and, if so, consider increasing the memory limit in the relevant `Deployment.yaml`.
	- Check the logs before the container restarted to see if there are `panic:` messages or similar using `kubectl logs -p worker`.
- **Docker Compose:**
	- Determine if the pod was OOM killed using `docker inspect -f '{{json .State}}' worker` (look for `"OOMKilled":true`) and, if so, consider increasing the memory limit of the worker container in `docker-compose.yml`.
	- Check the logs before the container restarted to see if there are `panic:` messages or similar using `docker logs worker` (note this will include logs from the previous and currently running container).

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/worker/worker?viewPanel=102500` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `count by(name) ((time() - container_last_seen{name=~"^worker.*"}) > 60)`

</details>

<br />

#### worker: container_cpu_usage

<p class="subtitle">Container cpu usage total (1m average) across all cores by instance</p>

Refer to the [alert solutions reference](./alert_solutions.md#worker-container-cpu-usage) for 1 alert related to this panel.

To see this panel, visit `/-/debug/grafana/d/worker/worker?viewPanel=102501` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `cadvisor_container_cpu_usage_percentage_total{name=~"^worker.*"}`

</details>

<br />

#### worker: container_memory_usage

<p class="subtitle">Container memory usage by instance</p>

Refer to the [alert solutions reference](./alert_solutions.md#worker-container-memory-usage) for 1 alert related to this panel.

To see this panel, visit `/-/debug/grafana/d/worker/worker?viewPanel=102502` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `cadvisor_container_memory_usage_percentage_total{name=~"^worker.*"}`

</details>

<br />

#### worker: fs_io_operations

<p class="subtitle">Filesystem reads and writes rate by instance over 1h</p>

This value indicates the number of filesystem read and write operations by containers of this service.
When extremely high, this can indicate a resource usage problem, or can cause problems with the service itself, especially if high values or spikes correlate with {{CONTAINER_NAME}} issues.

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/worker/worker?viewPanel=102503` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum by(name) (rate(container_fs_reads_total{name=~"^worker.*"}[1h]) + rate(container_fs_writes_total{name=~"^worker.*"}[1h]))`

</details>

<br />

### Worker: Provisioning indicators (not available on server)

#### worker: provisioning_container_cpu_usage_long_term

<p class="subtitle">Container cpu usage total (90th percentile over 1d) across all cores by instance</p>

Refer to the [alert solutions reference](./alert_solutions.md#worker-provisioning-container-cpu-usage-long-term) for 1 alert related to this panel.

To see this panel, visit `/-/debug/grafana/d/worker/worker?viewPanel=102600` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `quantile_over_time(0.9, cadvisor_container_cpu_usage_percentage_total{name=~"^worker.*"}[1d])`

</details>

<br />

#### worker: provisioning_container_memory_usage_long_term

<p class="subtitle">Container memory usage (1d maximum) by instance</p>

Refer to the [alert solutions reference](./alert_solutions.md#worker-provisioning-container-memory-usage-long-term) for 1 alert related to this panel.

To see this panel, visit `/-/debug/grafana/d/worker/worker?viewPanel=102601` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `max_over_time(cadvisor_container_memory_usage_percentage_total{name=~"^worker.*"}[1d])`

</details>

<br />

#### worker: provisioning_container_cpu_usage_short_term

<p class="subtitle">Container cpu usage total (5m maximum) across all cores by instance</p>

Refer to the [alert solutions reference](./alert_solutions.md#worker-provisioning-container-cpu-usage-short-term) for 1 alert related to this panel.

To see this panel, visit `/-/debug/grafana/d/worker/worker?viewPanel=102610` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `max_over_time(cadvisor_container_cpu_usage_percentage_total{name=~"^worker.*"}[5m])`

</details>

<br />

#### worker: provisioning_container_memory_usage_short_term

<p class="subtitle">Container memory usage (5m maximum) by instance</p>

Refer to the [alert solutions reference](./alert_solutions.md#worker-provisioning-container-memory-usage-short-term) for 1 alert related to this panel.

To see this panel, visit `/-/debug/grafana/d/worker/worker?viewPanel=102611` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `max_over_time(cadvisor_container_memory_usage_percentage_total{name=~"^worker.*"}[5m])`

</details>

<br />

### Worker: Golang runtime monitoring

#### worker: go_goroutines

<p class="subtitle">Maximum active goroutines</p>

A high value here indicates a possible goroutine leak.

Refer to the [alert solutions reference](./alert_solutions.md#worker-go-goroutines) for 1 alert related to this panel.

To see this panel, visit `/-/debug/grafana/d/worker/worker?viewPanel=102700` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `max by(instance) (go_goroutines{job=~".*worker"})`

</details>

<br />

#### worker: go_gc_duration_seconds

<p class="subtitle">Maximum go garbage collection duration</p>

Refer to the [alert solutions reference](./alert_solutions.md#worker-go-gc-duration-seconds) for 1 alert related to this panel.

To see this panel, visit `/-/debug/grafana/d/worker/worker?viewPanel=102701` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `max by(instance) (go_gc_duration_seconds{job=~".*worker"})`

</details>

<br />

### Worker: Kubernetes monitoring (only available on Kubernetes)

#### worker: pods_available_percentage

<p class="subtitle">Percentage pods available</p>

Refer to the [alert solutions reference](./alert_solutions.md#worker-pods-available-percentage) for 1 alert related to this panel.

To see this panel, visit `/-/debug/grafana/d/worker/worker?viewPanel=102800` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum by(app) (up{app=~".*worker"}) / count by (app) (up{app=~".*worker"}) * 100`

</details>

<br />

## Repo Updater

<p class="subtitle">Manages interaction with code hosts, instructs Gitserver to update repositories.</p>

To see this dashboard, visit `/-/debug/grafana/d/repo-updater/repo-updater` on your Sourcegraph instance.

### Repo Updater: Repositories

#### repo-updater: syncer_sync_last_time

<p class="subtitle">Time since last sync</p>

A high value here indicates issues synchronizing repo metadata.
If the value is persistently high, make sure all external services have valid tokens.

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/repo-updater/repo-updater?viewPanel=100000` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<details>
<summary>Technical details</summary>

Query: `max(timestamp(vector(time()))) - max(src_repoupdater_syncer_sync_last_time)`

</details>

<br />

#### repo-updater: src_repoupdater_max_sync_backoff

<p class="subtitle">Time since oldest sync</p>

Refer to the [alert solutions reference](./alert_solutions.md#repo-updater-src-repoupdater-max-sync-backoff) for 1 alert related to this panel.

To see this panel, visit `/-/debug/grafana/d/repo-updater/repo-updater?viewPanel=100001` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<details>
<summary>Technical details</summary>

Query: `max(src_repoupdater_max_sync_backoff)`

</details>

<br />

#### repo-updater: src_repoupdater_syncer_sync_errors_total

<p class="subtitle">Site level external service sync error rate</p>

Refer to the [alert solutions reference](./alert_solutions.md#repo-updater-src-repoupdater-syncer-sync-errors-total) for 2 alerts related to this panel.

To see this panel, visit `/-/debug/grafana/d/repo-updater/repo-updater?viewPanel=100002` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<details>
<summary>Technical details</summary>

Query: `max by (family) (rate(src_repoupdater_syncer_sync_errors_total{owner!="user"}[5m]))`

</details>

<br />

#### repo-updater: syncer_sync_start

<p class="subtitle">Repo metadata sync was started</p>

Refer to the [alert solutions reference](./alert_solutions.md#repo-updater-syncer-sync-start) for 1 alert related to this panel.

To see this panel, visit `/-/debug/grafana/d/repo-updater/repo-updater?viewPanel=100010` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<details>
<summary>Technical details</summary>

Query: `max by (family) (rate(src_repoupdater_syncer_start_sync{family="Syncer.SyncExternalService"}[9h0m0s]))`

</details>

<br />

#### repo-updater: syncer_sync_duration

<p class="subtitle">95th repositories sync duration</p>

Refer to the [alert solutions reference](./alert_solutions.md#repo-updater-syncer-sync-duration) for 1 alert related to this panel.

To see this panel, visit `/-/debug/grafana/d/repo-updater/repo-updater?viewPanel=100011` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<details>
<summary>Technical details</summary>

Query: `histogram_quantile(0.95, max by (le, family, success) (rate(src_repoupdater_syncer_sync_duration_seconds_bucket[1m])))`

</details>

<br />

#### repo-updater: source_duration

<p class="subtitle">95th repositories source duration</p>

Refer to the [alert solutions reference](./alert_solutions.md#repo-updater-source-duration) for 1 alert related to this panel.

To see this panel, visit `/-/debug/grafana/d/repo-updater/repo-updater?viewPanel=100012` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<details>
<summary>Technical details</summary>

Query: `histogram_quantile(0.95, max by (le) (rate(src_repoupdater_source_duration_seconds_bucket[1m])))`

</details>

<br />

#### repo-updater: syncer_synced_repos

<p class="subtitle">Repositories synced</p>

Refer to the [alert solutions reference](./alert_solutions.md#repo-updater-syncer-synced-repos) for 1 alert related to this panel.

To see this panel, visit `/-/debug/grafana/d/repo-updater/repo-updater?viewPanel=100020` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<details>
<summary>Technical details</summary>

Query: `max by (state) (rate(src_repoupdater_syncer_synced_repos_total[1m]))`

</details>

<br />

#### repo-updater: sourced_repos

<p class="subtitle">Repositories sourced</p>

Refer to the [alert solutions reference](./alert_solutions.md#repo-updater-sourced-repos) for 1 alert related to this panel.

To see this panel, visit `/-/debug/grafana/d/repo-updater/repo-updater?viewPanel=100021` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<details>
<summary>Technical details</summary>

Query: `max(rate(src_repoupdater_source_repos_total[1m]))`

</details>

<br />

#### repo-updater: user_added_repos

<p class="subtitle">Total number of user added repos</p>

Refer to the [alert solutions reference](./alert_solutions.md#repo-updater-user-added-repos) for 1 alert related to this panel.

To see this panel, visit `/-/debug/grafana/d/repo-updater/repo-updater?viewPanel=100022` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<details>
<summary>Technical details</summary>

Query: `max(src_repoupdater_user_repos_total)`

</details>

<br />

#### repo-updater: purge_failed

<p class="subtitle">Repositories purge failed</p>

Refer to the [alert solutions reference](./alert_solutions.md#repo-updater-purge-failed) for 1 alert related to this panel.

To see this panel, visit `/-/debug/grafana/d/repo-updater/repo-updater?viewPanel=100030` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<details>
<summary>Technical details</summary>

Query: `max(rate(src_repoupdater_purge_failed[1m]))`

</details>

<br />

#### repo-updater: sched_auto_fetch

<p class="subtitle">Repositories scheduled due to hitting a deadline</p>

Refer to the [alert solutions reference](./alert_solutions.md#repo-updater-sched-auto-fetch) for 1 alert related to this panel.

To see this panel, visit `/-/debug/grafana/d/repo-updater/repo-updater?viewPanel=100040` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<details>
<summary>Technical details</summary>

Query: `max(rate(src_repoupdater_sched_auto_fetch[1m]))`

</details>

<br />

#### repo-updater: sched_manual_fetch

<p class="subtitle">Repositories scheduled due to user traffic</p>

Check repo-updater logs if this value is persistently high.
This does not indicate anything if there are no user added code hosts.

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/repo-updater/repo-updater?viewPanel=100041` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<details>
<summary>Technical details</summary>

Query: `max(rate(src_repoupdater_sched_manual_fetch[1m]))`

</details>

<br />

#### repo-updater: sched_known_repos

<p class="subtitle">Repositories managed by the scheduler</p>

Refer to the [alert solutions reference](./alert_solutions.md#repo-updater-sched-known-repos) for 1 alert related to this panel.

To see this panel, visit `/-/debug/grafana/d/repo-updater/repo-updater?viewPanel=100050` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<details>
<summary>Technical details</summary>

Query: `max(src_repoupdater_sched_known_repos)`

</details>

<br />

#### repo-updater: sched_update_queue_length

<p class="subtitle">Rate of growth of update queue length over 5 minutes</p>

Refer to the [alert solutions reference](./alert_solutions.md#repo-updater-sched-update-queue-length) for 1 alert related to this panel.

To see this panel, visit `/-/debug/grafana/d/repo-updater/repo-updater?viewPanel=100051` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<details>
<summary>Technical details</summary>

Query: `max(deriv(src_repoupdater_sched_update_queue_length[5m]))`

</details>

<br />

#### repo-updater: sched_loops

<p class="subtitle">Scheduler loops</p>

Refer to the [alert solutions reference](./alert_solutions.md#repo-updater-sched-loops) for 1 alert related to this panel.

To see this panel, visit `/-/debug/grafana/d/repo-updater/repo-updater?viewPanel=100052` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<details>
<summary>Technical details</summary>

Query: `max(rate(src_repoupdater_sched_loops[1m]))`

</details>

<br />

#### repo-updater: sched_error

<p class="subtitle">Repositories schedule error rate</p>

Refer to the [alert solutions reference](./alert_solutions.md#repo-updater-sched-error) for 1 alert related to this panel.

To see this panel, visit `/-/debug/grafana/d/repo-updater/repo-updater?viewPanel=100060` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<details>
<summary>Technical details</summary>

Query: `max(rate(src_repoupdater_sched_error[1m]))`

</details>

<br />

### Repo Updater: Permissions

#### repo-updater: perms_syncer_perms

<p class="subtitle">Time gap between least and most up to date permissions</p>

Refer to the [alert solutions reference](./alert_solutions.md#repo-updater-perms-syncer-perms) for 1 alert related to this panel.

To see this panel, visit `/-/debug/grafana/d/repo-updater/repo-updater?viewPanel=100100` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<details>
<summary>Technical details</summary>

Query: `max by (type) (src_repoupdater_perms_syncer_perms_gap_seconds)`

</details>

<br />

#### repo-updater: perms_syncer_stale_perms

<p class="subtitle">Number of entities with stale permissions</p>

Refer to the [alert solutions reference](./alert_solutions.md#repo-updater-perms-syncer-stale-perms) for 1 alert related to this panel.

To see this panel, visit `/-/debug/grafana/d/repo-updater/repo-updater?viewPanel=100101` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<details>
<summary>Technical details</summary>

Query: `max by (type) (src_repoupdater_perms_syncer_stale_perms)`

</details>

<br />

#### repo-updater: perms_syncer_no_perms

<p class="subtitle">Number of entities with no permissions</p>

Refer to the [alert solutions reference](./alert_solutions.md#repo-updater-perms-syncer-no-perms) for 1 alert related to this panel.

To see this panel, visit `/-/debug/grafana/d/repo-updater/repo-updater?viewPanel=100110` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<details>
<summary>Technical details</summary>

Query: `max by (type) (src_repoupdater_perms_syncer_no_perms)`

</details>

<br />

#### repo-updater: perms_syncer_outdated_perms

<p class="subtitle">Number of entities with outdated permissions</p>

Refer to the [alert solutions reference](./alert_solutions.md#repo-updater-perms-syncer-outdated-perms) for 1 alert related to this panel.

To see this panel, visit `/-/debug/grafana/d/repo-updater/repo-updater?viewPanel=100111` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<details>
<summary>Technical details</summary>

Query: `max by (type) (src_repoupdater_perms_syncer_outdated_perms)`

</details>

<br />

#### repo-updater: perms_syncer_sync_duration

<p class="subtitle">95th permissions sync duration</p>

Refer to the [alert solutions reference](./alert_solutions.md#repo-updater-perms-syncer-sync-duration) for 1 alert related to this panel.

To see this panel, visit `/-/debug/grafana/d/repo-updater/repo-updater?viewPanel=100120` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<details>
<summary>Technical details</summary>

Query: `histogram_quantile(0.95, max by (le, type) (rate(src_repoupdater_perms_syncer_sync_duration_seconds_bucket[1m])))`

</details>

<br />

#### repo-updater: perms_syncer_queue_size

<p class="subtitle">Permissions sync queued items</p>

Refer to the [alert solutions reference](./alert_solutions.md#repo-updater-perms-syncer-queue-size) for 1 alert related to this panel.

To see this panel, visit `/-/debug/grafana/d/repo-updater/repo-updater?viewPanel=100121` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<details>
<summary>Technical details</summary>

Query: `max(src_repoupdater_perms_syncer_queue_size)`

</details>

<br />

#### repo-updater: perms_syncer_sync_errors

<p class="subtitle">Permissions sync error rate</p>

Refer to the [alert solutions reference](./alert_solutions.md#repo-updater-perms-syncer-sync-errors) for 1 alert related to this panel.

To see this panel, visit `/-/debug/grafana/d/repo-updater/repo-updater?viewPanel=100130` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<details>
<summary>Technical details</summary>

Query: `max by (type) (ceil(rate(src_repoupdater_perms_syncer_sync_errors_total[1m])))`

</details>

<br />

### Repo Updater: External services

#### repo-updater: src_repoupdater_external_services_total

<p class="subtitle">The total number of external services</p>

Refer to the [alert solutions reference](./alert_solutions.md#repo-updater-src-repoupdater-external-services-total) for 1 alert related to this panel.

To see this panel, visit `/-/debug/grafana/d/repo-updater/repo-updater?viewPanel=100200` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<details>
<summary>Technical details</summary>

Query: `max(src_repoupdater_external_services_total)`

</details>

<br />

#### repo-updater: src_repoupdater_user_external_services_total

<p class="subtitle">The total number of user added external services</p>

Refer to the [alert solutions reference](./alert_solutions.md#repo-updater-src-repoupdater-user-external-services-total) for 1 alert related to this panel.

To see this panel, visit `/-/debug/grafana/d/repo-updater/repo-updater?viewPanel=100201` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<details>
<summary>Technical details</summary>

Query: `max(src_repoupdater_user_external_services_total)`

</details>

<br />

#### repo-updater: repoupdater_queued_sync_jobs_total

<p class="subtitle">The total number of queued sync jobs</p>

Refer to the [alert solutions reference](./alert_solutions.md#repo-updater-repoupdater-queued-sync-jobs-total) for 1 alert related to this panel.

To see this panel, visit `/-/debug/grafana/d/repo-updater/repo-updater?viewPanel=100210` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<details>
<summary>Technical details</summary>

Query: `max(src_repoupdater_queued_sync_jobs_total)`

</details>

<br />

#### repo-updater: repoupdater_completed_sync_jobs_total

<p class="subtitle">The total number of completed sync jobs</p>

Refer to the [alert solutions reference](./alert_solutions.md#repo-updater-repoupdater-completed-sync-jobs-total) for 1 alert related to this panel.

To see this panel, visit `/-/debug/grafana/d/repo-updater/repo-updater?viewPanel=100211` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<details>
<summary>Technical details</summary>

Query: `max(src_repoupdater_completed_sync_jobs_total)`

</details>

<br />

#### repo-updater: repoupdater_errored_sync_jobs_percentage

<p class="subtitle">The percentage of external services that have failed their most recent sync</p>

Refer to the [alert solutions reference](./alert_solutions.md#repo-updater-repoupdater-errored-sync-jobs-percentage) for 1 alert related to this panel.

To see this panel, visit `/-/debug/grafana/d/repo-updater/repo-updater?viewPanel=100212` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<details>
<summary>Technical details</summary>

Query: `max(src_repoupdater_errored_sync_jobs_percentage)`

</details>

<br />

#### repo-updater: github_graphql_rate_limit_remaining

<p class="subtitle">Remaining calls to GitHub graphql API before hitting the rate limit</p>

Refer to the [alert solutions reference](./alert_solutions.md#repo-updater-github-graphql-rate-limit-remaining) for 1 alert related to this panel.

To see this panel, visit `/-/debug/grafana/d/repo-updater/repo-updater?viewPanel=100220` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<details>
<summary>Technical details</summary>

Query: `max by (name) (src_github_rate_limit_remaining_v2{resource="graphql"})`

</details>

<br />

#### repo-updater: github_rest_rate_limit_remaining

<p class="subtitle">Remaining calls to GitHub rest API before hitting the rate limit</p>

Refer to the [alert solutions reference](./alert_solutions.md#repo-updater-github-rest-rate-limit-remaining) for 1 alert related to this panel.

To see this panel, visit `/-/debug/grafana/d/repo-updater/repo-updater?viewPanel=100221` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<details>
<summary>Technical details</summary>

Query: `max by (name) (src_github_rate_limit_remaining_v2{resource="rest"})`

</details>

<br />

#### repo-updater: github_search_rate_limit_remaining

<p class="subtitle">Remaining calls to GitHub search API before hitting the rate limit</p>

Refer to the [alert solutions reference](./alert_solutions.md#repo-updater-github-search-rate-limit-remaining) for 1 alert related to this panel.

To see this panel, visit `/-/debug/grafana/d/repo-updater/repo-updater?viewPanel=100222` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<details>
<summary>Technical details</summary>

Query: `max by (name) (src_github_rate_limit_remaining_v2{resource="search"})`

</details>

<br />

#### repo-updater: github_graphql_rate_limit_wait_duration

<p class="subtitle">Time spent waiting for the GitHub graphql API rate limiter</p>

Indicates how long we`re waiting on the rate limit once it has been exceeded

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/repo-updater/repo-updater?viewPanel=100230` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<details>
<summary>Technical details</summary>

Query: `max by(name) (rate(src_github_rate_limit_wait_duration_seconds{resource="graphql"}[5m]))`

</details>

<br />

#### repo-updater: github_rest_rate_limit_wait_duration

<p class="subtitle">Time spent waiting for the GitHub rest API rate limiter</p>

Indicates how long we`re waiting on the rate limit once it has been exceeded

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/repo-updater/repo-updater?viewPanel=100231` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<details>
<summary>Technical details</summary>

Query: `max by(name) (rate(src_github_rate_limit_wait_duration_seconds{resource="rest"}[5m]))`

</details>

<br />

#### repo-updater: github_search_rate_limit_wait_duration

<p class="subtitle">Time spent waiting for the GitHub search API rate limiter</p>

Indicates how long we`re waiting on the rate limit once it has been exceeded

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/repo-updater/repo-updater?viewPanel=100232` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<details>
<summary>Technical details</summary>

Query: `max by(name) (rate(src_github_rate_limit_wait_duration_seconds{resource="search"}[5m]))`

</details>

<br />

#### repo-updater: gitlab_rest_rate_limit_remaining

<p class="subtitle">Remaining calls to GitLab rest API before hitting the rate limit</p>

Refer to the [alert solutions reference](./alert_solutions.md#repo-updater-gitlab-rest-rate-limit-remaining) for 1 alert related to this panel.

To see this panel, visit `/-/debug/grafana/d/repo-updater/repo-updater?viewPanel=100240` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<details>
<summary>Technical details</summary>

Query: `max by (name) (src_gitlab_rate_limit_remaining{resource="rest"})`

</details>

<br />

#### repo-updater: gitlab_rest_rate_limit_wait_duration

<p class="subtitle">Time spent waiting for the GitLab rest API rate limiter</p>

Indicates how long we`re waiting on the rate limit once it has been exceeded

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/repo-updater/repo-updater?viewPanel=100241` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<details>
<summary>Technical details</summary>

Query: `max by(name) (rate(src_gitlab_rate_limit_wait_duration_seconds{resource="rest"}[5m]))`

</details>

<br />

### Repo Updater: Batches: dbstore stats

#### repo-updater: batches_dbstore_total

<p class="subtitle">Aggregate store operations every 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/repo-updater/repo-updater?viewPanel=100300` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Batches team](https://about.sourcegraph.com/handbook/engineering/batches).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum(increase(src_batches_dbstore_total{job=~"^repo-updater.*"}[5m]))`

</details>

<br />

#### repo-updater: batches_dbstore_99th_percentile_duration

<p class="subtitle">99th percentile successful aggregate store operation duration over 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/repo-updater/repo-updater?viewPanel=100301` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Batches team](https://about.sourcegraph.com/handbook/engineering/batches).*</sub>

<details>
<summary>Technical details</summary>

Query: `histogram_quantile(0.99, sum  by (le)(rate(src_batches_dbstore_duration_seconds_bucket{job=~"^repo-updater.*"}[5m])))`

</details>

<br />

#### repo-updater: batches_dbstore_errors_total

<p class="subtitle">Aggregate store operation errors every 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/repo-updater/repo-updater?viewPanel=100302` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Batches team](https://about.sourcegraph.com/handbook/engineering/batches).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum(increase(src_batches_dbstore_errors_total{job=~"^repo-updater.*"}[5m]))`

</details>

<br />

#### repo-updater: batches_dbstore_error_rate

<p class="subtitle">Aggregate store operation error rate over 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/repo-updater/repo-updater?viewPanel=100303` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Batches team](https://about.sourcegraph.com/handbook/engineering/batches).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum(increase(src_batches_dbstore_errors_total{job=~"^repo-updater.*"}[5m])) / (sum(increase(src_batches_dbstore_total{job=~"^repo-updater.*"}[5m])) + sum(increase(src_batches_dbstore_errors_total{job=~"^repo-updater.*"}[5m]))) * 100`

</details>

<br />

#### repo-updater: batches_dbstore_total

<p class="subtitle">Store operations every 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/repo-updater/repo-updater?viewPanel=100310` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Batches team](https://about.sourcegraph.com/handbook/engineering/batches).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum by (op)(increase(src_batches_dbstore_total{job=~"^repo-updater.*"}[5m]))`

</details>

<br />

#### repo-updater: batches_dbstore_99th_percentile_duration

<p class="subtitle">99th percentile successful store operation duration over 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/repo-updater/repo-updater?viewPanel=100311` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Batches team](https://about.sourcegraph.com/handbook/engineering/batches).*</sub>

<details>
<summary>Technical details</summary>

Query: `histogram_quantile(0.99, sum  by (le,op)(rate(src_batches_dbstore_duration_seconds_bucket{job=~"^repo-updater.*"}[5m])))`

</details>

<br />

#### repo-updater: batches_dbstore_errors_total

<p class="subtitle">Store operation errors every 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/repo-updater/repo-updater?viewPanel=100312` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Batches team](https://about.sourcegraph.com/handbook/engineering/batches).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum by (op)(increase(src_batches_dbstore_errors_total{job=~"^repo-updater.*"}[5m]))`

</details>

<br />

#### repo-updater: batches_dbstore_error_rate

<p class="subtitle">Store operation error rate over 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/repo-updater/repo-updater?viewPanel=100313` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Batches team](https://about.sourcegraph.com/handbook/engineering/batches).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum by (op)(increase(src_batches_dbstore_errors_total{job=~"^repo-updater.*"}[5m])) / (sum by (op)(increase(src_batches_dbstore_total{job=~"^repo-updater.*"}[5m])) + sum by (op)(increase(src_batches_dbstore_errors_total{job=~"^repo-updater.*"}[5m]))) * 100`

</details>

<br />

### Repo Updater: Batches: service stats

#### repo-updater: batches_service_total

<p class="subtitle">Aggregate service operations every 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/repo-updater/repo-updater?viewPanel=100400` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Batches team](https://about.sourcegraph.com/handbook/engineering/batches).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum(increase(src_batches_service_total{job=~"^repo-updater.*"}[5m]))`

</details>

<br />

#### repo-updater: batches_service_99th_percentile_duration

<p class="subtitle">99th percentile successful aggregate service operation duration over 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/repo-updater/repo-updater?viewPanel=100401` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Batches team](https://about.sourcegraph.com/handbook/engineering/batches).*</sub>

<details>
<summary>Technical details</summary>

Query: `histogram_quantile(0.99, sum  by (le)(rate(src_batches_service_duration_seconds_bucket{job=~"^repo-updater.*"}[5m])))`

</details>

<br />

#### repo-updater: batches_service_errors_total

<p class="subtitle">Aggregate service operation errors every 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/repo-updater/repo-updater?viewPanel=100402` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Batches team](https://about.sourcegraph.com/handbook/engineering/batches).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum(increase(src_batches_service_errors_total{job=~"^repo-updater.*"}[5m]))`

</details>

<br />

#### repo-updater: batches_service_error_rate

<p class="subtitle">Aggregate service operation error rate over 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/repo-updater/repo-updater?viewPanel=100403` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Batches team](https://about.sourcegraph.com/handbook/engineering/batches).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum(increase(src_batches_service_errors_total{job=~"^repo-updater.*"}[5m])) / (sum(increase(src_batches_service_total{job=~"^repo-updater.*"}[5m])) + sum(increase(src_batches_service_errors_total{job=~"^repo-updater.*"}[5m]))) * 100`

</details>

<br />

#### repo-updater: batches_service_total

<p class="subtitle">Service operations every 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/repo-updater/repo-updater?viewPanel=100410` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Batches team](https://about.sourcegraph.com/handbook/engineering/batches).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum by (op)(increase(src_batches_service_total{job=~"^repo-updater.*"}[5m]))`

</details>

<br />

#### repo-updater: batches_service_99th_percentile_duration

<p class="subtitle">99th percentile successful service operation duration over 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/repo-updater/repo-updater?viewPanel=100411` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Batches team](https://about.sourcegraph.com/handbook/engineering/batches).*</sub>

<details>
<summary>Technical details</summary>

Query: `histogram_quantile(0.99, sum  by (le,op)(rate(src_batches_service_duration_seconds_bucket{job=~"^repo-updater.*"}[5m])))`

</details>

<br />

#### repo-updater: batches_service_errors_total

<p class="subtitle">Service operation errors every 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/repo-updater/repo-updater?viewPanel=100412` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Batches team](https://about.sourcegraph.com/handbook/engineering/batches).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum by (op)(increase(src_batches_service_errors_total{job=~"^repo-updater.*"}[5m]))`

</details>

<br />

#### repo-updater: batches_service_error_rate

<p class="subtitle">Service operation error rate over 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/repo-updater/repo-updater?viewPanel=100413` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Batches team](https://about.sourcegraph.com/handbook/engineering/batches).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum by (op)(increase(src_batches_service_errors_total{job=~"^repo-updater.*"}[5m])) / (sum by (op)(increase(src_batches_service_total{job=~"^repo-updater.*"}[5m])) + sum by (op)(increase(src_batches_service_errors_total{job=~"^repo-updater.*"}[5m]))) * 100`

</details>

<br />

### Repo Updater: Codeintel: Coursier invocation stats

#### repo-updater: codeintel_coursier_total

<p class="subtitle">Aggregate invocations operations every 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/repo-updater/repo-updater?viewPanel=100500` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum(increase(src_codeintel_coursier_total{op!="RunCommand",job=~"^repo-updater.*"}[5m]))`

</details>

<br />

#### repo-updater: codeintel_coursier_99th_percentile_duration

<p class="subtitle">99th percentile successful aggregate invocations operation duration over 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/repo-updater/repo-updater?viewPanel=100501` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `histogram_quantile(0.99, sum  by (le)(rate(src_codeintel_coursier_duration_seconds_bucket{op!="RunCommand",job=~"^repo-updater.*"}[5m])))`

</details>

<br />

#### repo-updater: codeintel_coursier_errors_total

<p class="subtitle">Aggregate invocations operation errors every 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/repo-updater/repo-updater?viewPanel=100502` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum(increase(src_codeintel_coursier_errors_total{op!="RunCommand",job=~"^repo-updater.*"}[5m]))`

</details>

<br />

#### repo-updater: codeintel_coursier_error_rate

<p class="subtitle">Aggregate invocations operation error rate over 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/repo-updater/repo-updater?viewPanel=100503` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum(increase(src_codeintel_coursier_errors_total{op!="RunCommand",job=~"^repo-updater.*"}[5m])) / (sum(increase(src_codeintel_coursier_total{op!="RunCommand",job=~"^repo-updater.*"}[5m])) + sum(increase(src_codeintel_coursier_errors_total{op!="RunCommand",job=~"^repo-updater.*"}[5m]))) * 100`

</details>

<br />

#### repo-updater: codeintel_coursier_total

<p class="subtitle">Invocations operations every 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/repo-updater/repo-updater?viewPanel=100510` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum by (op)(increase(src_codeintel_coursier_total{op!="RunCommand",job=~"^repo-updater.*"}[5m]))`

</details>

<br />

#### repo-updater: codeintel_coursier_99th_percentile_duration

<p class="subtitle">99th percentile successful invocations operation duration over 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/repo-updater/repo-updater?viewPanel=100511` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `histogram_quantile(0.99, sum  by (le,op)(rate(src_codeintel_coursier_duration_seconds_bucket{op!="RunCommand",job=~"^repo-updater.*"}[5m])))`

</details>

<br />

#### repo-updater: codeintel_coursier_errors_total

<p class="subtitle">Invocations operation errors every 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/repo-updater/repo-updater?viewPanel=100512` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum by (op)(increase(src_codeintel_coursier_errors_total{op!="RunCommand",job=~"^repo-updater.*"}[5m]))`

</details>

<br />

#### repo-updater: codeintel_coursier_error_rate

<p class="subtitle">Invocations operation error rate over 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/repo-updater/repo-updater?viewPanel=100513` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum by (op)(increase(src_codeintel_coursier_errors_total{op!="RunCommand",job=~"^repo-updater.*"}[5m])) / (sum by (op)(increase(src_codeintel_coursier_total{op!="RunCommand",job=~"^repo-updater.*"}[5m])) + sum by (op)(increase(src_codeintel_coursier_errors_total{op!="RunCommand",job=~"^repo-updater.*"}[5m]))) * 100`

</details>

<br />

### Repo Updater: Internal service requests

#### repo-updater: frontend_internal_api_error_responses

<p class="subtitle">Frontend-internal API error responses every 5m by route</p>

Refer to the [alert solutions reference](./alert_solutions.md#repo-updater-frontend-internal-api-error-responses) for 1 alert related to this panel.

To see this panel, visit `/-/debug/grafana/d/repo-updater/repo-updater?viewPanel=100600` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum by (category)(increase(src_frontend_internal_request_duration_seconds_count{job="repo-updater",code!~"2.."}[5m])) / ignoring(category) group_left sum(increase(src_frontend_internal_request_duration_seconds_count{job="repo-updater"}[5m]))`

</details>

<br />

### Repo Updater: Database connections

#### repo-updater: max_open_conns

<p class="subtitle">Maximum open</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/repo-updater/repo-updater?viewPanel=100700` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum by (app_name, db_name) (src_pgsql_conns_max_open{app_name="repo-updater"})`

</details>

<br />

#### repo-updater: open_conns

<p class="subtitle">Established</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/repo-updater/repo-updater?viewPanel=100701` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum by (app_name, db_name) (src_pgsql_conns_open{app_name="repo-updater"})`

</details>

<br />

#### repo-updater: in_use

<p class="subtitle">Used</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/repo-updater/repo-updater?viewPanel=100710` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum by (app_name, db_name) (src_pgsql_conns_in_use{app_name="repo-updater"})`

</details>

<br />

#### repo-updater: idle

<p class="subtitle">Idle</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/repo-updater/repo-updater?viewPanel=100711` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum by (app_name, db_name) (src_pgsql_conns_idle{app_name="repo-updater"})`

</details>

<br />

#### repo-updater: mean_blocked_seconds_per_conn_request

<p class="subtitle">Mean blocked seconds per conn request</p>

Refer to the [alert solutions reference](./alert_solutions.md#repo-updater-mean-blocked-seconds-per-conn-request) for 2 alerts related to this panel.

To see this panel, visit `/-/debug/grafana/d/repo-updater/repo-updater?viewPanel=100720` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum by (app_name, db_name) (increase(src_pgsql_conns_blocked_seconds{app_name="repo-updater"}[5m])) / sum by (app_name, db_name) (increase(src_pgsql_conns_waited_for{app_name="repo-updater"}[5m]))`

</details>

<br />

#### repo-updater: closed_max_idle

<p class="subtitle">Closed by SetMaxIdleConns</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/repo-updater/repo-updater?viewPanel=100730` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum by (app_name, db_name) (increase(src_pgsql_conns_closed_max_idle{app_name="repo-updater"}[5m]))`

</details>

<br />

#### repo-updater: closed_max_lifetime

<p class="subtitle">Closed by SetConnMaxLifetime</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/repo-updater/repo-updater?viewPanel=100731` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum by (app_name, db_name) (increase(src_pgsql_conns_closed_max_lifetime{app_name="repo-updater"}[5m]))`

</details>

<br />

#### repo-updater: closed_max_idle_time

<p class="subtitle">Closed by SetConnMaxIdleTime</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/repo-updater/repo-updater?viewPanel=100732` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum by (app_name, db_name) (increase(src_pgsql_conns_closed_max_idle_time{app_name="repo-updater"}[5m]))`

</details>

<br />

### Repo Updater: Container monitoring (not available on server)

#### repo-updater: container_missing

<p class="subtitle">Container missing</p>

This value is the number of times a container has not been seen for more than one minute. If you observe this
value change independent of deployment events (such as an upgrade), it could indicate pods are being OOM killed or terminated for some other reasons.

- **Kubernetes:**
	- Determine if the pod was OOM killed using `kubectl describe pod repo-updater` (look for `OOMKilled: true`) and, if so, consider increasing the memory limit in the relevant `Deployment.yaml`.
	- Check the logs before the container restarted to see if there are `panic:` messages or similar using `kubectl logs -p repo-updater`.
- **Docker Compose:**
	- Determine if the pod was OOM killed using `docker inspect -f '{{json .State}}' repo-updater` (look for `"OOMKilled":true`) and, if so, consider increasing the memory limit of the repo-updater container in `docker-compose.yml`.
	- Check the logs before the container restarted to see if there are `panic:` messages or similar using `docker logs repo-updater` (note this will include logs from the previous and currently running container).

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/repo-updater/repo-updater?viewPanel=100800` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<details>
<summary>Technical details</summary>

Query: `count by(name) ((time() - container_last_seen{name=~"^repo-updater.*"}) > 60)`

</details>

<br />

#### repo-updater: container_cpu_usage

<p class="subtitle">Container cpu usage total (1m average) across all cores by instance</p>

Refer to the [alert solutions reference](./alert_solutions.md#repo-updater-container-cpu-usage) for 1 alert related to this panel.

To see this panel, visit `/-/debug/grafana/d/repo-updater/repo-updater?viewPanel=100801` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<details>
<summary>Technical details</summary>

Query: `cadvisor_container_cpu_usage_percentage_total{name=~"^repo-updater.*"}`

</details>

<br />

#### repo-updater: container_memory_usage

<p class="subtitle">Container memory usage by instance</p>

Refer to the [alert solutions reference](./alert_solutions.md#repo-updater-container-memory-usage) for 1 alert related to this panel.

To see this panel, visit `/-/debug/grafana/d/repo-updater/repo-updater?viewPanel=100802` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<details>
<summary>Technical details</summary>

Query: `cadvisor_container_memory_usage_percentage_total{name=~"^repo-updater.*"}`

</details>

<br />

#### repo-updater: fs_io_operations

<p class="subtitle">Filesystem reads and writes rate by instance over 1h</p>

This value indicates the number of filesystem read and write operations by containers of this service.
When extremely high, this can indicate a resource usage problem, or can cause problems with the service itself, especially if high values or spikes correlate with {{CONTAINER_NAME}} issues.

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/repo-updater/repo-updater?viewPanel=100803` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum by(name) (rate(container_fs_reads_total{name=~"^repo-updater.*"}[1h]) + rate(container_fs_writes_total{name=~"^repo-updater.*"}[1h]))`

</details>

<br />

### Repo Updater: Provisioning indicators (not available on server)

#### repo-updater: provisioning_container_cpu_usage_long_term

<p class="subtitle">Container cpu usage total (90th percentile over 1d) across all cores by instance</p>

Refer to the [alert solutions reference](./alert_solutions.md#repo-updater-provisioning-container-cpu-usage-long-term) for 1 alert related to this panel.

To see this panel, visit `/-/debug/grafana/d/repo-updater/repo-updater?viewPanel=100900` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<details>
<summary>Technical details</summary>

Query: `quantile_over_time(0.9, cadvisor_container_cpu_usage_percentage_total{name=~"^repo-updater.*"}[1d])`

</details>

<br />

#### repo-updater: provisioning_container_memory_usage_long_term

<p class="subtitle">Container memory usage (1d maximum) by instance</p>

Refer to the [alert solutions reference](./alert_solutions.md#repo-updater-provisioning-container-memory-usage-long-term) for 1 alert related to this panel.

To see this panel, visit `/-/debug/grafana/d/repo-updater/repo-updater?viewPanel=100901` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<details>
<summary>Technical details</summary>

Query: `max_over_time(cadvisor_container_memory_usage_percentage_total{name=~"^repo-updater.*"}[1d])`

</details>

<br />

#### repo-updater: provisioning_container_cpu_usage_short_term

<p class="subtitle">Container cpu usage total (5m maximum) across all cores by instance</p>

Refer to the [alert solutions reference](./alert_solutions.md#repo-updater-provisioning-container-cpu-usage-short-term) for 1 alert related to this panel.

To see this panel, visit `/-/debug/grafana/d/repo-updater/repo-updater?viewPanel=100910` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<details>
<summary>Technical details</summary>

Query: `max_over_time(cadvisor_container_cpu_usage_percentage_total{name=~"^repo-updater.*"}[5m])`

</details>

<br />

#### repo-updater: provisioning_container_memory_usage_short_term

<p class="subtitle">Container memory usage (5m maximum) by instance</p>

Refer to the [alert solutions reference](./alert_solutions.md#repo-updater-provisioning-container-memory-usage-short-term) for 1 alert related to this panel.

To see this panel, visit `/-/debug/grafana/d/repo-updater/repo-updater?viewPanel=100911` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<details>
<summary>Technical details</summary>

Query: `max_over_time(cadvisor_container_memory_usage_percentage_total{name=~"^repo-updater.*"}[5m])`

</details>

<br />

### Repo Updater: Golang runtime monitoring

#### repo-updater: go_goroutines

<p class="subtitle">Maximum active goroutines</p>

A high value here indicates a possible goroutine leak.

Refer to the [alert solutions reference](./alert_solutions.md#repo-updater-go-goroutines) for 1 alert related to this panel.

To see this panel, visit `/-/debug/grafana/d/repo-updater/repo-updater?viewPanel=101000` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<details>
<summary>Technical details</summary>

Query: `max by(instance) (go_goroutines{job=~".*repo-updater"})`

</details>

<br />

#### repo-updater: go_gc_duration_seconds

<p class="subtitle">Maximum go garbage collection duration</p>

Refer to the [alert solutions reference](./alert_solutions.md#repo-updater-go-gc-duration-seconds) for 1 alert related to this panel.

To see this panel, visit `/-/debug/grafana/d/repo-updater/repo-updater?viewPanel=101001` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<details>
<summary>Technical details</summary>

Query: `max by(instance) (go_gc_duration_seconds{job=~".*repo-updater"})`

</details>

<br />

### Repo Updater: Kubernetes monitoring (only available on Kubernetes)

#### repo-updater: pods_available_percentage

<p class="subtitle">Percentage pods available</p>

Refer to the [alert solutions reference](./alert_solutions.md#repo-updater-pods-available-percentage) for 1 alert related to this panel.

To see this panel, visit `/-/debug/grafana/d/repo-updater/repo-updater?viewPanel=101100` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum by(app) (up{app=~".*repo-updater"}) / count by (app) (up{app=~".*repo-updater"}) * 100`

</details>

<br />

## Searcher

<p class="subtitle">Performs unindexed searches (diff and commit search, text search for unindexed branches).</p>

To see this dashboard, visit `/-/debug/grafana/d/searcher/searcher` on your Sourcegraph instance.

#### searcher: unindexed_search_request_errors

<p class="subtitle">Unindexed search request errors every 5m by code</p>

Refer to the [alert solutions reference](./alert_solutions.md#searcher-unindexed-search-request-errors) for 1 alert related to this panel.

To see this panel, visit `/-/debug/grafana/d/searcher/searcher?viewPanel=100000` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Search-core team](https://about.sourcegraph.com/handbook/engineering/search/core).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum by (code)(increase(searcher_service_request_total{code!="200",code!="canceled"}[5m])) / ignoring(code) group_left sum(increase(searcher_service_request_total[5m])) * 100`

</details>

<br />

#### searcher: replica_traffic

<p class="subtitle">Requests per second over 10m</p>

Refer to the [alert solutions reference](./alert_solutions.md#searcher-replica-traffic) for 1 alert related to this panel.

To see this panel, visit `/-/debug/grafana/d/searcher/searcher?viewPanel=100001` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Search-core team](https://about.sourcegraph.com/handbook/engineering/search/core).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum by(instance) (rate(searcher_service_request_total[10m]))`

</details>

<br />

### Searcher: Internal service requests

#### searcher: frontend_internal_api_error_responses

<p class="subtitle">Frontend-internal API error responses every 5m by route</p>

Refer to the [alert solutions reference](./alert_solutions.md#searcher-frontend-internal-api-error-responses) for 1 alert related to this panel.

To see this panel, visit `/-/debug/grafana/d/searcher/searcher?viewPanel=100100` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Search-core team](https://about.sourcegraph.com/handbook/engineering/search/core).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum by (category)(increase(src_frontend_internal_request_duration_seconds_count{job="searcher",code!~"2.."}[5m])) / ignoring(category) group_left sum(increase(src_frontend_internal_request_duration_seconds_count{job="searcher"}[5m]))`

</details>

<br />

### Searcher: Container monitoring (not available on server)

#### searcher: container_missing

<p class="subtitle">Container missing</p>

This value is the number of times a container has not been seen for more than one minute. If you observe this
value change independent of deployment events (such as an upgrade), it could indicate pods are being OOM killed or terminated for some other reasons.

- **Kubernetes:**
	- Determine if the pod was OOM killed using `kubectl describe pod searcher` (look for `OOMKilled: true`) and, if so, consider increasing the memory limit in the relevant `Deployment.yaml`.
	- Check the logs before the container restarted to see if there are `panic:` messages or similar using `kubectl logs -p searcher`.
- **Docker Compose:**
	- Determine if the pod was OOM killed using `docker inspect -f '{{json .State}}' searcher` (look for `"OOMKilled":true`) and, if so, consider increasing the memory limit of the searcher container in `docker-compose.yml`.
	- Check the logs before the container restarted to see if there are `panic:` messages or similar using `docker logs searcher` (note this will include logs from the previous and currently running container).

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/searcher/searcher?viewPanel=100200` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Search-core team](https://about.sourcegraph.com/handbook/engineering/search/core).*</sub>

<details>
<summary>Technical details</summary>

Query: `count by(name) ((time() - container_last_seen{name=~"^searcher.*"}) > 60)`

</details>

<br />

#### searcher: container_cpu_usage

<p class="subtitle">Container cpu usage total (1m average) across all cores by instance</p>

Refer to the [alert solutions reference](./alert_solutions.md#searcher-container-cpu-usage) for 1 alert related to this panel.

To see this panel, visit `/-/debug/grafana/d/searcher/searcher?viewPanel=100201` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Search-core team](https://about.sourcegraph.com/handbook/engineering/search/core).*</sub>

<details>
<summary>Technical details</summary>

Query: `cadvisor_container_cpu_usage_percentage_total{name=~"^searcher.*"}`

</details>

<br />

#### searcher: container_memory_usage

<p class="subtitle">Container memory usage by instance</p>

Refer to the [alert solutions reference](./alert_solutions.md#searcher-container-memory-usage) for 1 alert related to this panel.

To see this panel, visit `/-/debug/grafana/d/searcher/searcher?viewPanel=100202` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Search-core team](https://about.sourcegraph.com/handbook/engineering/search/core).*</sub>

<details>
<summary>Technical details</summary>

Query: `cadvisor_container_memory_usage_percentage_total{name=~"^searcher.*"}`

</details>

<br />

#### searcher: fs_io_operations

<p class="subtitle">Filesystem reads and writes rate by instance over 1h</p>

This value indicates the number of filesystem read and write operations by containers of this service.
When extremely high, this can indicate a resource usage problem, or can cause problems with the service itself, especially if high values or spikes correlate with {{CONTAINER_NAME}} issues.

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/searcher/searcher?viewPanel=100203` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum by(name) (rate(container_fs_reads_total{name=~"^searcher.*"}[1h]) + rate(container_fs_writes_total{name=~"^searcher.*"}[1h]))`

</details>

<br />

### Searcher: Provisioning indicators (not available on server)

#### searcher: provisioning_container_cpu_usage_long_term

<p class="subtitle">Container cpu usage total (90th percentile over 1d) across all cores by instance</p>

Refer to the [alert solutions reference](./alert_solutions.md#searcher-provisioning-container-cpu-usage-long-term) for 1 alert related to this panel.

To see this panel, visit `/-/debug/grafana/d/searcher/searcher?viewPanel=100300` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Search-core team](https://about.sourcegraph.com/handbook/engineering/search/core).*</sub>

<details>
<summary>Technical details</summary>

Query: `quantile_over_time(0.9, cadvisor_container_cpu_usage_percentage_total{name=~"^searcher.*"}[1d])`

</details>

<br />

#### searcher: provisioning_container_memory_usage_long_term

<p class="subtitle">Container memory usage (1d maximum) by instance</p>

Refer to the [alert solutions reference](./alert_solutions.md#searcher-provisioning-container-memory-usage-long-term) for 1 alert related to this panel.

To see this panel, visit `/-/debug/grafana/d/searcher/searcher?viewPanel=100301` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Search-core team](https://about.sourcegraph.com/handbook/engineering/search/core).*</sub>

<details>
<summary>Technical details</summary>

Query: `max_over_time(cadvisor_container_memory_usage_percentage_total{name=~"^searcher.*"}[1d])`

</details>

<br />

#### searcher: provisioning_container_cpu_usage_short_term

<p class="subtitle">Container cpu usage total (5m maximum) across all cores by instance</p>

Refer to the [alert solutions reference](./alert_solutions.md#searcher-provisioning-container-cpu-usage-short-term) for 1 alert related to this panel.

To see this panel, visit `/-/debug/grafana/d/searcher/searcher?viewPanel=100310` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Search-core team](https://about.sourcegraph.com/handbook/engineering/search/core).*</sub>

<details>
<summary>Technical details</summary>

Query: `max_over_time(cadvisor_container_cpu_usage_percentage_total{name=~"^searcher.*"}[5m])`

</details>

<br />

#### searcher: provisioning_container_memory_usage_short_term

<p class="subtitle">Container memory usage (5m maximum) by instance</p>

Refer to the [alert solutions reference](./alert_solutions.md#searcher-provisioning-container-memory-usage-short-term) for 1 alert related to this panel.

To see this panel, visit `/-/debug/grafana/d/searcher/searcher?viewPanel=100311` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Search-core team](https://about.sourcegraph.com/handbook/engineering/search/core).*</sub>

<details>
<summary>Technical details</summary>

Query: `max_over_time(cadvisor_container_memory_usage_percentage_total{name=~"^searcher.*"}[5m])`

</details>

<br />

### Searcher: Golang runtime monitoring

#### searcher: go_goroutines

<p class="subtitle">Maximum active goroutines</p>

A high value here indicates a possible goroutine leak.

Refer to the [alert solutions reference](./alert_solutions.md#searcher-go-goroutines) for 1 alert related to this panel.

To see this panel, visit `/-/debug/grafana/d/searcher/searcher?viewPanel=100400` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Search-core team](https://about.sourcegraph.com/handbook/engineering/search/core).*</sub>

<details>
<summary>Technical details</summary>

Query: `max by(instance) (go_goroutines{job=~".*searcher"})`

</details>

<br />

#### searcher: go_gc_duration_seconds

<p class="subtitle">Maximum go garbage collection duration</p>

Refer to the [alert solutions reference](./alert_solutions.md#searcher-go-gc-duration-seconds) for 1 alert related to this panel.

To see this panel, visit `/-/debug/grafana/d/searcher/searcher?viewPanel=100401` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Search-core team](https://about.sourcegraph.com/handbook/engineering/search/core).*</sub>

<details>
<summary>Technical details</summary>

Query: `max by(instance) (go_gc_duration_seconds{job=~".*searcher"})`

</details>

<br />

### Searcher: Kubernetes monitoring (only available on Kubernetes)

#### searcher: pods_available_percentage

<p class="subtitle">Percentage pods available</p>

Refer to the [alert solutions reference](./alert_solutions.md#searcher-pods-available-percentage) for 1 alert related to this panel.

To see this panel, visit `/-/debug/grafana/d/searcher/searcher?viewPanel=100500` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Search-core team](https://about.sourcegraph.com/handbook/engineering/search/core).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum by(app) (up{app=~".*searcher"}) / count by (app) (up{app=~".*searcher"}) * 100`

</details>

<br />

## Symbols

<p class="subtitle">Handles symbol searches for unindexed branches.</p>

To see this dashboard, visit `/-/debug/grafana/d/symbols/symbols` on your Sourcegraph instance.

#### symbols: store_fetch_failures

<p class="subtitle">Store fetch failures every 5m</p>

Refer to the [alert solutions reference](./alert_solutions.md#symbols-store-fetch-failures) for 1 alert related to this panel.

To see this panel, visit `/-/debug/grafana/d/symbols/symbols?viewPanel=100000` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum(increase(symbols_store_fetch_failed[5m]))`

</details>

<br />

#### symbols: current_fetch_queue_size

<p class="subtitle">Current fetch queue size</p>

Refer to the [alert solutions reference](./alert_solutions.md#symbols-current-fetch-queue-size) for 1 alert related to this panel.

To see this panel, visit `/-/debug/grafana/d/symbols/symbols?viewPanel=100001` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum(symbols_store_fetch_queue_size)`

</details>

<br />

### Symbols: Internal service requests

#### symbols: frontend_internal_api_error_responses

<p class="subtitle">Frontend-internal API error responses every 5m by route</p>

Refer to the [alert solutions reference](./alert_solutions.md#symbols-frontend-internal-api-error-responses) for 1 alert related to this panel.

To see this panel, visit `/-/debug/grafana/d/symbols/symbols?viewPanel=100100` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum by (category)(increase(src_frontend_internal_request_duration_seconds_count{job="symbols",code!~"2.."}[5m])) / ignoring(category) group_left sum(increase(src_frontend_internal_request_duration_seconds_count{job="symbols"}[5m]))`

</details>

<br />

### Symbols: Container monitoring (not available on server)

#### symbols: container_missing

<p class="subtitle">Container missing</p>

This value is the number of times a container has not been seen for more than one minute. If you observe this
value change independent of deployment events (such as an upgrade), it could indicate pods are being OOM killed or terminated for some other reasons.

- **Kubernetes:**
	- Determine if the pod was OOM killed using `kubectl describe pod symbols` (look for `OOMKilled: true`) and, if so, consider increasing the memory limit in the relevant `Deployment.yaml`.
	- Check the logs before the container restarted to see if there are `panic:` messages or similar using `kubectl logs -p symbols`.
- **Docker Compose:**
	- Determine if the pod was OOM killed using `docker inspect -f '{{json .State}}' symbols` (look for `"OOMKilled":true`) and, if so, consider increasing the memory limit of the symbols container in `docker-compose.yml`.
	- Check the logs before the container restarted to see if there are `panic:` messages or similar using `docker logs symbols` (note this will include logs from the previous and currently running container).

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/symbols/symbols?viewPanel=100200` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `count by(name) ((time() - container_last_seen{name=~"^symbols.*"}) > 60)`

</details>

<br />

#### symbols: container_cpu_usage

<p class="subtitle">Container cpu usage total (1m average) across all cores by instance</p>

Refer to the [alert solutions reference](./alert_solutions.md#symbols-container-cpu-usage) for 1 alert related to this panel.

To see this panel, visit `/-/debug/grafana/d/symbols/symbols?viewPanel=100201` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `cadvisor_container_cpu_usage_percentage_total{name=~"^symbols.*"}`

</details>

<br />

#### symbols: container_memory_usage

<p class="subtitle">Container memory usage by instance</p>

Refer to the [alert solutions reference](./alert_solutions.md#symbols-container-memory-usage) for 1 alert related to this panel.

To see this panel, visit `/-/debug/grafana/d/symbols/symbols?viewPanel=100202` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `cadvisor_container_memory_usage_percentage_total{name=~"^symbols.*"}`

</details>

<br />

#### symbols: fs_io_operations

<p class="subtitle">Filesystem reads and writes rate by instance over 1h</p>

This value indicates the number of filesystem read and write operations by containers of this service.
When extremely high, this can indicate a resource usage problem, or can cause problems with the service itself, especially if high values or spikes correlate with {{CONTAINER_NAME}} issues.

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/symbols/symbols?viewPanel=100203` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum by(name) (rate(container_fs_reads_total{name=~"^symbols.*"}[1h]) + rate(container_fs_writes_total{name=~"^symbols.*"}[1h]))`

</details>

<br />

### Symbols: Provisioning indicators (not available on server)

#### symbols: provisioning_container_cpu_usage_long_term

<p class="subtitle">Container cpu usage total (90th percentile over 1d) across all cores by instance</p>

Refer to the [alert solutions reference](./alert_solutions.md#symbols-provisioning-container-cpu-usage-long-term) for 1 alert related to this panel.

To see this panel, visit `/-/debug/grafana/d/symbols/symbols?viewPanel=100300` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `quantile_over_time(0.9, cadvisor_container_cpu_usage_percentage_total{name=~"^symbols.*"}[1d])`

</details>

<br />

#### symbols: provisioning_container_memory_usage_long_term

<p class="subtitle">Container memory usage (1d maximum) by instance</p>

Refer to the [alert solutions reference](./alert_solutions.md#symbols-provisioning-container-memory-usage-long-term) for 1 alert related to this panel.

To see this panel, visit `/-/debug/grafana/d/symbols/symbols?viewPanel=100301` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `max_over_time(cadvisor_container_memory_usage_percentage_total{name=~"^symbols.*"}[1d])`

</details>

<br />

#### symbols: provisioning_container_cpu_usage_short_term

<p class="subtitle">Container cpu usage total (5m maximum) across all cores by instance</p>

Refer to the [alert solutions reference](./alert_solutions.md#symbols-provisioning-container-cpu-usage-short-term) for 1 alert related to this panel.

To see this panel, visit `/-/debug/grafana/d/symbols/symbols?viewPanel=100310` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `max_over_time(cadvisor_container_cpu_usage_percentage_total{name=~"^symbols.*"}[5m])`

</details>

<br />

#### symbols: provisioning_container_memory_usage_short_term

<p class="subtitle">Container memory usage (5m maximum) by instance</p>

Refer to the [alert solutions reference](./alert_solutions.md#symbols-provisioning-container-memory-usage-short-term) for 1 alert related to this panel.

To see this panel, visit `/-/debug/grafana/d/symbols/symbols?viewPanel=100311` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `max_over_time(cadvisor_container_memory_usage_percentage_total{name=~"^symbols.*"}[5m])`

</details>

<br />

### Symbols: Golang runtime monitoring

#### symbols: go_goroutines

<p class="subtitle">Maximum active goroutines</p>

A high value here indicates a possible goroutine leak.

Refer to the [alert solutions reference](./alert_solutions.md#symbols-go-goroutines) for 1 alert related to this panel.

To see this panel, visit `/-/debug/grafana/d/symbols/symbols?viewPanel=100400` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `max by(instance) (go_goroutines{job=~".*symbols"})`

</details>

<br />

#### symbols: go_gc_duration_seconds

<p class="subtitle">Maximum go garbage collection duration</p>

Refer to the [alert solutions reference](./alert_solutions.md#symbols-go-gc-duration-seconds) for 1 alert related to this panel.

To see this panel, visit `/-/debug/grafana/d/symbols/symbols?viewPanel=100401` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `max by(instance) (go_gc_duration_seconds{job=~".*symbols"})`

</details>

<br />

### Symbols: Kubernetes monitoring (only available on Kubernetes)

#### symbols: pods_available_percentage

<p class="subtitle">Percentage pods available</p>

Refer to the [alert solutions reference](./alert_solutions.md#symbols-pods-available-percentage) for 1 alert related to this panel.

To see this panel, visit `/-/debug/grafana/d/symbols/symbols?viewPanel=100500` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum by(app) (up{app=~".*symbols"}) / count by (app) (up{app=~".*symbols"}) * 100`

</details>

<br />

## Syntect Server

<p class="subtitle">Handles syntax highlighting for code files.</p>

To see this dashboard, visit `/-/debug/grafana/d/syntect-server/syntect-server` on your Sourcegraph instance.

#### syntect-server: syntax_highlighting_errors

<p class="subtitle">Syntax highlighting errors every 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/syntect-server/syntect-server?viewPanel=100000` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum(increase(src_syntax_highlighting_requests{status="error"}[5m])) / sum(increase(src_syntax_highlighting_requests[5m])) * 100`

</details>

<br />

#### syntect-server: syntax_highlighting_timeouts

<p class="subtitle">Syntax highlighting timeouts every 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/syntect-server/syntect-server?viewPanel=100001` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum(increase(src_syntax_highlighting_requests{status="timeout"}[5m])) / sum(increase(src_syntax_highlighting_requests[5m])) * 100`

</details>

<br />

#### syntect-server: syntax_highlighting_panics

<p class="subtitle">Syntax highlighting panics every 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/syntect-server/syntect-server?viewPanel=100010` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum(increase(src_syntax_highlighting_requests{status="panic"}[5m]))`

</details>

<br />

#### syntect-server: syntax_highlighting_worker_deaths

<p class="subtitle">Syntax highlighter worker deaths every 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/syntect-server/syntect-server?viewPanel=100011` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum(increase(src_syntax_highlighting_requests{status="hss_worker_timeout"}[5m]))`

</details>

<br />

### Syntect Server: Container monitoring (not available on server)

#### syntect-server: container_missing

<p class="subtitle">Container missing</p>

This value is the number of times a container has not been seen for more than one minute. If you observe this
value change independent of deployment events (such as an upgrade), it could indicate pods are being OOM killed or terminated for some other reasons.

- **Kubernetes:**
	- Determine if the pod was OOM killed using `kubectl describe pod syntect-server` (look for `OOMKilled: true`) and, if so, consider increasing the memory limit in the relevant `Deployment.yaml`.
	- Check the logs before the container restarted to see if there are `panic:` messages or similar using `kubectl logs -p syntect-server`.
- **Docker Compose:**
	- Determine if the pod was OOM killed using `docker inspect -f '{{json .State}}' syntect-server` (look for `"OOMKilled":true`) and, if so, consider increasing the memory limit of the syntect-server container in `docker-compose.yml`.
	- Check the logs before the container restarted to see if there are `panic:` messages or similar using `docker logs syntect-server` (note this will include logs from the previous and currently running container).

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/syntect-server/syntect-server?viewPanel=100100` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<details>
<summary>Technical details</summary>

Query: `count by(name) ((time() - container_last_seen{name=~"^syntect-server.*"}) > 60)`

</details>

<br />

#### syntect-server: container_cpu_usage

<p class="subtitle">Container cpu usage total (1m average) across all cores by instance</p>

Refer to the [alert solutions reference](./alert_solutions.md#syntect-server-container-cpu-usage) for 1 alert related to this panel.

To see this panel, visit `/-/debug/grafana/d/syntect-server/syntect-server?viewPanel=100101` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<details>
<summary>Technical details</summary>

Query: `cadvisor_container_cpu_usage_percentage_total{name=~"^syntect-server.*"}`

</details>

<br />

#### syntect-server: container_memory_usage

<p class="subtitle">Container memory usage by instance</p>

Refer to the [alert solutions reference](./alert_solutions.md#syntect-server-container-memory-usage) for 1 alert related to this panel.

To see this panel, visit `/-/debug/grafana/d/syntect-server/syntect-server?viewPanel=100102` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<details>
<summary>Technical details</summary>

Query: `cadvisor_container_memory_usage_percentage_total{name=~"^syntect-server.*"}`

</details>

<br />

#### syntect-server: fs_io_operations

<p class="subtitle">Filesystem reads and writes rate by instance over 1h</p>

This value indicates the number of filesystem read and write operations by containers of this service.
When extremely high, this can indicate a resource usage problem, or can cause problems with the service itself, especially if high values or spikes correlate with {{CONTAINER_NAME}} issues.

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/syntect-server/syntect-server?viewPanel=100103` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum by(name) (rate(container_fs_reads_total{name=~"^syntect-server.*"}[1h]) + rate(container_fs_writes_total{name=~"^syntect-server.*"}[1h]))`

</details>

<br />

### Syntect Server: Provisioning indicators (not available on server)

#### syntect-server: provisioning_container_cpu_usage_long_term

<p class="subtitle">Container cpu usage total (90th percentile over 1d) across all cores by instance</p>

Refer to the [alert solutions reference](./alert_solutions.md#syntect-server-provisioning-container-cpu-usage-long-term) for 1 alert related to this panel.

To see this panel, visit `/-/debug/grafana/d/syntect-server/syntect-server?viewPanel=100200` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<details>
<summary>Technical details</summary>

Query: `quantile_over_time(0.9, cadvisor_container_cpu_usage_percentage_total{name=~"^syntect-server.*"}[1d])`

</details>

<br />

#### syntect-server: provisioning_container_memory_usage_long_term

<p class="subtitle">Container memory usage (1d maximum) by instance</p>

Refer to the [alert solutions reference](./alert_solutions.md#syntect-server-provisioning-container-memory-usage-long-term) for 1 alert related to this panel.

To see this panel, visit `/-/debug/grafana/d/syntect-server/syntect-server?viewPanel=100201` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<details>
<summary>Technical details</summary>

Query: `max_over_time(cadvisor_container_memory_usage_percentage_total{name=~"^syntect-server.*"}[1d])`

</details>

<br />

#### syntect-server: provisioning_container_cpu_usage_short_term

<p class="subtitle">Container cpu usage total (5m maximum) across all cores by instance</p>

Refer to the [alert solutions reference](./alert_solutions.md#syntect-server-provisioning-container-cpu-usage-short-term) for 1 alert related to this panel.

To see this panel, visit `/-/debug/grafana/d/syntect-server/syntect-server?viewPanel=100210` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<details>
<summary>Technical details</summary>

Query: `max_over_time(cadvisor_container_cpu_usage_percentage_total{name=~"^syntect-server.*"}[5m])`

</details>

<br />

#### syntect-server: provisioning_container_memory_usage_short_term

<p class="subtitle">Container memory usage (5m maximum) by instance</p>

Refer to the [alert solutions reference](./alert_solutions.md#syntect-server-provisioning-container-memory-usage-short-term) for 1 alert related to this panel.

To see this panel, visit `/-/debug/grafana/d/syntect-server/syntect-server?viewPanel=100211` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<details>
<summary>Technical details</summary>

Query: `max_over_time(cadvisor_container_memory_usage_percentage_total{name=~"^syntect-server.*"}[5m])`

</details>

<br />

### Syntect Server: Kubernetes monitoring (only available on Kubernetes)

#### syntect-server: pods_available_percentage

<p class="subtitle">Percentage pods available</p>

Refer to the [alert solutions reference](./alert_solutions.md#syntect-server-pods-available-percentage) for 1 alert related to this panel.

To see this panel, visit `/-/debug/grafana/d/syntect-server/syntect-server?viewPanel=100300` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum by(app) (up{app=~".*syntect-server"}) / count by (app) (up{app=~".*syntect-server"}) * 100`

</details>

<br />

## Zoekt Index Server

<p class="subtitle">Indexes repositories and populates the search index.</p>

To see this dashboard, visit `/-/debug/grafana/d/zoekt-indexserver/zoekt-indexserver` on your Sourcegraph instance.

#### zoekt-indexserver: repos_assigned

<p class="subtitle">Total number of repos</p>

Sudden changes should be caused by indexing configuration changes.

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/zoekt-indexserver/zoekt-indexserver?viewPanel=100000` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Search-core team](https://about.sourcegraph.com/handbook/engineering/search/core).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum(index_num_assigned)`

</details>

<br />

#### zoekt-indexserver: repo_index_state

<p class="subtitle">Indexing results over 5m (noop=no changes, empty=no branches to index)</p>

A persistent failing state indicates some repositories cannot be indexed, perhaps due to size and timeouts.

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/zoekt-indexserver/zoekt-indexserver?viewPanel=100010` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Search-core team](https://about.sourcegraph.com/handbook/engineering/search/core).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum by (state) (increase(index_repo_seconds_count[5m]))`

</details>

<br />

#### zoekt-indexserver: repo_index_success_speed

<p class="subtitle">Successful indexing durations</p>

Latency increases can indicate bottlenecks in the indexserver.

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/zoekt-indexserver/zoekt-indexserver?viewPanel=100011` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Search-core team](https://about.sourcegraph.com/handbook/engineering/search/core).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum by (le, state) (increase(index_repo_seconds_bucket{state="success"}[$__rate_interval]))`

</details>

<br />

#### zoekt-indexserver: repo_index_fail_speed

<p class="subtitle">Failed indexing durations</p>

Failures happening after a long time indicates timeouts.

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/zoekt-indexserver/zoekt-indexserver?viewPanel=100012` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Search-core team](https://about.sourcegraph.com/handbook/engineering/search/core).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum by (le, state) (increase(index_repo_seconds_bucket{state="fail"}[$__rate_interval]))`

</details>

<br />

#### zoekt-indexserver: average_resolve_revision_duration

<p class="subtitle">Average resolve revision duration over 5m</p>

Refer to the [alert solutions reference](./alert_solutions.md#zoekt-indexserver-average-resolve-revision-duration) for 2 alerts related to this panel.

To see this panel, visit `/-/debug/grafana/d/zoekt-indexserver/zoekt-indexserver?viewPanel=100020` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Search-core team](https://about.sourcegraph.com/handbook/engineering/search/core).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum(rate(resolve_revision_seconds_sum[5m])) / sum(rate(resolve_revision_seconds_count[5m]))`

</details>

<br />

#### zoekt-indexserver: get_index_options_error_increase

<p class="subtitle">The number of repositories we failed to get indexing options over 5m</p>

When considering indexing a repository we ask for the index configuration
from frontend per repository. The most likely reason this would fail is
failing to resolve branch names to git SHAs.

This value can spike up during deployments/etc. Only if you encounter
sustained periods of errors is there an underlying issue. When sustained
this indicates repositories will not get updated indexes.

Refer to the [alert solutions reference](./alert_solutions.md#zoekt-indexserver-get-index-options-error-increase) for 2 alerts related to this panel.

To see this panel, visit `/-/debug/grafana/d/zoekt-indexserver/zoekt-indexserver?viewPanel=100021` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Search-core team](https://about.sourcegraph.com/handbook/engineering/search/core).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum(increase(get_index_options_error_total[5m]))`

</details>

<br />

### Zoekt Index Server: Container monitoring (not available on server)

#### zoekt-indexserver: container_missing

<p class="subtitle">Container missing</p>

This value is the number of times a container has not been seen for more than one minute. If you observe this
value change independent of deployment events (such as an upgrade), it could indicate pods are being OOM killed or terminated for some other reasons.

- **Kubernetes:**
	- Determine if the pod was OOM killed using `kubectl describe pod zoekt-indexserver` (look for `OOMKilled: true`) and, if so, consider increasing the memory limit in the relevant `Deployment.yaml`.
	- Check the logs before the container restarted to see if there are `panic:` messages or similar using `kubectl logs -p zoekt-indexserver`.
- **Docker Compose:**
	- Determine if the pod was OOM killed using `docker inspect -f '{{json .State}}' zoekt-indexserver` (look for `"OOMKilled":true`) and, if so, consider increasing the memory limit of the zoekt-indexserver container in `docker-compose.yml`.
	- Check the logs before the container restarted to see if there are `panic:` messages or similar using `docker logs zoekt-indexserver` (note this will include logs from the previous and currently running container).

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/zoekt-indexserver/zoekt-indexserver?viewPanel=100100` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Search-core team](https://about.sourcegraph.com/handbook/engineering/search/core).*</sub>

<details>
<summary>Technical details</summary>

Query: `count by(name) ((time() - container_last_seen{name=~"^zoekt-indexserver.*"}) > 60)`

</details>

<br />

#### zoekt-indexserver: container_cpu_usage

<p class="subtitle">Container cpu usage total (1m average) across all cores by instance</p>

Refer to the [alert solutions reference](./alert_solutions.md#zoekt-indexserver-container-cpu-usage) for 1 alert related to this panel.

To see this panel, visit `/-/debug/grafana/d/zoekt-indexserver/zoekt-indexserver?viewPanel=100101` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Search-core team](https://about.sourcegraph.com/handbook/engineering/search/core).*</sub>

<details>
<summary>Technical details</summary>

Query: `cadvisor_container_cpu_usage_percentage_total{name=~"^zoekt-indexserver.*"}`

</details>

<br />

#### zoekt-indexserver: container_memory_usage

<p class="subtitle">Container memory usage by instance</p>

Refer to the [alert solutions reference](./alert_solutions.md#zoekt-indexserver-container-memory-usage) for 1 alert related to this panel.

To see this panel, visit `/-/debug/grafana/d/zoekt-indexserver/zoekt-indexserver?viewPanel=100102` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Search-core team](https://about.sourcegraph.com/handbook/engineering/search/core).*</sub>

<details>
<summary>Technical details</summary>

Query: `cadvisor_container_memory_usage_percentage_total{name=~"^zoekt-indexserver.*"}`

</details>

<br />

#### zoekt-indexserver: fs_io_operations

<p class="subtitle">Filesystem reads and writes rate by instance over 1h</p>

This value indicates the number of filesystem read and write operations by containers of this service.
When extremely high, this can indicate a resource usage problem, or can cause problems with the service itself, especially if high values or spikes correlate with {{CONTAINER_NAME}} issues.

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/zoekt-indexserver/zoekt-indexserver?viewPanel=100103` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum by(name) (rate(container_fs_reads_total{name=~"^zoekt-indexserver.*"}[1h]) + rate(container_fs_writes_total{name=~"^zoekt-indexserver.*"}[1h]))`

</details>

<br />

### Zoekt Index Server: Provisioning indicators (not available on server)

#### zoekt-indexserver: provisioning_container_cpu_usage_long_term

<p class="subtitle">Container cpu usage total (90th percentile over 1d) across all cores by instance</p>

Refer to the [alert solutions reference](./alert_solutions.md#zoekt-indexserver-provisioning-container-cpu-usage-long-term) for 1 alert related to this panel.

To see this panel, visit `/-/debug/grafana/d/zoekt-indexserver/zoekt-indexserver?viewPanel=100200` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Search-core team](https://about.sourcegraph.com/handbook/engineering/search/core).*</sub>

<details>
<summary>Technical details</summary>

Query: `quantile_over_time(0.9, cadvisor_container_cpu_usage_percentage_total{name=~"^zoekt-indexserver.*"}[1d])`

</details>

<br />

#### zoekt-indexserver: provisioning_container_memory_usage_long_term

<p class="subtitle">Container memory usage (1d maximum) by instance</p>

Refer to the [alert solutions reference](./alert_solutions.md#zoekt-indexserver-provisioning-container-memory-usage-long-term) for 1 alert related to this panel.

To see this panel, visit `/-/debug/grafana/d/zoekt-indexserver/zoekt-indexserver?viewPanel=100201` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Search-core team](https://about.sourcegraph.com/handbook/engineering/search/core).*</sub>

<details>
<summary>Technical details</summary>

Query: `max_over_time(cadvisor_container_memory_usage_percentage_total{name=~"^zoekt-indexserver.*"}[1d])`

</details>

<br />

#### zoekt-indexserver: provisioning_container_cpu_usage_short_term

<p class="subtitle">Container cpu usage total (5m maximum) across all cores by instance</p>

Refer to the [alert solutions reference](./alert_solutions.md#zoekt-indexserver-provisioning-container-cpu-usage-short-term) for 1 alert related to this panel.

To see this panel, visit `/-/debug/grafana/d/zoekt-indexserver/zoekt-indexserver?viewPanel=100210` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Search-core team](https://about.sourcegraph.com/handbook/engineering/search/core).*</sub>

<details>
<summary>Technical details</summary>

Query: `max_over_time(cadvisor_container_cpu_usage_percentage_total{name=~"^zoekt-indexserver.*"}[5m])`

</details>

<br />

#### zoekt-indexserver: provisioning_container_memory_usage_short_term

<p class="subtitle">Container memory usage (5m maximum) by instance</p>

Refer to the [alert solutions reference](./alert_solutions.md#zoekt-indexserver-provisioning-container-memory-usage-short-term) for 1 alert related to this panel.

To see this panel, visit `/-/debug/grafana/d/zoekt-indexserver/zoekt-indexserver?viewPanel=100211` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Search-core team](https://about.sourcegraph.com/handbook/engineering/search/core).*</sub>

<details>
<summary>Technical details</summary>

Query: `max_over_time(cadvisor_container_memory_usage_percentage_total{name=~"^zoekt-indexserver.*"}[5m])`

</details>

<br />

### Zoekt Index Server: Kubernetes monitoring (only available on Kubernetes)

#### zoekt-indexserver: pods_available_percentage

<p class="subtitle">Percentage pods available</p>

Refer to the [alert solutions reference](./alert_solutions.md#zoekt-indexserver-pods-available-percentage) for 1 alert related to this panel.

To see this panel, visit `/-/debug/grafana/d/zoekt-indexserver/zoekt-indexserver?viewPanel=100300` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Search-core team](https://about.sourcegraph.com/handbook/engineering/search/core).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum by(app) (up{app=~".*indexed-search"}) / count by (app) (up{app=~".*indexed-search"}) * 100`

</details>

<br />

## Zoekt Web Server

<p class="subtitle">Serves indexed search requests using the search index.</p>

To see this dashboard, visit `/-/debug/grafana/d/zoekt-webserver/zoekt-webserver` on your Sourcegraph instance.

#### zoekt-webserver: indexed_search_request_errors

<p class="subtitle">Indexed search request errors every 5m by code</p>

Refer to the [alert solutions reference](./alert_solutions.md#zoekt-webserver-indexed-search-request-errors) for 1 alert related to this panel.

To see this panel, visit `/-/debug/grafana/d/zoekt-webserver/zoekt-webserver?viewPanel=100000` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Search-core team](https://about.sourcegraph.com/handbook/engineering/search/core).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum by (code)(increase(src_zoekt_request_duration_seconds_count{code!~"2.."}[5m])) / ignoring(code) group_left sum(increase(src_zoekt_request_duration_seconds_count[5m])) * 100`

</details>

<br />

### Zoekt Web Server: Container monitoring (not available on server)

#### zoekt-webserver: container_missing

<p class="subtitle">Container missing</p>

This value is the number of times a container has not been seen for more than one minute. If you observe this
value change independent of deployment events (such as an upgrade), it could indicate pods are being OOM killed or terminated for some other reasons.

- **Kubernetes:**
	- Determine if the pod was OOM killed using `kubectl describe pod zoekt-webserver` (look for `OOMKilled: true`) and, if so, consider increasing the memory limit in the relevant `Deployment.yaml`.
	- Check the logs before the container restarted to see if there are `panic:` messages or similar using `kubectl logs -p zoekt-webserver`.
- **Docker Compose:**
	- Determine if the pod was OOM killed using `docker inspect -f '{{json .State}}' zoekt-webserver` (look for `"OOMKilled":true`) and, if so, consider increasing the memory limit of the zoekt-webserver container in `docker-compose.yml`.
	- Check the logs before the container restarted to see if there are `panic:` messages or similar using `docker logs zoekt-webserver` (note this will include logs from the previous and currently running container).

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/zoekt-webserver/zoekt-webserver?viewPanel=100100` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Search-core team](https://about.sourcegraph.com/handbook/engineering/search/core).*</sub>

<details>
<summary>Technical details</summary>

Query: `count by(name) ((time() - container_last_seen{name=~"^zoekt-webserver.*"}) > 60)`

</details>

<br />

#### zoekt-webserver: container_cpu_usage

<p class="subtitle">Container cpu usage total (1m average) across all cores by instance</p>

Refer to the [alert solutions reference](./alert_solutions.md#zoekt-webserver-container-cpu-usage) for 1 alert related to this panel.

To see this panel, visit `/-/debug/grafana/d/zoekt-webserver/zoekt-webserver?viewPanel=100101` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Search-core team](https://about.sourcegraph.com/handbook/engineering/search/core).*</sub>

<details>
<summary>Technical details</summary>

Query: `cadvisor_container_cpu_usage_percentage_total{name=~"^zoekt-webserver.*"}`

</details>

<br />

#### zoekt-webserver: container_memory_usage

<p class="subtitle">Container memory usage by instance</p>

Refer to the [alert solutions reference](./alert_solutions.md#zoekt-webserver-container-memory-usage) for 1 alert related to this panel.

To see this panel, visit `/-/debug/grafana/d/zoekt-webserver/zoekt-webserver?viewPanel=100102` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Search-core team](https://about.sourcegraph.com/handbook/engineering/search/core).*</sub>

<details>
<summary>Technical details</summary>

Query: `cadvisor_container_memory_usage_percentage_total{name=~"^zoekt-webserver.*"}`

</details>

<br />

#### zoekt-webserver: fs_io_operations

<p class="subtitle">Filesystem reads and writes rate by instance over 1h</p>

This value indicates the number of filesystem read and write operations by containers of this service.
When extremely high, this can indicate a resource usage problem, or can cause problems with the service itself, especially if high values or spikes correlate with {{CONTAINER_NAME}} issues.

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/zoekt-webserver/zoekt-webserver?viewPanel=100103` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum by(name) (rate(container_fs_reads_total{name=~"^zoekt-webserver.*"}[1h]) + rate(container_fs_writes_total{name=~"^zoekt-webserver.*"}[1h]))`

</details>

<br />

### Zoekt Web Server: Provisioning indicators (not available on server)

#### zoekt-webserver: provisioning_container_cpu_usage_long_term

<p class="subtitle">Container cpu usage total (90th percentile over 1d) across all cores by instance</p>

Refer to the [alert solutions reference](./alert_solutions.md#zoekt-webserver-provisioning-container-cpu-usage-long-term) for 1 alert related to this panel.

To see this panel, visit `/-/debug/grafana/d/zoekt-webserver/zoekt-webserver?viewPanel=100200` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Search-core team](https://about.sourcegraph.com/handbook/engineering/search/core).*</sub>

<details>
<summary>Technical details</summary>

Query: `quantile_over_time(0.9, cadvisor_container_cpu_usage_percentage_total{name=~"^zoekt-webserver.*"}[1d])`

</details>

<br />

#### zoekt-webserver: provisioning_container_memory_usage_long_term

<p class="subtitle">Container memory usage (1d maximum) by instance</p>

Refer to the [alert solutions reference](./alert_solutions.md#zoekt-webserver-provisioning-container-memory-usage-long-term) for 1 alert related to this panel.

To see this panel, visit `/-/debug/grafana/d/zoekt-webserver/zoekt-webserver?viewPanel=100201` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Search-core team](https://about.sourcegraph.com/handbook/engineering/search/core).*</sub>

<details>
<summary>Technical details</summary>

Query: `max_over_time(cadvisor_container_memory_usage_percentage_total{name=~"^zoekt-webserver.*"}[1d])`

</details>

<br />

#### zoekt-webserver: provisioning_container_cpu_usage_short_term

<p class="subtitle">Container cpu usage total (5m maximum) across all cores by instance</p>

Refer to the [alert solutions reference](./alert_solutions.md#zoekt-webserver-provisioning-container-cpu-usage-short-term) for 1 alert related to this panel.

To see this panel, visit `/-/debug/grafana/d/zoekt-webserver/zoekt-webserver?viewPanel=100210` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Search-core team](https://about.sourcegraph.com/handbook/engineering/search/core).*</sub>

<details>
<summary>Technical details</summary>

Query: `max_over_time(cadvisor_container_cpu_usage_percentage_total{name=~"^zoekt-webserver.*"}[5m])`

</details>

<br />

#### zoekt-webserver: provisioning_container_memory_usage_short_term

<p class="subtitle">Container memory usage (5m maximum) by instance</p>

Refer to the [alert solutions reference](./alert_solutions.md#zoekt-webserver-provisioning-container-memory-usage-short-term) for 1 alert related to this panel.

To see this panel, visit `/-/debug/grafana/d/zoekt-webserver/zoekt-webserver?viewPanel=100211` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Search-core team](https://about.sourcegraph.com/handbook/engineering/search/core).*</sub>

<details>
<summary>Technical details</summary>

Query: `max_over_time(cadvisor_container_memory_usage_percentage_total{name=~"^zoekt-webserver.*"}[5m])`

</details>

<br />

## Prometheus

<p class="subtitle">Sourcegraph's all-in-one Prometheus and Alertmanager service.</p>

To see this dashboard, visit `/-/debug/grafana/d/prometheus/prometheus` on your Sourcegraph instance.

### Prometheus: Metrics

#### prometheus: prometheus_rule_eval_duration

<p class="subtitle">Average prometheus rule group evaluation duration over 10m by rule group</p>

A high value here indicates Prometheus rule evaluation is taking longer than expected.
It might indicate that certain rule groups are taking too long to evaluate, or Prometheus is underprovisioned.

Rules that Sourcegraph ships with are grouped under `/sg_config_prometheus`. [Custom rules are grouped under `/sg_prometheus_addons`](https://docs.sourcegraph.com/admin/observability/metrics#prometheus-configuration).

Refer to the [alert solutions reference](./alert_solutions.md#prometheus-prometheus-rule-eval-duration) for 1 alert related to this panel.

To see this panel, visit `/-/debug/grafana/d/prometheus/prometheus?viewPanel=100000` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Distribution team](https://about.sourcegraph.com/handbook/engineering/distribution).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum by(rule_group) (avg_over_time(prometheus_rule_group_last_duration_seconds[10m]))`

</details>

<br />

#### prometheus: prometheus_rule_eval_failures

<p class="subtitle">Failed prometheus rule evaluations over 5m by rule group</p>

Rules that Sourcegraph ships with are grouped under `/sg_config_prometheus`. [Custom rules are grouped under `/sg_prometheus_addons`](https://docs.sourcegraph.com/admin/observability/metrics#prometheus-configuration).

Refer to the [alert solutions reference](./alert_solutions.md#prometheus-prometheus-rule-eval-failures) for 1 alert related to this panel.

To see this panel, visit `/-/debug/grafana/d/prometheus/prometheus?viewPanel=100001` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Distribution team](https://about.sourcegraph.com/handbook/engineering/distribution).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum by(rule_group) (rate(prometheus_rule_evaluation_failures_total[5m]))`

</details>

<br />

### Prometheus: Alerts

#### prometheus: alertmanager_notification_latency

<p class="subtitle">Alertmanager notification latency over 1m by integration</p>

Refer to the [alert solutions reference](./alert_solutions.md#prometheus-alertmanager-notification-latency) for 1 alert related to this panel.

To see this panel, visit `/-/debug/grafana/d/prometheus/prometheus?viewPanel=100100` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Distribution team](https://about.sourcegraph.com/handbook/engineering/distribution).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum by(integration) (rate(alertmanager_notification_latency_seconds_sum[1m]))`

</details>

<br />

#### prometheus: alertmanager_notification_failures

<p class="subtitle">Failed alertmanager notifications over 1m by integration</p>

Refer to the [alert solutions reference](./alert_solutions.md#prometheus-alertmanager-notification-failures) for 1 alert related to this panel.

To see this panel, visit `/-/debug/grafana/d/prometheus/prometheus?viewPanel=100101` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Distribution team](https://about.sourcegraph.com/handbook/engineering/distribution).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum by(integration) (rate(alertmanager_notifications_failed_total[1m]))`

</details>

<br />

### Prometheus: Internals

#### prometheus: prometheus_config_status

<p class="subtitle">Prometheus configuration reload status</p>

A `1` indicates Prometheus reloaded its configuration successfully.

Refer to the [alert solutions reference](./alert_solutions.md#prometheus-prometheus-config-status) for 1 alert related to this panel.

To see this panel, visit `/-/debug/grafana/d/prometheus/prometheus?viewPanel=100200` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Distribution team](https://about.sourcegraph.com/handbook/engineering/distribution).*</sub>

<details>
<summary>Technical details</summary>

Query: `prometheus_config_last_reload_successful`

</details>

<br />

#### prometheus: alertmanager_config_status

<p class="subtitle">Alertmanager configuration reload status</p>

A `1` indicates Alertmanager reloaded its configuration successfully.

Refer to the [alert solutions reference](./alert_solutions.md#prometheus-alertmanager-config-status) for 1 alert related to this panel.

To see this panel, visit `/-/debug/grafana/d/prometheus/prometheus?viewPanel=100201` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Distribution team](https://about.sourcegraph.com/handbook/engineering/distribution).*</sub>

<details>
<summary>Technical details</summary>

Query: `alertmanager_config_last_reload_successful`

</details>

<br />

#### prometheus: prometheus_tsdb_op_failure

<p class="subtitle">Prometheus tsdb failures by operation over 1m by operation</p>

Refer to the [alert solutions reference](./alert_solutions.md#prometheus-prometheus-tsdb-op-failure) for 1 alert related to this panel.

To see this panel, visit `/-/debug/grafana/d/prometheus/prometheus?viewPanel=100210` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Distribution team](https://about.sourcegraph.com/handbook/engineering/distribution).*</sub>

<details>
<summary>Technical details</summary>

Query: `increase(label_replace({__name__=~"prometheus_tsdb_(.*)_failed_total"}, "operation", "$1", "__name__", "(.+)s_failed_total")[5m:1m])`

</details>

<br />

#### prometheus: prometheus_target_sample_exceeded

<p class="subtitle">Prometheus scrapes that exceed the sample limit over 10m</p>

Refer to the [alert solutions reference](./alert_solutions.md#prometheus-prometheus-target-sample-exceeded) for 1 alert related to this panel.

To see this panel, visit `/-/debug/grafana/d/prometheus/prometheus?viewPanel=100211` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Distribution team](https://about.sourcegraph.com/handbook/engineering/distribution).*</sub>

<details>
<summary>Technical details</summary>

Query: `increase(prometheus_target_scrapes_exceeded_sample_limit_total[10m])`

</details>

<br />

#### prometheus: prometheus_target_sample_duplicate

<p class="subtitle">Prometheus scrapes rejected due to duplicate timestamps over 10m</p>

Refer to the [alert solutions reference](./alert_solutions.md#prometheus-prometheus-target-sample-duplicate) for 1 alert related to this panel.

To see this panel, visit `/-/debug/grafana/d/prometheus/prometheus?viewPanel=100212` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Distribution team](https://about.sourcegraph.com/handbook/engineering/distribution).*</sub>

<details>
<summary>Technical details</summary>

Query: `increase(prometheus_target_scrapes_sample_duplicate_timestamp_total[10m])`

</details>

<br />

### Prometheus: Container monitoring (not available on server)

#### prometheus: container_missing

<p class="subtitle">Container missing</p>

This value is the number of times a container has not been seen for more than one minute. If you observe this
value change independent of deployment events (such as an upgrade), it could indicate pods are being OOM killed or terminated for some other reasons.

- **Kubernetes:**
	- Determine if the pod was OOM killed using `kubectl describe pod prometheus` (look for `OOMKilled: true`) and, if so, consider increasing the memory limit in the relevant `Deployment.yaml`.
	- Check the logs before the container restarted to see if there are `panic:` messages or similar using `kubectl logs -p prometheus`.
- **Docker Compose:**
	- Determine if the pod was OOM killed using `docker inspect -f '{{json .State}}' prometheus` (look for `"OOMKilled":true`) and, if so, consider increasing the memory limit of the prometheus container in `docker-compose.yml`.
	- Check the logs before the container restarted to see if there are `panic:` messages or similar using `docker logs prometheus` (note this will include logs from the previous and currently running container).

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/prometheus/prometheus?viewPanel=100300` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Distribution team](https://about.sourcegraph.com/handbook/engineering/distribution).*</sub>

<details>
<summary>Technical details</summary>

Query: `count by(name) ((time() - container_last_seen{name=~"^prometheus.*"}) > 60)`

</details>

<br />

#### prometheus: container_cpu_usage

<p class="subtitle">Container cpu usage total (1m average) across all cores by instance</p>

Refer to the [alert solutions reference](./alert_solutions.md#prometheus-container-cpu-usage) for 1 alert related to this panel.

To see this panel, visit `/-/debug/grafana/d/prometheus/prometheus?viewPanel=100301` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Distribution team](https://about.sourcegraph.com/handbook/engineering/distribution).*</sub>

<details>
<summary>Technical details</summary>

Query: `cadvisor_container_cpu_usage_percentage_total{name=~"^prometheus.*"}`

</details>

<br />

#### prometheus: container_memory_usage

<p class="subtitle">Container memory usage by instance</p>

Refer to the [alert solutions reference](./alert_solutions.md#prometheus-container-memory-usage) for 1 alert related to this panel.

To see this panel, visit `/-/debug/grafana/d/prometheus/prometheus?viewPanel=100302` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Distribution team](https://about.sourcegraph.com/handbook/engineering/distribution).*</sub>

<details>
<summary>Technical details</summary>

Query: `cadvisor_container_memory_usage_percentage_total{name=~"^prometheus.*"}`

</details>

<br />

#### prometheus: fs_io_operations

<p class="subtitle">Filesystem reads and writes rate by instance over 1h</p>

This value indicates the number of filesystem read and write operations by containers of this service.
When extremely high, this can indicate a resource usage problem, or can cause problems with the service itself, especially if high values or spikes correlate with {{CONTAINER_NAME}} issues.

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/prometheus/prometheus?viewPanel=100303` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum by(name) (rate(container_fs_reads_total{name=~"^prometheus.*"}[1h]) + rate(container_fs_writes_total{name=~"^prometheus.*"}[1h]))`

</details>

<br />

### Prometheus: Provisioning indicators (not available on server)

#### prometheus: provisioning_container_cpu_usage_long_term

<p class="subtitle">Container cpu usage total (90th percentile over 1d) across all cores by instance</p>

Refer to the [alert solutions reference](./alert_solutions.md#prometheus-provisioning-container-cpu-usage-long-term) for 1 alert related to this panel.

To see this panel, visit `/-/debug/grafana/d/prometheus/prometheus?viewPanel=100400` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Distribution team](https://about.sourcegraph.com/handbook/engineering/distribution).*</sub>

<details>
<summary>Technical details</summary>

Query: `quantile_over_time(0.9, cadvisor_container_cpu_usage_percentage_total{name=~"^prometheus.*"}[1d])`

</details>

<br />

#### prometheus: provisioning_container_memory_usage_long_term

<p class="subtitle">Container memory usage (1d maximum) by instance</p>

Refer to the [alert solutions reference](./alert_solutions.md#prometheus-provisioning-container-memory-usage-long-term) for 1 alert related to this panel.

To see this panel, visit `/-/debug/grafana/d/prometheus/prometheus?viewPanel=100401` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Distribution team](https://about.sourcegraph.com/handbook/engineering/distribution).*</sub>

<details>
<summary>Technical details</summary>

Query: `max_over_time(cadvisor_container_memory_usage_percentage_total{name=~"^prometheus.*"}[1d])`

</details>

<br />

#### prometheus: provisioning_container_cpu_usage_short_term

<p class="subtitle">Container cpu usage total (5m maximum) across all cores by instance</p>

Refer to the [alert solutions reference](./alert_solutions.md#prometheus-provisioning-container-cpu-usage-short-term) for 1 alert related to this panel.

To see this panel, visit `/-/debug/grafana/d/prometheus/prometheus?viewPanel=100410` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Distribution team](https://about.sourcegraph.com/handbook/engineering/distribution).*</sub>

<details>
<summary>Technical details</summary>

Query: `max_over_time(cadvisor_container_cpu_usage_percentage_total{name=~"^prometheus.*"}[5m])`

</details>

<br />

#### prometheus: provisioning_container_memory_usage_short_term

<p class="subtitle">Container memory usage (5m maximum) by instance</p>

Refer to the [alert solutions reference](./alert_solutions.md#prometheus-provisioning-container-memory-usage-short-term) for 1 alert related to this panel.

To see this panel, visit `/-/debug/grafana/d/prometheus/prometheus?viewPanel=100411` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Distribution team](https://about.sourcegraph.com/handbook/engineering/distribution).*</sub>

<details>
<summary>Technical details</summary>

Query: `max_over_time(cadvisor_container_memory_usage_percentage_total{name=~"^prometheus.*"}[5m])`

</details>

<br />

### Prometheus: Kubernetes monitoring (only available on Kubernetes)

#### prometheus: pods_available_percentage

<p class="subtitle">Percentage pods available</p>

Refer to the [alert solutions reference](./alert_solutions.md#prometheus-pods-available-percentage) for 1 alert related to this panel.

To see this panel, visit `/-/debug/grafana/d/prometheus/prometheus?viewPanel=100500` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Distribution team](https://about.sourcegraph.com/handbook/engineering/distribution).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum by(app) (up{app=~".*prometheus"}) / count by (app) (up{app=~".*prometheus"}) * 100`

</details>

<br />

## Executor

<p class="subtitle">Executes jobs in an isolated environment.</p>

To see this dashboard, visit `/-/debug/grafana/d/executor/executor` on your Sourcegraph instance.

### Executor: Executor: Executor jobs

#### executor: executor_queue_size

<p class="subtitle">Unprocessed executor job queue size</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/executor/executor?viewPanel=100000` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `max by (queue)(src_executor_total{job=~"^(executor|sourcegraph-code-intel-indexers|executor-batches|frontend|sourcegraph-frontend|worker).*"})`

</details>

<br />

#### executor: executor_queue_growth_rate

<p class="subtitle">Unprocessed executor job queue growth rate over 30m</p>

This value compares the rate of enqueues against the rate of finished jobs for the selected queue.

	- A value < than 1 indicates that process rate > enqueue rate
	- A value = than 1 indicates that process rate = enqueue rate
	- A value > than 1 indicates that process rate < enqueue rate

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/executor/executor?viewPanel=100001` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum by (queue)(increase(src_executor_total{job=~"^(executor|sourcegraph-code-intel-indexers|executor-batches|frontend|sourcegraph-frontend|worker).*"}[30m])) / sum by (queue)(increase(src_executor_processor_total{job=~"^(executor|sourcegraph-code-intel-indexers|executor-batches|frontend|sourcegraph-frontend|worker).*"}[30m]))`

</details>

<br />

### Executor: Executor: Executor jobs

#### executor: executor_handlers

<p class="subtitle">Handler active handlers</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/executor/executor?viewPanel=100100` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum(src_executor_processor_handlers{queue=~"${queue:regex}",job=~"^(executor|sourcegraph-code-intel-indexers|executor-batches).*"})`

</details>

<br />

#### executor: executor_processor_total

<p class="subtitle">Handler operations every 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/executor/executor?viewPanel=100110` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum(increase(src_executor_processor_total{queue=~"${queue:regex}",job=~"^(executor|sourcegraph-code-intel-indexers|executor-batches).*"}[5m]))`

</details>

<br />

#### executor: executor_processor_99th_percentile_duration

<p class="subtitle">99th percentile successful handler operation duration over 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/executor/executor?viewPanel=100111` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `histogram_quantile(0.99, sum  by (le)(rate(src_executor_processor_duration_seconds_bucket{queue=~"${queue:regex}",job=~"^(executor|sourcegraph-code-intel-indexers|executor-batches).*"}[5m])))`

</details>

<br />

#### executor: executor_processor_errors_total

<p class="subtitle">Handler operation errors every 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/executor/executor?viewPanel=100112` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum(increase(src_executor_processor_errors_total{queue=~"${queue:regex}",job=~"^(executor|sourcegraph-code-intel-indexers|executor-batches).*"}[5m]))`

</details>

<br />

#### executor: executor_processor_error_rate

<p class="subtitle">Handler operation error rate over 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/executor/executor?viewPanel=100113` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum(increase(src_executor_processor_errors_total{queue=~"${queue:regex}",job=~"^(executor|sourcegraph-code-intel-indexers|executor-batches).*"}[5m])) / (sum(increase(src_executor_processor_total{queue=~"${queue:regex}",job=~"^(executor|sourcegraph-code-intel-indexers|executor-batches).*"}[5m])) + sum(increase(src_executor_processor_errors_total{queue=~"${queue:regex}",job=~"^(executor|sourcegraph-code-intel-indexers|executor-batches).*"}[5m]))) * 100`

</details>

<br />

### Executor: Run lock contention

#### executor: executor_run_lock_wait_total

<p class="subtitle">Milliseconds wait every 5m</p>

Number of milliseconds spent waiting for the run lock every 5m

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/executor/executor?viewPanel=100200` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum(increase(src_executor_run_lock_wait_total{job=~"^(executor|sourcegraph-code-intel-indexers|executor-batches).*"}[5m]))`

</details>

<br />

#### executor: executor_run_lock_held_total

<p class="subtitle">Milliseconds held every 5m</p>

Number of milliseconds spent holding for the run lock every 5m

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/executor/executor?viewPanel=100201` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum(increase(src_executor_run_lock_held_total{job=~"^(executor|sourcegraph-code-intel-indexers|executor-batches).*"}[5m]))`

</details>

<br />

### Executor: Executor: Queue API client

#### executor: apiworker_apiclient_total

<p class="subtitle">Aggregate client operations every 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/executor/executor?viewPanel=100300` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum(increase(src_apiworker_apiclient_total{job=~"^(executor|sourcegraph-code-intel-indexers|executor-batches).*"}[5m]))`

</details>

<br />

#### executor: apiworker_apiclient_99th_percentile_duration

<p class="subtitle">99th percentile successful aggregate client operation duration over 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/executor/executor?viewPanel=100301` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `histogram_quantile(0.99, sum  by (le)(rate(src_apiworker_apiclient_duration_seconds_bucket{job=~"^(executor|sourcegraph-code-intel-indexers|executor-batches).*"}[5m])))`

</details>

<br />

#### executor: apiworker_apiclient_errors_total

<p class="subtitle">Aggregate client operation errors every 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/executor/executor?viewPanel=100302` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum(increase(src_apiworker_apiclient_errors_total{job=~"^(executor|sourcegraph-code-intel-indexers|executor-batches).*"}[5m]))`

</details>

<br />

#### executor: apiworker_apiclient_error_rate

<p class="subtitle">Aggregate client operation error rate over 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/executor/executor?viewPanel=100303` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum(increase(src_apiworker_apiclient_errors_total{job=~"^(executor|sourcegraph-code-intel-indexers|executor-batches).*"}[5m])) / (sum(increase(src_apiworker_apiclient_total{job=~"^(executor|sourcegraph-code-intel-indexers|executor-batches).*"}[5m])) + sum(increase(src_apiworker_apiclient_errors_total{job=~"^(executor|sourcegraph-code-intel-indexers|executor-batches).*"}[5m]))) * 100`

</details>

<br />

#### executor: apiworker_apiclient_total

<p class="subtitle">Client operations every 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/executor/executor?viewPanel=100310` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum by (op)(increase(src_apiworker_apiclient_total{job=~"^(executor|sourcegraph-code-intel-indexers|executor-batches).*"}[5m]))`

</details>

<br />

#### executor: apiworker_apiclient_99th_percentile_duration

<p class="subtitle">99th percentile successful client operation duration over 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/executor/executor?viewPanel=100311` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `histogram_quantile(0.99, sum  by (le,op)(rate(src_apiworker_apiclient_duration_seconds_bucket{job=~"^(executor|sourcegraph-code-intel-indexers|executor-batches).*"}[5m])))`

</details>

<br />

#### executor: apiworker_apiclient_errors_total

<p class="subtitle">Client operation errors every 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/executor/executor?viewPanel=100312` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum by (op)(increase(src_apiworker_apiclient_errors_total{job=~"^(executor|sourcegraph-code-intel-indexers|executor-batches).*"}[5m]))`

</details>

<br />

#### executor: apiworker_apiclient_error_rate

<p class="subtitle">Client operation error rate over 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/executor/executor?viewPanel=100313` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum by (op)(increase(src_apiworker_apiclient_errors_total{job=~"^(executor|sourcegraph-code-intel-indexers|executor-batches).*"}[5m])) / (sum by (op)(increase(src_apiworker_apiclient_total{job=~"^(executor|sourcegraph-code-intel-indexers|executor-batches).*"}[5m])) + sum by (op)(increase(src_apiworker_apiclient_errors_total{job=~"^(executor|sourcegraph-code-intel-indexers|executor-batches).*"}[5m]))) * 100`

</details>

<br />

### Executor: Executor: Job setup

#### executor: apiworker_command_total

<p class="subtitle">Aggregate command operations every 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/executor/executor?viewPanel=100400` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum(increase(src_apiworker_command_total{op=~"setup.*",job=~"^(executor|sourcegraph-code-intel-indexers|executor-batches).*"}[5m]))`

</details>

<br />

#### executor: apiworker_command_99th_percentile_duration

<p class="subtitle">99th percentile successful aggregate command operation duration over 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/executor/executor?viewPanel=100401` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `histogram_quantile(0.99, sum  by (le)(rate(src_apiworker_command_duration_seconds_bucket{op=~"setup.*",job=~"^(executor|sourcegraph-code-intel-indexers|executor-batches).*"}[5m])))`

</details>

<br />

#### executor: apiworker_command_errors_total

<p class="subtitle">Aggregate command operation errors every 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/executor/executor?viewPanel=100402` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum(increase(src_apiworker_command_errors_total{op=~"setup.*",job=~"^(executor|sourcegraph-code-intel-indexers|executor-batches).*"}[5m]))`

</details>

<br />

#### executor: apiworker_command_error_rate

<p class="subtitle">Aggregate command operation error rate over 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/executor/executor?viewPanel=100403` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum(increase(src_apiworker_command_errors_total{op=~"setup.*",job=~"^(executor|sourcegraph-code-intel-indexers|executor-batches).*"}[5m])) / (sum(increase(src_apiworker_command_total{op=~"setup.*",job=~"^(executor|sourcegraph-code-intel-indexers|executor-batches).*"}[5m])) + sum(increase(src_apiworker_command_errors_total{op=~"setup.*",job=~"^(executor|sourcegraph-code-intel-indexers|executor-batches).*"}[5m]))) * 100`

</details>

<br />

#### executor: apiworker_command_total

<p class="subtitle">Command operations every 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/executor/executor?viewPanel=100410` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum by (op)(increase(src_apiworker_command_total{op=~"setup.*",job=~"^(executor|sourcegraph-code-intel-indexers|executor-batches).*"}[5m]))`

</details>

<br />

#### executor: apiworker_command_99th_percentile_duration

<p class="subtitle">99th percentile successful command operation duration over 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/executor/executor?viewPanel=100411` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `histogram_quantile(0.99, sum  by (le,op)(rate(src_apiworker_command_duration_seconds_bucket{op=~"setup.*",job=~"^(executor|sourcegraph-code-intel-indexers|executor-batches).*"}[5m])))`

</details>

<br />

#### executor: apiworker_command_errors_total

<p class="subtitle">Command operation errors every 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/executor/executor?viewPanel=100412` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum by (op)(increase(src_apiworker_command_errors_total{op=~"setup.*",job=~"^(executor|sourcegraph-code-intel-indexers|executor-batches).*"}[5m]))`

</details>

<br />

#### executor: apiworker_command_error_rate

<p class="subtitle">Command operation error rate over 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/executor/executor?viewPanel=100413` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum by (op)(increase(src_apiworker_command_errors_total{op=~"setup.*",job=~"^(executor|sourcegraph-code-intel-indexers|executor-batches).*"}[5m])) / (sum by (op)(increase(src_apiworker_command_total{op=~"setup.*",job=~"^(executor|sourcegraph-code-intel-indexers|executor-batches).*"}[5m])) + sum by (op)(increase(src_apiworker_command_errors_total{op=~"setup.*",job=~"^(executor|sourcegraph-code-intel-indexers|executor-batches).*"}[5m]))) * 100`

</details>

<br />

### Executor: Executor: Job execution

#### executor: apiworker_command_total

<p class="subtitle">Aggregate command operations every 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/executor/executor?viewPanel=100500` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum(increase(src_apiworker_command_total{op=~"exec.*",job=~"^(executor|sourcegraph-code-intel-indexers|executor-batches).*"}[5m]))`

</details>

<br />

#### executor: apiworker_command_99th_percentile_duration

<p class="subtitle">99th percentile successful aggregate command operation duration over 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/executor/executor?viewPanel=100501` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `histogram_quantile(0.99, sum  by (le)(rate(src_apiworker_command_duration_seconds_bucket{op=~"exec.*",job=~"^(executor|sourcegraph-code-intel-indexers|executor-batches).*"}[5m])))`

</details>

<br />

#### executor: apiworker_command_errors_total

<p class="subtitle">Aggregate command operation errors every 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/executor/executor?viewPanel=100502` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum(increase(src_apiworker_command_errors_total{op=~"exec.*",job=~"^(executor|sourcegraph-code-intel-indexers|executor-batches).*"}[5m]))`

</details>

<br />

#### executor: apiworker_command_error_rate

<p class="subtitle">Aggregate command operation error rate over 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/executor/executor?viewPanel=100503` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum(increase(src_apiworker_command_errors_total{op=~"exec.*",job=~"^(executor|sourcegraph-code-intel-indexers|executor-batches).*"}[5m])) / (sum(increase(src_apiworker_command_total{op=~"exec.*",job=~"^(executor|sourcegraph-code-intel-indexers|executor-batches).*"}[5m])) + sum(increase(src_apiworker_command_errors_total{op=~"exec.*",job=~"^(executor|sourcegraph-code-intel-indexers|executor-batches).*"}[5m]))) * 100`

</details>

<br />

#### executor: apiworker_command_total

<p class="subtitle">Command operations every 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/executor/executor?viewPanel=100510` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum by (op)(increase(src_apiworker_command_total{op=~"exec.*",job=~"^(executor|sourcegraph-code-intel-indexers|executor-batches).*"}[5m]))`

</details>

<br />

#### executor: apiworker_command_99th_percentile_duration

<p class="subtitle">99th percentile successful command operation duration over 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/executor/executor?viewPanel=100511` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `histogram_quantile(0.99, sum  by (le,op)(rate(src_apiworker_command_duration_seconds_bucket{op=~"exec.*",job=~"^(executor|sourcegraph-code-intel-indexers|executor-batches).*"}[5m])))`

</details>

<br />

#### executor: apiworker_command_errors_total

<p class="subtitle">Command operation errors every 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/executor/executor?viewPanel=100512` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum by (op)(increase(src_apiworker_command_errors_total{op=~"exec.*",job=~"^(executor|sourcegraph-code-intel-indexers|executor-batches).*"}[5m]))`

</details>

<br />

#### executor: apiworker_command_error_rate

<p class="subtitle">Command operation error rate over 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/executor/executor?viewPanel=100513` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum by (op)(increase(src_apiworker_command_errors_total{op=~"exec.*",job=~"^(executor|sourcegraph-code-intel-indexers|executor-batches).*"}[5m])) / (sum by (op)(increase(src_apiworker_command_total{op=~"exec.*",job=~"^(executor|sourcegraph-code-intel-indexers|executor-batches).*"}[5m])) + sum by (op)(increase(src_apiworker_command_errors_total{op=~"exec.*",job=~"^(executor|sourcegraph-code-intel-indexers|executor-batches).*"}[5m]))) * 100`

</details>

<br />

### Executor: Executor: Job teardown

#### executor: apiworker_command_total

<p class="subtitle">Aggregate command operations every 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/executor/executor?viewPanel=100600` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum(increase(src_apiworker_command_total{op=~"teardown.*",job=~"^(executor|sourcegraph-code-intel-indexers|executor-batches).*"}[5m]))`

</details>

<br />

#### executor: apiworker_command_99th_percentile_duration

<p class="subtitle">99th percentile successful aggregate command operation duration over 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/executor/executor?viewPanel=100601` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `histogram_quantile(0.99, sum  by (le)(rate(src_apiworker_command_duration_seconds_bucket{op=~"teardown.*",job=~"^(executor|sourcegraph-code-intel-indexers|executor-batches).*"}[5m])))`

</details>

<br />

#### executor: apiworker_command_errors_total

<p class="subtitle">Aggregate command operation errors every 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/executor/executor?viewPanel=100602` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum(increase(src_apiworker_command_errors_total{op=~"teardown.*",job=~"^(executor|sourcegraph-code-intel-indexers|executor-batches).*"}[5m]))`

</details>

<br />

#### executor: apiworker_command_error_rate

<p class="subtitle">Aggregate command operation error rate over 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/executor/executor?viewPanel=100603` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum(increase(src_apiworker_command_errors_total{op=~"teardown.*",job=~"^(executor|sourcegraph-code-intel-indexers|executor-batches).*"}[5m])) / (sum(increase(src_apiworker_command_total{op=~"teardown.*",job=~"^(executor|sourcegraph-code-intel-indexers|executor-batches).*"}[5m])) + sum(increase(src_apiworker_command_errors_total{op=~"teardown.*",job=~"^(executor|sourcegraph-code-intel-indexers|executor-batches).*"}[5m]))) * 100`

</details>

<br />

#### executor: apiworker_command_total

<p class="subtitle">Command operations every 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/executor/executor?viewPanel=100610` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum by (op)(increase(src_apiworker_command_total{op=~"teardown.*",job=~"^(executor|sourcegraph-code-intel-indexers|executor-batches).*"}[5m]))`

</details>

<br />

#### executor: apiworker_command_99th_percentile_duration

<p class="subtitle">99th percentile successful command operation duration over 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/executor/executor?viewPanel=100611` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `histogram_quantile(0.99, sum  by (le,op)(rate(src_apiworker_command_duration_seconds_bucket{op=~"teardown.*",job=~"^(executor|sourcegraph-code-intel-indexers|executor-batches).*"}[5m])))`

</details>

<br />

#### executor: apiworker_command_errors_total

<p class="subtitle">Command operation errors every 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/executor/executor?viewPanel=100612` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum by (op)(increase(src_apiworker_command_errors_total{op=~"teardown.*",job=~"^(executor|sourcegraph-code-intel-indexers|executor-batches).*"}[5m]))`

</details>

<br />

#### executor: apiworker_command_error_rate

<p class="subtitle">Command operation error rate over 5m</p>

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/executor/executor?viewPanel=100613` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum by (op)(increase(src_apiworker_command_errors_total{op=~"teardown.*",job=~"^(executor|sourcegraph-code-intel-indexers|executor-batches).*"}[5m])) / (sum by (op)(increase(src_apiworker_command_total{op=~"teardown.*",job=~"^(executor|sourcegraph-code-intel-indexers|executor-batches).*"}[5m])) + sum by (op)(increase(src_apiworker_command_errors_total{op=~"teardown.*",job=~"^(executor|sourcegraph-code-intel-indexers|executor-batches).*"}[5m]))) * 100`

</details>

<br />

### Executor: Container monitoring (not available on server)

#### executor: container_missing

<p class="subtitle">Container missing</p>

This value is the number of times a container has not been seen for more than one minute. If you observe this
value change independent of deployment events (such as an upgrade), it could indicate pods are being OOM killed or terminated for some other reasons.

- **Kubernetes:**
	- Determine if the pod was OOM killed using `kubectl describe pod (executor|sourcegraph-code-intel-indexers|executor-batches)` (look for `OOMKilled: true`) and, if so, consider increasing the memory limit in the relevant `Deployment.yaml`.
	- Check the logs before the container restarted to see if there are `panic:` messages or similar using `kubectl logs -p (executor|sourcegraph-code-intel-indexers|executor-batches)`.
- **Docker Compose:**
	- Determine if the pod was OOM killed using `docker inspect -f '{{json .State}}' (executor|sourcegraph-code-intel-indexers|executor-batches)` (look for `"OOMKilled":true`) and, if so, consider increasing the memory limit of the (executor|sourcegraph-code-intel-indexers|executor-batches) container in `docker-compose.yml`.
	- Check the logs before the container restarted to see if there are `panic:` messages or similar using `docker logs (executor|sourcegraph-code-intel-indexers|executor-batches)` (note this will include logs from the previous and currently running container).

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/executor/executor?viewPanel=100700` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `count by(name) ((time() - container_last_seen{name=~"^(executor|sourcegraph-code-intel-indexers|executor-batches).*"}) > 60)`

</details>

<br />

#### executor: container_cpu_usage

<p class="subtitle">Container cpu usage total (1m average) across all cores by instance</p>

Refer to the [alert solutions reference](./alert_solutions.md#executor-container-cpu-usage) for 1 alert related to this panel.

To see this panel, visit `/-/debug/grafana/d/executor/executor?viewPanel=100701` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `cadvisor_container_cpu_usage_percentage_total{name=~"^(executor|sourcegraph-code-intel-indexers|executor-batches).*"}`

</details>

<br />

#### executor: container_memory_usage

<p class="subtitle">Container memory usage by instance</p>

Refer to the [alert solutions reference](./alert_solutions.md#executor-container-memory-usage) for 1 alert related to this panel.

To see this panel, visit `/-/debug/grafana/d/executor/executor?viewPanel=100702` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `cadvisor_container_memory_usage_percentage_total{name=~"^(executor|sourcegraph-code-intel-indexers|executor-batches).*"}`

</details>

<br />

#### executor: fs_io_operations

<p class="subtitle">Filesystem reads and writes rate by instance over 1h</p>

This value indicates the number of filesystem read and write operations by containers of this service.
When extremely high, this can indicate a resource usage problem, or can cause problems with the service itself, especially if high values or spikes correlate with {{CONTAINER_NAME}} issues.

This panel has no related alerts.

To see this panel, visit `/-/debug/grafana/d/executor/executor?viewPanel=100703` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum by(name) (rate(container_fs_reads_total{name=~"^(executor|sourcegraph-code-intel-indexers|executor-batches).*"}[1h]) + rate(container_fs_writes_total{name=~"^(executor|sourcegraph-code-intel-indexers|executor-batches).*"}[1h]))`

</details>

<br />

### Executor: Provisioning indicators (not available on server)

#### executor: provisioning_container_cpu_usage_long_term

<p class="subtitle">Container cpu usage total (90th percentile over 1d) across all cores by instance</p>

Refer to the [alert solutions reference](./alert_solutions.md#executor-provisioning-container-cpu-usage-long-term) for 1 alert related to this panel.

To see this panel, visit `/-/debug/grafana/d/executor/executor?viewPanel=100800` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `quantile_over_time(0.9, cadvisor_container_cpu_usage_percentage_total{name=~"^(executor|sourcegraph-code-intel-indexers|executor-batches).*"}[1d])`

</details>

<br />

#### executor: provisioning_container_memory_usage_long_term

<p class="subtitle">Container memory usage (1d maximum) by instance</p>

Refer to the [alert solutions reference](./alert_solutions.md#executor-provisioning-container-memory-usage-long-term) for 1 alert related to this panel.

To see this panel, visit `/-/debug/grafana/d/executor/executor?viewPanel=100801` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `max_over_time(cadvisor_container_memory_usage_percentage_total{name=~"^(executor|sourcegraph-code-intel-indexers|executor-batches).*"}[1d])`

</details>

<br />

#### executor: provisioning_container_cpu_usage_short_term

<p class="subtitle">Container cpu usage total (5m maximum) across all cores by instance</p>

Refer to the [alert solutions reference](./alert_solutions.md#executor-provisioning-container-cpu-usage-short-term) for 1 alert related to this panel.

To see this panel, visit `/-/debug/grafana/d/executor/executor?viewPanel=100810` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `max_over_time(cadvisor_container_cpu_usage_percentage_total{name=~"^(executor|sourcegraph-code-intel-indexers|executor-batches).*"}[5m])`

</details>

<br />

#### executor: provisioning_container_memory_usage_short_term

<p class="subtitle">Container memory usage (5m maximum) by instance</p>

Refer to the [alert solutions reference](./alert_solutions.md#executor-provisioning-container-memory-usage-short-term) for 1 alert related to this panel.

To see this panel, visit `/-/debug/grafana/d/executor/executor?viewPanel=100811` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `max_over_time(cadvisor_container_memory_usage_percentage_total{name=~"^(executor|sourcegraph-code-intel-indexers|executor-batches).*"}[5m])`

</details>

<br />

### Executor: Golang runtime monitoring

#### executor: go_goroutines

<p class="subtitle">Maximum active goroutines</p>

A high value here indicates a possible goroutine leak.

Refer to the [alert solutions reference](./alert_solutions.md#executor-go-goroutines) for 1 alert related to this panel.

To see this panel, visit `/-/debug/grafana/d/executor/executor?viewPanel=100900` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `max by(instance) (go_goroutines{job=~".*(executor|sourcegraph-code-intel-indexers|executor-batches)"})`

</details>

<br />

#### executor: go_gc_duration_seconds

<p class="subtitle">Maximum go garbage collection duration</p>

Refer to the [alert solutions reference](./alert_solutions.md#executor-go-gc-duration-seconds) for 1 alert related to this panel.

To see this panel, visit `/-/debug/grafana/d/executor/executor?viewPanel=100901` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `max by(instance) (go_gc_duration_seconds{job=~".*(executor|sourcegraph-code-intel-indexers|executor-batches)"})`

</details>

<br />

### Executor: Kubernetes monitoring (only available on Kubernetes)

#### executor: pods_available_percentage

<p class="subtitle">Percentage pods available</p>

Refer to the [alert solutions reference](./alert_solutions.md#executor-pods-available-percentage) for 1 alert related to this panel.

To see this panel, visit `/-/debug/grafana/d/executor/executor?viewPanel=101000` on your Sourcegraph instance.

<sub>*Managed by the [Sourcegraph Code-intel team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<details>
<summary>Technical details</summary>

Query: `sum by(app) (up{app=~".*(executor|sourcegraph-code-intel-indexers|executor-batches)"}) / count by (app) (up{app=~".*(executor|sourcegraph-code-intel-indexers|executor-batches)"}) * 100`

</details>

<br />

