# Alert solutions

<!-- DO NOT EDIT: generated via: go generate ./monitoring -->

This document contains possible solutions for when you find alerts are firing in Sourcegraph's monitoring.
If your alert isn't mentioned here, or if the solution doesn't help, [contact us](mailto:support@sourcegraph.com) for assistance.

To learn more about Sourcegraph's alerting and how to set up alerts, see [our alerting guide](https://docs.sourcegraph.com/admin/observability/alerting).

## frontend: 99th_percentile_search_request_duration

<p class="subtitle">99th percentile successful search request duration over 5m</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> frontend: 20s+ 99th percentile successful search request duration over 5m

**Possible solutions**

- **Get details on the exact queries that are slow** by configuring `"observability.logSlowSearches": 20,` in the site configuration and looking for `frontend` warning logs prefixed with `slow search request` for additional details.
- **Check that most repositories are indexed** by visiting https://sourcegraph.example.com/site-admin/repositories?filter=needs-index (it should show few or no results.)
- **Kubernetes:** Check CPU usage of zoekt-webserver in the indexed-search pod, consider increasing CPU limits in the `indexed-search.Deployment.yaml` if regularly hitting max CPU utilization.
- **Docker Compose:** Check CPU usage on the Zoekt Web Server dashboard, consider increasing `cpus:` of the zoekt-webserver container in `docker-compose.yml` if regularly hitting max CPU utilization.
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_frontend_99th_percentile_search_request_duration"
]
```

<sub>*Managed by the [Sourcegraph Search team](https://about.sourcegraph.com/handbook/engineering/search).*</sub>

<br />

## frontend: 90th_percentile_search_request_duration

<p class="subtitle">90th percentile successful search request duration over 5m</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> frontend: 15s+ 90th percentile successful search request duration over 5m

**Possible solutions**

- **Get details on the exact queries that are slow** by configuring `"observability.logSlowSearches": 15,` in the site configuration and looking for `frontend` warning logs prefixed with `slow search request` for additional details.
- **Check that most repositories are indexed** by visiting https://sourcegraph.example.com/site-admin/repositories?filter=needs-index (it should show few or no results.)
- **Kubernetes:** Check CPU usage of zoekt-webserver in the indexed-search pod, consider increasing CPU limits in the `indexed-search.Deployment.yaml` if regularly hitting max CPU utilization.
- **Docker Compose:** Check CPU usage on the Zoekt Web Server dashboard, consider increasing `cpus:` of the zoekt-webserver container in `docker-compose.yml` if regularly hitting max CPU utilization.
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_frontend_90th_percentile_search_request_duration"
]
```

<sub>*Managed by the [Sourcegraph Search team](https://about.sourcegraph.com/handbook/engineering/search).*</sub>

<br />

## frontend: hard_timeout_search_responses

<p class="subtitle">hard timeout search responses every 5m</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> frontend: 2%+ hard timeout search responses every 5m for 15m0s
- <span class="badge badge-critical">critical</span> frontend: 5%+ hard timeout search responses every 5m for 15m0s

**Possible solutions**

- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_frontend_hard_timeout_search_responses",
  "critical_frontend_hard_timeout_search_responses"
]
```

<sub>*Managed by the [Sourcegraph Search team](https://about.sourcegraph.com/handbook/engineering/search).*</sub>

<br />

## frontend: hard_error_search_responses

<p class="subtitle">hard error search responses every 5m</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> frontend: 2%+ hard error search responses every 5m for 15m0s
- <span class="badge badge-critical">critical</span> frontend: 5%+ hard error search responses every 5m for 15m0s

**Possible solutions**

- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_frontend_hard_error_search_responses",
  "critical_frontend_hard_error_search_responses"
]
```

<sub>*Managed by the [Sourcegraph Search team](https://about.sourcegraph.com/handbook/engineering/search).*</sub>

<br />

## frontend: partial_timeout_search_responses

<p class="subtitle">partial timeout search responses every 5m</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> frontend: 5%+ partial timeout search responses every 5m for 15m0s

**Possible solutions**

- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_frontend_partial_timeout_search_responses"
]
```

<sub>*Managed by the [Sourcegraph Search team](https://about.sourcegraph.com/handbook/engineering/search).*</sub>

<br />

## frontend: search_alert_user_suggestions

<p class="subtitle">search alert user suggestions shown every 5m</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> frontend: 5%+ search alert user suggestions shown every 5m for 15m0s

**Possible solutions**

- This indicates your user`s are making syntax errors or similar user errors.
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_frontend_search_alert_user_suggestions"
]
```

<sub>*Managed by the [Sourcegraph Search team](https://about.sourcegraph.com/handbook/engineering/search).*</sub>

<br />

## frontend: page_load_latency

<p class="subtitle">90th percentile page load latency over all routes over 10m</p>

**Descriptions**

- <span class="badge badge-critical">critical</span> frontend: 2s+ 90th percentile page load latency over all routes over 10m

**Possible solutions**

- Confirm that the Sourcegraph frontend has enough CPU/memory using the provisioning panels.
- Trace a request to see what the slowest part is: https://docs.sourcegraph.com/admin/observability/tracing
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "critical_frontend_page_load_latency"
]
```

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<br />

## frontend: blob_load_latency

<p class="subtitle">90th percentile blob load latency over 10m</p>

**Descriptions**

- <span class="badge badge-critical">critical</span> frontend: 5s+ 90th percentile blob load latency over 10m

**Possible solutions**

- Confirm that the Sourcegraph frontend has enough CPU/memory using the provisioning panels.
- Trace a request to see what the slowest part is: https://docs.sourcegraph.com/admin/observability/tracing
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "critical_frontend_blob_load_latency"
]
```

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<br />

## frontend: 99th_percentile_search_codeintel_request_duration

<p class="subtitle">99th percentile code-intel successful search request duration over 5m</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> frontend: 20s+ 99th percentile code-intel successful search request duration over 5m

**Possible solutions**

- **Get details on the exact queries that are slow** by configuring `"observability.logSlowSearches": 20,` in the site configuration and looking for `frontend` warning logs prefixed with `slow search request` for additional details.
- **Check that most repositories are indexed** by visiting https://sourcegraph.example.com/site-admin/repositories?filter=needs-index (it should show few or no results.)
- **Kubernetes:** Check CPU usage of zoekt-webserver in the indexed-search pod, consider increasing CPU limits in the `indexed-search.Deployment.yaml` if regularly hitting max CPU utilization.
- **Docker Compose:** Check CPU usage on the Zoekt Web Server dashboard, consider increasing `cpus:` of the zoekt-webserver container in `docker-compose.yml` if regularly hitting max CPU utilization.
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_frontend_99th_percentile_search_codeintel_request_duration"
]
```

<sub>*Managed by the [Sourcegraph Code-intelligence team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<br />

## frontend: 90th_percentile_search_codeintel_request_duration

<p class="subtitle">90th percentile code-intel successful search request duration over 5m</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> frontend: 15s+ 90th percentile code-intel successful search request duration over 5m

**Possible solutions**

- **Get details on the exact queries that are slow** by configuring `"observability.logSlowSearches": 15,` in the site configuration and looking for `frontend` warning logs prefixed with `slow search request` for additional details.
- **Check that most repositories are indexed** by visiting https://sourcegraph.example.com/site-admin/repositories?filter=needs-index (it should show few or no results.)
- **Kubernetes:** Check CPU usage of zoekt-webserver in the indexed-search pod, consider increasing CPU limits in the `indexed-search.Deployment.yaml` if regularly hitting max CPU utilization.
- **Docker Compose:** Check CPU usage on the Zoekt Web Server dashboard, consider increasing `cpus:` of the zoekt-webserver container in `docker-compose.yml` if regularly hitting max CPU utilization.
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_frontend_90th_percentile_search_codeintel_request_duration"
]
```

<sub>*Managed by the [Sourcegraph Code-intelligence team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<br />

## frontend: hard_timeout_search_codeintel_responses

<p class="subtitle">hard timeout search code-intel responses every 5m</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> frontend: 2%+ hard timeout search code-intel responses every 5m for 15m0s
- <span class="badge badge-critical">critical</span> frontend: 5%+ hard timeout search code-intel responses every 5m for 15m0s

**Possible solutions**

- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_frontend_hard_timeout_search_codeintel_responses",
  "critical_frontend_hard_timeout_search_codeintel_responses"
]
```

<sub>*Managed by the [Sourcegraph Code-intelligence team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<br />

## frontend: hard_error_search_codeintel_responses

<p class="subtitle">hard error search code-intel responses every 5m</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> frontend: 2%+ hard error search code-intel responses every 5m for 15m0s
- <span class="badge badge-critical">critical</span> frontend: 5%+ hard error search code-intel responses every 5m for 15m0s

**Possible solutions**

- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_frontend_hard_error_search_codeintel_responses",
  "critical_frontend_hard_error_search_codeintel_responses"
]
```

<sub>*Managed by the [Sourcegraph Code-intelligence team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<br />

## frontend: partial_timeout_search_codeintel_responses

<p class="subtitle">partial timeout search code-intel responses every 5m</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> frontend: 5%+ partial timeout search code-intel responses every 5m for 15m0s

**Possible solutions**

- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_frontend_partial_timeout_search_codeintel_responses"
]
```

<sub>*Managed by the [Sourcegraph Code-intelligence team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<br />

## frontend: search_codeintel_alert_user_suggestions

<p class="subtitle">search code-intel alert user suggestions shown every 5m</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> frontend: 5%+ search code-intel alert user suggestions shown every 5m for 15m0s

**Possible solutions**

- This indicates a bug in Sourcegraph, please [open an issue](https://github.com/sourcegraph/sourcegraph/issues/new/choose).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_frontend_search_codeintel_alert_user_suggestions"
]
```

<sub>*Managed by the [Sourcegraph Code-intelligence team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<br />

## frontend: 99th_percentile_search_api_request_duration

<p class="subtitle">99th percentile successful search API request duration over 5m</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> frontend: 50s+ 99th percentile successful search API request duration over 5m

**Possible solutions**

- **Get details on the exact queries that are slow** by configuring `"observability.logSlowSearches": 20,` in the site configuration and looking for `frontend` warning logs prefixed with `slow search request` for additional details.
- **Check that most repositories are indexed** by visiting https://sourcegraph.example.com/site-admin/repositories?filter=needs-index (it should show few or no results.)
- **Kubernetes:** Check CPU usage of zoekt-webserver in the indexed-search pod, consider increasing CPU limits in the `indexed-search.Deployment.yaml` if regularly hitting max CPU utilization.
- **Docker Compose:** Check CPU usage on the Zoekt Web Server dashboard, consider increasing `cpus:` of the zoekt-webserver container in `docker-compose.yml` if regularly hitting max CPU utilization.
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_frontend_99th_percentile_search_api_request_duration"
]
```

<sub>*Managed by the [Sourcegraph Search team](https://about.sourcegraph.com/handbook/engineering/search).*</sub>

<br />

## frontend: 90th_percentile_search_api_request_duration

<p class="subtitle">90th percentile successful search API request duration over 5m</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> frontend: 40s+ 90th percentile successful search API request duration over 5m

**Possible solutions**

- **Get details on the exact queries that are slow** by configuring `"observability.logSlowSearches": 15,` in the site configuration and looking for `frontend` warning logs prefixed with `slow search request` for additional details.
- **Check that most repositories are indexed** by visiting https://sourcegraph.example.com/site-admin/repositories?filter=needs-index (it should show few or no results.)
- **Kubernetes:** Check CPU usage of zoekt-webserver in the indexed-search pod, consider increasing CPU limits in the `indexed-search.Deployment.yaml` if regularly hitting max CPU utilization.
- **Docker Compose:** Check CPU usage on the Zoekt Web Server dashboard, consider increasing `cpus:` of the zoekt-webserver container in `docker-compose.yml` if regularly hitting max CPU utilization.
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_frontend_90th_percentile_search_api_request_duration"
]
```

<sub>*Managed by the [Sourcegraph Search team](https://about.sourcegraph.com/handbook/engineering/search).*</sub>

<br />

## frontend: hard_timeout_search_api_responses

<p class="subtitle">hard timeout search API responses every 5m</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> frontend: 2%+ hard timeout search API responses every 5m for 15m0s
- <span class="badge badge-critical">critical</span> frontend: 5%+ hard timeout search API responses every 5m for 15m0s

**Possible solutions**

- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_frontend_hard_timeout_search_api_responses",
  "critical_frontend_hard_timeout_search_api_responses"
]
```

<sub>*Managed by the [Sourcegraph Search team](https://about.sourcegraph.com/handbook/engineering/search).*</sub>

<br />

## frontend: hard_error_search_api_responses

<p class="subtitle">hard error search API responses every 5m</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> frontend: 2%+ hard error search API responses every 5m for 15m0s
- <span class="badge badge-critical">critical</span> frontend: 5%+ hard error search API responses every 5m for 15m0s

**Possible solutions**

- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_frontend_hard_error_search_api_responses",
  "critical_frontend_hard_error_search_api_responses"
]
```

<sub>*Managed by the [Sourcegraph Search team](https://about.sourcegraph.com/handbook/engineering/search).*</sub>

<br />

## frontend: partial_timeout_search_api_responses

<p class="subtitle">partial timeout search API responses every 5m</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> frontend: 5%+ partial timeout search API responses every 5m for 15m0s

**Possible solutions**

- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_frontend_partial_timeout_search_api_responses"
]
```

<sub>*Managed by the [Sourcegraph Search team](https://about.sourcegraph.com/handbook/engineering/search).*</sub>

<br />

## frontend: search_api_alert_user_suggestions

<p class="subtitle">search API alert user suggestions shown every 5m</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> frontend: 5%+ search API alert user suggestions shown every 5m

**Possible solutions**

- This indicates your user`s search API requests have syntax errors or a similar user error. Check the responses the API sends back for an explanation.
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_frontend_search_api_alert_user_suggestions"
]
```

<sub>*Managed by the [Sourcegraph Search team](https://about.sourcegraph.com/handbook/engineering/search).*</sub>

<br />

## frontend: internal_indexed_search_error_responses

<p class="subtitle">internal indexed search error responses every 5m</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> frontend: 5%+ internal indexed search error responses every 5m for 15m0s

**Possible solutions**

- Check the Zoekt Web Server dashboard for indications it might be unhealthy.
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_frontend_internal_indexed_search_error_responses"
]
```

<sub>*Managed by the [Sourcegraph Search team](https://about.sourcegraph.com/handbook/engineering/search).*</sub>

<br />

## frontend: internal_unindexed_search_error_responses

<p class="subtitle">internal unindexed search error responses every 5m</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> frontend: 5%+ internal unindexed search error responses every 5m for 15m0s

**Possible solutions**

- Check the Searcher dashboard for indications it might be unhealthy.
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_frontend_internal_unindexed_search_error_responses"
]
```

<sub>*Managed by the [Sourcegraph Search team](https://about.sourcegraph.com/handbook/engineering/search).*</sub>

<br />

## frontend: internal_api_error_responses

<p class="subtitle">internal API error responses every 5m by route</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> frontend: 5%+ internal API error responses every 5m by route for 15m0s

**Possible solutions**

- May not be a substantial issue, check the `frontend` logs for potential causes.
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_frontend_internal_api_error_responses"
]
```

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<br />

## frontend: 99th_percentile_gitserver_duration

<p class="subtitle">99th percentile successful gitserver query duration over 5m</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> frontend: 20s+ 99th percentile successful gitserver query duration over 5m

**Possible solutions**

- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_frontend_99th_percentile_gitserver_duration"
]
```

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<br />

## frontend: gitserver_error_responses

<p class="subtitle">gitserver error responses every 5m</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> frontend: 5%+ gitserver error responses every 5m for 15m0s

**Possible solutions**

- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_frontend_gitserver_error_responses"
]
```

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<br />

## frontend: observability_test_alert_warning

<p class="subtitle">warning test alert metric</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> frontend: 1+ warning test alert metric

**Possible solutions**

- This alert is triggered via the `triggerObservabilityTestAlert` GraphQL endpoint, and will automatically resolve itself.
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_frontend_observability_test_alert_warning"
]
```

<sub>*Managed by the [Sourcegraph Distribution team](https://about.sourcegraph.com/handbook/engineering/distribution).*</sub>

<br />

## frontend: observability_test_alert_critical

<p class="subtitle">critical test alert metric</p>

**Descriptions**

- <span class="badge badge-critical">critical</span> frontend: 1+ critical test alert metric

**Possible solutions**

- This alert is triggered via the `triggerObservabilityTestAlert` GraphQL endpoint, and will automatically resolve itself.
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "critical_frontend_observability_test_alert_critical"
]
```

<sub>*Managed by the [Sourcegraph Distribution team](https://about.sourcegraph.com/handbook/engineering/distribution).*</sub>

<br />

## frontend: mean_blocked_seconds_per_conn_request

<p class="subtitle">mean blocked seconds per conn request</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> frontend: 0.05s+ mean blocked seconds per conn request for 5m0s
- <span class="badge badge-critical">critical</span> frontend: 0.1s+ mean blocked seconds per conn request for 10m0s

**Possible solutions**

- Increase SRC_PGSQL_MAX_OPEN together with giving more memory to the database if needed
- Scale up Postgres memory / cpus [See our scaling guide](https://docs.sourcegraph.com/admin/config/postgres-conf)
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_frontend_mean_blocked_seconds_per_conn_request",
  "critical_frontend_mean_blocked_seconds_per_conn_request"
]
```

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<br />

## frontend: container_cpu_usage

<p class="subtitle">container cpu usage total (1m average) across all cores by instance</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> frontend: 99%+ container cpu usage total (1m average) across all cores by instance

**Possible solutions**

- **Kubernetes:** Consider increasing CPU limits in the the relevant `Deployment.yaml`.
- **Docker Compose:** Consider increasing `cpus:` of the (frontend|sourcegraph-frontend) container in `docker-compose.yml`.
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_frontend_container_cpu_usage"
]
```

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<br />

## frontend: container_memory_usage

<p class="subtitle">container memory usage by instance</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> frontend: 99%+ container memory usage by instance

**Possible solutions**

- **Kubernetes:** Consider increasing memory limit in relevant `Deployment.yaml`.
- **Docker Compose:** Consider increasing `memory:` of (frontend|sourcegraph-frontend) container in `docker-compose.yml`.
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_frontend_container_memory_usage"
]
```

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<br />

## frontend: provisioning_container_cpu_usage_long_term

<p class="subtitle">container cpu usage total (90th percentile over 1d) across all cores by instance</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> frontend: 80%+ container cpu usage total (90th percentile over 1d) across all cores by instance for 336h0m0s

**Possible solutions**

- **Kubernetes:** Consider increasing CPU limits in the `Deployment.yaml` for the (frontend|sourcegraph-frontend) service.
- **Docker Compose:** Consider increasing `cpus:` of the (frontend|sourcegraph-frontend) container in `docker-compose.yml`.
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_frontend_provisioning_container_cpu_usage_long_term"
]
```

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<br />

## frontend: provisioning_container_memory_usage_long_term

<p class="subtitle">container memory usage (1d maximum) by instance</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> frontend: 80%+ container memory usage (1d maximum) by instance for 336h0m0s

**Possible solutions**

- **Kubernetes:** Consider increasing memory limits in the `Deployment.yaml` for the (frontend|sourcegraph-frontend) service.
- **Docker Compose:** Consider increasing `memory:` of the (frontend|sourcegraph-frontend) container in `docker-compose.yml`.
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_frontend_provisioning_container_memory_usage_long_term"
]
```

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<br />

## frontend: provisioning_container_cpu_usage_short_term

<p class="subtitle">container cpu usage total (5m maximum) across all cores by instance</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> frontend: 90%+ container cpu usage total (5m maximum) across all cores by instance for 30m0s

**Possible solutions**

- **Kubernetes:** Consider increasing CPU limits in the the relevant `Deployment.yaml`.
- **Docker Compose:** Consider increasing `cpus:` of the (frontend|sourcegraph-frontend) container in `docker-compose.yml`.
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_frontend_provisioning_container_cpu_usage_short_term"
]
```

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<br />

## frontend: provisioning_container_memory_usage_short_term

<p class="subtitle">container memory usage (5m maximum) by instance</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> frontend: 90%+ container memory usage (5m maximum) by instance

**Possible solutions**

- **Kubernetes:** Consider increasing memory limit in relevant `Deployment.yaml`.
- **Docker Compose:** Consider increasing `memory:` of (frontend|sourcegraph-frontend) container in `docker-compose.yml`.
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_frontend_provisioning_container_memory_usage_short_term"
]
```

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<br />

## frontend: go_goroutines

<p class="subtitle">maximum active goroutines</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> frontend: 10000+ maximum active goroutines for 10m0s

**Possible solutions**

- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_frontend_go_goroutines"
]
```

> NOTE: More help interpreting this metric is available in the [dashboards reference](./dashboards.md#frontend-go-goroutines).

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<br />

## frontend: go_gc_duration_seconds

<p class="subtitle">maximum go garbage collection duration</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> frontend: 2s+ maximum go garbage collection duration

**Possible solutions**

- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_frontend_go_gc_duration_seconds"
]
```

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<br />

## frontend: pods_available_percentage

<p class="subtitle">percentage pods available</p>

**Descriptions**

- <span class="badge badge-critical">critical</span> frontend: less than 90% percentage pods available for 10m0s

**Possible solutions**

- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "critical_frontend_pods_available_percentage"
]
```

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<br />

## frontend: mean_successful_sentinel_duration_5m

<p class="subtitle">mean successful sentinel search duration over 5m</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> frontend: 5s+ mean successful sentinel search duration over 5m for 15m0s
- <span class="badge badge-critical">critical</span> frontend: 8s+ mean successful sentinel search duration over 5m for 30m0s

**Possible solutions**

- Look at the breakdown by query to determine if a specific query type is being affected
- Check for high CPU usage on zoekt-webserver
- Check Honeycomb for unusual activity
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_frontend_mean_successful_sentinel_duration_5m",
  "critical_frontend_mean_successful_sentinel_duration_5m"
]
```

<sub>*Managed by the [Sourcegraph Search team](https://about.sourcegraph.com/handbook/engineering/search).*</sub>

<br />

## frontend: mean_sentinel_stream_latency_5m

<p class="subtitle">mean sentinel stream latency over 5m</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> frontend: 2s+ mean sentinel stream latency over 5m for 15m0s
- <span class="badge badge-critical">critical</span> frontend: 3s+ mean sentinel stream latency over 5m for 30m0s

**Possible solutions**

- Look at the breakdown by query to determine if a specific query type is being affected
- Check for high CPU usage on zoekt-webserver
- Check Honeycomb for unusual activity
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_frontend_mean_sentinel_stream_latency_5m",
  "critical_frontend_mean_sentinel_stream_latency_5m"
]
```

<sub>*Managed by the [Sourcegraph Search team](https://about.sourcegraph.com/handbook/engineering/search).*</sub>

<br />

## frontend: 90th_percentile_successful_sentinel_duration_5m

<p class="subtitle">90th percentile successful sentinel search duration over 5m</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> frontend: 5s+ 90th percentile successful sentinel search duration over 5m for 15m0s
- <span class="badge badge-critical">critical</span> frontend: 10s+ 90th percentile successful sentinel search duration over 5m for 30m0s

**Possible solutions**

- Look at the breakdown by query to determine if a specific query type is being affected
- Check for high CPU usage on zoekt-webserver
- Check Honeycomb for unusual activity
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_frontend_90th_percentile_successful_sentinel_duration_5m",
  "critical_frontend_90th_percentile_successful_sentinel_duration_5m"
]
```

<sub>*Managed by the [Sourcegraph Search team](https://about.sourcegraph.com/handbook/engineering/search).*</sub>

<br />

## frontend: 90th_percentile_sentinel_stream_latency_5m

<p class="subtitle">90th percentile sentinel stream latency over 5m</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> frontend: 4s+ 90th percentile sentinel stream latency over 5m for 15m0s
- <span class="badge badge-critical">critical</span> frontend: 6s+ 90th percentile sentinel stream latency over 5m for 30m0s

**Possible solutions**

- Look at the breakdown by query to determine if a specific query type is being affected
- Check for high CPU usage on zoekt-webserver
- Check Honeycomb for unusual activity
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_frontend_90th_percentile_sentinel_stream_latency_5m",
  "critical_frontend_90th_percentile_sentinel_stream_latency_5m"
]
```

<sub>*Managed by the [Sourcegraph Search team](https://about.sourcegraph.com/handbook/engineering/search).*</sub>

<br />

## gitserver: disk_space_remaining

<p class="subtitle">disk space remaining by instance</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> gitserver: less than 25% disk space remaining by instance
- <span class="badge badge-critical">critical</span> gitserver: less than 15% disk space remaining by instance

**Possible solutions**

- **Provision more disk space:** Sourcegraph will begin deleting least-used repository clones at 10% disk space remaining which may result in decreased performance, users having to wait for repositories to clone, etc.
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_gitserver_disk_space_remaining",
  "critical_gitserver_disk_space_remaining"
]
```

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<br />

## gitserver: running_git_commands

<p class="subtitle">git commands running on each gitserver instance</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> gitserver: 50+ git commands running on each gitserver instance for 2m0s
- <span class="badge badge-critical">critical</span> gitserver: 100+ git commands running on each gitserver instance for 5m0s

**Possible solutions**

- **Check if the problem may be an intermittent and temporary peak** using the "Container monitoring" section at the bottom of the Git Server dashboard.
- **Single container deployments:** Consider upgrading to a [Docker Compose deployment](../install/docker-compose/migrate.md) which offers better scalability and resource isolation.
- **Kubernetes and Docker Compose:** Check that you are running a similar number of git server replicas and that their CPU/memory limits are allocated according to what is shown in the [Sourcegraph resource estimator](../install/resource_estimator.md).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_gitserver_running_git_commands",
  "critical_gitserver_running_git_commands"
]
```

> NOTE: More help interpreting this metric is available in the [dashboards reference](./dashboards.md#gitserver-running-git-commands).

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<br />

## gitserver: repository_clone_queue_size

<p class="subtitle">repository clone queue size</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> gitserver: 25+ repository clone queue size

**Possible solutions**

- **If you just added several repositories**, the warning may be expected.
- **Check which repositories need cloning**, by visiting e.g. https://sourcegraph.example.com/site-admin/repositories?filter=not-cloned
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_gitserver_repository_clone_queue_size"
]
```

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<br />

## gitserver: repository_existence_check_queue_size

<p class="subtitle">repository existence check queue size</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> gitserver: 25+ repository existence check queue size

**Possible solutions**

- **Check the code host status indicator for errors:** on the Sourcegraph app homepage, when signed in as an admin click the cloud icon in the top right corner of the page.
- **Check if the issue continues to happen after 30 minutes**, it may be temporary.
- **Check the gitserver logs for more information.**
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_gitserver_repository_existence_check_queue_size"
]
```

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<br />

## gitserver: frontend_internal_api_error_responses

<p class="subtitle">frontend-internal API error responses every 5m by route</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> gitserver: 2%+ frontend-internal API error responses every 5m by route for 5m0s

**Possible solutions**

- **Single-container deployments:** Check `docker logs $CONTAINER_ID` for logs starting with `repo-updater` that indicate requests to the frontend service are failing.
- **Kubernetes:**
	- Confirm that `kubectl get pods` shows the `frontend` pods are healthy.
	- Check `kubectl logs gitserver` for logs indicate request failures to `frontend` or `frontend-internal`.
- **Docker Compose:**
	- Confirm that `docker ps` shows the `frontend-internal` container is healthy.
	- Check `docker logs gitserver` for logs indicating request failures to `frontend` or `frontend-internal`.
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_gitserver_frontend_internal_api_error_responses"
]
```

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<br />

## gitserver: mean_blocked_seconds_per_conn_request

<p class="subtitle">mean blocked seconds per conn request</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> gitserver: 0.05s+ mean blocked seconds per conn request for 5m0s
- <span class="badge badge-critical">critical</span> gitserver: 0.1s+ mean blocked seconds per conn request for 10m0s

**Possible solutions**

- Increase SRC_PGSQL_MAX_OPEN together with giving more memory to the database if needed
- Scale up Postgres memory / cpus [See our scaling guide](https://docs.sourcegraph.com/admin/config/postgres-conf)
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_gitserver_mean_blocked_seconds_per_conn_request",
  "critical_gitserver_mean_blocked_seconds_per_conn_request"
]
```

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<br />

## gitserver: container_cpu_usage

<p class="subtitle">container cpu usage total (1m average) across all cores by instance</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> gitserver: 99%+ container cpu usage total (1m average) across all cores by instance

**Possible solutions**

- **Kubernetes:** Consider increasing CPU limits in the the relevant `Deployment.yaml`.
- **Docker Compose:** Consider increasing `cpus:` of the gitserver container in `docker-compose.yml`.
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_gitserver_container_cpu_usage"
]
```

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<br />

## gitserver: container_memory_usage

<p class="subtitle">container memory usage by instance</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> gitserver: 99%+ container memory usage by instance

**Possible solutions**

- **Kubernetes:** Consider increasing memory limit in relevant `Deployment.yaml`.
- **Docker Compose:** Consider increasing `memory:` of gitserver container in `docker-compose.yml`.
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_gitserver_container_memory_usage"
]
```

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<br />

## gitserver: provisioning_container_cpu_usage_long_term

<p class="subtitle">container cpu usage total (90th percentile over 1d) across all cores by instance</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> gitserver: 80%+ container cpu usage total (90th percentile over 1d) across all cores by instance for 336h0m0s

**Possible solutions**

- **Kubernetes:** Consider increasing CPU limits in the `Deployment.yaml` for the gitserver service.
- **Docker Compose:** Consider increasing `cpus:` of the gitserver container in `docker-compose.yml`.
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_gitserver_provisioning_container_cpu_usage_long_term"
]
```

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<br />

## gitserver: provisioning_container_cpu_usage_short_term

<p class="subtitle">container cpu usage total (5m maximum) across all cores by instance</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> gitserver: 90%+ container cpu usage total (5m maximum) across all cores by instance for 30m0s

**Possible solutions**

- **Kubernetes:** Consider increasing CPU limits in the the relevant `Deployment.yaml`.
- **Docker Compose:** Consider increasing `cpus:` of the gitserver container in `docker-compose.yml`.
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_gitserver_provisioning_container_cpu_usage_short_term"
]
```

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<br />

## gitserver: go_goroutines

<p class="subtitle">maximum active goroutines</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> gitserver: 10000+ maximum active goroutines for 10m0s

**Possible solutions**

- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_gitserver_go_goroutines"
]
```

> NOTE: More help interpreting this metric is available in the [dashboards reference](./dashboards.md#gitserver-go-goroutines).

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<br />

## gitserver: go_gc_duration_seconds

<p class="subtitle">maximum go garbage collection duration</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> gitserver: 2s+ maximum go garbage collection duration

**Possible solutions**

- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_gitserver_go_gc_duration_seconds"
]
```

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<br />

## gitserver: pods_available_percentage

<p class="subtitle">percentage pods available</p>

**Descriptions**

- <span class="badge badge-critical">critical</span> gitserver: less than 90% percentage pods available for 10m0s

**Possible solutions**

- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "critical_gitserver_pods_available_percentage"
]
```

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<br />

## github-proxy: github_proxy_waiting_requests

<p class="subtitle">number of requests waiting on the global mutex</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> github-proxy: 100+ number of requests waiting on the global mutex for 5m0s

**Possible solutions**

- 								- **Check github-proxy logs for network connection issues.
								- **Check github status.
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_github-proxy_github_proxy_waiting_requests"
]
```

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<br />

## github-proxy: container_cpu_usage

<p class="subtitle">container cpu usage total (1m average) across all cores by instance</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> github-proxy: 99%+ container cpu usage total (1m average) across all cores by instance

**Possible solutions**

- **Kubernetes:** Consider increasing CPU limits in the the relevant `Deployment.yaml`.
- **Docker Compose:** Consider increasing `cpus:` of the github-proxy container in `docker-compose.yml`.
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_github-proxy_container_cpu_usage"
]
```

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<br />

## github-proxy: container_memory_usage

<p class="subtitle">container memory usage by instance</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> github-proxy: 99%+ container memory usage by instance

**Possible solutions**

- **Kubernetes:** Consider increasing memory limit in relevant `Deployment.yaml`.
- **Docker Compose:** Consider increasing `memory:` of github-proxy container in `docker-compose.yml`.
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_github-proxy_container_memory_usage"
]
```

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<br />

## github-proxy: provisioning_container_cpu_usage_long_term

<p class="subtitle">container cpu usage total (90th percentile over 1d) across all cores by instance</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> github-proxy: 80%+ container cpu usage total (90th percentile over 1d) across all cores by instance for 336h0m0s

**Possible solutions**

- **Kubernetes:** Consider increasing CPU limits in the `Deployment.yaml` for the github-proxy service.
- **Docker Compose:** Consider increasing `cpus:` of the github-proxy container in `docker-compose.yml`.
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_github-proxy_provisioning_container_cpu_usage_long_term"
]
```

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<br />

## github-proxy: provisioning_container_memory_usage_long_term

<p class="subtitle">container memory usage (1d maximum) by instance</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> github-proxy: 80%+ container memory usage (1d maximum) by instance for 336h0m0s

**Possible solutions**

- **Kubernetes:** Consider increasing memory limits in the `Deployment.yaml` for the github-proxy service.
- **Docker Compose:** Consider increasing `memory:` of the github-proxy container in `docker-compose.yml`.
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_github-proxy_provisioning_container_memory_usage_long_term"
]
```

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<br />

## github-proxy: provisioning_container_cpu_usage_short_term

<p class="subtitle">container cpu usage total (5m maximum) across all cores by instance</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> github-proxy: 90%+ container cpu usage total (5m maximum) across all cores by instance for 30m0s

**Possible solutions**

- **Kubernetes:** Consider increasing CPU limits in the the relevant `Deployment.yaml`.
- **Docker Compose:** Consider increasing `cpus:` of the github-proxy container in `docker-compose.yml`.
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_github-proxy_provisioning_container_cpu_usage_short_term"
]
```

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<br />

## github-proxy: provisioning_container_memory_usage_short_term

<p class="subtitle">container memory usage (5m maximum) by instance</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> github-proxy: 90%+ container memory usage (5m maximum) by instance

**Possible solutions**

- **Kubernetes:** Consider increasing memory limit in relevant `Deployment.yaml`.
- **Docker Compose:** Consider increasing `memory:` of github-proxy container in `docker-compose.yml`.
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_github-proxy_provisioning_container_memory_usage_short_term"
]
```

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<br />

## github-proxy: go_goroutines

<p class="subtitle">maximum active goroutines</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> github-proxy: 10000+ maximum active goroutines for 10m0s

**Possible solutions**

- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_github-proxy_go_goroutines"
]
```

> NOTE: More help interpreting this metric is available in the [dashboards reference](./dashboards.md#github-proxy-go-goroutines).

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<br />

## github-proxy: go_gc_duration_seconds

<p class="subtitle">maximum go garbage collection duration</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> github-proxy: 2s+ maximum go garbage collection duration

**Possible solutions**

- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_github-proxy_go_gc_duration_seconds"
]
```

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<br />

## github-proxy: pods_available_percentage

<p class="subtitle">percentage pods available</p>

**Descriptions**

- <span class="badge badge-critical">critical</span> github-proxy: less than 90% percentage pods available for 10m0s

**Possible solutions**

- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "critical_github-proxy_pods_available_percentage"
]
```

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<br />

## postgres: connections

<p class="subtitle">active connections</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> postgres: less than 5 active connections for 5m0s

**Possible solutions**

- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_postgres_connections"
]
```

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<br />

## postgres: transaction_durations

<p class="subtitle">maximum transaction durations</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> postgres: 300ms+ maximum transaction durations for 5m0s
- <span class="badge badge-critical">critical</span> postgres: 500ms+ maximum transaction durations for 10m0s

**Possible solutions**

- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_postgres_transaction_durations",
  "critical_postgres_transaction_durations"
]
```

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<br />

## postgres: postgres_up

<p class="subtitle">database availability</p>

**Descriptions**

- <span class="badge badge-critical">critical</span> postgres: less than 0 database availability for 5m0s

**Possible solutions**

- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "critical_postgres_postgres_up"
]
```

> NOTE: More help interpreting this metric is available in the [dashboards reference](./dashboards.md#postgres-postgres-up).

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<br />

## postgres: invalid_indexes

<p class="subtitle">invalid indexes (unusable by the query planner)</p>

**Descriptions**

- <span class="badge badge-critical">critical</span> postgres: 1+ invalid indexes (unusable by the query planner)

**Possible solutions**

- Drop and re-create the invalid trigger - please contact Sourcegraph to supply the trigger definition.
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "critical_postgres_invalid_indexes"
]
```

> NOTE: More help interpreting this metric is available in the [dashboards reference](./dashboards.md#postgres-invalid-indexes).

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<br />

## postgres: pg_exporter_err

<p class="subtitle">errors scraping postgres exporter</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> postgres: 1+ errors scraping postgres exporter for 5m0s

**Possible solutions**

- Ensure the Postgres exporter can access the Postgres database. Also, check the Postgres exporter logs for errors.
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_postgres_pg_exporter_err"
]
```

> NOTE: More help interpreting this metric is available in the [dashboards reference](./dashboards.md#postgres-pg-exporter-err).

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<br />

## postgres: migration_in_progress

<p class="subtitle">active schema migration</p>

**Descriptions**

- <span class="badge badge-critical">critical</span> postgres: 1+ active schema migration for 5m0s

**Possible solutions**

- The database migration has been in progress for 5 or more minutes - please contact Sourcegraph if this persists.
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "critical_postgres_migration_in_progress"
]
```

> NOTE: More help interpreting this metric is available in the [dashboards reference](./dashboards.md#postgres-migration-in-progress).

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<br />

## postgres: provisioning_container_cpu_usage_long_term

<p class="subtitle">container cpu usage total (90th percentile over 1d) across all cores by instance</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> postgres: 80%+ container cpu usage total (90th percentile over 1d) across all cores by instance for 336h0m0s

**Possible solutions**

- **Kubernetes:** Consider increasing CPU limits in the `Deployment.yaml` for the (pgsql|codeintel-db) service.
- **Docker Compose:** Consider increasing `cpus:` of the (pgsql|codeintel-db) container in `docker-compose.yml`.
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_postgres_provisioning_container_cpu_usage_long_term"
]
```

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<br />

## postgres: provisioning_container_memory_usage_long_term

<p class="subtitle">container memory usage (1d maximum) by instance</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> postgres: 80%+ container memory usage (1d maximum) by instance for 336h0m0s

**Possible solutions**

- **Kubernetes:** Consider increasing memory limits in the `Deployment.yaml` for the (pgsql|codeintel-db) service.
- **Docker Compose:** Consider increasing `memory:` of the (pgsql|codeintel-db) container in `docker-compose.yml`.
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_postgres_provisioning_container_memory_usage_long_term"
]
```

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<br />

## postgres: provisioning_container_cpu_usage_short_term

<p class="subtitle">container cpu usage total (5m maximum) across all cores by instance</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> postgres: 90%+ container cpu usage total (5m maximum) across all cores by instance for 30m0s

**Possible solutions**

- **Kubernetes:** Consider increasing CPU limits in the the relevant `Deployment.yaml`.
- **Docker Compose:** Consider increasing `cpus:` of the (pgsql|codeintel-db) container in `docker-compose.yml`.
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_postgres_provisioning_container_cpu_usage_short_term"
]
```

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<br />

## postgres: provisioning_container_memory_usage_short_term

<p class="subtitle">container memory usage (5m maximum) by instance</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> postgres: 90%+ container memory usage (5m maximum) by instance

**Possible solutions**

- **Kubernetes:** Consider increasing memory limit in relevant `Deployment.yaml`.
- **Docker Compose:** Consider increasing `memory:` of (pgsql|codeintel-db) container in `docker-compose.yml`.
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_postgres_provisioning_container_memory_usage_short_term"
]
```

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<br />

## postgres: pods_available_percentage

<p class="subtitle">percentage pods available</p>

**Descriptions**

- <span class="badge badge-critical">critical</span> postgres: less than 90% percentage pods available for 10m0s

**Possible solutions**

- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "critical_postgres_pods_available_percentage"
]
```

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<br />

## precise-code-intel-worker: frontend_internal_api_error_responses

<p class="subtitle">frontend-internal API error responses every 5m by route</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> precise-code-intel-worker: 2%+ frontend-internal API error responses every 5m by route for 5m0s

**Possible solutions**

- **Single-container deployments:** Check `docker logs $CONTAINER_ID` for logs starting with `repo-updater` that indicate requests to the frontend service are failing.
- **Kubernetes:**
	- Confirm that `kubectl get pods` shows the `frontend` pods are healthy.
	- Check `kubectl logs precise-code-intel-worker` for logs indicate request failures to `frontend` or `frontend-internal`.
- **Docker Compose:**
	- Confirm that `docker ps` shows the `frontend-internal` container is healthy.
	- Check `docker logs precise-code-intel-worker` for logs indicating request failures to `frontend` or `frontend-internal`.
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_precise-code-intel-worker_frontend_internal_api_error_responses"
]
```

<sub>*Managed by the [Sourcegraph Code-intelligence team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<br />

## precise-code-intel-worker: mean_blocked_seconds_per_conn_request

<p class="subtitle">mean blocked seconds per conn request</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> precise-code-intel-worker: 0.05s+ mean blocked seconds per conn request for 5m0s
- <span class="badge badge-critical">critical</span> precise-code-intel-worker: 0.1s+ mean blocked seconds per conn request for 10m0s

**Possible solutions**

- Increase SRC_PGSQL_MAX_OPEN together with giving more memory to the database if needed
- Scale up Postgres memory / cpus [See our scaling guide](https://docs.sourcegraph.com/admin/config/postgres-conf)
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_precise-code-intel-worker_mean_blocked_seconds_per_conn_request",
  "critical_precise-code-intel-worker_mean_blocked_seconds_per_conn_request"
]
```

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<br />

## precise-code-intel-worker: container_cpu_usage

<p class="subtitle">container cpu usage total (1m average) across all cores by instance</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> precise-code-intel-worker: 99%+ container cpu usage total (1m average) across all cores by instance

**Possible solutions**

- **Kubernetes:** Consider increasing CPU limits in the the relevant `Deployment.yaml`.
- **Docker Compose:** Consider increasing `cpus:` of the precise-code-intel-worker container in `docker-compose.yml`.
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_precise-code-intel-worker_container_cpu_usage"
]
```

<sub>*Managed by the [Sourcegraph Code-intelligence team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<br />

## precise-code-intel-worker: container_memory_usage

<p class="subtitle">container memory usage by instance</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> precise-code-intel-worker: 99%+ container memory usage by instance

**Possible solutions**

- **Kubernetes:** Consider increasing memory limit in relevant `Deployment.yaml`.
- **Docker Compose:** Consider increasing `memory:` of precise-code-intel-worker container in `docker-compose.yml`.
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_precise-code-intel-worker_container_memory_usage"
]
```

<sub>*Managed by the [Sourcegraph Code-intelligence team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<br />

## precise-code-intel-worker: provisioning_container_cpu_usage_long_term

<p class="subtitle">container cpu usage total (90th percentile over 1d) across all cores by instance</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> precise-code-intel-worker: 80%+ container cpu usage total (90th percentile over 1d) across all cores by instance for 336h0m0s

**Possible solutions**

- **Kubernetes:** Consider increasing CPU limits in the `Deployment.yaml` for the precise-code-intel-worker service.
- **Docker Compose:** Consider increasing `cpus:` of the precise-code-intel-worker container in `docker-compose.yml`.
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_precise-code-intel-worker_provisioning_container_cpu_usage_long_term"
]
```

<sub>*Managed by the [Sourcegraph Code-intelligence team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<br />

## precise-code-intel-worker: provisioning_container_memory_usage_long_term

<p class="subtitle">container memory usage (1d maximum) by instance</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> precise-code-intel-worker: 80%+ container memory usage (1d maximum) by instance for 336h0m0s

**Possible solutions**

- **Kubernetes:** Consider increasing memory limits in the `Deployment.yaml` for the precise-code-intel-worker service.
- **Docker Compose:** Consider increasing `memory:` of the precise-code-intel-worker container in `docker-compose.yml`.
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_precise-code-intel-worker_provisioning_container_memory_usage_long_term"
]
```

<sub>*Managed by the [Sourcegraph Code-intelligence team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<br />

## precise-code-intel-worker: provisioning_container_cpu_usage_short_term

<p class="subtitle">container cpu usage total (5m maximum) across all cores by instance</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> precise-code-intel-worker: 90%+ container cpu usage total (5m maximum) across all cores by instance for 30m0s

**Possible solutions**

- **Kubernetes:** Consider increasing CPU limits in the the relevant `Deployment.yaml`.
- **Docker Compose:** Consider increasing `cpus:` of the precise-code-intel-worker container in `docker-compose.yml`.
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_precise-code-intel-worker_provisioning_container_cpu_usage_short_term"
]
```

<sub>*Managed by the [Sourcegraph Code-intelligence team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<br />

## precise-code-intel-worker: provisioning_container_memory_usage_short_term

<p class="subtitle">container memory usage (5m maximum) by instance</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> precise-code-intel-worker: 90%+ container memory usage (5m maximum) by instance

**Possible solutions**

- **Kubernetes:** Consider increasing memory limit in relevant `Deployment.yaml`.
- **Docker Compose:** Consider increasing `memory:` of precise-code-intel-worker container in `docker-compose.yml`.
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_precise-code-intel-worker_provisioning_container_memory_usage_short_term"
]
```

<sub>*Managed by the [Sourcegraph Code-intelligence team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<br />

## precise-code-intel-worker: go_goroutines

<p class="subtitle">maximum active goroutines</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> precise-code-intel-worker: 10000+ maximum active goroutines for 10m0s

**Possible solutions**

- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_precise-code-intel-worker_go_goroutines"
]
```

> NOTE: More help interpreting this metric is available in the [dashboards reference](./dashboards.md#precise-code-intel-worker-go-goroutines).

<sub>*Managed by the [Sourcegraph Code-intelligence team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<br />

## precise-code-intel-worker: go_gc_duration_seconds

<p class="subtitle">maximum go garbage collection duration</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> precise-code-intel-worker: 2s+ maximum go garbage collection duration

**Possible solutions**

- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_precise-code-intel-worker_go_gc_duration_seconds"
]
```

<sub>*Managed by the [Sourcegraph Code-intelligence team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<br />

## precise-code-intel-worker: pods_available_percentage

<p class="subtitle">percentage pods available</p>

**Descriptions**

- <span class="badge badge-critical">critical</span> precise-code-intel-worker: less than 90% percentage pods available for 10m0s

**Possible solutions**

- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "critical_precise-code-intel-worker_pods_available_percentage"
]
```

<sub>*Managed by the [Sourcegraph Code-intelligence team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<br />

## query-runner: frontend_internal_api_error_responses

<p class="subtitle">frontend-internal API error responses every 5m by route</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> query-runner: 2%+ frontend-internal API error responses every 5m by route for 5m0s

**Possible solutions**

- **Single-container deployments:** Check `docker logs $CONTAINER_ID` for logs starting with `repo-updater` that indicate requests to the frontend service are failing.
- **Kubernetes:**
	- Confirm that `kubectl get pods` shows the `frontend` pods are healthy.
	- Check `kubectl logs query-runner` for logs indicate request failures to `frontend` or `frontend-internal`.
- **Docker Compose:**
	- Confirm that `docker ps` shows the `frontend-internal` container is healthy.
	- Check `docker logs query-runner` for logs indicating request failures to `frontend` or `frontend-internal`.
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_query-runner_frontend_internal_api_error_responses"
]
```

<sub>*Managed by the [Sourcegraph Search team](https://about.sourcegraph.com/handbook/engineering/search).*</sub>

<br />

## query-runner: container_cpu_usage

<p class="subtitle">container cpu usage total (1m average) across all cores by instance</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> query-runner: 99%+ container cpu usage total (1m average) across all cores by instance

**Possible solutions**

- **Kubernetes:** Consider increasing CPU limits in the the relevant `Deployment.yaml`.
- **Docker Compose:** Consider increasing `cpus:` of the query-runner container in `docker-compose.yml`.
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_query-runner_container_cpu_usage"
]
```

<sub>*Managed by the [Sourcegraph Search team](https://about.sourcegraph.com/handbook/engineering/search).*</sub>

<br />

## query-runner: container_memory_usage

<p class="subtitle">container memory usage by instance</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> query-runner: 99%+ container memory usage by instance

**Possible solutions**

- **Kubernetes:** Consider increasing memory limit in relevant `Deployment.yaml`.
- **Docker Compose:** Consider increasing `memory:` of query-runner container in `docker-compose.yml`.
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_query-runner_container_memory_usage"
]
```

<sub>*Managed by the [Sourcegraph Search team](https://about.sourcegraph.com/handbook/engineering/search).*</sub>

<br />

## query-runner: provisioning_container_cpu_usage_long_term

<p class="subtitle">container cpu usage total (90th percentile over 1d) across all cores by instance</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> query-runner: 80%+ container cpu usage total (90th percentile over 1d) across all cores by instance for 336h0m0s

**Possible solutions**

- **Kubernetes:** Consider increasing CPU limits in the `Deployment.yaml` for the query-runner service.
- **Docker Compose:** Consider increasing `cpus:` of the query-runner container in `docker-compose.yml`.
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_query-runner_provisioning_container_cpu_usage_long_term"
]
```

<sub>*Managed by the [Sourcegraph Search team](https://about.sourcegraph.com/handbook/engineering/search).*</sub>

<br />

## query-runner: provisioning_container_memory_usage_long_term

<p class="subtitle">container memory usage (1d maximum) by instance</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> query-runner: 80%+ container memory usage (1d maximum) by instance for 336h0m0s

**Possible solutions**

- **Kubernetes:** Consider increasing memory limits in the `Deployment.yaml` for the query-runner service.
- **Docker Compose:** Consider increasing `memory:` of the query-runner container in `docker-compose.yml`.
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_query-runner_provisioning_container_memory_usage_long_term"
]
```

<sub>*Managed by the [Sourcegraph Search team](https://about.sourcegraph.com/handbook/engineering/search).*</sub>

<br />

## query-runner: provisioning_container_cpu_usage_short_term

<p class="subtitle">container cpu usage total (5m maximum) across all cores by instance</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> query-runner: 90%+ container cpu usage total (5m maximum) across all cores by instance for 30m0s

**Possible solutions**

- **Kubernetes:** Consider increasing CPU limits in the the relevant `Deployment.yaml`.
- **Docker Compose:** Consider increasing `cpus:` of the query-runner container in `docker-compose.yml`.
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_query-runner_provisioning_container_cpu_usage_short_term"
]
```

<sub>*Managed by the [Sourcegraph Search team](https://about.sourcegraph.com/handbook/engineering/search).*</sub>

<br />

## query-runner: provisioning_container_memory_usage_short_term

<p class="subtitle">container memory usage (5m maximum) by instance</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> query-runner: 90%+ container memory usage (5m maximum) by instance

**Possible solutions**

- **Kubernetes:** Consider increasing memory limit in relevant `Deployment.yaml`.
- **Docker Compose:** Consider increasing `memory:` of query-runner container in `docker-compose.yml`.
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_query-runner_provisioning_container_memory_usage_short_term"
]
```

<sub>*Managed by the [Sourcegraph Search team](https://about.sourcegraph.com/handbook/engineering/search).*</sub>

<br />

## query-runner: go_goroutines

<p class="subtitle">maximum active goroutines</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> query-runner: 10000+ maximum active goroutines for 10m0s

**Possible solutions**

- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_query-runner_go_goroutines"
]
```

> NOTE: More help interpreting this metric is available in the [dashboards reference](./dashboards.md#query-runner-go-goroutines).

<sub>*Managed by the [Sourcegraph Search team](https://about.sourcegraph.com/handbook/engineering/search).*</sub>

<br />

## query-runner: go_gc_duration_seconds

<p class="subtitle">maximum go garbage collection duration</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> query-runner: 2s+ maximum go garbage collection duration

**Possible solutions**

- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_query-runner_go_gc_duration_seconds"
]
```

<sub>*Managed by the [Sourcegraph Search team](https://about.sourcegraph.com/handbook/engineering/search).*</sub>

<br />

## query-runner: pods_available_percentage

<p class="subtitle">percentage pods available</p>

**Descriptions**

- <span class="badge badge-critical">critical</span> query-runner: less than 90% percentage pods available for 10m0s

**Possible solutions**

- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "critical_query-runner_pods_available_percentage"
]
```

<sub>*Managed by the [Sourcegraph Search team](https://about.sourcegraph.com/handbook/engineering/search).*</sub>

<br />

## worker: worker_job_codeintel-janitor_count

<p class="subtitle">number of worker instances running the codeintel-janitor job</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> worker: less than 1 number of worker instances running the codeintel-janitor job for 1m0s
- <span class="badge badge-critical">critical</span> worker: less than 1 number of worker instances running the codeintel-janitor job for 5m0s

**Possible solutions**

- Ensure your instance defines a worker container such that:
	- `WORKER_JOB_ALLOWLIST` contains "codeintel-janitor" (or "all"), and
	- `WORKER_JOB_BLOCKLIST` does not contain "codeintel-janitor"
- Ensure that such a container is not failing to start or stay active
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_worker_worker_job_codeintel-janitor_count",
  "critical_worker_worker_job_codeintel-janitor_count"
]
```

<sub>*Managed by the [Sourcegraph Code-intelligence team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<br />

## worker: worker_job_codeintel-commitgraph_count

<p class="subtitle">number of worker instances running the codeintel-commitgraph job</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> worker: less than 1 number of worker instances running the codeintel-commitgraph job for 1m0s
- <span class="badge badge-critical">critical</span> worker: less than 1 number of worker instances running the codeintel-commitgraph job for 5m0s

**Possible solutions**

- Ensure your instance defines a worker container such that:
	- `WORKER_JOB_ALLOWLIST` contains "codeintel-commitgraph" (or "all"), and
	- `WORKER_JOB_BLOCKLIST` does not contain "codeintel-commitgraph"
- Ensure that such a container is not failing to start or stay active
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_worker_worker_job_codeintel-commitgraph_count",
  "critical_worker_worker_job_codeintel-commitgraph_count"
]
```

<sub>*Managed by the [Sourcegraph Code-intelligence team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<br />

## worker: worker_job_codeintel-auto-indexing_count

<p class="subtitle">number of worker instances running the codeintel-auto-indexing job</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> worker: less than 1 number of worker instances running the codeintel-auto-indexing job for 1m0s
- <span class="badge badge-critical">critical</span> worker: less than 1 number of worker instances running the codeintel-auto-indexing job for 5m0s

**Possible solutions**

- Ensure your instance defines a worker container such that:
	- `WORKER_JOB_ALLOWLIST` contains "codeintel-auto-indexing" (or "all"), and
	- `WORKER_JOB_BLOCKLIST` does not contain "codeintel-auto-indexing"
- Ensure that such a container is not failing to start or stay active
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_worker_worker_job_codeintel-auto-indexing_count",
  "critical_worker_worker_job_codeintel-auto-indexing_count"
]
```

<sub>*Managed by the [Sourcegraph Code-intelligence team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<br />

## worker: frontend_internal_api_error_responses

<p class="subtitle">frontend-internal API error responses every 5m by route</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> worker: 2%+ frontend-internal API error responses every 5m by route for 5m0s

**Possible solutions**

- **Single-container deployments:** Check `docker logs $CONTAINER_ID` for logs starting with `repo-updater` that indicate requests to the frontend service are failing.
- **Kubernetes:**
	- Confirm that `kubectl get pods` shows the `frontend` pods are healthy.
	- Check `kubectl logs worker` for logs indicate request failures to `frontend` or `frontend-internal`.
- **Docker Compose:**
	- Confirm that `docker ps` shows the `frontend-internal` container is healthy.
	- Check `docker logs worker` for logs indicating request failures to `frontend` or `frontend-internal`.
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_worker_frontend_internal_api_error_responses"
]
```

<sub>*Managed by the [Sourcegraph Code-intelligence team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<br />

## worker: mean_blocked_seconds_per_conn_request

<p class="subtitle">mean blocked seconds per conn request</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> worker: 0.05s+ mean blocked seconds per conn request for 5m0s
- <span class="badge badge-critical">critical</span> worker: 0.1s+ mean blocked seconds per conn request for 10m0s

**Possible solutions**

- Increase SRC_PGSQL_MAX_OPEN together with giving more memory to the database if needed
- Scale up Postgres memory / cpus [See our scaling guide](https://docs.sourcegraph.com/admin/config/postgres-conf)
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_worker_mean_blocked_seconds_per_conn_request",
  "critical_worker_mean_blocked_seconds_per_conn_request"
]
```

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<br />

## worker: container_cpu_usage

<p class="subtitle">container cpu usage total (1m average) across all cores by instance</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> worker: 99%+ container cpu usage total (1m average) across all cores by instance

**Possible solutions**

- **Kubernetes:** Consider increasing CPU limits in the the relevant `Deployment.yaml`.
- **Docker Compose:** Consider increasing `cpus:` of the worker container in `docker-compose.yml`.
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_worker_container_cpu_usage"
]
```

<sub>*Managed by the [Sourcegraph Code-intelligence team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<br />

## worker: container_memory_usage

<p class="subtitle">container memory usage by instance</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> worker: 99%+ container memory usage by instance

**Possible solutions**

- **Kubernetes:** Consider increasing memory limit in relevant `Deployment.yaml`.
- **Docker Compose:** Consider increasing `memory:` of worker container in `docker-compose.yml`.
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_worker_container_memory_usage"
]
```

<sub>*Managed by the [Sourcegraph Code-intelligence team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<br />

## worker: provisioning_container_cpu_usage_long_term

<p class="subtitle">container cpu usage total (90th percentile over 1d) across all cores by instance</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> worker: 80%+ container cpu usage total (90th percentile over 1d) across all cores by instance for 336h0m0s

**Possible solutions**

- **Kubernetes:** Consider increasing CPU limits in the `Deployment.yaml` for the worker service.
- **Docker Compose:** Consider increasing `cpus:` of the worker container in `docker-compose.yml`.
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_worker_provisioning_container_cpu_usage_long_term"
]
```

<sub>*Managed by the [Sourcegraph Code-intelligence team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<br />

## worker: provisioning_container_memory_usage_long_term

<p class="subtitle">container memory usage (1d maximum) by instance</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> worker: 80%+ container memory usage (1d maximum) by instance for 336h0m0s

**Possible solutions**

- **Kubernetes:** Consider increasing memory limits in the `Deployment.yaml` for the worker service.
- **Docker Compose:** Consider increasing `memory:` of the worker container in `docker-compose.yml`.
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_worker_provisioning_container_memory_usage_long_term"
]
```

<sub>*Managed by the [Sourcegraph Code-intelligence team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<br />

## worker: provisioning_container_cpu_usage_short_term

<p class="subtitle">container cpu usage total (5m maximum) across all cores by instance</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> worker: 90%+ container cpu usage total (5m maximum) across all cores by instance for 30m0s

**Possible solutions**

- **Kubernetes:** Consider increasing CPU limits in the the relevant `Deployment.yaml`.
- **Docker Compose:** Consider increasing `cpus:` of the worker container in `docker-compose.yml`.
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_worker_provisioning_container_cpu_usage_short_term"
]
```

<sub>*Managed by the [Sourcegraph Code-intelligence team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<br />

## worker: provisioning_container_memory_usage_short_term

<p class="subtitle">container memory usage (5m maximum) by instance</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> worker: 90%+ container memory usage (5m maximum) by instance

**Possible solutions**

- **Kubernetes:** Consider increasing memory limit in relevant `Deployment.yaml`.
- **Docker Compose:** Consider increasing `memory:` of worker container in `docker-compose.yml`.
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_worker_provisioning_container_memory_usage_short_term"
]
```

<sub>*Managed by the [Sourcegraph Code-intelligence team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<br />

## worker: go_goroutines

<p class="subtitle">maximum active goroutines</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> worker: 10000+ maximum active goroutines for 10m0s

**Possible solutions**

- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_worker_go_goroutines"
]
```

> NOTE: More help interpreting this metric is available in the [dashboards reference](./dashboards.md#worker-go-goroutines).

<sub>*Managed by the [Sourcegraph Code-intelligence team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<br />

## worker: go_gc_duration_seconds

<p class="subtitle">maximum go garbage collection duration</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> worker: 2s+ maximum go garbage collection duration

**Possible solutions**

- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_worker_go_gc_duration_seconds"
]
```

<sub>*Managed by the [Sourcegraph Code-intelligence team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<br />

## worker: pods_available_percentage

<p class="subtitle">percentage pods available</p>

**Descriptions**

- <span class="badge badge-critical">critical</span> worker: less than 90% percentage pods available for 10m0s

**Possible solutions**

- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "critical_worker_pods_available_percentage"
]
```

<sub>*Managed by the [Sourcegraph Code-intelligence team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<br />

## repo-updater: src_repoupdater_max_sync_backoff

<p class="subtitle">time since oldest sync</p>

**Descriptions**

- <span class="badge badge-critical">critical</span> repo-updater: 32400s+ time since oldest sync for 10m0s

**Possible solutions**

- An alert here indicates that no code host connections have synced in at least 9h0m0s. This indicates that there could be a configuration issue
with your code hosts connections or networking issues affecting communication with your code hosts.
- Check the code host status indicator (cloud icon in top right of Sourcegraph homepage) for errors.
- Make sure external services do not have invalid tokens by navigating to them in the web UI and clicking save. If there are no errors, they are valid.
- Check the repo-updater logs for errors about syncing.
- Confirm that outbound network connections are allowed where repo-updater is deployed.
- Check back in an hour to see if the issue has resolved itself.
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "critical_repo-updater_src_repoupdater_max_sync_backoff"
]
```

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<br />

## repo-updater: src_repoupdater_syncer_sync_errors_total

<p class="subtitle">site level external service sync error rate</p>

**Descriptions**

- <span class="badge badge-critical">critical</span> repo-updater: 0+ site level external service sync error rate for 10m0s

**Possible solutions**

- An alert here indicates errors syncing site level repo metadata with code hosts. This indicates that there could be a configuration issue
with your code hosts connections or networking issues affecting communication with your code hosts.
- Check the code host status indicator (cloud icon in top right of Sourcegraph homepage) for errors.
- Make sure external services do not have invalid tokens by navigating to them in the web UI and clicking save. If there are no errors, they are valid.
- Check the repo-updater logs for errors about syncing.
- Confirm that outbound network connections are allowed where repo-updater is deployed.
- Check back in an hour to see if the issue has resolved itself.
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "critical_repo-updater_src_repoupdater_syncer_sync_errors_total"
]
```

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<br />

## repo-updater: syncer_sync_start

<p class="subtitle">repo metadata sync was started</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> repo-updater: less than 0 repo metadata sync was started for 9h0m0s

**Possible solutions**

- Check repo-updater logs for errors.
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_repo-updater_syncer_sync_start"
]
```

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<br />

## repo-updater: syncer_sync_duration

<p class="subtitle">95th repositories sync duration</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> repo-updater: 30s+ 95th repositories sync duration for 5m0s

**Possible solutions**

- Check the network latency is reasonable (<50ms) between the Sourcegraph and the code host
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_repo-updater_syncer_sync_duration"
]
```

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<br />

## repo-updater: source_duration

<p class="subtitle">95th repositories source duration</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> repo-updater: 30s+ 95th repositories source duration for 5m0s

**Possible solutions**

- Check the network latency is reasonable (<50ms) between the Sourcegraph and the code host
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_repo-updater_source_duration"
]
```

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<br />

## repo-updater: syncer_synced_repos

<p class="subtitle">repositories synced</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> repo-updater: less than 0 repositories synced for 9h0m0s

**Possible solutions**

- Check network connectivity to code hosts
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_repo-updater_syncer_synced_repos"
]
```

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<br />

## repo-updater: sourced_repos

<p class="subtitle">repositories sourced</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> repo-updater: less than 0 repositories sourced for 9h0m0s

**Possible solutions**

- Check network connectivity to code hosts
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_repo-updater_sourced_repos"
]
```

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<br />

## repo-updater: user_added_repos

<p class="subtitle">total number of user added repos</p>

**Descriptions**

- <span class="badge badge-critical">critical</span> repo-updater: 180000+ total number of user added repos for 5m0s

**Possible solutions**

- Check for unusual spikes in user added repos. Each user is only allowed to add 2000
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "critical_repo-updater_user_added_repos"
]
```

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<br />

## repo-updater: purge_failed

<p class="subtitle">repositories purge failed</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> repo-updater: 0+ repositories purge failed for 5m0s

**Possible solutions**

- Check repo-updater`s connectivity with gitserver and gitserver logs
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_repo-updater_purge_failed"
]
```

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<br />

## repo-updater: sched_auto_fetch

<p class="subtitle">repositories scheduled due to hitting a deadline</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> repo-updater: less than 0 repositories scheduled due to hitting a deadline for 9h0m0s

**Possible solutions**

- Check repo-updater logs. This is expected to fire if there are no user added code hosts
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_repo-updater_sched_auto_fetch"
]
```

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<br />

## repo-updater: sched_known_repos

<p class="subtitle">repositories managed by the scheduler</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> repo-updater: less than 0 repositories managed by the scheduler for 10m0s

**Possible solutions**

- Check repo-updater logs. This is expected to fire if there are no user added code hosts
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_repo-updater_sched_known_repos"
]
```

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<br />

## repo-updater: sched_update_queue_length

<p class="subtitle">rate of growth of update queue length over 5 minutes</p>

**Descriptions**

- <span class="badge badge-critical">critical</span> repo-updater: 0+ rate of growth of update queue length over 5 minutes for 2h0m0s

**Possible solutions**

- Check repo-updater logs for indications that the queue is not being processed. The queue length should trend downwards over time as items are sent to GitServer
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "critical_repo-updater_sched_update_queue_length"
]
```

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<br />

## repo-updater: sched_loops

<p class="subtitle">scheduler loops</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> repo-updater: less than 0 scheduler loops for 9h0m0s

**Possible solutions**

- Check repo-updater logs for errors. This is expected to fire if there are no user added code hosts
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_repo-updater_sched_loops"
]
```

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<br />

## repo-updater: sched_error

<p class="subtitle">repositories schedule error rate</p>

**Descriptions**

- <span class="badge badge-critical">critical</span> repo-updater: 1+ repositories schedule error rate for 1m0s

**Possible solutions**

- Check repo-updater logs for errors
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "critical_repo-updater_sched_error"
]
```

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<br />

## repo-updater: perms_syncer_perms

<p class="subtitle">time gap between least and most up to date permissions</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> repo-updater: 259200s+ time gap between least and most up to date permissions for 5m0s

**Possible solutions**

- Increase the API rate limit to [GitHub](https://docs.sourcegraph.com/admin/external_service/github#github-com-rate-limits), [GitLab](https://docs.sourcegraph.com/admin/external_service/gitlab#internal-rate-limits) or [Bitbucket Server](https://docs.sourcegraph.com/admin/external_service/bitbucket_server#internal-rate-limits).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_repo-updater_perms_syncer_perms"
]
```

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<br />

## repo-updater: perms_syncer_stale_perms

<p class="subtitle">number of entities with stale permissions</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> repo-updater: 100+ number of entities with stale permissions for 5m0s

**Possible solutions**

- Increase the API rate limit to [GitHub](https://docs.sourcegraph.com/admin/external_service/github#github-com-rate-limits), [GitLab](https://docs.sourcegraph.com/admin/external_service/gitlab#internal-rate-limits) or [Bitbucket Server](https://docs.sourcegraph.com/admin/external_service/bitbucket_server#internal-rate-limits).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_repo-updater_perms_syncer_stale_perms"
]
```

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<br />

## repo-updater: perms_syncer_no_perms

<p class="subtitle">number of entities with no permissions</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> repo-updater: 100+ number of entities with no permissions for 5m0s

**Possible solutions**

- **Enabled permissions for the first time:** Wait for few minutes and see if the number goes down.
- **Otherwise:** Increase the API rate limit to [GitHub](https://docs.sourcegraph.com/admin/external_service/github#github-com-rate-limits), [GitLab](https://docs.sourcegraph.com/admin/external_service/gitlab#internal-rate-limits) or [Bitbucket Server](https://docs.sourcegraph.com/admin/external_service/bitbucket_server#internal-rate-limits).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_repo-updater_perms_syncer_no_perms"
]
```

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<br />

## repo-updater: perms_syncer_sync_duration

<p class="subtitle">95th permissions sync duration</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> repo-updater: 30s+ 95th permissions sync duration for 5m0s

**Possible solutions**

- Check the network latency is reasonable (<50ms) between the Sourcegraph and the code host.
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_repo-updater_perms_syncer_sync_duration"
]
```

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<br />

## repo-updater: perms_syncer_queue_size

<p class="subtitle">permissions sync queued items</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> repo-updater: 100+ permissions sync queued items for 5m0s

**Possible solutions**

- **Enabled permissions for the first time:** Wait for few minutes and see if the number goes down.
- **Otherwise:** Increase the API rate limit to [GitHub](https://docs.sourcegraph.com/admin/external_service/github#github-com-rate-limits), [GitLab](https://docs.sourcegraph.com/admin/external_service/gitlab#internal-rate-limits) or [Bitbucket Server](https://docs.sourcegraph.com/admin/external_service/bitbucket_server#internal-rate-limits).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_repo-updater_perms_syncer_queue_size"
]
```

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<br />

## repo-updater: perms_syncer_sync_errors

<p class="subtitle">permissions sync error rate</p>

**Descriptions**

- <span class="badge badge-critical">critical</span> repo-updater: 1+ permissions sync error rate for 1m0s

**Possible solutions**

- Check the network connectivity the Sourcegraph and the code host.
- Check if API rate limit quota is exhausted on the code host.
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "critical_repo-updater_perms_syncer_sync_errors"
]
```

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<br />

## repo-updater: src_repoupdater_external_services_total

<p class="subtitle">the total number of external services</p>

**Descriptions**

- <span class="badge badge-critical">critical</span> repo-updater: 20000+ the total number of external services for 1h0m0s

**Possible solutions**

- Check for spikes in external services, could be abuse
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "critical_repo-updater_src_repoupdater_external_services_total"
]
```

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<br />

## repo-updater: src_repoupdater_user_external_services_total

<p class="subtitle">the total number of user added external services</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> repo-updater: 20000+ the total number of user added external services for 1h0m0s

**Possible solutions**

- Check for spikes in external services, could be abuse
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_repo-updater_src_repoupdater_user_external_services_total"
]
```

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<br />

## repo-updater: repoupdater_queued_sync_jobs_total

<p class="subtitle">the total number of queued sync jobs</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> repo-updater: 100+ the total number of queued sync jobs for 1h0m0s

**Possible solutions**

- **Check if jobs are failing to sync:** "SELECT * FROM external_service_sync_jobs WHERE state = `errored`";
- **Increase the number of workers** using the `repoConcurrentExternalServiceSyncers` site config.
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_repo-updater_repoupdater_queued_sync_jobs_total"
]
```

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<br />

## repo-updater: repoupdater_completed_sync_jobs_total

<p class="subtitle">the total number of completed sync jobs</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> repo-updater: 100000+ the total number of completed sync jobs for 1h0m0s

**Possible solutions**

- Check repo-updater logs. Jobs older than 1 day should have been removed.
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_repo-updater_repoupdater_completed_sync_jobs_total"
]
```

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<br />

## repo-updater: repoupdater_errored_sync_jobs_percentage

<p class="subtitle">the percentage of external services that have failed their most recent sync</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> repo-updater: 10%+ the percentage of external services that have failed their most recent sync for 1h0m0s

**Possible solutions**

- Check repo-updater logs. Check code host connectivity
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_repo-updater_repoupdater_errored_sync_jobs_percentage"
]
```

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<br />

## repo-updater: github_graphql_rate_limit_remaining

<p class="subtitle">remaining calls to GitHub graphql API before hitting the rate limit</p>

**Descriptions**

- <span class="badge badge-critical">critical</span> repo-updater: less than 250 remaining calls to GitHub graphql API before hitting the rate limit

**Possible solutions**

- Try restarting the pod to get a different public IP.
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "critical_repo-updater_github_graphql_rate_limit_remaining"
]
```

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<br />

## repo-updater: github_rest_rate_limit_remaining

<p class="subtitle">remaining calls to GitHub rest API before hitting the rate limit</p>

**Descriptions**

- <span class="badge badge-critical">critical</span> repo-updater: less than 250 remaining calls to GitHub rest API before hitting the rate limit

**Possible solutions**

- Try restarting the pod to get a different public IP.
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "critical_repo-updater_github_rest_rate_limit_remaining"
]
```

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<br />

## repo-updater: github_search_rate_limit_remaining

<p class="subtitle">remaining calls to GitHub search API before hitting the rate limit</p>

**Descriptions**

- <span class="badge badge-critical">critical</span> repo-updater: less than 5 remaining calls to GitHub search API before hitting the rate limit

**Possible solutions**

- Try restarting the pod to get a different public IP.
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "critical_repo-updater_github_search_rate_limit_remaining"
]
```

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<br />

## repo-updater: gitlab_rest_rate_limit_remaining

<p class="subtitle">remaining calls to GitLab rest API before hitting the rate limit</p>

**Descriptions**

- <span class="badge badge-critical">critical</span> repo-updater: less than 30 remaining calls to GitLab rest API before hitting the rate limit

**Possible solutions**

- Try restarting the pod to get a different public IP.
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "critical_repo-updater_gitlab_rest_rate_limit_remaining"
]
```

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<br />

## repo-updater: frontend_internal_api_error_responses

<p class="subtitle">frontend-internal API error responses every 5m by route</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> repo-updater: 2%+ frontend-internal API error responses every 5m by route for 5m0s

**Possible solutions**

- **Single-container deployments:** Check `docker logs $CONTAINER_ID` for logs starting with `repo-updater` that indicate requests to the frontend service are failing.
- **Kubernetes:**
	- Confirm that `kubectl get pods` shows the `frontend` pods are healthy.
	- Check `kubectl logs repo-updater` for logs indicate request failures to `frontend` or `frontend-internal`.
- **Docker Compose:**
	- Confirm that `docker ps` shows the `frontend-internal` container is healthy.
	- Check `docker logs repo-updater` for logs indicating request failures to `frontend` or `frontend-internal`.
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_repo-updater_frontend_internal_api_error_responses"
]
```

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<br />

## repo-updater: mean_blocked_seconds_per_conn_request

<p class="subtitle">mean blocked seconds per conn request</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> repo-updater: 0.05s+ mean blocked seconds per conn request for 5m0s
- <span class="badge badge-critical">critical</span> repo-updater: 0.1s+ mean blocked seconds per conn request for 10m0s

**Possible solutions**

- Increase SRC_PGSQL_MAX_OPEN together with giving more memory to the database if needed
- Scale up Postgres memory / cpus [See our scaling guide](https://docs.sourcegraph.com/admin/config/postgres-conf)
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_repo-updater_mean_blocked_seconds_per_conn_request",
  "critical_repo-updater_mean_blocked_seconds_per_conn_request"
]
```

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<br />

## repo-updater: container_cpu_usage

<p class="subtitle">container cpu usage total (1m average) across all cores by instance</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> repo-updater: 99%+ container cpu usage total (1m average) across all cores by instance

**Possible solutions**

- **Kubernetes:** Consider increasing CPU limits in the the relevant `Deployment.yaml`.
- **Docker Compose:** Consider increasing `cpus:` of the repo-updater container in `docker-compose.yml`.
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_repo-updater_container_cpu_usage"
]
```

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<br />

## repo-updater: container_memory_usage

<p class="subtitle">container memory usage by instance</p>

**Descriptions**

- <span class="badge badge-critical">critical</span> repo-updater: 90%+ container memory usage by instance for 10m0s

**Possible solutions**

- **Kubernetes:** Consider increasing memory limit in relevant `Deployment.yaml`.
- **Docker Compose:** Consider increasing `memory:` of repo-updater container in `docker-compose.yml`.
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "critical_repo-updater_container_memory_usage"
]
```

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<br />

## repo-updater: provisioning_container_cpu_usage_long_term

<p class="subtitle">container cpu usage total (90th percentile over 1d) across all cores by instance</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> repo-updater: 80%+ container cpu usage total (90th percentile over 1d) across all cores by instance for 336h0m0s

**Possible solutions**

- **Kubernetes:** Consider increasing CPU limits in the `Deployment.yaml` for the repo-updater service.
- **Docker Compose:** Consider increasing `cpus:` of the repo-updater container in `docker-compose.yml`.
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_repo-updater_provisioning_container_cpu_usage_long_term"
]
```

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<br />

## repo-updater: provisioning_container_memory_usage_long_term

<p class="subtitle">container memory usage (1d maximum) by instance</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> repo-updater: 80%+ container memory usage (1d maximum) by instance for 336h0m0s

**Possible solutions**

- **Kubernetes:** Consider increasing memory limits in the `Deployment.yaml` for the repo-updater service.
- **Docker Compose:** Consider increasing `memory:` of the repo-updater container in `docker-compose.yml`.
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_repo-updater_provisioning_container_memory_usage_long_term"
]
```

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<br />

## repo-updater: provisioning_container_cpu_usage_short_term

<p class="subtitle">container cpu usage total (5m maximum) across all cores by instance</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> repo-updater: 90%+ container cpu usage total (5m maximum) across all cores by instance for 30m0s

**Possible solutions**

- **Kubernetes:** Consider increasing CPU limits in the the relevant `Deployment.yaml`.
- **Docker Compose:** Consider increasing `cpus:` of the repo-updater container in `docker-compose.yml`.
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_repo-updater_provisioning_container_cpu_usage_short_term"
]
```

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<br />

## repo-updater: provisioning_container_memory_usage_short_term

<p class="subtitle">container memory usage (5m maximum) by instance</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> repo-updater: 90%+ container memory usage (5m maximum) by instance

**Possible solutions**

- **Kubernetes:** Consider increasing memory limit in relevant `Deployment.yaml`.
- **Docker Compose:** Consider increasing `memory:` of repo-updater container in `docker-compose.yml`.
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_repo-updater_provisioning_container_memory_usage_short_term"
]
```

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<br />

## repo-updater: go_goroutines

<p class="subtitle">maximum active goroutines</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> repo-updater: 10000+ maximum active goroutines for 10m0s

**Possible solutions**

- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_repo-updater_go_goroutines"
]
```

> NOTE: More help interpreting this metric is available in the [dashboards reference](./dashboards.md#repo-updater-go-goroutines).

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<br />

## repo-updater: go_gc_duration_seconds

<p class="subtitle">maximum go garbage collection duration</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> repo-updater: 2s+ maximum go garbage collection duration

**Possible solutions**

- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_repo-updater_go_gc_duration_seconds"
]
```

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<br />

## repo-updater: pods_available_percentage

<p class="subtitle">percentage pods available</p>

**Descriptions**

- <span class="badge badge-critical">critical</span> repo-updater: less than 90% percentage pods available for 10m0s

**Possible solutions**

- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "critical_repo-updater_pods_available_percentage"
]
```

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<br />

## searcher: unindexed_search_request_errors

<p class="subtitle">unindexed search request errors every 5m by code</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> searcher: 5%+ unindexed search request errors every 5m by code for 5m0s

**Possible solutions**

- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_searcher_unindexed_search_request_errors"
]
```

<sub>*Managed by the [Sourcegraph Search team](https://about.sourcegraph.com/handbook/engineering/search).*</sub>

<br />

## searcher: replica_traffic

<p class="subtitle">requests per second over 10m</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> searcher: 5+ requests per second over 10m

**Possible solutions**

- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_searcher_replica_traffic"
]
```

<sub>*Managed by the [Sourcegraph Search team](https://about.sourcegraph.com/handbook/engineering/search).*</sub>

<br />

## searcher: frontend_internal_api_error_responses

<p class="subtitle">frontend-internal API error responses every 5m by route</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> searcher: 2%+ frontend-internal API error responses every 5m by route for 5m0s

**Possible solutions**

- **Single-container deployments:** Check `docker logs $CONTAINER_ID` for logs starting with `repo-updater` that indicate requests to the frontend service are failing.
- **Kubernetes:**
	- Confirm that `kubectl get pods` shows the `frontend` pods are healthy.
	- Check `kubectl logs searcher` for logs indicate request failures to `frontend` or `frontend-internal`.
- **Docker Compose:**
	- Confirm that `docker ps` shows the `frontend-internal` container is healthy.
	- Check `docker logs searcher` for logs indicating request failures to `frontend` or `frontend-internal`.
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_searcher_frontend_internal_api_error_responses"
]
```

<sub>*Managed by the [Sourcegraph Search team](https://about.sourcegraph.com/handbook/engineering/search).*</sub>

<br />

## searcher: container_cpu_usage

<p class="subtitle">container cpu usage total (1m average) across all cores by instance</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> searcher: 99%+ container cpu usage total (1m average) across all cores by instance

**Possible solutions**

- **Kubernetes:** Consider increasing CPU limits in the the relevant `Deployment.yaml`.
- **Docker Compose:** Consider increasing `cpus:` of the searcher container in `docker-compose.yml`.
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_searcher_container_cpu_usage"
]
```

<sub>*Managed by the [Sourcegraph Search team](https://about.sourcegraph.com/handbook/engineering/search).*</sub>

<br />

## searcher: container_memory_usage

<p class="subtitle">container memory usage by instance</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> searcher: 99%+ container memory usage by instance

**Possible solutions**

- **Kubernetes:** Consider increasing memory limit in relevant `Deployment.yaml`.
- **Docker Compose:** Consider increasing `memory:` of searcher container in `docker-compose.yml`.
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_searcher_container_memory_usage"
]
```

<sub>*Managed by the [Sourcegraph Search team](https://about.sourcegraph.com/handbook/engineering/search).*</sub>

<br />

## searcher: provisioning_container_cpu_usage_long_term

<p class="subtitle">container cpu usage total (90th percentile over 1d) across all cores by instance</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> searcher: 80%+ container cpu usage total (90th percentile over 1d) across all cores by instance for 336h0m0s

**Possible solutions**

- **Kubernetes:** Consider increasing CPU limits in the `Deployment.yaml` for the searcher service.
- **Docker Compose:** Consider increasing `cpus:` of the searcher container in `docker-compose.yml`.
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_searcher_provisioning_container_cpu_usage_long_term"
]
```

<sub>*Managed by the [Sourcegraph Search team](https://about.sourcegraph.com/handbook/engineering/search).*</sub>

<br />

## searcher: provisioning_container_memory_usage_long_term

<p class="subtitle">container memory usage (1d maximum) by instance</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> searcher: 80%+ container memory usage (1d maximum) by instance for 336h0m0s

**Possible solutions**

- **Kubernetes:** Consider increasing memory limits in the `Deployment.yaml` for the searcher service.
- **Docker Compose:** Consider increasing `memory:` of the searcher container in `docker-compose.yml`.
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_searcher_provisioning_container_memory_usage_long_term"
]
```

<sub>*Managed by the [Sourcegraph Search team](https://about.sourcegraph.com/handbook/engineering/search).*</sub>

<br />

## searcher: provisioning_container_cpu_usage_short_term

<p class="subtitle">container cpu usage total (5m maximum) across all cores by instance</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> searcher: 90%+ container cpu usage total (5m maximum) across all cores by instance for 30m0s

**Possible solutions**

- **Kubernetes:** Consider increasing CPU limits in the the relevant `Deployment.yaml`.
- **Docker Compose:** Consider increasing `cpus:` of the searcher container in `docker-compose.yml`.
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_searcher_provisioning_container_cpu_usage_short_term"
]
```

<sub>*Managed by the [Sourcegraph Search team](https://about.sourcegraph.com/handbook/engineering/search).*</sub>

<br />

## searcher: provisioning_container_memory_usage_short_term

<p class="subtitle">container memory usage (5m maximum) by instance</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> searcher: 90%+ container memory usage (5m maximum) by instance

**Possible solutions**

- **Kubernetes:** Consider increasing memory limit in relevant `Deployment.yaml`.
- **Docker Compose:** Consider increasing `memory:` of searcher container in `docker-compose.yml`.
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_searcher_provisioning_container_memory_usage_short_term"
]
```

<sub>*Managed by the [Sourcegraph Search team](https://about.sourcegraph.com/handbook/engineering/search).*</sub>

<br />

## searcher: go_goroutines

<p class="subtitle">maximum active goroutines</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> searcher: 10000+ maximum active goroutines for 10m0s

**Possible solutions**

- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_searcher_go_goroutines"
]
```

> NOTE: More help interpreting this metric is available in the [dashboards reference](./dashboards.md#searcher-go-goroutines).

<sub>*Managed by the [Sourcegraph Search team](https://about.sourcegraph.com/handbook/engineering/search).*</sub>

<br />

## searcher: go_gc_duration_seconds

<p class="subtitle">maximum go garbage collection duration</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> searcher: 2s+ maximum go garbage collection duration

**Possible solutions**

- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_searcher_go_gc_duration_seconds"
]
```

<sub>*Managed by the [Sourcegraph Search team](https://about.sourcegraph.com/handbook/engineering/search).*</sub>

<br />

## searcher: pods_available_percentage

<p class="subtitle">percentage pods available</p>

**Descriptions**

- <span class="badge badge-critical">critical</span> searcher: less than 90% percentage pods available for 10m0s

**Possible solutions**

- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "critical_searcher_pods_available_percentage"
]
```

<sub>*Managed by the [Sourcegraph Search team](https://about.sourcegraph.com/handbook/engineering/search).*</sub>

<br />

## symbols: store_fetch_failures

<p class="subtitle">store fetch failures every 5m</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> symbols: 5+ store fetch failures every 5m

**Possible solutions**

- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_symbols_store_fetch_failures"
]
```

<sub>*Managed by the [Sourcegraph Code-intelligence team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<br />

## symbols: current_fetch_queue_size

<p class="subtitle">current fetch queue size</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> symbols: 25+ current fetch queue size

**Possible solutions**

- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_symbols_current_fetch_queue_size"
]
```

<sub>*Managed by the [Sourcegraph Code-intelligence team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<br />

## symbols: frontend_internal_api_error_responses

<p class="subtitle">frontend-internal API error responses every 5m by route</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> symbols: 2%+ frontend-internal API error responses every 5m by route for 5m0s

**Possible solutions**

- **Single-container deployments:** Check `docker logs $CONTAINER_ID` for logs starting with `repo-updater` that indicate requests to the frontend service are failing.
- **Kubernetes:**
	- Confirm that `kubectl get pods` shows the `frontend` pods are healthy.
	- Check `kubectl logs symbols` for logs indicate request failures to `frontend` or `frontend-internal`.
- **Docker Compose:**
	- Confirm that `docker ps` shows the `frontend-internal` container is healthy.
	- Check `docker logs symbols` for logs indicating request failures to `frontend` or `frontend-internal`.
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_symbols_frontend_internal_api_error_responses"
]
```

<sub>*Managed by the [Sourcegraph Code-intelligence team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<br />

## symbols: container_cpu_usage

<p class="subtitle">container cpu usage total (1m average) across all cores by instance</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> symbols: 99%+ container cpu usage total (1m average) across all cores by instance

**Possible solutions**

- **Kubernetes:** Consider increasing CPU limits in the the relevant `Deployment.yaml`.
- **Docker Compose:** Consider increasing `cpus:` of the symbols container in `docker-compose.yml`.
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_symbols_container_cpu_usage"
]
```

<sub>*Managed by the [Sourcegraph Code-intelligence team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<br />

## symbols: container_memory_usage

<p class="subtitle">container memory usage by instance</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> symbols: 99%+ container memory usage by instance

**Possible solutions**

- **Kubernetes:** Consider increasing memory limit in relevant `Deployment.yaml`.
- **Docker Compose:** Consider increasing `memory:` of symbols container in `docker-compose.yml`.
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_symbols_container_memory_usage"
]
```

<sub>*Managed by the [Sourcegraph Code-intelligence team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<br />

## symbols: provisioning_container_cpu_usage_long_term

<p class="subtitle">container cpu usage total (90th percentile over 1d) across all cores by instance</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> symbols: 80%+ container cpu usage total (90th percentile over 1d) across all cores by instance for 336h0m0s

**Possible solutions**

- **Kubernetes:** Consider increasing CPU limits in the `Deployment.yaml` for the symbols service.
- **Docker Compose:** Consider increasing `cpus:` of the symbols container in `docker-compose.yml`.
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_symbols_provisioning_container_cpu_usage_long_term"
]
```

<sub>*Managed by the [Sourcegraph Code-intelligence team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<br />

## symbols: provisioning_container_memory_usage_long_term

<p class="subtitle">container memory usage (1d maximum) by instance</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> symbols: 80%+ container memory usage (1d maximum) by instance for 336h0m0s

**Possible solutions**

- **Kubernetes:** Consider increasing memory limits in the `Deployment.yaml` for the symbols service.
- **Docker Compose:** Consider increasing `memory:` of the symbols container in `docker-compose.yml`.
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_symbols_provisioning_container_memory_usage_long_term"
]
```

<sub>*Managed by the [Sourcegraph Code-intelligence team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<br />

## symbols: provisioning_container_cpu_usage_short_term

<p class="subtitle">container cpu usage total (5m maximum) across all cores by instance</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> symbols: 90%+ container cpu usage total (5m maximum) across all cores by instance for 30m0s

**Possible solutions**

- **Kubernetes:** Consider increasing CPU limits in the the relevant `Deployment.yaml`.
- **Docker Compose:** Consider increasing `cpus:` of the symbols container in `docker-compose.yml`.
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_symbols_provisioning_container_cpu_usage_short_term"
]
```

<sub>*Managed by the [Sourcegraph Code-intelligence team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<br />

## symbols: provisioning_container_memory_usage_short_term

<p class="subtitle">container memory usage (5m maximum) by instance</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> symbols: 90%+ container memory usage (5m maximum) by instance

**Possible solutions**

- **Kubernetes:** Consider increasing memory limit in relevant `Deployment.yaml`.
- **Docker Compose:** Consider increasing `memory:` of symbols container in `docker-compose.yml`.
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_symbols_provisioning_container_memory_usage_short_term"
]
```

<sub>*Managed by the [Sourcegraph Code-intelligence team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<br />

## symbols: go_goroutines

<p class="subtitle">maximum active goroutines</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> symbols: 10000+ maximum active goroutines for 10m0s

**Possible solutions**

- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_symbols_go_goroutines"
]
```

> NOTE: More help interpreting this metric is available in the [dashboards reference](./dashboards.md#symbols-go-goroutines).

<sub>*Managed by the [Sourcegraph Code-intelligence team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<br />

## symbols: go_gc_duration_seconds

<p class="subtitle">maximum go garbage collection duration</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> symbols: 2s+ maximum go garbage collection duration

**Possible solutions**

- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_symbols_go_gc_duration_seconds"
]
```

<sub>*Managed by the [Sourcegraph Code-intelligence team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<br />

## symbols: pods_available_percentage

<p class="subtitle">percentage pods available</p>

**Descriptions**

- <span class="badge badge-critical">critical</span> symbols: less than 90% percentage pods available for 10m0s

**Possible solutions**

- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "critical_symbols_pods_available_percentage"
]
```

<sub>*Managed by the [Sourcegraph Code-intelligence team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<br />

## syntect-server: container_cpu_usage

<p class="subtitle">container cpu usage total (1m average) across all cores by instance</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> syntect-server: 99%+ container cpu usage total (1m average) across all cores by instance

**Possible solutions**

- **Kubernetes:** Consider increasing CPU limits in the the relevant `Deployment.yaml`.
- **Docker Compose:** Consider increasing `cpus:` of the syntect-server container in `docker-compose.yml`.
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_syntect-server_container_cpu_usage"
]
```

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<br />

## syntect-server: container_memory_usage

<p class="subtitle">container memory usage by instance</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> syntect-server: 99%+ container memory usage by instance

**Possible solutions**

- **Kubernetes:** Consider increasing memory limit in relevant `Deployment.yaml`.
- **Docker Compose:** Consider increasing `memory:` of syntect-server container in `docker-compose.yml`.
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_syntect-server_container_memory_usage"
]
```

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<br />

## syntect-server: provisioning_container_cpu_usage_long_term

<p class="subtitle">container cpu usage total (90th percentile over 1d) across all cores by instance</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> syntect-server: 80%+ container cpu usage total (90th percentile over 1d) across all cores by instance for 336h0m0s

**Possible solutions**

- **Kubernetes:** Consider increasing CPU limits in the `Deployment.yaml` for the syntect-server service.
- **Docker Compose:** Consider increasing `cpus:` of the syntect-server container in `docker-compose.yml`.
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_syntect-server_provisioning_container_cpu_usage_long_term"
]
```

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<br />

## syntect-server: provisioning_container_memory_usage_long_term

<p class="subtitle">container memory usage (1d maximum) by instance</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> syntect-server: 80%+ container memory usage (1d maximum) by instance for 336h0m0s

**Possible solutions**

- **Kubernetes:** Consider increasing memory limits in the `Deployment.yaml` for the syntect-server service.
- **Docker Compose:** Consider increasing `memory:` of the syntect-server container in `docker-compose.yml`.
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_syntect-server_provisioning_container_memory_usage_long_term"
]
```

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<br />

## syntect-server: provisioning_container_cpu_usage_short_term

<p class="subtitle">container cpu usage total (5m maximum) across all cores by instance</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> syntect-server: 90%+ container cpu usage total (5m maximum) across all cores by instance for 30m0s

**Possible solutions**

- **Kubernetes:** Consider increasing CPU limits in the the relevant `Deployment.yaml`.
- **Docker Compose:** Consider increasing `cpus:` of the syntect-server container in `docker-compose.yml`.
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_syntect-server_provisioning_container_cpu_usage_short_term"
]
```

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<br />

## syntect-server: provisioning_container_memory_usage_short_term

<p class="subtitle">container memory usage (5m maximum) by instance</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> syntect-server: 90%+ container memory usage (5m maximum) by instance

**Possible solutions**

- **Kubernetes:** Consider increasing memory limit in relevant `Deployment.yaml`.
- **Docker Compose:** Consider increasing `memory:` of syntect-server container in `docker-compose.yml`.
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_syntect-server_provisioning_container_memory_usage_short_term"
]
```

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<br />

## syntect-server: pods_available_percentage

<p class="subtitle">percentage pods available</p>

**Descriptions**

- <span class="badge badge-critical">critical</span> syntect-server: less than 90% percentage pods available for 10m0s

**Possible solutions**

- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "critical_syntect-server_pods_available_percentage"
]
```

<sub>*Managed by the [Sourcegraph Core application team](https://about.sourcegraph.com/handbook/engineering/core-application).*</sub>

<br />

## zoekt-indexserver: average_resolve_revision_duration

<p class="subtitle">average resolve revision duration over 5m</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> zoekt-indexserver: 15s+ average resolve revision duration over 5m
- <span class="badge badge-critical">critical</span> zoekt-indexserver: 30s+ average resolve revision duration over 5m

**Possible solutions**

- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_zoekt-indexserver_average_resolve_revision_duration",
  "critical_zoekt-indexserver_average_resolve_revision_duration"
]
```

<sub>*Managed by the [Sourcegraph Search team](https://about.sourcegraph.com/handbook/engineering/search).*</sub>

<br />

## zoekt-indexserver: container_cpu_usage

<p class="subtitle">container cpu usage total (1m average) across all cores by instance</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> zoekt-indexserver: 99%+ container cpu usage total (1m average) across all cores by instance

**Possible solutions**

- **Kubernetes:** Consider increasing CPU limits in the the relevant `Deployment.yaml`.
- **Docker Compose:** Consider increasing `cpus:` of the zoekt-indexserver container in `docker-compose.yml`.
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_zoekt-indexserver_container_cpu_usage"
]
```

<sub>*Managed by the [Sourcegraph Search team](https://about.sourcegraph.com/handbook/engineering/search).*</sub>

<br />

## zoekt-indexserver: container_memory_usage

<p class="subtitle">container memory usage by instance</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> zoekt-indexserver: 99%+ container memory usage by instance

**Possible solutions**

- **Kubernetes:** Consider increasing memory limit in relevant `Deployment.yaml`.
- **Docker Compose:** Consider increasing `memory:` of zoekt-indexserver container in `docker-compose.yml`.
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_zoekt-indexserver_container_memory_usage"
]
```

<sub>*Managed by the [Sourcegraph Search team](https://about.sourcegraph.com/handbook/engineering/search).*</sub>

<br />

## zoekt-indexserver: provisioning_container_cpu_usage_long_term

<p class="subtitle">container cpu usage total (90th percentile over 1d) across all cores by instance</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> zoekt-indexserver: 80%+ container cpu usage total (90th percentile over 1d) across all cores by instance for 336h0m0s

**Possible solutions**

- **Kubernetes:** Consider increasing CPU limits in the `Deployment.yaml` for the zoekt-indexserver service.
- **Docker Compose:** Consider increasing `cpus:` of the zoekt-indexserver container in `docker-compose.yml`.
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_zoekt-indexserver_provisioning_container_cpu_usage_long_term"
]
```

<sub>*Managed by the [Sourcegraph Search team](https://about.sourcegraph.com/handbook/engineering/search).*</sub>

<br />

## zoekt-indexserver: provisioning_container_memory_usage_long_term

<p class="subtitle">container memory usage (1d maximum) by instance</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> zoekt-indexserver: 80%+ container memory usage (1d maximum) by instance for 336h0m0s

**Possible solutions**

- **Kubernetes:** Consider increasing memory limits in the `Deployment.yaml` for the zoekt-indexserver service.
- **Docker Compose:** Consider increasing `memory:` of the zoekt-indexserver container in `docker-compose.yml`.
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_zoekt-indexserver_provisioning_container_memory_usage_long_term"
]
```

<sub>*Managed by the [Sourcegraph Search team](https://about.sourcegraph.com/handbook/engineering/search).*</sub>

<br />

## zoekt-indexserver: provisioning_container_cpu_usage_short_term

<p class="subtitle">container cpu usage total (5m maximum) across all cores by instance</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> zoekt-indexserver: 90%+ container cpu usage total (5m maximum) across all cores by instance for 30m0s

**Possible solutions**

- **Kubernetes:** Consider increasing CPU limits in the the relevant `Deployment.yaml`.
- **Docker Compose:** Consider increasing `cpus:` of the zoekt-indexserver container in `docker-compose.yml`.
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_zoekt-indexserver_provisioning_container_cpu_usage_short_term"
]
```

<sub>*Managed by the [Sourcegraph Search team](https://about.sourcegraph.com/handbook/engineering/search).*</sub>

<br />

## zoekt-indexserver: provisioning_container_memory_usage_short_term

<p class="subtitle">container memory usage (5m maximum) by instance</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> zoekt-indexserver: 90%+ container memory usage (5m maximum) by instance

**Possible solutions**

- **Kubernetes:** Consider increasing memory limit in relevant `Deployment.yaml`.
- **Docker Compose:** Consider increasing `memory:` of zoekt-indexserver container in `docker-compose.yml`.
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_zoekt-indexserver_provisioning_container_memory_usage_short_term"
]
```

<sub>*Managed by the [Sourcegraph Search team](https://about.sourcegraph.com/handbook/engineering/search).*</sub>

<br />

## zoekt-indexserver: pods_available_percentage

<p class="subtitle">percentage pods available</p>

**Descriptions**

- <span class="badge badge-critical">critical</span> zoekt-indexserver: less than 90% percentage pods available for 10m0s

**Possible solutions**

- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "critical_zoekt-indexserver_pods_available_percentage"
]
```

<sub>*Managed by the [Sourcegraph Search team](https://about.sourcegraph.com/handbook/engineering/search).*</sub>

<br />

## zoekt-webserver: indexed_search_request_errors

<p class="subtitle">indexed search request errors every 5m by code</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> zoekt-webserver: 5%+ indexed search request errors every 5m by code for 5m0s

**Possible solutions**

- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_zoekt-webserver_indexed_search_request_errors"
]
```

<sub>*Managed by the [Sourcegraph Search team](https://about.sourcegraph.com/handbook/engineering/search).*</sub>

<br />

## zoekt-webserver: container_cpu_usage

<p class="subtitle">container cpu usage total (1m average) across all cores by instance</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> zoekt-webserver: 99%+ container cpu usage total (1m average) across all cores by instance

**Possible solutions**

- **Kubernetes:** Consider increasing CPU limits in the the relevant `Deployment.yaml`.
- **Docker Compose:** Consider increasing `cpus:` of the zoekt-webserver container in `docker-compose.yml`.
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_zoekt-webserver_container_cpu_usage"
]
```

<sub>*Managed by the [Sourcegraph Search team](https://about.sourcegraph.com/handbook/engineering/search).*</sub>

<br />

## zoekt-webserver: container_memory_usage

<p class="subtitle">container memory usage by instance</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> zoekt-webserver: 99%+ container memory usage by instance

**Possible solutions**

- **Kubernetes:** Consider increasing memory limit in relevant `Deployment.yaml`.
- **Docker Compose:** Consider increasing `memory:` of zoekt-webserver container in `docker-compose.yml`.
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_zoekt-webserver_container_memory_usage"
]
```

<sub>*Managed by the [Sourcegraph Search team](https://about.sourcegraph.com/handbook/engineering/search).*</sub>

<br />

## zoekt-webserver: provisioning_container_cpu_usage_long_term

<p class="subtitle">container cpu usage total (90th percentile over 1d) across all cores by instance</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> zoekt-webserver: 80%+ container cpu usage total (90th percentile over 1d) across all cores by instance for 336h0m0s

**Possible solutions**

- **Kubernetes:** Consider increasing CPU limits in the `Deployment.yaml` for the zoekt-webserver service.
- **Docker Compose:** Consider increasing `cpus:` of the zoekt-webserver container in `docker-compose.yml`.
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_zoekt-webserver_provisioning_container_cpu_usage_long_term"
]
```

<sub>*Managed by the [Sourcegraph Search team](https://about.sourcegraph.com/handbook/engineering/search).*</sub>

<br />

## zoekt-webserver: provisioning_container_memory_usage_long_term

<p class="subtitle">container memory usage (1d maximum) by instance</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> zoekt-webserver: 80%+ container memory usage (1d maximum) by instance for 336h0m0s

**Possible solutions**

- **Kubernetes:** Consider increasing memory limits in the `Deployment.yaml` for the zoekt-webserver service.
- **Docker Compose:** Consider increasing `memory:` of the zoekt-webserver container in `docker-compose.yml`.
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_zoekt-webserver_provisioning_container_memory_usage_long_term"
]
```

<sub>*Managed by the [Sourcegraph Search team](https://about.sourcegraph.com/handbook/engineering/search).*</sub>

<br />

## zoekt-webserver: provisioning_container_cpu_usage_short_term

<p class="subtitle">container cpu usage total (5m maximum) across all cores by instance</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> zoekt-webserver: 90%+ container cpu usage total (5m maximum) across all cores by instance for 30m0s

**Possible solutions**

- **Kubernetes:** Consider increasing CPU limits in the the relevant `Deployment.yaml`.
- **Docker Compose:** Consider increasing `cpus:` of the zoekt-webserver container in `docker-compose.yml`.
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_zoekt-webserver_provisioning_container_cpu_usage_short_term"
]
```

<sub>*Managed by the [Sourcegraph Search team](https://about.sourcegraph.com/handbook/engineering/search).*</sub>

<br />

## zoekt-webserver: provisioning_container_memory_usage_short_term

<p class="subtitle">container memory usage (5m maximum) by instance</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> zoekt-webserver: 90%+ container memory usage (5m maximum) by instance

**Possible solutions**

- **Kubernetes:** Consider increasing memory limit in relevant `Deployment.yaml`.
- **Docker Compose:** Consider increasing `memory:` of zoekt-webserver container in `docker-compose.yml`.
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_zoekt-webserver_provisioning_container_memory_usage_short_term"
]
```

<sub>*Managed by the [Sourcegraph Search team](https://about.sourcegraph.com/handbook/engineering/search).*</sub>

<br />

## prometheus: prometheus_rule_eval_duration

<p class="subtitle">average prometheus rule group evaluation duration over 10m by rule group</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> prometheus: 30s+ average prometheus rule group evaluation duration over 10m by rule group

**Possible solutions**

- Check the Container monitoring (not available on server) panels and try increasing resources for Prometheus if necessary.
- If the rule group taking a long time to evaluate belongs to `/sg_prometheus_addons`, try reducing the complexity of any custom Prometheus rules provided.
- If the rule group taking a long time to evaluate belongs to `/sg_config_prometheus`, please [open an issue](https://github.com/sourcegraph/sourcegraph/issues/new?assignees=&labels=&template=bug_report.md&title=).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_prometheus_prometheus_rule_eval_duration"
]
```

> NOTE: More help interpreting this metric is available in the [dashboards reference](./dashboards.md#prometheus-prometheus-rule-eval-duration).

<sub>*Managed by the [Sourcegraph Distribution team](https://about.sourcegraph.com/handbook/engineering/distribution).*</sub>

<br />

## prometheus: prometheus_rule_eval_failures

<p class="subtitle">failed prometheus rule evaluations over 5m by rule group</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> prometheus: 0+ failed prometheus rule evaluations over 5m by rule group

**Possible solutions**

- Check Prometheus logs for messages related to rule group evaluation (generally with log field `component="rule manager"`).
- If the rule group failing to evaluate belongs to `/sg_prometheus_addons`, ensure any custom Prometheus configuration provided is valid.
- If the rule group taking a long time to evaluate belongs to `/sg_config_prometheus`, please [open an issue](https://github.com/sourcegraph/sourcegraph/issues/new?assignees=&labels=&template=bug_report.md&title=).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_prometheus_prometheus_rule_eval_failures"
]
```

> NOTE: More help interpreting this metric is available in the [dashboards reference](./dashboards.md#prometheus-prometheus-rule-eval-failures).

<sub>*Managed by the [Sourcegraph Distribution team](https://about.sourcegraph.com/handbook/engineering/distribution).*</sub>

<br />

## prometheus: alertmanager_notification_latency

<p class="subtitle">alertmanager notification latency over 1m by integration</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> prometheus: 1s+ alertmanager notification latency over 1m by integration

**Possible solutions**

- Check the Container monitoring (not available on server) panels and try increasing resources for Prometheus if necessary.
- Ensure that your [`observability.alerts` configuration](https://docs.sourcegraph.com/admin/observability/alerting#setting-up-alerting) (in site configuration) is valid.
- Check if the relevant alert integration service is experiencing downtime or issues.
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_prometheus_alertmanager_notification_latency"
]
```

<sub>*Managed by the [Sourcegraph Distribution team](https://about.sourcegraph.com/handbook/engineering/distribution).*</sub>

<br />

## prometheus: alertmanager_notification_failures

<p class="subtitle">failed alertmanager notifications over 1m by integration</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> prometheus: 0+ failed alertmanager notifications over 1m by integration

**Possible solutions**

- Ensure that your [`observability.alerts` configuration](https://docs.sourcegraph.com/admin/observability/alerting#setting-up-alerting) (in site configuration) is valid.
- Check if the relevant alert integration service is experiencing downtime or issues.
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_prometheus_alertmanager_notification_failures"
]
```

<sub>*Managed by the [Sourcegraph Distribution team](https://about.sourcegraph.com/handbook/engineering/distribution).*</sub>

<br />

## prometheus: prometheus_config_status

<p class="subtitle">prometheus configuration reload status</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> prometheus: less than 1 prometheus configuration reload status

**Possible solutions**

- Check Prometheus logs for messages related to configuration loading.
- Ensure any [custom configuration you have provided Prometheus](https://docs.sourcegraph.com/admin/observability/metrics#prometheus-configuration) is valid.
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_prometheus_prometheus_config_status"
]
```

> NOTE: More help interpreting this metric is available in the [dashboards reference](./dashboards.md#prometheus-prometheus-config-status).

<sub>*Managed by the [Sourcegraph Distribution team](https://about.sourcegraph.com/handbook/engineering/distribution).*</sub>

<br />

## prometheus: alertmanager_config_status

<p class="subtitle">alertmanager configuration reload status</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> prometheus: less than 1 alertmanager configuration reload status

**Possible solutions**

- Ensure that your [`observability.alerts` configuration](https://docs.sourcegraph.com/admin/observability/alerting#setting-up-alerting) (in site configuration) is valid.
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_prometheus_alertmanager_config_status"
]
```

> NOTE: More help interpreting this metric is available in the [dashboards reference](./dashboards.md#prometheus-alertmanager-config-status).

<sub>*Managed by the [Sourcegraph Distribution team](https://about.sourcegraph.com/handbook/engineering/distribution).*</sub>

<br />

## prometheus: prometheus_tsdb_op_failure

<p class="subtitle">prometheus tsdb failures by operation over 1m by operation</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> prometheus: 0+ prometheus tsdb failures by operation over 1m by operation

**Possible solutions**

- Check Prometheus logs for messages related to the failing operation.
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_prometheus_prometheus_tsdb_op_failure"
]
```

<sub>*Managed by the [Sourcegraph Distribution team](https://about.sourcegraph.com/handbook/engineering/distribution).*</sub>

<br />

## prometheus: prometheus_target_sample_exceeded

<p class="subtitle">prometheus scrapes that exceed the sample limit over 10m</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> prometheus: 0+ prometheus scrapes that exceed the sample limit over 10m

**Possible solutions**

- Check Prometheus logs for messages related to target scrape failures.
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_prometheus_prometheus_target_sample_exceeded"
]
```

<sub>*Managed by the [Sourcegraph Distribution team](https://about.sourcegraph.com/handbook/engineering/distribution).*</sub>

<br />

## prometheus: prometheus_target_sample_duplicate

<p class="subtitle">prometheus scrapes rejected due to duplicate timestamps over 10m</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> prometheus: 0+ prometheus scrapes rejected due to duplicate timestamps over 10m

**Possible solutions**

- Check Prometheus logs for messages related to target scrape failures.
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_prometheus_prometheus_target_sample_duplicate"
]
```

<sub>*Managed by the [Sourcegraph Distribution team](https://about.sourcegraph.com/handbook/engineering/distribution).*</sub>

<br />

## prometheus: container_cpu_usage

<p class="subtitle">container cpu usage total (1m average) across all cores by instance</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> prometheus: 99%+ container cpu usage total (1m average) across all cores by instance

**Possible solutions**

- **Kubernetes:** Consider increasing CPU limits in the the relevant `Deployment.yaml`.
- **Docker Compose:** Consider increasing `cpus:` of the prometheus container in `docker-compose.yml`.
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_prometheus_container_cpu_usage"
]
```

<sub>*Managed by the [Sourcegraph Distribution team](https://about.sourcegraph.com/handbook/engineering/distribution).*</sub>

<br />

## prometheus: container_memory_usage

<p class="subtitle">container memory usage by instance</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> prometheus: 99%+ container memory usage by instance

**Possible solutions**

- **Kubernetes:** Consider increasing memory limit in relevant `Deployment.yaml`.
- **Docker Compose:** Consider increasing `memory:` of prometheus container in `docker-compose.yml`.
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_prometheus_container_memory_usage"
]
```

<sub>*Managed by the [Sourcegraph Distribution team](https://about.sourcegraph.com/handbook/engineering/distribution).*</sub>

<br />

## prometheus: provisioning_container_cpu_usage_long_term

<p class="subtitle">container cpu usage total (90th percentile over 1d) across all cores by instance</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> prometheus: 80%+ container cpu usage total (90th percentile over 1d) across all cores by instance for 336h0m0s

**Possible solutions**

- **Kubernetes:** Consider increasing CPU limits in the `Deployment.yaml` for the prometheus service.
- **Docker Compose:** Consider increasing `cpus:` of the prometheus container in `docker-compose.yml`.
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_prometheus_provisioning_container_cpu_usage_long_term"
]
```

<sub>*Managed by the [Sourcegraph Distribution team](https://about.sourcegraph.com/handbook/engineering/distribution).*</sub>

<br />

## prometheus: provisioning_container_memory_usage_long_term

<p class="subtitle">container memory usage (1d maximum) by instance</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> prometheus: 80%+ container memory usage (1d maximum) by instance for 336h0m0s

**Possible solutions**

- **Kubernetes:** Consider increasing memory limits in the `Deployment.yaml` for the prometheus service.
- **Docker Compose:** Consider increasing `memory:` of the prometheus container in `docker-compose.yml`.
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_prometheus_provisioning_container_memory_usage_long_term"
]
```

<sub>*Managed by the [Sourcegraph Distribution team](https://about.sourcegraph.com/handbook/engineering/distribution).*</sub>

<br />

## prometheus: provisioning_container_cpu_usage_short_term

<p class="subtitle">container cpu usage total (5m maximum) across all cores by instance</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> prometheus: 90%+ container cpu usage total (5m maximum) across all cores by instance for 30m0s

**Possible solutions**

- **Kubernetes:** Consider increasing CPU limits in the the relevant `Deployment.yaml`.
- **Docker Compose:** Consider increasing `cpus:` of the prometheus container in `docker-compose.yml`.
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_prometheus_provisioning_container_cpu_usage_short_term"
]
```

<sub>*Managed by the [Sourcegraph Distribution team](https://about.sourcegraph.com/handbook/engineering/distribution).*</sub>

<br />

## prometheus: provisioning_container_memory_usage_short_term

<p class="subtitle">container memory usage (5m maximum) by instance</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> prometheus: 90%+ container memory usage (5m maximum) by instance

**Possible solutions**

- **Kubernetes:** Consider increasing memory limit in relevant `Deployment.yaml`.
- **Docker Compose:** Consider increasing `memory:` of prometheus container in `docker-compose.yml`.
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_prometheus_provisioning_container_memory_usage_short_term"
]
```

<sub>*Managed by the [Sourcegraph Distribution team](https://about.sourcegraph.com/handbook/engineering/distribution).*</sub>

<br />

## prometheus: pods_available_percentage

<p class="subtitle">percentage pods available</p>

**Descriptions**

- <span class="badge badge-critical">critical</span> prometheus: less than 90% percentage pods available for 10m0s

**Possible solutions**

- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "critical_prometheus_pods_available_percentage"
]
```

<sub>*Managed by the [Sourcegraph Distribution team](https://about.sourcegraph.com/handbook/engineering/distribution).*</sub>

<br />

## executor: container_cpu_usage

<p class="subtitle">container cpu usage total (1m average) across all cores by instance</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> executor: 99%+ container cpu usage total (1m average) across all cores by instance

**Possible solutions**

- **Kubernetes:** Consider increasing CPU limits in the the relevant `Deployment.yaml`.
- **Docker Compose:** Consider increasing `cpus:` of the (executor|sourcegraph-code-intel-indexers|executor-batches) container in `docker-compose.yml`.
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_executor_container_cpu_usage"
]
```

<sub>*Managed by the [Sourcegraph Code-intelligence team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<br />

## executor: container_memory_usage

<p class="subtitle">container memory usage by instance</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> executor: 99%+ container memory usage by instance

**Possible solutions**

- **Kubernetes:** Consider increasing memory limit in relevant `Deployment.yaml`.
- **Docker Compose:** Consider increasing `memory:` of (executor|sourcegraph-code-intel-indexers|executor-batches) container in `docker-compose.yml`.
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_executor_container_memory_usage"
]
```

<sub>*Managed by the [Sourcegraph Code-intelligence team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<br />

## executor: provisioning_container_cpu_usage_long_term

<p class="subtitle">container cpu usage total (90th percentile over 1d) across all cores by instance</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> executor: 80%+ container cpu usage total (90th percentile over 1d) across all cores by instance for 336h0m0s

**Possible solutions**

- **Kubernetes:** Consider increasing CPU limits in the `Deployment.yaml` for the (executor|sourcegraph-code-intel-indexers|executor-batches) service.
- **Docker Compose:** Consider increasing `cpus:` of the (executor|sourcegraph-code-intel-indexers|executor-batches) container in `docker-compose.yml`.
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_executor_provisioning_container_cpu_usage_long_term"
]
```

<sub>*Managed by the [Sourcegraph Code-intelligence team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<br />

## executor: provisioning_container_memory_usage_long_term

<p class="subtitle">container memory usage (1d maximum) by instance</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> executor: 80%+ container memory usage (1d maximum) by instance for 336h0m0s

**Possible solutions**

- **Kubernetes:** Consider increasing memory limits in the `Deployment.yaml` for the (executor|sourcegraph-code-intel-indexers|executor-batches) service.
- **Docker Compose:** Consider increasing `memory:` of the (executor|sourcegraph-code-intel-indexers|executor-batches) container in `docker-compose.yml`.
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_executor_provisioning_container_memory_usage_long_term"
]
```

<sub>*Managed by the [Sourcegraph Code-intelligence team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<br />

## executor: provisioning_container_cpu_usage_short_term

<p class="subtitle">container cpu usage total (5m maximum) across all cores by instance</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> executor: 90%+ container cpu usage total (5m maximum) across all cores by instance for 30m0s

**Possible solutions**

- **Kubernetes:** Consider increasing CPU limits in the the relevant `Deployment.yaml`.
- **Docker Compose:** Consider increasing `cpus:` of the (executor|sourcegraph-code-intel-indexers|executor-batches) container in `docker-compose.yml`.
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_executor_provisioning_container_cpu_usage_short_term"
]
```

<sub>*Managed by the [Sourcegraph Code-intelligence team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<br />

## executor: provisioning_container_memory_usage_short_term

<p class="subtitle">container memory usage (5m maximum) by instance</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> executor: 90%+ container memory usage (5m maximum) by instance

**Possible solutions**

- **Kubernetes:** Consider increasing memory limit in relevant `Deployment.yaml`.
- **Docker Compose:** Consider increasing `memory:` of (executor|sourcegraph-code-intel-indexers|executor-batches) container in `docker-compose.yml`.
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_executor_provisioning_container_memory_usage_short_term"
]
```

<sub>*Managed by the [Sourcegraph Code-intelligence team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<br />

## executor: go_goroutines

<p class="subtitle">maximum active goroutines</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> executor: 10000+ maximum active goroutines for 10m0s

**Possible solutions**

- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_executor_go_goroutines"
]
```

> NOTE: More help interpreting this metric is available in the [dashboards reference](./dashboards.md#executor-go-goroutines).

<sub>*Managed by the [Sourcegraph Code-intelligence team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<br />

## executor: go_gc_duration_seconds

<p class="subtitle">maximum go garbage collection duration</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> executor: 2s+ maximum go garbage collection duration

**Possible solutions**

- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_executor_go_gc_duration_seconds"
]
```

<sub>*Managed by the [Sourcegraph Code-intelligence team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<br />

## executor: pods_available_percentage

<p class="subtitle">percentage pods available</p>

**Descriptions**

- <span class="badge badge-critical">critical</span> executor: less than 90% percentage pods available for 10m0s

**Possible solutions**

- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "critical_executor_pods_available_percentage"
]
```

<sub>*Managed by the [Sourcegraph Code-intelligence team](https://about.sourcegraph.com/handbook/engineering/code-intelligence).*</sub>

<br />

