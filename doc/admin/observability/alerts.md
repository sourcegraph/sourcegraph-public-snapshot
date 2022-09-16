# Alerts reference

<!-- DO NOT EDIT: generated via: go generate ./monitoring -->

This document contains a complete reference of all alerts in Sourcegraph's monitoring, and next steps for when you find alerts that are firing.
If your alert isn't mentioned here, or if the next steps don't help, [contact us](mailto:support@sourcegraph.com) for assistance.

To learn more about Sourcegraph's alerting and how to set up alerts, see [our alerting guide](https://docs.sourcegraph.com/admin/observability/alerting).

## frontend: 99th_percentile_search_request_duration

<p class="subtitle">99th percentile successful search request duration over 5m</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> frontend: 20s+ 99th percentile successful search request duration over 5m

**Next steps**

- **Get details on the exact queries that are slow** by configuring `"observability.logSlowSearches": 20,` in the site configuration and looking for `frontend` warning logs prefixed with `slow search request` for additional details.
- **Check that most repositories are indexed** by visiting https://sourcegraph.example.com/site-admin/repositories?filter=needs-index (it should show few or no results.)
- **Kubernetes:** Check CPU usage of zoekt-webserver in the indexed-search pod, consider increasing CPU limits in the `indexed-search.Deployment.yaml` if regularly hitting max CPU utilization.
- **Docker Compose:** Check CPU usage on the Zoekt Web Server dashboard, consider increasing `cpus:` of the zoekt-webserver container in `docker-compose.yml` if regularly hitting max CPU utilization.
- Learn more about the related dashboard panel in the [dashboards reference](./dashboards.md#frontend-99th-percentile-search-request-duration).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_frontend_99th_percentile_search_request_duration"
]
```

<sub>*Managed by the [Sourcegraph Search team](https://handbook.sourcegraph.com/departments/engineering/teams/search/product).*</sub>

<br />

## frontend: 90th_percentile_search_request_duration

<p class="subtitle">90th percentile successful search request duration over 5m</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> frontend: 15s+ 90th percentile successful search request duration over 5m

**Next steps**

- **Get details on the exact queries that are slow** by configuring `"observability.logSlowSearches": 15,` in the site configuration and looking for `frontend` warning logs prefixed with `slow search request` for additional details.
- **Check that most repositories are indexed** by visiting https://sourcegraph.example.com/site-admin/repositories?filter=needs-index (it should show few or no results.)
- **Kubernetes:** Check CPU usage of zoekt-webserver in the indexed-search pod, consider increasing CPU limits in the `indexed-search.Deployment.yaml` if regularly hitting max CPU utilization.
- **Docker Compose:** Check CPU usage on the Zoekt Web Server dashboard, consider increasing `cpus:` of the zoekt-webserver container in `docker-compose.yml` if regularly hitting max CPU utilization.
- Learn more about the related dashboard panel in the [dashboards reference](./dashboards.md#frontend-90th-percentile-search-request-duration).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_frontend_90th_percentile_search_request_duration"
]
```

<sub>*Managed by the [Sourcegraph Search team](https://handbook.sourcegraph.com/departments/engineering/teams/search/product).*</sub>

<br />

## frontend: hard_timeout_search_responses

<p class="subtitle">hard timeout search responses every 5m</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> frontend: 2%+ hard timeout search responses every 5m for 15m0s

**Next steps**

- Learn more about the related dashboard panel in the [dashboards reference](./dashboards.md#frontend-hard-timeout-search-responses).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_frontend_hard_timeout_search_responses"
]
```

<sub>*Managed by the [Sourcegraph Search team](https://handbook.sourcegraph.com/departments/engineering/teams/search/product).*</sub>

<br />

## frontend: hard_error_search_responses

<p class="subtitle">hard error search responses every 5m</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> frontend: 2%+ hard error search responses every 5m for 15m0s

**Next steps**

- Learn more about the related dashboard panel in the [dashboards reference](./dashboards.md#frontend-hard-error-search-responses).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_frontend_hard_error_search_responses"
]
```

<sub>*Managed by the [Sourcegraph Search team](https://handbook.sourcegraph.com/departments/engineering/teams/search/product).*</sub>

<br />

## frontend: partial_timeout_search_responses

<p class="subtitle">partial timeout search responses every 5m</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> frontend: 5%+ partial timeout search responses every 5m for 15m0s

**Next steps**

- Learn more about the related dashboard panel in the [dashboards reference](./dashboards.md#frontend-partial-timeout-search-responses).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_frontend_partial_timeout_search_responses"
]
```

<sub>*Managed by the [Sourcegraph Search team](https://handbook.sourcegraph.com/departments/engineering/teams/search/product).*</sub>

<br />

## frontend: search_alert_user_suggestions

<p class="subtitle">search alert user suggestions shown every 5m</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> frontend: 5%+ search alert user suggestions shown every 5m for 15m0s

**Next steps**

- This indicates your user`s are making syntax errors or similar user errors.
- Learn more about the related dashboard panel in the [dashboards reference](./dashboards.md#frontend-search-alert-user-suggestions).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_frontend_search_alert_user_suggestions"
]
```

<sub>*Managed by the [Sourcegraph Search team](https://handbook.sourcegraph.com/departments/engineering/teams/search/product).*</sub>

<br />

## frontend: page_load_latency

<p class="subtitle">90th percentile page load latency over all routes over 10m</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> frontend: 2s+ 90th percentile page load latency over all routes over 10m

**Next steps**

- Confirm that the Sourcegraph frontend has enough CPU/memory using the provisioning panels.
- Investigate potential sources of latency by selecting Explore and modifying the `sum by(le)` section to include additional labels: for example, `sum by(le, job)` or `sum by (le, instance)`.
- Trace a request to see what the slowest part is: https://docs.sourcegraph.com/admin/observability/tracing
- Learn more about the related dashboard panel in the [dashboards reference](./dashboards.md#frontend-page-load-latency).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_frontend_page_load_latency"
]
```

<sub>*Managed by the [Sourcegraph Identity and Access Management team](https://handbook.sourcegraph.com/departments/engineering/teams/iam).*</sub>

<br />

## frontend: blob_load_latency

<p class="subtitle">90th percentile blob load latency over 10m</p>

**Descriptions**

- <span class="badge badge-critical">critical</span> frontend: 5s+ 90th percentile blob load latency over 10m

**Next steps**

- Confirm that the Sourcegraph frontend has enough CPU/memory using the provisioning panels.
- Trace a request to see what the slowest part is: https://docs.sourcegraph.com/admin/observability/tracing
- Check that gitserver containers have enough CPU/memory and are not getting throttled.
- Learn more about the related dashboard panel in the [dashboards reference](./dashboards.md#frontend-blob-load-latency).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "critical_frontend_blob_load_latency"
]
```

<sub>*Managed by the [Sourcegraph Repo Management team](https://handbook.sourcegraph.com/departments/engineering/teams/repo-management).*</sub>

<br />

## frontend: 99th_percentile_search_codeintel_request_duration

<p class="subtitle">99th percentile code-intel successful search request duration over 5m</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> frontend: 20s+ 99th percentile code-intel successful search request duration over 5m

**Next steps**

- **Get details on the exact queries that are slow** by configuring `"observability.logSlowSearches": 20,` in the site configuration and looking for `frontend` warning logs prefixed with `slow search request` for additional details.
- **Check that most repositories are indexed** by visiting https://sourcegraph.example.com/site-admin/repositories?filter=needs-index (it should show few or no results.)
- **Kubernetes:** Check CPU usage of zoekt-webserver in the indexed-search pod, consider increasing CPU limits in the `indexed-search.Deployment.yaml` if regularly hitting max CPU utilization.
- **Docker Compose:** Check CPU usage on the Zoekt Web Server dashboard, consider increasing `cpus:` of the zoekt-webserver container in `docker-compose.yml` if regularly hitting max CPU utilization.
- This alert may indicate that your instance is struggling to process symbols queries on a monorepo, [learn more here](../how-to/monorepo-issues.md).
- Learn more about the related dashboard panel in the [dashboards reference](./dashboards.md#frontend-99th-percentile-search-codeintel-request-duration).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_frontend_99th_percentile_search_codeintel_request_duration"
]
```

<sub>*Managed by the [Sourcegraph Search team](https://handbook.sourcegraph.com/departments/engineering/teams/search/product).*</sub>

<br />

## frontend: 90th_percentile_search_codeintel_request_duration

<p class="subtitle">90th percentile code-intel successful search request duration over 5m</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> frontend: 15s+ 90th percentile code-intel successful search request duration over 5m

**Next steps**

- **Get details on the exact queries that are slow** by configuring `"observability.logSlowSearches": 15,` in the site configuration and looking for `frontend` warning logs prefixed with `slow search request` for additional details.
- **Check that most repositories are indexed** by visiting https://sourcegraph.example.com/site-admin/repositories?filter=needs-index (it should show few or no results.)
- **Kubernetes:** Check CPU usage of zoekt-webserver in the indexed-search pod, consider increasing CPU limits in the `indexed-search.Deployment.yaml` if regularly hitting max CPU utilization.
- **Docker Compose:** Check CPU usage on the Zoekt Web Server dashboard, consider increasing `cpus:` of the zoekt-webserver container in `docker-compose.yml` if regularly hitting max CPU utilization.
- This alert may indicate that your instance is struggling to process symbols queries on a monorepo, [learn more here](../how-to/monorepo-issues.md).
- Learn more about the related dashboard panel in the [dashboards reference](./dashboards.md#frontend-90th-percentile-search-codeintel-request-duration).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_frontend_90th_percentile_search_codeintel_request_duration"
]
```

<sub>*Managed by the [Sourcegraph Search team](https://handbook.sourcegraph.com/departments/engineering/teams/search/product).*</sub>

<br />

## frontend: hard_timeout_search_codeintel_responses

<p class="subtitle">hard timeout search code-intel responses every 5m</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> frontend: 2%+ hard timeout search code-intel responses every 5m for 15m0s

**Next steps**

- Learn more about the related dashboard panel in the [dashboards reference](./dashboards.md#frontend-hard-timeout-search-codeintel-responses).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_frontend_hard_timeout_search_codeintel_responses"
]
```

<sub>*Managed by the [Sourcegraph Search team](https://handbook.sourcegraph.com/departments/engineering/teams/search/product).*</sub>

<br />

## frontend: hard_error_search_codeintel_responses

<p class="subtitle">hard error search code-intel responses every 5m</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> frontend: 2%+ hard error search code-intel responses every 5m for 15m0s

**Next steps**

- Learn more about the related dashboard panel in the [dashboards reference](./dashboards.md#frontend-hard-error-search-codeintel-responses).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_frontend_hard_error_search_codeintel_responses"
]
```

<sub>*Managed by the [Sourcegraph Search team](https://handbook.sourcegraph.com/departments/engineering/teams/search/product).*</sub>

<br />

## frontend: partial_timeout_search_codeintel_responses

<p class="subtitle">partial timeout search code-intel responses every 5m</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> frontend: 5%+ partial timeout search code-intel responses every 5m for 15m0s

**Next steps**

- Learn more about the related dashboard panel in the [dashboards reference](./dashboards.md#frontend-partial-timeout-search-codeintel-responses).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_frontend_partial_timeout_search_codeintel_responses"
]
```

<sub>*Managed by the [Sourcegraph Search team](https://handbook.sourcegraph.com/departments/engineering/teams/search/product).*</sub>

<br />

## frontend: search_codeintel_alert_user_suggestions

<p class="subtitle">search code-intel alert user suggestions shown every 5m</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> frontend: 5%+ search code-intel alert user suggestions shown every 5m for 15m0s

**Next steps**

- This indicates a bug in Sourcegraph, please [open an issue](https://github.com/sourcegraph/sourcegraph/issues/new/choose).
- Learn more about the related dashboard panel in the [dashboards reference](./dashboards.md#frontend-search-codeintel-alert-user-suggestions).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_frontend_search_codeintel_alert_user_suggestions"
]
```

<sub>*Managed by the [Sourcegraph Search team](https://handbook.sourcegraph.com/departments/engineering/teams/search/product).*</sub>

<br />

## frontend: 99th_percentile_search_api_request_duration

<p class="subtitle">99th percentile successful search API request duration over 5m</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> frontend: 50s+ 99th percentile successful search API request duration over 5m

**Next steps**

- **Get details on the exact queries that are slow** by configuring `"observability.logSlowSearches": 20,` in the site configuration and looking for `frontend` warning logs prefixed with `slow search request` for additional details.
- **Check that most repositories are indexed** by visiting https://sourcegraph.example.com/site-admin/repositories?filter=needs-index (it should show few or no results.)
- **Kubernetes:** Check CPU usage of zoekt-webserver in the indexed-search pod, consider increasing CPU limits in the `indexed-search.Deployment.yaml` if regularly hitting max CPU utilization.
- **Docker Compose:** Check CPU usage on the Zoekt Web Server dashboard, consider increasing `cpus:` of the zoekt-webserver container in `docker-compose.yml` if regularly hitting max CPU utilization.
- Learn more about the related dashboard panel in the [dashboards reference](./dashboards.md#frontend-99th-percentile-search-api-request-duration).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_frontend_99th_percentile_search_api_request_duration"
]
```

<sub>*Managed by the [Sourcegraph Search team](https://handbook.sourcegraph.com/departments/engineering/teams/search/product).*</sub>

<br />

## frontend: 90th_percentile_search_api_request_duration

<p class="subtitle">90th percentile successful search API request duration over 5m</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> frontend: 40s+ 90th percentile successful search API request duration over 5m

**Next steps**

- **Get details on the exact queries that are slow** by configuring `"observability.logSlowSearches": 15,` in the site configuration and looking for `frontend` warning logs prefixed with `slow search request` for additional details.
- **Check that most repositories are indexed** by visiting https://sourcegraph.example.com/site-admin/repositories?filter=needs-index (it should show few or no results.)
- **Kubernetes:** Check CPU usage of zoekt-webserver in the indexed-search pod, consider increasing CPU limits in the `indexed-search.Deployment.yaml` if regularly hitting max CPU utilization.
- **Docker Compose:** Check CPU usage on the Zoekt Web Server dashboard, consider increasing `cpus:` of the zoekt-webserver container in `docker-compose.yml` if regularly hitting max CPU utilization.
- Learn more about the related dashboard panel in the [dashboards reference](./dashboards.md#frontend-90th-percentile-search-api-request-duration).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_frontend_90th_percentile_search_api_request_duration"
]
```

<sub>*Managed by the [Sourcegraph Search team](https://handbook.sourcegraph.com/departments/engineering/teams/search/product).*</sub>

<br />

## frontend: hard_error_search_api_responses

<p class="subtitle">hard error search API responses every 5m</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> frontend: 2%+ hard error search API responses every 5m for 15m0s

**Next steps**

- Learn more about the related dashboard panel in the [dashboards reference](./dashboards.md#frontend-hard-error-search-api-responses).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_frontend_hard_error_search_api_responses"
]
```

<sub>*Managed by the [Sourcegraph Search team](https://handbook.sourcegraph.com/departments/engineering/teams/search/product).*</sub>

<br />

## frontend: partial_timeout_search_api_responses

<p class="subtitle">partial timeout search API responses every 5m</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> frontend: 5%+ partial timeout search API responses every 5m for 15m0s

**Next steps**

- Learn more about the related dashboard panel in the [dashboards reference](./dashboards.md#frontend-partial-timeout-search-api-responses).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_frontend_partial_timeout_search_api_responses"
]
```

<sub>*Managed by the [Sourcegraph Search team](https://handbook.sourcegraph.com/departments/engineering/teams/search/product).*</sub>

<br />

## frontend: search_api_alert_user_suggestions

<p class="subtitle">search API alert user suggestions shown every 5m</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> frontend: 5%+ search API alert user suggestions shown every 5m

**Next steps**

- This indicates your user`s search API requests have syntax errors or a similar user error. Check the responses the API sends back for an explanation.
- Learn more about the related dashboard panel in the [dashboards reference](./dashboards.md#frontend-search-api-alert-user-suggestions).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_frontend_search_api_alert_user_suggestions"
]
```

<sub>*Managed by the [Sourcegraph Search team](https://handbook.sourcegraph.com/departments/engineering/teams/search/product).*</sub>

<br />

## frontend: internal_indexed_search_error_responses

<p class="subtitle">internal indexed search error responses every 5m</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> frontend: 5%+ internal indexed search error responses every 5m for 15m0s

**Next steps**

- Check the Zoekt Web Server dashboard for indications it might be unhealthy.
- Learn more about the related dashboard panel in the [dashboards reference](./dashboards.md#frontend-internal-indexed-search-error-responses).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_frontend_internal_indexed_search_error_responses"
]
```

<sub>*Managed by the [Sourcegraph Search team](https://handbook.sourcegraph.com/departments/engineering/teams/search/product).*</sub>

<br />

## frontend: internal_unindexed_search_error_responses

<p class="subtitle">internal unindexed search error responses every 5m</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> frontend: 5%+ internal unindexed search error responses every 5m for 15m0s

**Next steps**

- Check the Searcher dashboard for indications it might be unhealthy.
- Learn more about the related dashboard panel in the [dashboards reference](./dashboards.md#frontend-internal-unindexed-search-error-responses).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_frontend_internal_unindexed_search_error_responses"
]
```

<sub>*Managed by the [Sourcegraph Search team](https://handbook.sourcegraph.com/departments/engineering/teams/search/product).*</sub>

<br />

## frontend: internalapi_error_responses

<p class="subtitle">internal API error responses every 5m by route</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> frontend: 5%+ internal API error responses every 5m by route for 15m0s

**Next steps**

- May not be a substantial issue, check the `frontend` logs for potential causes.
- Learn more about the related dashboard panel in the [dashboards reference](./dashboards.md#frontend-internalapi-error-responses).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_frontend_internalapi_error_responses"
]
```

<sub>*Managed by the [Sourcegraph Identity and Access Management team](https://handbook.sourcegraph.com/departments/engineering/teams/iam).*</sub>

<br />

## frontend: 99th_percentile_gitserver_duration

<p class="subtitle">99th percentile successful gitserver query duration over 5m</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> frontend: 20s+ 99th percentile successful gitserver query duration over 5m

**Next steps**

- Learn more about the related dashboard panel in the [dashboards reference](./dashboards.md#frontend-99th-percentile-gitserver-duration).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_frontend_99th_percentile_gitserver_duration"
]
```

<sub>*Managed by the [Sourcegraph Repo Management team](https://handbook.sourcegraph.com/departments/engineering/teams/repo-management).*</sub>

<br />

## frontend: gitserver_error_responses

<p class="subtitle">gitserver error responses every 5m</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> frontend: 5%+ gitserver error responses every 5m for 15m0s

**Next steps**

- Learn more about the related dashboard panel in the [dashboards reference](./dashboards.md#frontend-gitserver-error-responses).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_frontend_gitserver_error_responses"
]
```

<sub>*Managed by the [Sourcegraph Repo Management team](https://handbook.sourcegraph.com/departments/engineering/teams/repo-management).*</sub>

<br />

## frontend: observability_test_alert_warning

<p class="subtitle">warning test alert metric</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> frontend: 1+ warning test alert metric

**Next steps**

- This alert is triggered via the `triggerObservabilityTestAlert` GraphQL endpoint, and will automatically resolve itself.
- Learn more about the related dashboard panel in the [dashboards reference](./dashboards.md#frontend-observability-test-alert-warning).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_frontend_observability_test_alert_warning"
]
```

<sub>*Managed by the [Sourcegraph Cloud DevOps team](https://handbook.sourcegraph.com/departments/engineering/teams/devops).*</sub>

<br />

## frontend: observability_test_alert_critical

<p class="subtitle">critical test alert metric</p>

**Descriptions**

- <span class="badge badge-critical">critical</span> frontend: 1+ critical test alert metric

**Next steps**

- This alert is triggered via the `triggerObservabilityTestAlert` GraphQL endpoint, and will automatically resolve itself.
- Learn more about the related dashboard panel in the [dashboards reference](./dashboards.md#frontend-observability-test-alert-critical).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "critical_frontend_observability_test_alert_critical"
]
```

<sub>*Managed by the [Sourcegraph Cloud DevOps team](https://handbook.sourcegraph.com/departments/engineering/teams/devops).*</sub>

<br />

## frontend: cloudkms_cryptographic_requests

<p class="subtitle">cryptographic requests to Cloud KMS every 1m</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> frontend: 15000+ cryptographic requests to Cloud KMS every 1m for 5m0s
- <span class="badge badge-critical">critical</span> frontend: 30000+ cryptographic requests to Cloud KMS every 1m for 5m0s

**Next steps**

- Revert recent commits that cause extensive listing from "external_services" and/or "user_external_accounts" tables.
- Learn more about the related dashboard panel in the [dashboards reference](./dashboards.md#frontend-cloudkms-cryptographic-requests).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_frontend_cloudkms_cryptographic_requests",
  "critical_frontend_cloudkms_cryptographic_requests"
]
```

<sub>*Managed by the [Sourcegraph Repo Management team](https://handbook.sourcegraph.com/departments/engineering/teams/repo-management).*</sub>

<br />

## frontend: mean_blocked_seconds_per_conn_request

<p class="subtitle">mean blocked seconds per conn request</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> frontend: 0.05s+ mean blocked seconds per conn request for 10m0s
- <span class="badge badge-critical">critical</span> frontend: 0.1s+ mean blocked seconds per conn request for 15m0s

**Next steps**

- Increase SRC_PGSQL_MAX_OPEN together with giving more memory to the database if needed
- Scale up Postgres memory / cpus [See our scaling guide](https://docs.sourcegraph.com/admin/config/postgres-conf)
- Learn more about the related dashboard panel in the [dashboards reference](./dashboards.md#frontend-mean-blocked-seconds-per-conn-request).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_frontend_mean_blocked_seconds_per_conn_request",
  "critical_frontend_mean_blocked_seconds_per_conn_request"
]
```

<sub>*Managed by the [Sourcegraph Cloud DevOps team](https://handbook.sourcegraph.com/departments/engineering/teams/devops).*</sub>

<br />

## frontend: container_cpu_usage

<p class="subtitle">container cpu usage total (1m average) across all cores by instance</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> frontend: 99%+ container cpu usage total (1m average) across all cores by instance

**Next steps**

- **Kubernetes:** Consider increasing CPU limits in the the relevant `Deployment.yaml`.
- **Docker Compose:** Consider increasing `cpus:` of the (frontend|sourcegraph-frontend) container in `docker-compose.yml`.
- Learn more about the related dashboard panel in the [dashboards reference](./dashboards.md#frontend-container-cpu-usage).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_frontend_container_cpu_usage"
]
```

<sub>*Managed by the [Sourcegraph Cloud DevOps team](https://handbook.sourcegraph.com/departments/engineering/teams/devops).*</sub>

<br />

## frontend: container_memory_usage

<p class="subtitle">container memory usage by instance</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> frontend: 99%+ container memory usage by instance

**Next steps**

- **Kubernetes:** Consider increasing memory limit in relevant `Deployment.yaml`.
- **Docker Compose:** Consider increasing `memory:` of (frontend|sourcegraph-frontend) container in `docker-compose.yml`.
- Learn more about the related dashboard panel in the [dashboards reference](./dashboards.md#frontend-container-memory-usage).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_frontend_container_memory_usage"
]
```

<sub>*Managed by the [Sourcegraph Cloud DevOps team](https://handbook.sourcegraph.com/departments/engineering/teams/devops).*</sub>

<br />

## frontend: provisioning_container_cpu_usage_long_term

<p class="subtitle">container cpu usage total (90th percentile over 1d) across all cores by instance</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> frontend: 80%+ container cpu usage total (90th percentile over 1d) across all cores by instance for 336h0m0s

**Next steps**

- **Kubernetes:** Consider increasing CPU limits in the `Deployment.yaml` for the (frontend|sourcegraph-frontend) service.
- **Docker Compose:** Consider increasing `cpus:` of the (frontend|sourcegraph-frontend) container in `docker-compose.yml`.
- Learn more about the related dashboard panel in the [dashboards reference](./dashboards.md#frontend-provisioning-container-cpu-usage-long-term).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_frontend_provisioning_container_cpu_usage_long_term"
]
```

<sub>*Managed by the [Sourcegraph Cloud DevOps team](https://handbook.sourcegraph.com/departments/engineering/teams/devops).*</sub>

<br />

## frontend: provisioning_container_memory_usage_long_term

<p class="subtitle">container memory usage (1d maximum) by instance</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> frontend: 80%+ container memory usage (1d maximum) by instance for 336h0m0s

**Next steps**

- **Kubernetes:** Consider increasing memory limits in the `Deployment.yaml` for the (frontend|sourcegraph-frontend) service.
- **Docker Compose:** Consider increasing `memory:` of the (frontend|sourcegraph-frontend) container in `docker-compose.yml`.
- Learn more about the related dashboard panel in the [dashboards reference](./dashboards.md#frontend-provisioning-container-memory-usage-long-term).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_frontend_provisioning_container_memory_usage_long_term"
]
```

<sub>*Managed by the [Sourcegraph Cloud DevOps team](https://handbook.sourcegraph.com/departments/engineering/teams/devops).*</sub>

<br />

## frontend: provisioning_container_cpu_usage_short_term

<p class="subtitle">container cpu usage total (5m maximum) across all cores by instance</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> frontend: 90%+ container cpu usage total (5m maximum) across all cores by instance for 30m0s

**Next steps**

- **Kubernetes:** Consider increasing CPU limits in the the relevant `Deployment.yaml`.
- **Docker Compose:** Consider increasing `cpus:` of the (frontend|sourcegraph-frontend) container in `docker-compose.yml`.
- Learn more about the related dashboard panel in the [dashboards reference](./dashboards.md#frontend-provisioning-container-cpu-usage-short-term).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_frontend_provisioning_container_cpu_usage_short_term"
]
```

<sub>*Managed by the [Sourcegraph Cloud DevOps team](https://handbook.sourcegraph.com/departments/engineering/teams/devops).*</sub>

<br />

## frontend: provisioning_container_memory_usage_short_term

<p class="subtitle">container memory usage (5m maximum) by instance</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> frontend: 90%+ container memory usage (5m maximum) by instance

**Next steps**

- **Kubernetes:** Consider increasing memory limit in relevant `Deployment.yaml`.
- **Docker Compose:** Consider increasing `memory:` of (frontend|sourcegraph-frontend) container in `docker-compose.yml`.
- Learn more about the related dashboard panel in the [dashboards reference](./dashboards.md#frontend-provisioning-container-memory-usage-short-term).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_frontend_provisioning_container_memory_usage_short_term"
]
```

<sub>*Managed by the [Sourcegraph Cloud DevOps team](https://handbook.sourcegraph.com/departments/engineering/teams/devops).*</sub>

<br />

## frontend: container_oomkill_events_total

<p class="subtitle">container OOMKILL events total by instance</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> frontend: 1+ container OOMKILL events total by instance

**Next steps**

- **Kubernetes:** Consider increasing memory limit in relevant `Deployment.yaml`.
- **Docker Compose:** Consider increasing `memory:` of (frontend|sourcegraph-frontend) container in `docker-compose.yml`.
- More help interpreting this metric is available in the [dashboards reference](./dashboards.md#frontend-container-oomkill-events-total).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_frontend_container_oomkill_events_total"
]
```

<sub>*Managed by the [Sourcegraph Cloud DevOps team](https://handbook.sourcegraph.com/departments/engineering/teams/devops).*</sub>

<br />

## frontend: go_goroutines

<p class="subtitle">maximum active goroutines</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> frontend: 10000+ maximum active goroutines for 10m0s

**Next steps**

- More help interpreting this metric is available in the [dashboards reference](./dashboards.md#frontend-go-goroutines).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_frontend_go_goroutines"
]
```

<sub>*Managed by the [Sourcegraph Cloud DevOps team](https://handbook.sourcegraph.com/departments/engineering/teams/devops).*</sub>

<br />

## frontend: go_gc_duration_seconds

<p class="subtitle">maximum go garbage collection duration</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> frontend: 2s+ maximum go garbage collection duration

**Next steps**

- Learn more about the related dashboard panel in the [dashboards reference](./dashboards.md#frontend-go-gc-duration-seconds).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_frontend_go_gc_duration_seconds"
]
```

<sub>*Managed by the [Sourcegraph Cloud DevOps team](https://handbook.sourcegraph.com/departments/engineering/teams/devops).*</sub>

<br />

## frontend: pods_available_percentage

<p class="subtitle">percentage pods available</p>

**Descriptions**

- <span class="badge badge-critical">critical</span> frontend: less than 90% percentage pods available for 10m0s

**Next steps**

- Determine if the pod was OOM killed using `kubectl describe pod (frontend|sourcegraph-frontend)` (look for `OOMKilled: true`) and, if so, consider increasing the memory limit in the relevant `Deployment.yaml`.
- Check the logs before the container restarted to see if there are `panic:` messages or similar using `kubectl logs -p (frontend|sourcegraph-frontend)`.
- Learn more about the related dashboard panel in the [dashboards reference](./dashboards.md#frontend-pods-available-percentage).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "critical_frontend_pods_available_percentage"
]
```

<sub>*Managed by the [Sourcegraph Cloud DevOps team](https://handbook.sourcegraph.com/departments/engineering/teams/devops).*</sub>

<br />

## frontend: mean_successful_sentinel_duration_over_2h

<p class="subtitle">mean successful sentinel search duration over 2h</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> frontend: 5s+ mean successful sentinel search duration over 2h for 15m0s
- <span class="badge badge-critical">critical</span> frontend: 8s+ mean successful sentinel search duration over 2h for 30m0s

**Next steps**

- Look at the breakdown by query to determine if a specific query type is being affected
- Check for high CPU usage on zoekt-webserver
- Check Honeycomb for unusual activity
- More help interpreting this metric is available in the [dashboards reference](./dashboards.md#frontend-mean-successful-sentinel-duration-over-2h).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_frontend_mean_successful_sentinel_duration_over_2h",
  "critical_frontend_mean_successful_sentinel_duration_over_2h"
]
```

<sub>*Managed by the [Sourcegraph Search team](https://handbook.sourcegraph.com/departments/engineering/teams/search/product).*</sub>

<br />

## frontend: mean_sentinel_stream_latency_over_2h

<p class="subtitle">mean successful sentinel stream latency over 2h</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> frontend: 2s+ mean successful sentinel stream latency over 2h for 15m0s
- <span class="badge badge-critical">critical</span> frontend: 3s+ mean successful sentinel stream latency over 2h for 30m0s

**Next steps**

- Look at the breakdown by query to determine if a specific query type is being affected
- Check for high CPU usage on zoekt-webserver
- Check Honeycomb for unusual activity
- More help interpreting this metric is available in the [dashboards reference](./dashboards.md#frontend-mean-sentinel-stream-latency-over-2h).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_frontend_mean_sentinel_stream_latency_over_2h",
  "critical_frontend_mean_sentinel_stream_latency_over_2h"
]
```

<sub>*Managed by the [Sourcegraph Search team](https://handbook.sourcegraph.com/departments/engineering/teams/search/product).*</sub>

<br />

## frontend: 90th_percentile_successful_sentinel_duration_over_2h

<p class="subtitle">90th percentile successful sentinel search duration over 2h</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> frontend: 5s+ 90th percentile successful sentinel search duration over 2h for 15m0s
- <span class="badge badge-critical">critical</span> frontend: 10s+ 90th percentile successful sentinel search duration over 2h for 3h30m0s

**Next steps**

- Look at the breakdown by query to determine if a specific query type is being affected
- Check for high CPU usage on zoekt-webserver
- Check Honeycomb for unusual activity
- More help interpreting this metric is available in the [dashboards reference](./dashboards.md#frontend-90th-percentile-successful-sentinel-duration-over-2h).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_frontend_90th_percentile_successful_sentinel_duration_over_2h",
  "critical_frontend_90th_percentile_successful_sentinel_duration_over_2h"
]
```

<sub>*Managed by the [Sourcegraph Search team](https://handbook.sourcegraph.com/departments/engineering/teams/search/product).*</sub>

<br />

## frontend: 90th_percentile_sentinel_stream_latency_over_2h

<p class="subtitle">90th percentile successful sentinel stream latency over 2h</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> frontend: 4s+ 90th percentile successful sentinel stream latency over 2h for 15m0s
- <span class="badge badge-critical">critical</span> frontend: 6s+ 90th percentile successful sentinel stream latency over 2h for 3h30m0s

**Next steps**

- Look at the breakdown by query to determine if a specific query type is being affected
- Check for high CPU usage on zoekt-webserver
- Check Honeycomb for unusual activity
- More help interpreting this metric is available in the [dashboards reference](./dashboards.md#frontend-90th-percentile-sentinel-stream-latency-over-2h).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_frontend_90th_percentile_sentinel_stream_latency_over_2h",
  "critical_frontend_90th_percentile_sentinel_stream_latency_over_2h"
]
```

<sub>*Managed by the [Sourcegraph Search team](https://handbook.sourcegraph.com/departments/engineering/teams/search/product).*</sub>

<br />

## gitserver: disk_space_remaining

<p class="subtitle">disk space remaining by instance</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> gitserver: less than 25% disk space remaining by instance
- <span class="badge badge-critical">critical</span> gitserver: less than 15% disk space remaining by instance

**Next steps**

- **Provision more disk space:** Sourcegraph will begin deleting least-used repository clones at 10% disk space remaining which may result in decreased performance, users having to wait for repositories to clone, etc.
- Learn more about the related dashboard panel in the [dashboards reference](./dashboards.md#gitserver-disk-space-remaining).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_gitserver_disk_space_remaining",
  "critical_gitserver_disk_space_remaining"
]
```

<sub>*Managed by the [Sourcegraph Repo Management team](https://handbook.sourcegraph.com/departments/engineering/teams/repo-management).*</sub>

<br />

## gitserver: running_git_commands

<p class="subtitle">git commands running on each gitserver instance</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> gitserver: 50+ git commands running on each gitserver instance for 2m0s
- <span class="badge badge-critical">critical</span> gitserver: 100+ git commands running on each gitserver instance for 5m0s

**Next steps**

- **Check if the problem may be an intermittent and temporary peak** using the "Container monitoring" section at the bottom of the Git Server dashboard.
- **Single container deployments:** Consider upgrading to a [Docker Compose deployment](../deploy/docker-compose/migrate.md) which offers better scalability and resource isolation.
- **Kubernetes and Docker Compose:** Check that you are running a similar number of git server replicas and that their CPU/memory limits are allocated according to what is shown in the [Sourcegraph resource estimator](../deploy/resource_estimator.md).
- More help interpreting this metric is available in the [dashboards reference](./dashboards.md#gitserver-running-git-commands).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_gitserver_running_git_commands",
  "critical_gitserver_running_git_commands"
]
```

<sub>*Managed by the [Sourcegraph Repo Management team](https://handbook.sourcegraph.com/departments/engineering/teams/repo-management).*</sub>

<br />

## gitserver: repository_clone_queue_size

<p class="subtitle">repository clone queue size</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> gitserver: 25+ repository clone queue size

**Next steps**

- **If you just added several repositories**, the warning may be expected.
- **Check which repositories need cloning**, by visiting e.g. https://sourcegraph.example.com/site-admin/repositories?filter=not-cloned
- Learn more about the related dashboard panel in the [dashboards reference](./dashboards.md#gitserver-repository-clone-queue-size).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_gitserver_repository_clone_queue_size"
]
```

<sub>*Managed by the [Sourcegraph Repo Management team](https://handbook.sourcegraph.com/departments/engineering/teams/repo-management).*</sub>

<br />

## gitserver: repository_existence_check_queue_size

<p class="subtitle">repository existence check queue size</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> gitserver: 25+ repository existence check queue size

**Next steps**

- **Check the code host status indicator for errors:** on the Sourcegraph app homepage, when signed in as an admin click the cloud icon in the top right corner of the page.
- **Check if the issue continues to happen after 30 minutes**, it may be temporary.
- **Check the gitserver logs for more information.**
- Learn more about the related dashboard panel in the [dashboards reference](./dashboards.md#gitserver-repository-existence-check-queue-size).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_gitserver_repository_existence_check_queue_size"
]
```

<sub>*Managed by the [Sourcegraph Repo Management team](https://handbook.sourcegraph.com/departments/engineering/teams/repo-management).*</sub>

<br />

## gitserver: frontend_internal_api_error_responses

<p class="subtitle">frontend-internal API error responses every 5m by route</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> gitserver: 2%+ frontend-internal API error responses every 5m by route for 5m0s

**Next steps**

- **Single-container deployments:** Check `docker logs $CONTAINER_ID` for logs starting with `repo-updater` that indicate requests to the frontend service are failing.
- **Kubernetes:**
	- Confirm that `kubectl get pods` shows the `frontend` pods are healthy.
	- Check `kubectl logs gitserver` for logs indicate request failures to `frontend` or `frontend-internal`.
- **Docker Compose:**
	- Confirm that `docker ps` shows the `frontend-internal` container is healthy.
	- Check `docker logs gitserver` for logs indicating request failures to `frontend` or `frontend-internal`.
- Learn more about the related dashboard panel in the [dashboards reference](./dashboards.md#gitserver-frontend-internal-api-error-responses).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_gitserver_frontend_internal_api_error_responses"
]
```

<sub>*Managed by the [Sourcegraph Repo Management team](https://handbook.sourcegraph.com/departments/engineering/teams/repo-management).*</sub>

<br />

## gitserver: mean_blocked_seconds_per_conn_request

<p class="subtitle">mean blocked seconds per conn request</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> gitserver: 0.05s+ mean blocked seconds per conn request for 10m0s
- <span class="badge badge-critical">critical</span> gitserver: 0.1s+ mean blocked seconds per conn request for 15m0s

**Next steps**

- Increase SRC_PGSQL_MAX_OPEN together with giving more memory to the database if needed
- Scale up Postgres memory / cpus [See our scaling guide](https://docs.sourcegraph.com/admin/config/postgres-conf)
- Learn more about the related dashboard panel in the [dashboards reference](./dashboards.md#gitserver-mean-blocked-seconds-per-conn-request).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_gitserver_mean_blocked_seconds_per_conn_request",
  "critical_gitserver_mean_blocked_seconds_per_conn_request"
]
```

<sub>*Managed by the [Sourcegraph Cloud DevOps team](https://handbook.sourcegraph.com/departments/engineering/teams/devops).*</sub>

<br />

## gitserver: container_cpu_usage

<p class="subtitle">container cpu usage total (1m average) across all cores by instance</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> gitserver: 99%+ container cpu usage total (1m average) across all cores by instance

**Next steps**

- **Kubernetes:** Consider increasing CPU limits in the the relevant `Deployment.yaml`.
- **Docker Compose:** Consider increasing `cpus:` of the gitserver container in `docker-compose.yml`.
- Learn more about the related dashboard panel in the [dashboards reference](./dashboards.md#gitserver-container-cpu-usage).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_gitserver_container_cpu_usage"
]
```

<sub>*Managed by the [Sourcegraph Repo Management team](https://handbook.sourcegraph.com/departments/engineering/teams/repo-management).*</sub>

<br />

## gitserver: container_memory_usage

<p class="subtitle">container memory usage by instance</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> gitserver: 99%+ container memory usage by instance

**Next steps**

- **Kubernetes:** Consider increasing memory limit in relevant `Deployment.yaml`.
- **Docker Compose:** Consider increasing `memory:` of gitserver container in `docker-compose.yml`.
- Learn more about the related dashboard panel in the [dashboards reference](./dashboards.md#gitserver-container-memory-usage).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_gitserver_container_memory_usage"
]
```

<sub>*Managed by the [Sourcegraph Repo Management team](https://handbook.sourcegraph.com/departments/engineering/teams/repo-management).*</sub>

<br />

## gitserver: provisioning_container_cpu_usage_long_term

<p class="subtitle">container cpu usage total (90th percentile over 1d) across all cores by instance</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> gitserver: 80%+ container cpu usage total (90th percentile over 1d) across all cores by instance for 336h0m0s

**Next steps**

- **Kubernetes:** Consider increasing CPU limits in the `Deployment.yaml` for the gitserver service.
- **Docker Compose:** Consider increasing `cpus:` of the gitserver container in `docker-compose.yml`.
- Learn more about the related dashboard panel in the [dashboards reference](./dashboards.md#gitserver-provisioning-container-cpu-usage-long-term).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_gitserver_provisioning_container_cpu_usage_long_term"
]
```

<sub>*Managed by the [Sourcegraph Repo Management team](https://handbook.sourcegraph.com/departments/engineering/teams/repo-management).*</sub>

<br />

## gitserver: provisioning_container_cpu_usage_short_term

<p class="subtitle">container cpu usage total (5m maximum) across all cores by instance</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> gitserver: 90%+ container cpu usage total (5m maximum) across all cores by instance for 30m0s

**Next steps**

- **Kubernetes:** Consider increasing CPU limits in the the relevant `Deployment.yaml`.
- **Docker Compose:** Consider increasing `cpus:` of the gitserver container in `docker-compose.yml`.
- Learn more about the related dashboard panel in the [dashboards reference](./dashboards.md#gitserver-provisioning-container-cpu-usage-short-term).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_gitserver_provisioning_container_cpu_usage_short_term"
]
```

<sub>*Managed by the [Sourcegraph Repo Management team](https://handbook.sourcegraph.com/departments/engineering/teams/repo-management).*</sub>

<br />

## gitserver: container_oomkill_events_total

<p class="subtitle">container OOMKILL events total by instance</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> gitserver: 1+ container OOMKILL events total by instance

**Next steps**

- **Kubernetes:** Consider increasing memory limit in relevant `Deployment.yaml`.
- **Docker Compose:** Consider increasing `memory:` of gitserver container in `docker-compose.yml`.
- More help interpreting this metric is available in the [dashboards reference](./dashboards.md#gitserver-container-oomkill-events-total).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_gitserver_container_oomkill_events_total"
]
```

<sub>*Managed by the [Sourcegraph Repo Management team](https://handbook.sourcegraph.com/departments/engineering/teams/repo-management).*</sub>

<br />

## gitserver: go_goroutines

<p class="subtitle">maximum active goroutines</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> gitserver: 10000+ maximum active goroutines for 10m0s

**Next steps**

- More help interpreting this metric is available in the [dashboards reference](./dashboards.md#gitserver-go-goroutines).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_gitserver_go_goroutines"
]
```

<sub>*Managed by the [Sourcegraph Repo Management team](https://handbook.sourcegraph.com/departments/engineering/teams/repo-management).*</sub>

<br />

## gitserver: go_gc_duration_seconds

<p class="subtitle">maximum go garbage collection duration</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> gitserver: 2s+ maximum go garbage collection duration

**Next steps**

- Learn more about the related dashboard panel in the [dashboards reference](./dashboards.md#gitserver-go-gc-duration-seconds).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_gitserver_go_gc_duration_seconds"
]
```

<sub>*Managed by the [Sourcegraph Repo Management team](https://handbook.sourcegraph.com/departments/engineering/teams/repo-management).*</sub>

<br />

## gitserver: pods_available_percentage

<p class="subtitle">percentage pods available</p>

**Descriptions**

- <span class="badge badge-critical">critical</span> gitserver: less than 90% percentage pods available for 10m0s

**Next steps**

- Determine if the pod was OOM killed using `kubectl describe pod gitserver` (look for `OOMKilled: true`) and, if so, consider increasing the memory limit in the relevant `Deployment.yaml`.
- Check the logs before the container restarted to see if there are `panic:` messages or similar using `kubectl logs -p gitserver`.
- Learn more about the related dashboard panel in the [dashboards reference](./dashboards.md#gitserver-pods-available-percentage).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "critical_gitserver_pods_available_percentage"
]
```

<sub>*Managed by the [Sourcegraph Repo Management team](https://handbook.sourcegraph.com/departments/engineering/teams/repo-management).*</sub>

<br />

## github-proxy: github_proxy_waiting_requests

<p class="subtitle">number of requests waiting on the global mutex</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> github-proxy: 100+ number of requests waiting on the global mutex for 5m0s

**Next steps**

- 								- **Check github-proxy logs for network connection issues.
								- **Check github status.
- Learn more about the related dashboard panel in the [dashboards reference](./dashboards.md#github-proxy-github-proxy-waiting-requests).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_github-proxy_github_proxy_waiting_requests"
]
```

<sub>*Managed by the [Sourcegraph Repo Management team](https://handbook.sourcegraph.com/departments/engineering/teams/repo-management).*</sub>

<br />

## github-proxy: container_cpu_usage

<p class="subtitle">container cpu usage total (1m average) across all cores by instance</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> github-proxy: 99%+ container cpu usage total (1m average) across all cores by instance

**Next steps**

- **Kubernetes:** Consider increasing CPU limits in the the relevant `Deployment.yaml`.
- **Docker Compose:** Consider increasing `cpus:` of the github-proxy container in `docker-compose.yml`.
- Learn more about the related dashboard panel in the [dashboards reference](./dashboards.md#github-proxy-container-cpu-usage).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_github-proxy_container_cpu_usage"
]
```

<sub>*Managed by the [Sourcegraph Cloud DevOps team](https://handbook.sourcegraph.com/departments/engineering/teams/devops).*</sub>

<br />

## github-proxy: container_memory_usage

<p class="subtitle">container memory usage by instance</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> github-proxy: 99%+ container memory usage by instance

**Next steps**

- **Kubernetes:** Consider increasing memory limit in relevant `Deployment.yaml`.
- **Docker Compose:** Consider increasing `memory:` of github-proxy container in `docker-compose.yml`.
- Learn more about the related dashboard panel in the [dashboards reference](./dashboards.md#github-proxy-container-memory-usage).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_github-proxy_container_memory_usage"
]
```

<sub>*Managed by the [Sourcegraph Cloud DevOps team](https://handbook.sourcegraph.com/departments/engineering/teams/devops).*</sub>

<br />

## github-proxy: provisioning_container_cpu_usage_long_term

<p class="subtitle">container cpu usage total (90th percentile over 1d) across all cores by instance</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> github-proxy: 80%+ container cpu usage total (90th percentile over 1d) across all cores by instance for 336h0m0s

**Next steps**

- **Kubernetes:** Consider increasing CPU limits in the `Deployment.yaml` for the github-proxy service.
- **Docker Compose:** Consider increasing `cpus:` of the github-proxy container in `docker-compose.yml`.
- Learn more about the related dashboard panel in the [dashboards reference](./dashboards.md#github-proxy-provisioning-container-cpu-usage-long-term).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_github-proxy_provisioning_container_cpu_usage_long_term"
]
```

<sub>*Managed by the [Sourcegraph Cloud DevOps team](https://handbook.sourcegraph.com/departments/engineering/teams/devops).*</sub>

<br />

## github-proxy: provisioning_container_memory_usage_long_term

<p class="subtitle">container memory usage (1d maximum) by instance</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> github-proxy: 80%+ container memory usage (1d maximum) by instance for 336h0m0s

**Next steps**

- **Kubernetes:** Consider increasing memory limits in the `Deployment.yaml` for the github-proxy service.
- **Docker Compose:** Consider increasing `memory:` of the github-proxy container in `docker-compose.yml`.
- Learn more about the related dashboard panel in the [dashboards reference](./dashboards.md#github-proxy-provisioning-container-memory-usage-long-term).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_github-proxy_provisioning_container_memory_usage_long_term"
]
```

<sub>*Managed by the [Sourcegraph Cloud DevOps team](https://handbook.sourcegraph.com/departments/engineering/teams/devops).*</sub>

<br />

## github-proxy: provisioning_container_cpu_usage_short_term

<p class="subtitle">container cpu usage total (5m maximum) across all cores by instance</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> github-proxy: 90%+ container cpu usage total (5m maximum) across all cores by instance for 30m0s

**Next steps**

- **Kubernetes:** Consider increasing CPU limits in the the relevant `Deployment.yaml`.
- **Docker Compose:** Consider increasing `cpus:` of the github-proxy container in `docker-compose.yml`.
- Learn more about the related dashboard panel in the [dashboards reference](./dashboards.md#github-proxy-provisioning-container-cpu-usage-short-term).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_github-proxy_provisioning_container_cpu_usage_short_term"
]
```

<sub>*Managed by the [Sourcegraph Cloud DevOps team](https://handbook.sourcegraph.com/departments/engineering/teams/devops).*</sub>

<br />

## github-proxy: provisioning_container_memory_usage_short_term

<p class="subtitle">container memory usage (5m maximum) by instance</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> github-proxy: 90%+ container memory usage (5m maximum) by instance

**Next steps**

- **Kubernetes:** Consider increasing memory limit in relevant `Deployment.yaml`.
- **Docker Compose:** Consider increasing `memory:` of github-proxy container in `docker-compose.yml`.
- Learn more about the related dashboard panel in the [dashboards reference](./dashboards.md#github-proxy-provisioning-container-memory-usage-short-term).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_github-proxy_provisioning_container_memory_usage_short_term"
]
```

<sub>*Managed by the [Sourcegraph Cloud DevOps team](https://handbook.sourcegraph.com/departments/engineering/teams/devops).*</sub>

<br />

## github-proxy: container_oomkill_events_total

<p class="subtitle">container OOMKILL events total by instance</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> github-proxy: 1+ container OOMKILL events total by instance

**Next steps**

- **Kubernetes:** Consider increasing memory limit in relevant `Deployment.yaml`.
- **Docker Compose:** Consider increasing `memory:` of github-proxy container in `docker-compose.yml`.
- More help interpreting this metric is available in the [dashboards reference](./dashboards.md#github-proxy-container-oomkill-events-total).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_github-proxy_container_oomkill_events_total"
]
```

<sub>*Managed by the [Sourcegraph Cloud DevOps team](https://handbook.sourcegraph.com/departments/engineering/teams/devops).*</sub>

<br />

## github-proxy: go_goroutines

<p class="subtitle">maximum active goroutines</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> github-proxy: 10000+ maximum active goroutines for 10m0s

**Next steps**

- More help interpreting this metric is available in the [dashboards reference](./dashboards.md#github-proxy-go-goroutines).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_github-proxy_go_goroutines"
]
```

<sub>*Managed by the [Sourcegraph Cloud DevOps team](https://handbook.sourcegraph.com/departments/engineering/teams/devops).*</sub>

<br />

## github-proxy: go_gc_duration_seconds

<p class="subtitle">maximum go garbage collection duration</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> github-proxy: 2s+ maximum go garbage collection duration

**Next steps**

- Learn more about the related dashboard panel in the [dashboards reference](./dashboards.md#github-proxy-go-gc-duration-seconds).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_github-proxy_go_gc_duration_seconds"
]
```

<sub>*Managed by the [Sourcegraph Cloud DevOps team](https://handbook.sourcegraph.com/departments/engineering/teams/devops).*</sub>

<br />

## github-proxy: pods_available_percentage

<p class="subtitle">percentage pods available</p>

**Descriptions**

- <span class="badge badge-critical">critical</span> github-proxy: less than 90% percentage pods available for 10m0s

**Next steps**

- Determine if the pod was OOM killed using `kubectl describe pod github-proxy` (look for `OOMKilled: true`) and, if so, consider increasing the memory limit in the relevant `Deployment.yaml`.
- Check the logs before the container restarted to see if there are `panic:` messages or similar using `kubectl logs -p github-proxy`.
- Learn more about the related dashboard panel in the [dashboards reference](./dashboards.md#github-proxy-pods-available-percentage).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "critical_github-proxy_pods_available_percentage"
]
```

<sub>*Managed by the [Sourcegraph Cloud DevOps team](https://handbook.sourcegraph.com/departments/engineering/teams/devops).*</sub>

<br />

## postgres: connections

<p class="subtitle">active connections</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> postgres: less than 5 active connections for 5m0s

**Next steps**

- Learn more about the related dashboard panel in the [dashboards reference](./dashboards.md#postgres-connections).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_postgres_connections"
]
```

<sub>*Managed by the [Sourcegraph Cloud DevOps team](https://handbook.sourcegraph.com/departments/engineering/teams/devops).*</sub>

<br />

## postgres: usage_connections_percentage

<p class="subtitle">connection in use</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> postgres: 80%+ connection in use for 5m0s
- <span class="badge badge-critical">critical</span> postgres: 100%+ connection in use for 5m0s

**Next steps**

- Consider increasing [max_connections](https://www.postgresql.org/docs/current/runtime-config-connection.html#GUC-MAX-CONNECTIONS) of the database instance, [learn more](https://docs.sourcegraph.com/admin/config/postgres-conf)
- Learn more about the related dashboard panel in the [dashboards reference](./dashboards.md#postgres-usage-connections-percentage).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_postgres_usage_connections_percentage",
  "critical_postgres_usage_connections_percentage"
]
```

<sub>*Managed by the [Sourcegraph Cloud DevOps team](https://handbook.sourcegraph.com/departments/engineering/teams/devops).*</sub>

<br />

## postgres: transaction_durations

<p class="subtitle">maximum transaction durations</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> postgres: 0.3s+ maximum transaction durations for 5m0s

**Next steps**

- Learn more about the related dashboard panel in the [dashboards reference](./dashboards.md#postgres-transaction-durations).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_postgres_transaction_durations"
]
```

<sub>*Managed by the [Sourcegraph Cloud DevOps team](https://handbook.sourcegraph.com/departments/engineering/teams/devops).*</sub>

<br />

## postgres: postgres_up

<p class="subtitle">database availability</p>

**Descriptions**

- <span class="badge badge-critical">critical</span> postgres: less than 0 database availability for 5m0s

**Next steps**

- **Kubernetes:**
	- Determine if the pod was OOM killed using `kubectl describe pod (pgsql|codeintel-db|codeinsights)` (look for `OOMKilled: true`) and, if so, consider increasing the memory limit in the relevant `Deployment.yaml`.
	- Check the logs before the container restarted to see if there are `panic:` messages or similar using `kubectl logs -p (pgsql|codeintel-db|codeinsights)`.
	- Check if there is any OOMKILL event using the provisioning panels
	- Check kernel logs using `dmesg` for OOMKILL events on worker nodes
- **Docker Compose:**
	- Determine if the pod was OOM killed using `docker inspect -f '{{json .State}}' (pgsql|codeintel-db|codeinsights)` (look for `"OOMKilled":true`) and, if so, consider increasing the memory limit of the (pgsql|codeintel-db|codeinsights) container in `docker-compose.yml`.
	- Check the logs before the container restarted to see if there are `panic:` messages or similar using `docker logs (pgsql|codeintel-db|codeinsights)` (note this will include logs from the previous and currently running container).
	- Check if there is any OOMKILL event using the provisioning panels
	- Check kernel logs using `dmesg` for OOMKILL events
- More help interpreting this metric is available in the [dashboards reference](./dashboards.md#postgres-postgres-up).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "critical_postgres_postgres_up"
]
```

<sub>*Managed by the [Sourcegraph Cloud DevOps team](https://handbook.sourcegraph.com/departments/engineering/teams/devops).*</sub>

<br />

## postgres: invalid_indexes

<p class="subtitle">invalid indexes (unusable by the query planner)</p>

**Descriptions**

- <span class="badge badge-critical">critical</span> postgres: 1+ invalid indexes (unusable by the query planner)

**Next steps**

- Drop and re-create the invalid trigger - please contact Sourcegraph to supply the trigger definition.
- More help interpreting this metric is available in the [dashboards reference](./dashboards.md#postgres-invalid-indexes).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "critical_postgres_invalid_indexes"
]
```

<sub>*Managed by the [Sourcegraph Cloud DevOps team](https://handbook.sourcegraph.com/departments/engineering/teams/devops).*</sub>

<br />

## postgres: pg_exporter_err

<p class="subtitle">errors scraping postgres exporter</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> postgres: 1+ errors scraping postgres exporter for 5m0s

**Next steps**

- Ensure the Postgres exporter can access the Postgres database. Also, check the Postgres exporter logs for errors.
- More help interpreting this metric is available in the [dashboards reference](./dashboards.md#postgres-pg-exporter-err).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_postgres_pg_exporter_err"
]
```

<sub>*Managed by the [Sourcegraph Cloud DevOps team](https://handbook.sourcegraph.com/departments/engineering/teams/devops).*</sub>

<br />

## postgres: migration_in_progress

<p class="subtitle">active schema migration</p>

**Descriptions**

- <span class="badge badge-critical">critical</span> postgres: 1+ active schema migration for 5m0s

**Next steps**

- The database migration has been in progress for 5 or more minutes - please contact Sourcegraph if this persists.
- More help interpreting this metric is available in the [dashboards reference](./dashboards.md#postgres-migration-in-progress).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "critical_postgres_migration_in_progress"
]
```

<sub>*Managed by the [Sourcegraph Cloud DevOps team](https://handbook.sourcegraph.com/departments/engineering/teams/devops).*</sub>

<br />

## postgres: provisioning_container_cpu_usage_long_term

<p class="subtitle">container cpu usage total (90th percentile over 1d) across all cores by instance</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> postgres: 80%+ container cpu usage total (90th percentile over 1d) across all cores by instance for 336h0m0s

**Next steps**

- **Kubernetes:** Consider increasing CPU limits in the `Deployment.yaml` for the (pgsql|codeintel-db|codeinsights) service.
- **Docker Compose:** Consider increasing `cpus:` of the (pgsql|codeintel-db|codeinsights) container in `docker-compose.yml`.
- Learn more about the related dashboard panel in the [dashboards reference](./dashboards.md#postgres-provisioning-container-cpu-usage-long-term).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_postgres_provisioning_container_cpu_usage_long_term"
]
```

<sub>*Managed by the [Sourcegraph Cloud DevOps team](https://handbook.sourcegraph.com/departments/engineering/teams/devops).*</sub>

<br />

## postgres: provisioning_container_memory_usage_long_term

<p class="subtitle">container memory usage (1d maximum) by instance</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> postgres: 80%+ container memory usage (1d maximum) by instance for 336h0m0s

**Next steps**

- **Kubernetes:** Consider increasing memory limits in the `Deployment.yaml` for the (pgsql|codeintel-db|codeinsights) service.
- **Docker Compose:** Consider increasing `memory:` of the (pgsql|codeintel-db|codeinsights) container in `docker-compose.yml`.
- Learn more about the related dashboard panel in the [dashboards reference](./dashboards.md#postgres-provisioning-container-memory-usage-long-term).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_postgres_provisioning_container_memory_usage_long_term"
]
```

<sub>*Managed by the [Sourcegraph Cloud DevOps team](https://handbook.sourcegraph.com/departments/engineering/teams/devops).*</sub>

<br />

## postgres: provisioning_container_cpu_usage_short_term

<p class="subtitle">container cpu usage total (5m maximum) across all cores by instance</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> postgres: 90%+ container cpu usage total (5m maximum) across all cores by instance for 30m0s

**Next steps**

- **Kubernetes:** Consider increasing CPU limits in the the relevant `Deployment.yaml`.
- **Docker Compose:** Consider increasing `cpus:` of the (pgsql|codeintel-db|codeinsights) container in `docker-compose.yml`.
- Learn more about the related dashboard panel in the [dashboards reference](./dashboards.md#postgres-provisioning-container-cpu-usage-short-term).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_postgres_provisioning_container_cpu_usage_short_term"
]
```

<sub>*Managed by the [Sourcegraph Cloud DevOps team](https://handbook.sourcegraph.com/departments/engineering/teams/devops).*</sub>

<br />

## postgres: provisioning_container_memory_usage_short_term

<p class="subtitle">container memory usage (5m maximum) by instance</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> postgres: 90%+ container memory usage (5m maximum) by instance

**Next steps**

- **Kubernetes:** Consider increasing memory limit in relevant `Deployment.yaml`.
- **Docker Compose:** Consider increasing `memory:` of (pgsql|codeintel-db|codeinsights) container in `docker-compose.yml`.
- Learn more about the related dashboard panel in the [dashboards reference](./dashboards.md#postgres-provisioning-container-memory-usage-short-term).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_postgres_provisioning_container_memory_usage_short_term"
]
```

<sub>*Managed by the [Sourcegraph Cloud DevOps team](https://handbook.sourcegraph.com/departments/engineering/teams/devops).*</sub>

<br />

## postgres: container_oomkill_events_total

<p class="subtitle">container OOMKILL events total by instance</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> postgres: 1+ container OOMKILL events total by instance

**Next steps**

- **Kubernetes:** Consider increasing memory limit in relevant `Deployment.yaml`.
- **Docker Compose:** Consider increasing `memory:` of (pgsql|codeintel-db|codeinsights) container in `docker-compose.yml`.
- More help interpreting this metric is available in the [dashboards reference](./dashboards.md#postgres-container-oomkill-events-total).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_postgres_container_oomkill_events_total"
]
```

<sub>*Managed by the [Sourcegraph Cloud DevOps team](https://handbook.sourcegraph.com/departments/engineering/teams/devops).*</sub>

<br />

## postgres: pods_available_percentage

<p class="subtitle">percentage pods available</p>

**Descriptions**

- <span class="badge badge-critical">critical</span> postgres: less than 90% percentage pods available for 10m0s

**Next steps**

- Determine if the pod was OOM killed using `kubectl describe pod (pgsql|codeintel-db|codeinsights)` (look for `OOMKilled: true`) and, if so, consider increasing the memory limit in the relevant `Deployment.yaml`.
- Check the logs before the container restarted to see if there are `panic:` messages or similar using `kubectl logs -p (pgsql|codeintel-db|codeinsights)`.
- Learn more about the related dashboard panel in the [dashboards reference](./dashboards.md#postgres-pods-available-percentage).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "critical_postgres_pods_available_percentage"
]
```

<sub>*Managed by the [Sourcegraph Cloud DevOps team](https://handbook.sourcegraph.com/departments/engineering/teams/devops).*</sub>

<br />

## precise-code-intel-worker: codeintel_upload_queued_max_age

<p class="subtitle">unprocessed upload record queue longest time in queue</p>

**Descriptions**

- <span class="badge badge-critical">critical</span> precise-code-intel-worker: 18000s+ unprocessed upload record queue longest time in queue

**Next steps**

- An alert here could be indicative of a few things: an upload surfacing a pathological performance characteristic,
precise-code-intel-worker being underprovisioned for the required upload processing throughput, or a higher replica
count being required for the volume of uploads.
- Learn more about the related dashboard panel in the [dashboards reference](./dashboards.md#precise-code-intel-worker-codeintel-upload-queued-max-age).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "critical_precise-code-intel-worker_codeintel_upload_queued_max_age"
]
```

<sub>*Managed by the [Sourcegraph Code intelligence team](https://handbook.sourcegraph.com/departments/engineering/teams/code-intelligence).*</sub>

<br />

## precise-code-intel-worker: frontend_internal_api_error_responses

<p class="subtitle">frontend-internal API error responses every 5m by route</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> precise-code-intel-worker: 2%+ frontend-internal API error responses every 5m by route for 5m0s

**Next steps**

- **Single-container deployments:** Check `docker logs $CONTAINER_ID` for logs starting with `repo-updater` that indicate requests to the frontend service are failing.
- **Kubernetes:**
	- Confirm that `kubectl get pods` shows the `frontend` pods are healthy.
	- Check `kubectl logs precise-code-intel-worker` for logs indicate request failures to `frontend` or `frontend-internal`.
- **Docker Compose:**
	- Confirm that `docker ps` shows the `frontend-internal` container is healthy.
	- Check `docker logs precise-code-intel-worker` for logs indicating request failures to `frontend` or `frontend-internal`.
- Learn more about the related dashboard panel in the [dashboards reference](./dashboards.md#precise-code-intel-worker-frontend-internal-api-error-responses).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_precise-code-intel-worker_frontend_internal_api_error_responses"
]
```

<sub>*Managed by the [Sourcegraph Code intelligence team](https://handbook.sourcegraph.com/departments/engineering/teams/code-intelligence).*</sub>

<br />

## precise-code-intel-worker: mean_blocked_seconds_per_conn_request

<p class="subtitle">mean blocked seconds per conn request</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> precise-code-intel-worker: 0.05s+ mean blocked seconds per conn request for 10m0s
- <span class="badge badge-critical">critical</span> precise-code-intel-worker: 0.1s+ mean blocked seconds per conn request for 15m0s

**Next steps**

- Increase SRC_PGSQL_MAX_OPEN together with giving more memory to the database if needed
- Scale up Postgres memory / cpus [See our scaling guide](https://docs.sourcegraph.com/admin/config/postgres-conf)
- Learn more about the related dashboard panel in the [dashboards reference](./dashboards.md#precise-code-intel-worker-mean-blocked-seconds-per-conn-request).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_precise-code-intel-worker_mean_blocked_seconds_per_conn_request",
  "critical_precise-code-intel-worker_mean_blocked_seconds_per_conn_request"
]
```

<sub>*Managed by the [Sourcegraph Cloud DevOps team](https://handbook.sourcegraph.com/departments/engineering/teams/devops).*</sub>

<br />

## precise-code-intel-worker: container_cpu_usage

<p class="subtitle">container cpu usage total (1m average) across all cores by instance</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> precise-code-intel-worker: 99%+ container cpu usage total (1m average) across all cores by instance

**Next steps**

- **Kubernetes:** Consider increasing CPU limits in the the relevant `Deployment.yaml`.
- **Docker Compose:** Consider increasing `cpus:` of the precise-code-intel-worker container in `docker-compose.yml`.
- Learn more about the related dashboard panel in the [dashboards reference](./dashboards.md#precise-code-intel-worker-container-cpu-usage).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_precise-code-intel-worker_container_cpu_usage"
]
```

<sub>*Managed by the [Sourcegraph Code intelligence team](https://handbook.sourcegraph.com/departments/engineering/teams/code-intelligence).*</sub>

<br />

## precise-code-intel-worker: container_memory_usage

<p class="subtitle">container memory usage by instance</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> precise-code-intel-worker: 99%+ container memory usage by instance

**Next steps**

- **Kubernetes:** Consider increasing memory limit in relevant `Deployment.yaml`.
- **Docker Compose:** Consider increasing `memory:` of precise-code-intel-worker container in `docker-compose.yml`.
- Learn more about the related dashboard panel in the [dashboards reference](./dashboards.md#precise-code-intel-worker-container-memory-usage).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_precise-code-intel-worker_container_memory_usage"
]
```

<sub>*Managed by the [Sourcegraph Code intelligence team](https://handbook.sourcegraph.com/departments/engineering/teams/code-intelligence).*</sub>

<br />

## precise-code-intel-worker: provisioning_container_cpu_usage_long_term

<p class="subtitle">container cpu usage total (90th percentile over 1d) across all cores by instance</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> precise-code-intel-worker: 80%+ container cpu usage total (90th percentile over 1d) across all cores by instance for 336h0m0s

**Next steps**

- **Kubernetes:** Consider increasing CPU limits in the `Deployment.yaml` for the precise-code-intel-worker service.
- **Docker Compose:** Consider increasing `cpus:` of the precise-code-intel-worker container in `docker-compose.yml`.
- Learn more about the related dashboard panel in the [dashboards reference](./dashboards.md#precise-code-intel-worker-provisioning-container-cpu-usage-long-term).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_precise-code-intel-worker_provisioning_container_cpu_usage_long_term"
]
```

<sub>*Managed by the [Sourcegraph Code intelligence team](https://handbook.sourcegraph.com/departments/engineering/teams/code-intelligence).*</sub>

<br />

## precise-code-intel-worker: provisioning_container_memory_usage_long_term

<p class="subtitle">container memory usage (1d maximum) by instance</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> precise-code-intel-worker: 80%+ container memory usage (1d maximum) by instance for 336h0m0s

**Next steps**

- **Kubernetes:** Consider increasing memory limits in the `Deployment.yaml` for the precise-code-intel-worker service.
- **Docker Compose:** Consider increasing `memory:` of the precise-code-intel-worker container in `docker-compose.yml`.
- Learn more about the related dashboard panel in the [dashboards reference](./dashboards.md#precise-code-intel-worker-provisioning-container-memory-usage-long-term).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_precise-code-intel-worker_provisioning_container_memory_usage_long_term"
]
```

<sub>*Managed by the [Sourcegraph Code intelligence team](https://handbook.sourcegraph.com/departments/engineering/teams/code-intelligence).*</sub>

<br />

## precise-code-intel-worker: provisioning_container_cpu_usage_short_term

<p class="subtitle">container cpu usage total (5m maximum) across all cores by instance</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> precise-code-intel-worker: 90%+ container cpu usage total (5m maximum) across all cores by instance for 30m0s

**Next steps**

- **Kubernetes:** Consider increasing CPU limits in the the relevant `Deployment.yaml`.
- **Docker Compose:** Consider increasing `cpus:` of the precise-code-intel-worker container in `docker-compose.yml`.
- Learn more about the related dashboard panel in the [dashboards reference](./dashboards.md#precise-code-intel-worker-provisioning-container-cpu-usage-short-term).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_precise-code-intel-worker_provisioning_container_cpu_usage_short_term"
]
```

<sub>*Managed by the [Sourcegraph Code intelligence team](https://handbook.sourcegraph.com/departments/engineering/teams/code-intelligence).*</sub>

<br />

## precise-code-intel-worker: provisioning_container_memory_usage_short_term

<p class="subtitle">container memory usage (5m maximum) by instance</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> precise-code-intel-worker: 90%+ container memory usage (5m maximum) by instance

**Next steps**

- **Kubernetes:** Consider increasing memory limit in relevant `Deployment.yaml`.
- **Docker Compose:** Consider increasing `memory:` of precise-code-intel-worker container in `docker-compose.yml`.
- Learn more about the related dashboard panel in the [dashboards reference](./dashboards.md#precise-code-intel-worker-provisioning-container-memory-usage-short-term).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_precise-code-intel-worker_provisioning_container_memory_usage_short_term"
]
```

<sub>*Managed by the [Sourcegraph Code intelligence team](https://handbook.sourcegraph.com/departments/engineering/teams/code-intelligence).*</sub>

<br />

## precise-code-intel-worker: container_oomkill_events_total

<p class="subtitle">container OOMKILL events total by instance</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> precise-code-intel-worker: 1+ container OOMKILL events total by instance

**Next steps**

- **Kubernetes:** Consider increasing memory limit in relevant `Deployment.yaml`.
- **Docker Compose:** Consider increasing `memory:` of precise-code-intel-worker container in `docker-compose.yml`.
- More help interpreting this metric is available in the [dashboards reference](./dashboards.md#precise-code-intel-worker-container-oomkill-events-total).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_precise-code-intel-worker_container_oomkill_events_total"
]
```

<sub>*Managed by the [Sourcegraph Code intelligence team](https://handbook.sourcegraph.com/departments/engineering/teams/code-intelligence).*</sub>

<br />

## precise-code-intel-worker: go_goroutines

<p class="subtitle">maximum active goroutines</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> precise-code-intel-worker: 10000+ maximum active goroutines for 10m0s

**Next steps**

- More help interpreting this metric is available in the [dashboards reference](./dashboards.md#precise-code-intel-worker-go-goroutines).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_precise-code-intel-worker_go_goroutines"
]
```

<sub>*Managed by the [Sourcegraph Code intelligence team](https://handbook.sourcegraph.com/departments/engineering/teams/code-intelligence).*</sub>

<br />

## precise-code-intel-worker: go_gc_duration_seconds

<p class="subtitle">maximum go garbage collection duration</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> precise-code-intel-worker: 2s+ maximum go garbage collection duration

**Next steps**

- Learn more about the related dashboard panel in the [dashboards reference](./dashboards.md#precise-code-intel-worker-go-gc-duration-seconds).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_precise-code-intel-worker_go_gc_duration_seconds"
]
```

<sub>*Managed by the [Sourcegraph Code intelligence team](https://handbook.sourcegraph.com/departments/engineering/teams/code-intelligence).*</sub>

<br />

## precise-code-intel-worker: pods_available_percentage

<p class="subtitle">percentage pods available</p>

**Descriptions**

- <span class="badge badge-critical">critical</span> precise-code-intel-worker: less than 90% percentage pods available for 10m0s

**Next steps**

- Determine if the pod was OOM killed using `kubectl describe pod precise-code-intel-worker` (look for `OOMKilled: true`) and, if so, consider increasing the memory limit in the relevant `Deployment.yaml`.
- Check the logs before the container restarted to see if there are `panic:` messages or similar using `kubectl logs -p precise-code-intel-worker`.
- Learn more about the related dashboard panel in the [dashboards reference](./dashboards.md#precise-code-intel-worker-pods-available-percentage).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "critical_precise-code-intel-worker_pods_available_percentage"
]
```

<sub>*Managed by the [Sourcegraph Code intelligence team](https://handbook.sourcegraph.com/departments/engineering/teams/code-intelligence).*</sub>

<br />

## redis: redis-store_up

<p class="subtitle">redis-store availability</p>

**Descriptions**

- <span class="badge badge-critical">critical</span> redis: less than 1 redis-store availability for 10s

**Next steps**

- Ensure redis-store is running
- More help interpreting this metric is available in the [dashboards reference](./dashboards.md#redis-redis-store-up).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "critical_redis_redis-store_up"
]
```

<sub>*Managed by the [Sourcegraph Cloud DevOps team](https://handbook.sourcegraph.com/departments/engineering/teams/devops).*</sub>

<br />

## redis: redis-cache_up

<p class="subtitle">redis-cache availability</p>

**Descriptions**

- <span class="badge badge-critical">critical</span> redis: less than 1 redis-cache availability for 10s

**Next steps**

- Ensure redis-cache is running
- More help interpreting this metric is available in the [dashboards reference](./dashboards.md#redis-redis-cache-up).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "critical_redis_redis-cache_up"
]
```

<sub>*Managed by the [Sourcegraph Cloud DevOps team](https://handbook.sourcegraph.com/departments/engineering/teams/devops).*</sub>

<br />

## redis: provisioning_container_cpu_usage_long_term

<p class="subtitle">container cpu usage total (90th percentile over 1d) across all cores by instance</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> redis: 80%+ container cpu usage total (90th percentile over 1d) across all cores by instance for 336h0m0s

**Next steps**

- **Kubernetes:** Consider increasing CPU limits in the `Deployment.yaml` for the redis-cache service.
- **Docker Compose:** Consider increasing `cpus:` of the redis-cache container in `docker-compose.yml`.
- Learn more about the related dashboard panel in the [dashboards reference](./dashboards.md#redis-provisioning-container-cpu-usage-long-term).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_redis_provisioning_container_cpu_usage_long_term"
]
```

<sub>*Managed by the [Sourcegraph Cloud DevOps team](https://handbook.sourcegraph.com/departments/engineering/teams/devops).*</sub>

<br />

## redis: provisioning_container_memory_usage_long_term

<p class="subtitle">container memory usage (1d maximum) by instance</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> redis: 80%+ container memory usage (1d maximum) by instance for 336h0m0s

**Next steps**

- **Kubernetes:** Consider increasing memory limits in the `Deployment.yaml` for the redis-cache service.
- **Docker Compose:** Consider increasing `memory:` of the redis-cache container in `docker-compose.yml`.
- Learn more about the related dashboard panel in the [dashboards reference](./dashboards.md#redis-provisioning-container-memory-usage-long-term).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_redis_provisioning_container_memory_usage_long_term"
]
```

<sub>*Managed by the [Sourcegraph Cloud DevOps team](https://handbook.sourcegraph.com/departments/engineering/teams/devops).*</sub>

<br />

## redis: provisioning_container_cpu_usage_short_term

<p class="subtitle">container cpu usage total (5m maximum) across all cores by instance</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> redis: 90%+ container cpu usage total (5m maximum) across all cores by instance for 30m0s

**Next steps**

- **Kubernetes:** Consider increasing CPU limits in the the relevant `Deployment.yaml`.
- **Docker Compose:** Consider increasing `cpus:` of the redis-cache container in `docker-compose.yml`.
- Learn more about the related dashboard panel in the [dashboards reference](./dashboards.md#redis-provisioning-container-cpu-usage-short-term).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_redis_provisioning_container_cpu_usage_short_term"
]
```

<sub>*Managed by the [Sourcegraph Cloud DevOps team](https://handbook.sourcegraph.com/departments/engineering/teams/devops).*</sub>

<br />

## redis: provisioning_container_memory_usage_short_term

<p class="subtitle">container memory usage (5m maximum) by instance</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> redis: 90%+ container memory usage (5m maximum) by instance

**Next steps**

- **Kubernetes:** Consider increasing memory limit in relevant `Deployment.yaml`.
- **Docker Compose:** Consider increasing `memory:` of redis-cache container in `docker-compose.yml`.
- Learn more about the related dashboard panel in the [dashboards reference](./dashboards.md#redis-provisioning-container-memory-usage-short-term).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_redis_provisioning_container_memory_usage_short_term"
]
```

<sub>*Managed by the [Sourcegraph Cloud DevOps team](https://handbook.sourcegraph.com/departments/engineering/teams/devops).*</sub>

<br />

## redis: container_oomkill_events_total

<p class="subtitle">container OOMKILL events total by instance</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> redis: 1+ container OOMKILL events total by instance

**Next steps**

- **Kubernetes:** Consider increasing memory limit in relevant `Deployment.yaml`.
- **Docker Compose:** Consider increasing `memory:` of redis-cache container in `docker-compose.yml`.
- More help interpreting this metric is available in the [dashboards reference](./dashboards.md#redis-container-oomkill-events-total).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_redis_container_oomkill_events_total"
]
```

<sub>*Managed by the [Sourcegraph Cloud DevOps team](https://handbook.sourcegraph.com/departments/engineering/teams/devops).*</sub>

<br />

## redis: provisioning_container_cpu_usage_long_term

<p class="subtitle">container cpu usage total (90th percentile over 1d) across all cores by instance</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> redis: 80%+ container cpu usage total (90th percentile over 1d) across all cores by instance for 336h0m0s

**Next steps**

- **Kubernetes:** Consider increasing CPU limits in the `Deployment.yaml` for the redis-store service.
- **Docker Compose:** Consider increasing `cpus:` of the redis-store container in `docker-compose.yml`.
- Learn more about the related dashboard panel in the [dashboards reference](./dashboards.md#redis-provisioning-container-cpu-usage-long-term).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_redis_provisioning_container_cpu_usage_long_term"
]
```

<sub>*Managed by the [Sourcegraph Cloud DevOps team](https://handbook.sourcegraph.com/departments/engineering/teams/devops).*</sub>

<br />

## redis: provisioning_container_memory_usage_long_term

<p class="subtitle">container memory usage (1d maximum) by instance</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> redis: 80%+ container memory usage (1d maximum) by instance for 336h0m0s

**Next steps**

- **Kubernetes:** Consider increasing memory limits in the `Deployment.yaml` for the redis-store service.
- **Docker Compose:** Consider increasing `memory:` of the redis-store container in `docker-compose.yml`.
- Learn more about the related dashboard panel in the [dashboards reference](./dashboards.md#redis-provisioning-container-memory-usage-long-term).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_redis_provisioning_container_memory_usage_long_term"
]
```

<sub>*Managed by the [Sourcegraph Cloud DevOps team](https://handbook.sourcegraph.com/departments/engineering/teams/devops).*</sub>

<br />

## redis: provisioning_container_cpu_usage_short_term

<p class="subtitle">container cpu usage total (5m maximum) across all cores by instance</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> redis: 90%+ container cpu usage total (5m maximum) across all cores by instance for 30m0s

**Next steps**

- **Kubernetes:** Consider increasing CPU limits in the the relevant `Deployment.yaml`.
- **Docker Compose:** Consider increasing `cpus:` of the redis-store container in `docker-compose.yml`.
- Learn more about the related dashboard panel in the [dashboards reference](./dashboards.md#redis-provisioning-container-cpu-usage-short-term).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_redis_provisioning_container_cpu_usage_short_term"
]
```

<sub>*Managed by the [Sourcegraph Cloud DevOps team](https://handbook.sourcegraph.com/departments/engineering/teams/devops).*</sub>

<br />

## redis: provisioning_container_memory_usage_short_term

<p class="subtitle">container memory usage (5m maximum) by instance</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> redis: 90%+ container memory usage (5m maximum) by instance

**Next steps**

- **Kubernetes:** Consider increasing memory limit in relevant `Deployment.yaml`.
- **Docker Compose:** Consider increasing `memory:` of redis-store container in `docker-compose.yml`.
- Learn more about the related dashboard panel in the [dashboards reference](./dashboards.md#redis-provisioning-container-memory-usage-short-term).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_redis_provisioning_container_memory_usage_short_term"
]
```

<sub>*Managed by the [Sourcegraph Cloud DevOps team](https://handbook.sourcegraph.com/departments/engineering/teams/devops).*</sub>

<br />

## redis: container_oomkill_events_total

<p class="subtitle">container OOMKILL events total by instance</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> redis: 1+ container OOMKILL events total by instance

**Next steps**

- **Kubernetes:** Consider increasing memory limit in relevant `Deployment.yaml`.
- **Docker Compose:** Consider increasing `memory:` of redis-store container in `docker-compose.yml`.
- More help interpreting this metric is available in the [dashboards reference](./dashboards.md#redis-container-oomkill-events-total).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_redis_container_oomkill_events_total"
]
```

<sub>*Managed by the [Sourcegraph Cloud DevOps team](https://handbook.sourcegraph.com/departments/engineering/teams/devops).*</sub>

<br />

## redis: pods_available_percentage

<p class="subtitle">percentage pods available</p>

**Descriptions**

- <span class="badge badge-critical">critical</span> redis: less than 90% percentage pods available for 10m0s

**Next steps**

- Determine if the pod was OOM killed using `kubectl describe pod redis-cache` (look for `OOMKilled: true`) and, if so, consider increasing the memory limit in the relevant `Deployment.yaml`.
- Check the logs before the container restarted to see if there are `panic:` messages or similar using `kubectl logs -p redis-cache`.
- Learn more about the related dashboard panel in the [dashboards reference](./dashboards.md#redis-pods-available-percentage).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "critical_redis_pods_available_percentage"
]
```

<sub>*Managed by the [Sourcegraph Cloud DevOps team](https://handbook.sourcegraph.com/departments/engineering/teams/devops).*</sub>

<br />

## redis: pods_available_percentage

<p class="subtitle">percentage pods available</p>

**Descriptions**

- <span class="badge badge-critical">critical</span> redis: less than 90% percentage pods available for 10m0s

**Next steps**

- Determine if the pod was OOM killed using `kubectl describe pod redis-store` (look for `OOMKilled: true`) and, if so, consider increasing the memory limit in the relevant `Deployment.yaml`.
- Check the logs before the container restarted to see if there are `panic:` messages or similar using `kubectl logs -p redis-store`.
- Learn more about the related dashboard panel in the [dashboards reference](./dashboards.md#redis-pods-available-percentage).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "critical_redis_pods_available_percentage"
]
```

<sub>*Managed by the [Sourcegraph Cloud DevOps team](https://handbook.sourcegraph.com/departments/engineering/teams/devops).*</sub>

<br />

## worker: worker_job_codeintel-upload-janitor_count

<p class="subtitle">number of worker instances running the codeintel-upload-janitor job</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> worker: less than 1 number of worker instances running the codeintel-upload-janitor job for 1m0s
- <span class="badge badge-critical">critical</span> worker: less than 1 number of worker instances running the codeintel-upload-janitor job for 5m0s

**Next steps**

- Ensure your instance defines a worker container such that:
	- `WORKER_JOB_ALLOWLIST` contains "codeintel-upload-janitor" (or "all"), and
	- `WORKER_JOB_BLOCKLIST` does not contain "codeintel-upload-janitor"
- Ensure that such a container is not failing to start or stay active
- Learn more about the related dashboard panel in the [dashboards reference](./dashboards.md#worker-worker-job-codeintel-upload-janitor-count).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_worker_worker_job_codeintel-upload-janitor_count",
  "critical_worker_worker_job_codeintel-upload-janitor_count"
]
```

<sub>*Managed by the [Sourcegraph Code intelligence team](https://handbook.sourcegraph.com/departments/engineering/teams/code-intelligence).*</sub>

<br />

## worker: worker_job_codeintel-commitgraph-updater_count

<p class="subtitle">number of worker instances running the codeintel-commitgraph-updater job</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> worker: less than 1 number of worker instances running the codeintel-commitgraph-updater job for 1m0s
- <span class="badge badge-critical">critical</span> worker: less than 1 number of worker instances running the codeintel-commitgraph-updater job for 5m0s

**Next steps**

- Ensure your instance defines a worker container such that:
	- `WORKER_JOB_ALLOWLIST` contains "codeintel-commitgraph-updater" (or "all"), and
	- `WORKER_JOB_BLOCKLIST` does not contain "codeintel-commitgraph-updater"
- Ensure that such a container is not failing to start or stay active
- Learn more about the related dashboard panel in the [dashboards reference](./dashboards.md#worker-worker-job-codeintel-commitgraph-updater-count).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_worker_worker_job_codeintel-commitgraph-updater_count",
  "critical_worker_worker_job_codeintel-commitgraph-updater_count"
]
```

<sub>*Managed by the [Sourcegraph Code intelligence team](https://handbook.sourcegraph.com/departments/engineering/teams/code-intelligence).*</sub>

<br />

## worker: worker_job_codeintel-autoindexing-scheduler_count

<p class="subtitle">number of worker instances running the codeintel-autoindexing-scheduler job</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> worker: less than 1 number of worker instances running the codeintel-autoindexing-scheduler job for 1m0s
- <span class="badge badge-critical">critical</span> worker: less than 1 number of worker instances running the codeintel-autoindexing-scheduler job for 5m0s

**Next steps**

- Ensure your instance defines a worker container such that:
	- `WORKER_JOB_ALLOWLIST` contains "codeintel-autoindexing-scheduler" (or "all"), and
	- `WORKER_JOB_BLOCKLIST` does not contain "codeintel-autoindexing-scheduler"
- Ensure that such a container is not failing to start or stay active
- Learn more about the related dashboard panel in the [dashboards reference](./dashboards.md#worker-worker-job-codeintel-autoindexing-scheduler-count).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_worker_worker_job_codeintel-autoindexing-scheduler_count",
  "critical_worker_worker_job_codeintel-autoindexing-scheduler_count"
]
```

<sub>*Managed by the [Sourcegraph Code intelligence team](https://handbook.sourcegraph.com/departments/engineering/teams/code-intelligence).*</sub>

<br />

## worker: codeintel_commit_graph_queued_max_age

<p class="subtitle">repository queue longest time in queue</p>

**Descriptions**

- <span class="badge badge-critical">critical</span> worker: 3600s+ repository queue longest time in queue

**Next steps**

- An alert here is generally indicative of either underprovisioned worker instance(s) and/or
an underprovisioned main postgres instance.
- Learn more about the related dashboard panel in the [dashboards reference](./dashboards.md#worker-codeintel-commit-graph-queued-max-age).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "critical_worker_codeintel_commit_graph_queued_max_age"
]
```

<sub>*Managed by the [Sourcegraph Code intelligence team](https://handbook.sourcegraph.com/departments/engineering/teams/code-intelligence).*</sub>

<br />

## worker: insights_queue_unutilized_size

<p class="subtitle">insights queue size that is not utilized (not processing)</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> worker: 0+ insights queue size that is not utilized (not processing) for 30m0s

**Next steps**

- Verify code insights worker job has successfully started. Restart worker service and monitoring startup logs, looking for worker panics.
- More help interpreting this metric is available in the [dashboards reference](./dashboards.md#worker-insights-queue-unutilized-size).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_worker_insights_queue_unutilized_size"
]
```

<sub>*Managed by the [Sourcegraph Code Insights team](https://handbook.sourcegraph.com/departments/engineering/teams/code-insights).*</sub>

<br />

## worker: frontend_internal_api_error_responses

<p class="subtitle">frontend-internal API error responses every 5m by route</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> worker: 2%+ frontend-internal API error responses every 5m by route for 5m0s

**Next steps**

- **Single-container deployments:** Check `docker logs $CONTAINER_ID` for logs starting with `repo-updater` that indicate requests to the frontend service are failing.
- **Kubernetes:**
	- Confirm that `kubectl get pods` shows the `frontend` pods are healthy.
	- Check `kubectl logs worker` for logs indicate request failures to `frontend` or `frontend-internal`.
- **Docker Compose:**
	- Confirm that `docker ps` shows the `frontend-internal` container is healthy.
	- Check `docker logs worker` for logs indicating request failures to `frontend` or `frontend-internal`.
- Learn more about the related dashboard panel in the [dashboards reference](./dashboards.md#worker-frontend-internal-api-error-responses).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_worker_frontend_internal_api_error_responses"
]
```

<sub>*Managed by the [Sourcegraph Code intelligence team](https://handbook.sourcegraph.com/departments/engineering/teams/code-intelligence).*</sub>

<br />

## worker: mean_blocked_seconds_per_conn_request

<p class="subtitle">mean blocked seconds per conn request</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> worker: 0.05s+ mean blocked seconds per conn request for 10m0s
- <span class="badge badge-critical">critical</span> worker: 0.1s+ mean blocked seconds per conn request for 15m0s

**Next steps**

- Increase SRC_PGSQL_MAX_OPEN together with giving more memory to the database if needed
- Scale up Postgres memory / cpus [See our scaling guide](https://docs.sourcegraph.com/admin/config/postgres-conf)
- Learn more about the related dashboard panel in the [dashboards reference](./dashboards.md#worker-mean-blocked-seconds-per-conn-request).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_worker_mean_blocked_seconds_per_conn_request",
  "critical_worker_mean_blocked_seconds_per_conn_request"
]
```

<sub>*Managed by the [Sourcegraph Cloud DevOps team](https://handbook.sourcegraph.com/departments/engineering/teams/devops).*</sub>

<br />

## worker: container_cpu_usage

<p class="subtitle">container cpu usage total (1m average) across all cores by instance</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> worker: 99%+ container cpu usage total (1m average) across all cores by instance

**Next steps**

- **Kubernetes:** Consider increasing CPU limits in the the relevant `Deployment.yaml`.
- **Docker Compose:** Consider increasing `cpus:` of the worker container in `docker-compose.yml`.
- Learn more about the related dashboard panel in the [dashboards reference](./dashboards.md#worker-container-cpu-usage).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_worker_container_cpu_usage"
]
```

<sub>*Managed by the [Sourcegraph Code intelligence team](https://handbook.sourcegraph.com/departments/engineering/teams/code-intelligence).*</sub>

<br />

## worker: container_memory_usage

<p class="subtitle">container memory usage by instance</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> worker: 99%+ container memory usage by instance

**Next steps**

- **Kubernetes:** Consider increasing memory limit in relevant `Deployment.yaml`.
- **Docker Compose:** Consider increasing `memory:` of worker container in `docker-compose.yml`.
- Learn more about the related dashboard panel in the [dashboards reference](./dashboards.md#worker-container-memory-usage).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_worker_container_memory_usage"
]
```

<sub>*Managed by the [Sourcegraph Code intelligence team](https://handbook.sourcegraph.com/departments/engineering/teams/code-intelligence).*</sub>

<br />

## worker: provisioning_container_cpu_usage_long_term

<p class="subtitle">container cpu usage total (90th percentile over 1d) across all cores by instance</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> worker: 80%+ container cpu usage total (90th percentile over 1d) across all cores by instance for 336h0m0s

**Next steps**

- **Kubernetes:** Consider increasing CPU limits in the `Deployment.yaml` for the worker service.
- **Docker Compose:** Consider increasing `cpus:` of the worker container in `docker-compose.yml`.
- Learn more about the related dashboard panel in the [dashboards reference](./dashboards.md#worker-provisioning-container-cpu-usage-long-term).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_worker_provisioning_container_cpu_usage_long_term"
]
```

<sub>*Managed by the [Sourcegraph Code intelligence team](https://handbook.sourcegraph.com/departments/engineering/teams/code-intelligence).*</sub>

<br />

## worker: provisioning_container_memory_usage_long_term

<p class="subtitle">container memory usage (1d maximum) by instance</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> worker: 80%+ container memory usage (1d maximum) by instance for 336h0m0s

**Next steps**

- **Kubernetes:** Consider increasing memory limits in the `Deployment.yaml` for the worker service.
- **Docker Compose:** Consider increasing `memory:` of the worker container in `docker-compose.yml`.
- Learn more about the related dashboard panel in the [dashboards reference](./dashboards.md#worker-provisioning-container-memory-usage-long-term).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_worker_provisioning_container_memory_usage_long_term"
]
```

<sub>*Managed by the [Sourcegraph Code intelligence team](https://handbook.sourcegraph.com/departments/engineering/teams/code-intelligence).*</sub>

<br />

## worker: provisioning_container_cpu_usage_short_term

<p class="subtitle">container cpu usage total (5m maximum) across all cores by instance</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> worker: 90%+ container cpu usage total (5m maximum) across all cores by instance for 30m0s

**Next steps**

- **Kubernetes:** Consider increasing CPU limits in the the relevant `Deployment.yaml`.
- **Docker Compose:** Consider increasing `cpus:` of the worker container in `docker-compose.yml`.
- Learn more about the related dashboard panel in the [dashboards reference](./dashboards.md#worker-provisioning-container-cpu-usage-short-term).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_worker_provisioning_container_cpu_usage_short_term"
]
```

<sub>*Managed by the [Sourcegraph Code intelligence team](https://handbook.sourcegraph.com/departments/engineering/teams/code-intelligence).*</sub>

<br />

## worker: provisioning_container_memory_usage_short_term

<p class="subtitle">container memory usage (5m maximum) by instance</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> worker: 90%+ container memory usage (5m maximum) by instance

**Next steps**

- **Kubernetes:** Consider increasing memory limit in relevant `Deployment.yaml`.
- **Docker Compose:** Consider increasing `memory:` of worker container in `docker-compose.yml`.
- Learn more about the related dashboard panel in the [dashboards reference](./dashboards.md#worker-provisioning-container-memory-usage-short-term).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_worker_provisioning_container_memory_usage_short_term"
]
```

<sub>*Managed by the [Sourcegraph Code intelligence team](https://handbook.sourcegraph.com/departments/engineering/teams/code-intelligence).*</sub>

<br />

## worker: container_oomkill_events_total

<p class="subtitle">container OOMKILL events total by instance</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> worker: 1+ container OOMKILL events total by instance

**Next steps**

- **Kubernetes:** Consider increasing memory limit in relevant `Deployment.yaml`.
- **Docker Compose:** Consider increasing `memory:` of worker container in `docker-compose.yml`.
- More help interpreting this metric is available in the [dashboards reference](./dashboards.md#worker-container-oomkill-events-total).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_worker_container_oomkill_events_total"
]
```

<sub>*Managed by the [Sourcegraph Code intelligence team](https://handbook.sourcegraph.com/departments/engineering/teams/code-intelligence).*</sub>

<br />

## worker: go_goroutines

<p class="subtitle">maximum active goroutines</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> worker: 10000+ maximum active goroutines for 10m0s

**Next steps**

- More help interpreting this metric is available in the [dashboards reference](./dashboards.md#worker-go-goroutines).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_worker_go_goroutines"
]
```

<sub>*Managed by the [Sourcegraph Code intelligence team](https://handbook.sourcegraph.com/departments/engineering/teams/code-intelligence).*</sub>

<br />

## worker: go_gc_duration_seconds

<p class="subtitle">maximum go garbage collection duration</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> worker: 2s+ maximum go garbage collection duration

**Next steps**

- Learn more about the related dashboard panel in the [dashboards reference](./dashboards.md#worker-go-gc-duration-seconds).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_worker_go_gc_duration_seconds"
]
```

<sub>*Managed by the [Sourcegraph Code intelligence team](https://handbook.sourcegraph.com/departments/engineering/teams/code-intelligence).*</sub>

<br />

## worker: pods_available_percentage

<p class="subtitle">percentage pods available</p>

**Descriptions**

- <span class="badge badge-critical">critical</span> worker: less than 90% percentage pods available for 10m0s

**Next steps**

- Determine if the pod was OOM killed using `kubectl describe pod worker` (look for `OOMKilled: true`) and, if so, consider increasing the memory limit in the relevant `Deployment.yaml`.
- Check the logs before the container restarted to see if there are `panic:` messages or similar using `kubectl logs -p worker`.
- Learn more about the related dashboard panel in the [dashboards reference](./dashboards.md#worker-pods-available-percentage).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "critical_worker_pods_available_percentage"
]
```

<sub>*Managed by the [Sourcegraph Code intelligence team](https://handbook.sourcegraph.com/departments/engineering/teams/code-intelligence).*</sub>

<br />

## repo-updater: src_repoupdater_max_sync_backoff

<p class="subtitle">time since oldest sync</p>

**Descriptions**

- <span class="badge badge-critical">critical</span> repo-updater: 32400s+ time since oldest sync for 10m0s

**Next steps**

- An alert here indicates that no code host connections have synced in at least 9h0m0s. This indicates that there could be a configuration issue
with your code hosts connections or networking issues affecting communication with your code hosts.
- Check the code host status indicator (cloud icon in top right of Sourcegraph homepage) for errors.
- Make sure external services do not have invalid tokens by navigating to them in the web UI and clicking save. If there are no errors, they are valid.
- Check the repo-updater logs for errors about syncing.
- Confirm that outbound network connections are allowed where repo-updater is deployed.
- Check back in an hour to see if the issue has resolved itself.
- Learn more about the related dashboard panel in the [dashboards reference](./dashboards.md#repo-updater-src-repoupdater-max-sync-backoff).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "critical_repo-updater_src_repoupdater_max_sync_backoff"
]
```

<sub>*Managed by the [Sourcegraph Repo Management team](https://handbook.sourcegraph.com/departments/engineering/teams/repo-management).*</sub>

<br />

## repo-updater: src_repoupdater_syncer_sync_errors_total

<p class="subtitle">site level external service sync error rate</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> repo-updater: 0.5+ site level external service sync error rate for 10m0s
- <span class="badge badge-critical">critical</span> repo-updater: 1+ site level external service sync error rate for 10m0s

**Next steps**

- An alert here indicates errors syncing site level repo metadata with code hosts. This indicates that there could be a configuration issue
with your code hosts connections or networking issues affecting communication with your code hosts.
- Check the code host status indicator (cloud icon in top right of Sourcegraph homepage) for errors.
- Make sure external services do not have invalid tokens by navigating to them in the web UI and clicking save. If there are no errors, they are valid.
- Check the repo-updater logs for errors about syncing.
- Confirm that outbound network connections are allowed where repo-updater is deployed.
- Check back in an hour to see if the issue has resolved itself.
- Learn more about the related dashboard panel in the [dashboards reference](./dashboards.md#repo-updater-src-repoupdater-syncer-sync-errors-total).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_repo-updater_src_repoupdater_syncer_sync_errors_total",
  "critical_repo-updater_src_repoupdater_syncer_sync_errors_total"
]
```

<sub>*Managed by the [Sourcegraph Repo Management team](https://handbook.sourcegraph.com/departments/engineering/teams/repo-management).*</sub>

<br />

## repo-updater: syncer_sync_start

<p class="subtitle">repo metadata sync was started</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> repo-updater: less than 0 repo metadata sync was started for 9h0m0s

**Next steps**

- Check repo-updater logs for errors.
- Learn more about the related dashboard panel in the [dashboards reference](./dashboards.md#repo-updater-syncer-sync-start).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_repo-updater_syncer_sync_start"
]
```

<sub>*Managed by the [Sourcegraph Repo Management team](https://handbook.sourcegraph.com/departments/engineering/teams/repo-management).*</sub>

<br />

## repo-updater: syncer_sync_duration

<p class="subtitle">95th repositories sync duration</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> repo-updater: 30s+ 95th repositories sync duration for 5m0s

**Next steps**

- Check the network latency is reasonable (<50ms) between the Sourcegraph and the code host
- Learn more about the related dashboard panel in the [dashboards reference](./dashboards.md#repo-updater-syncer-sync-duration).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_repo-updater_syncer_sync_duration"
]
```

<sub>*Managed by the [Sourcegraph Repo Management team](https://handbook.sourcegraph.com/departments/engineering/teams/repo-management).*</sub>

<br />

## repo-updater: source_duration

<p class="subtitle">95th repositories source duration</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> repo-updater: 30s+ 95th repositories source duration for 5m0s

**Next steps**

- Check the network latency is reasonable (<50ms) between the Sourcegraph and the code host
- Learn more about the related dashboard panel in the [dashboards reference](./dashboards.md#repo-updater-source-duration).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_repo-updater_source_duration"
]
```

<sub>*Managed by the [Sourcegraph Repo Management team](https://handbook.sourcegraph.com/departments/engineering/teams/repo-management).*</sub>

<br />

## repo-updater: syncer_synced_repos

<p class="subtitle">repositories synced</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> repo-updater: less than 0 repositories synced for 9h0m0s

**Next steps**

- Check network connectivity to code hosts
- Learn more about the related dashboard panel in the [dashboards reference](./dashboards.md#repo-updater-syncer-synced-repos).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_repo-updater_syncer_synced_repos"
]
```

<sub>*Managed by the [Sourcegraph Repo Management team](https://handbook.sourcegraph.com/departments/engineering/teams/repo-management).*</sub>

<br />

## repo-updater: sourced_repos

<p class="subtitle">repositories sourced</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> repo-updater: less than 0 repositories sourced for 9h0m0s

**Next steps**

- Check network connectivity to code hosts
- Learn more about the related dashboard panel in the [dashboards reference](./dashboards.md#repo-updater-sourced-repos).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_repo-updater_sourced_repos"
]
```

<sub>*Managed by the [Sourcegraph Repo Management team](https://handbook.sourcegraph.com/departments/engineering/teams/repo-management).*</sub>

<br />

## repo-updater: user_added_repos

<p class="subtitle">total number of user added repos</p>

**Descriptions**

- <span class="badge badge-critical">critical</span> repo-updater: 720000+ total number of user added repos for 5m0s

**Next steps**

- Check for unusual spikes in user added repos. Each user is only allowed to add 2000 and we have a site wide limit of 800k.
- Learn more about the related dashboard panel in the [dashboards reference](./dashboards.md#repo-updater-user-added-repos).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "critical_repo-updater_user_added_repos"
]
```

<sub>*Managed by the [Sourcegraph Repo Management team](https://handbook.sourcegraph.com/departments/engineering/teams/repo-management).*</sub>

<br />

## repo-updater: purge_failed

<p class="subtitle">repositories purge failed</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> repo-updater: 0+ repositories purge failed for 5m0s

**Next steps**

- Check repo-updater`s connectivity with gitserver and gitserver logs
- Learn more about the related dashboard panel in the [dashboards reference](./dashboards.md#repo-updater-purge-failed).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_repo-updater_purge_failed"
]
```

<sub>*Managed by the [Sourcegraph Repo Management team](https://handbook.sourcegraph.com/departments/engineering/teams/repo-management).*</sub>

<br />

## repo-updater: sched_auto_fetch

<p class="subtitle">repositories scheduled due to hitting a deadline</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> repo-updater: less than 0 repositories scheduled due to hitting a deadline for 9h0m0s

**Next steps**

- Check repo-updater logs. This is expected to fire if there are no user added code hosts
- Learn more about the related dashboard panel in the [dashboards reference](./dashboards.md#repo-updater-sched-auto-fetch).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_repo-updater_sched_auto_fetch"
]
```

<sub>*Managed by the [Sourcegraph Repo Management team](https://handbook.sourcegraph.com/departments/engineering/teams/repo-management).*</sub>

<br />

## repo-updater: sched_known_repos

<p class="subtitle">repositories managed by the scheduler</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> repo-updater: less than 0 repositories managed by the scheduler for 10m0s

**Next steps**

- Check repo-updater logs. This is expected to fire if there are no user added code hosts
- Learn more about the related dashboard panel in the [dashboards reference](./dashboards.md#repo-updater-sched-known-repos).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_repo-updater_sched_known_repos"
]
```

<sub>*Managed by the [Sourcegraph Repo Management team](https://handbook.sourcegraph.com/departments/engineering/teams/repo-management).*</sub>

<br />

## repo-updater: sched_update_queue_length

<p class="subtitle">rate of growth of update queue length over 5 minutes</p>

**Descriptions**

- <span class="badge badge-critical">critical</span> repo-updater: 0+ rate of growth of update queue length over 5 minutes for 2h0m0s

**Next steps**

- Check repo-updater logs for indications that the queue is not being processed. The queue length should trend downwards over time as items are sent to GitServer
- Learn more about the related dashboard panel in the [dashboards reference](./dashboards.md#repo-updater-sched-update-queue-length).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "critical_repo-updater_sched_update_queue_length"
]
```

<sub>*Managed by the [Sourcegraph Repo Management team](https://handbook.sourcegraph.com/departments/engineering/teams/repo-management).*</sub>

<br />

## repo-updater: sched_loops

<p class="subtitle">scheduler loops</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> repo-updater: less than 0 scheduler loops for 9h0m0s

**Next steps**

- Check repo-updater logs for errors. This is expected to fire if there are no user added code hosts
- Learn more about the related dashboard panel in the [dashboards reference](./dashboards.md#repo-updater-sched-loops).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_repo-updater_sched_loops"
]
```

<sub>*Managed by the [Sourcegraph Repo Management team](https://handbook.sourcegraph.com/departments/engineering/teams/repo-management).*</sub>

<br />

## repo-updater: src_repoupdater_stale_repos

<p class="subtitle">repos that haven't been fetched in more than 8 hours</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> repo-updater: 1+ repos that haven't been fetched in more than 8 hours for 25m0s

**Next steps**

- 							Check repo-updater logs for errors.
							Check for rows in gitserver_repos where LastError is not an empty string.
- Learn more about the related dashboard panel in the [dashboards reference](./dashboards.md#repo-updater-src-repoupdater-stale-repos).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_repo-updater_src_repoupdater_stale_repos"
]
```

<sub>*Managed by the [Sourcegraph Repo Management team](https://handbook.sourcegraph.com/departments/engineering/teams/repo-management).*</sub>

<br />

## repo-updater: sched_error

<p class="subtitle">repositories schedule error rate</p>

**Descriptions**

- <span class="badge badge-critical">critical</span> repo-updater: 1+ repositories schedule error rate for 25m0s

**Next steps**

- Check repo-updater logs for errors
- Learn more about the related dashboard panel in the [dashboards reference](./dashboards.md#repo-updater-sched-error).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "critical_repo-updater_sched_error"
]
```

<sub>*Managed by the [Sourcegraph Repo Management team](https://handbook.sourcegraph.com/departments/engineering/teams/repo-management).*</sub>

<br />

## repo-updater: perms_syncer_perms

<p class="subtitle">time gap between least and most up to date permissions</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> repo-updater: 259200s+ time gap between least and most up to date permissions for 5m0s

**Next steps**

- Increase the API rate limit to [GitHub](https://docs.sourcegraph.com/admin/external_service/github#github-com-rate-limits), [GitLab](https://docs.sourcegraph.com/admin/external_service/gitlab#internal-rate-limits) or [Bitbucket Server](https://docs.sourcegraph.com/admin/external_service/bitbucket_server#internal-rate-limits).
- Learn more about the related dashboard panel in the [dashboards reference](./dashboards.md#repo-updater-perms-syncer-perms).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_repo-updater_perms_syncer_perms"
]
```

<sub>*Managed by the [Sourcegraph Repo Management team](https://handbook.sourcegraph.com/departments/engineering/teams/repo-management).*</sub>

<br />

## repo-updater: perms_syncer_stale_perms

<p class="subtitle">number of entities with stale permissions</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> repo-updater: 100+ number of entities with stale permissions for 5m0s

**Next steps**

- Increase the API rate limit to [GitHub](https://docs.sourcegraph.com/admin/external_service/github#github-com-rate-limits), [GitLab](https://docs.sourcegraph.com/admin/external_service/gitlab#internal-rate-limits) or [Bitbucket Server](https://docs.sourcegraph.com/admin/external_service/bitbucket_server#internal-rate-limits).
- Learn more about the related dashboard panel in the [dashboards reference](./dashboards.md#repo-updater-perms-syncer-stale-perms).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_repo-updater_perms_syncer_stale_perms"
]
```

<sub>*Managed by the [Sourcegraph Repo Management team](https://handbook.sourcegraph.com/departments/engineering/teams/repo-management).*</sub>

<br />

## repo-updater: perms_syncer_no_perms

<p class="subtitle">number of entities with no permissions</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> repo-updater: 100+ number of entities with no permissions for 5m0s

**Next steps**

- **Enabled permissions for the first time:** Wait for few minutes and see if the number goes down.
- **Otherwise:** Increase the API rate limit to [GitHub](https://docs.sourcegraph.com/admin/external_service/github#github-com-rate-limits), [GitLab](https://docs.sourcegraph.com/admin/external_service/gitlab#internal-rate-limits) or [Bitbucket Server](https://docs.sourcegraph.com/admin/external_service/bitbucket_server#internal-rate-limits).
- Learn more about the related dashboard panel in the [dashboards reference](./dashboards.md#repo-updater-perms-syncer-no-perms).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_repo-updater_perms_syncer_no_perms"
]
```

<sub>*Managed by the [Sourcegraph Repo Management team](https://handbook.sourcegraph.com/departments/engineering/teams/repo-management).*</sub>

<br />

## repo-updater: perms_syncer_outdated_perms

<p class="subtitle">number of entities with outdated permissions</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> repo-updater: 100+ number of entities with outdated permissions for 5m0s

**Next steps**

- **Enabled permissions for the first time:** Wait for few minutes and see if the number goes down.
- **Otherwise:** Increase the API rate limit to [GitHub](https://docs.sourcegraph.com/admin/external_service/github#github-com-rate-limits), [GitLab](https://docs.sourcegraph.com/admin/external_service/gitlab#internal-rate-limits) or [Bitbucket Server](https://docs.sourcegraph.com/admin/external_service/bitbucket_server#internal-rate-limits).
- Learn more about the related dashboard panel in the [dashboards reference](./dashboards.md#repo-updater-perms-syncer-outdated-perms).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_repo-updater_perms_syncer_outdated_perms"
]
```

<sub>*Managed by the [Sourcegraph Repo Management team](https://handbook.sourcegraph.com/departments/engineering/teams/repo-management).*</sub>

<br />

## repo-updater: perms_syncer_sync_duration

<p class="subtitle">95th permissions sync duration</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> repo-updater: 30s+ 95th permissions sync duration for 5m0s

**Next steps**

- Check the network latency is reasonable (<50ms) between the Sourcegraph and the code host.
- Learn more about the related dashboard panel in the [dashboards reference](./dashboards.md#repo-updater-perms-syncer-sync-duration).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_repo-updater_perms_syncer_sync_duration"
]
```

<sub>*Managed by the [Sourcegraph Repo Management team](https://handbook.sourcegraph.com/departments/engineering/teams/repo-management).*</sub>

<br />

## repo-updater: perms_syncer_queue_size

<p class="subtitle">permissions sync queued items</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> repo-updater: 100+ permissions sync queued items for 5m0s

**Next steps**

- **Enabled permissions for the first time:** Wait for few minutes and see if the number goes down.
- **Otherwise:** Increase the API rate limit to [GitHub](https://docs.sourcegraph.com/admin/external_service/github#github-com-rate-limits), [GitLab](https://docs.sourcegraph.com/admin/external_service/gitlab#internal-rate-limits) or [Bitbucket Server](https://docs.sourcegraph.com/admin/external_service/bitbucket_server#internal-rate-limits).
- Learn more about the related dashboard panel in the [dashboards reference](./dashboards.md#repo-updater-perms-syncer-queue-size).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_repo-updater_perms_syncer_queue_size"
]
```

<sub>*Managed by the [Sourcegraph Repo Management team](https://handbook.sourcegraph.com/departments/engineering/teams/repo-management).*</sub>

<br />

## repo-updater: perms_syncer_sync_errors

<p class="subtitle">permissions sync error rate</p>

**Descriptions**

- <span class="badge badge-critical">critical</span> repo-updater: 1+ permissions sync error rate for 1m0s

**Next steps**

- Check the network connectivity the Sourcegraph and the code host.
- Check if API rate limit quota is exhausted on the code host.
- Learn more about the related dashboard panel in the [dashboards reference](./dashboards.md#repo-updater-perms-syncer-sync-errors).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "critical_repo-updater_perms_syncer_sync_errors"
]
```

<sub>*Managed by the [Sourcegraph Repo Management team](https://handbook.sourcegraph.com/departments/engineering/teams/repo-management).*</sub>

<br />

## repo-updater: src_repoupdater_external_services_total

<p class="subtitle">the total number of external services</p>

**Descriptions**

- <span class="badge badge-critical">critical</span> repo-updater: 20000+ the total number of external services for 1h0m0s

**Next steps**

- Check for spikes in external services, could be abuse
- Learn more about the related dashboard panel in the [dashboards reference](./dashboards.md#repo-updater-src-repoupdater-external-services-total).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "critical_repo-updater_src_repoupdater_external_services_total"
]
```

<sub>*Managed by the [Sourcegraph Repo Management team](https://handbook.sourcegraph.com/departments/engineering/teams/repo-management).*</sub>

<br />

## repo-updater: src_repoupdater_user_external_services_total

<p class="subtitle">the total number of user added external services</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> repo-updater: 20000+ the total number of user added external services for 1h0m0s

**Next steps**

- Check for spikes in external services, could be abuse
- Learn more about the related dashboard panel in the [dashboards reference](./dashboards.md#repo-updater-src-repoupdater-user-external-services-total).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_repo-updater_src_repoupdater_user_external_services_total"
]
```

<sub>*Managed by the [Sourcegraph Repo Management team](https://handbook.sourcegraph.com/departments/engineering/teams/repo-management).*</sub>

<br />

## repo-updater: repoupdater_queued_sync_jobs_total

<p class="subtitle">the total number of queued sync jobs</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> repo-updater: 100+ the total number of queued sync jobs for 1h0m0s

**Next steps**

- **Check if jobs are failing to sync:** "SELECT * FROM external_service_sync_jobs WHERE state = `errored`";
- **Increase the number of workers** using the `repoConcurrentExternalServiceSyncers` site config.
- Learn more about the related dashboard panel in the [dashboards reference](./dashboards.md#repo-updater-repoupdater-queued-sync-jobs-total).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_repo-updater_repoupdater_queued_sync_jobs_total"
]
```

<sub>*Managed by the [Sourcegraph Repo Management team](https://handbook.sourcegraph.com/departments/engineering/teams/repo-management).*</sub>

<br />

## repo-updater: repoupdater_completed_sync_jobs_total

<p class="subtitle">the total number of completed sync jobs</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> repo-updater: 100000+ the total number of completed sync jobs for 1h0m0s

**Next steps**

- Check repo-updater logs. Jobs older than 1 day should have been removed.
- Learn more about the related dashboard panel in the [dashboards reference](./dashboards.md#repo-updater-repoupdater-completed-sync-jobs-total).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_repo-updater_repoupdater_completed_sync_jobs_total"
]
```

<sub>*Managed by the [Sourcegraph Repo Management team](https://handbook.sourcegraph.com/departments/engineering/teams/repo-management).*</sub>

<br />

## repo-updater: repoupdater_errored_sync_jobs_percentage

<p class="subtitle">the percentage of external services that have failed their most recent sync</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> repo-updater: 10%+ the percentage of external services that have failed their most recent sync for 1h0m0s

**Next steps**

- Check repo-updater logs. Check code host connectivity
- Learn more about the related dashboard panel in the [dashboards reference](./dashboards.md#repo-updater-repoupdater-errored-sync-jobs-percentage).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_repo-updater_repoupdater_errored_sync_jobs_percentage"
]
```

<sub>*Managed by the [Sourcegraph Repo Management team](https://handbook.sourcegraph.com/departments/engineering/teams/repo-management).*</sub>

<br />

## repo-updater: github_graphql_rate_limit_remaining

<p class="subtitle">remaining calls to GitHub graphql API before hitting the rate limit</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> repo-updater: less than 250 remaining calls to GitHub graphql API before hitting the rate limit

**Next steps**

- Consider creating a new token for the indicated resource (the `name` label for series below the threshold in the dashboard) under a dedicated machine user to reduce rate limit pressure.
- Learn more about the related dashboard panel in the [dashboards reference](./dashboards.md#repo-updater-github-graphql-rate-limit-remaining).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_repo-updater_github_graphql_rate_limit_remaining"
]
```

<sub>*Managed by the [Sourcegraph Repo Management team](https://handbook.sourcegraph.com/departments/engineering/teams/repo-management).*</sub>

<br />

## repo-updater: github_rest_rate_limit_remaining

<p class="subtitle">remaining calls to GitHub rest API before hitting the rate limit</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> repo-updater: less than 250 remaining calls to GitHub rest API before hitting the rate limit

**Next steps**

- Consider creating a new token for the indicated resource (the `name` label for series below the threshold in the dashboard) under a dedicated machine user to reduce rate limit pressure.
- Learn more about the related dashboard panel in the [dashboards reference](./dashboards.md#repo-updater-github-rest-rate-limit-remaining).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_repo-updater_github_rest_rate_limit_remaining"
]
```

<sub>*Managed by the [Sourcegraph Repo Management team](https://handbook.sourcegraph.com/departments/engineering/teams/repo-management).*</sub>

<br />

## repo-updater: github_search_rate_limit_remaining

<p class="subtitle">remaining calls to GitHub search API before hitting the rate limit</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> repo-updater: less than 5 remaining calls to GitHub search API before hitting the rate limit

**Next steps**

- Consider creating a new token for the indicated resource (the `name` label for series below the threshold in the dashboard) under a dedicated machine user to reduce rate limit pressure.
- Learn more about the related dashboard panel in the [dashboards reference](./dashboards.md#repo-updater-github-search-rate-limit-remaining).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_repo-updater_github_search_rate_limit_remaining"
]
```

<sub>*Managed by the [Sourcegraph Repo Management team](https://handbook.sourcegraph.com/departments/engineering/teams/repo-management).*</sub>

<br />

## repo-updater: gitlab_rest_rate_limit_remaining

<p class="subtitle">remaining calls to GitLab rest API before hitting the rate limit</p>

**Descriptions**

- <span class="badge badge-critical">critical</span> repo-updater: less than 30 remaining calls to GitLab rest API before hitting the rate limit

**Next steps**

- Try restarting the pod to get a different public IP.
- Learn more about the related dashboard panel in the [dashboards reference](./dashboards.md#repo-updater-gitlab-rest-rate-limit-remaining).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "critical_repo-updater_gitlab_rest_rate_limit_remaining"
]
```

<sub>*Managed by the [Sourcegraph Repo Management team](https://handbook.sourcegraph.com/departments/engineering/teams/repo-management).*</sub>

<br />

## repo-updater: frontend_internal_api_error_responses

<p class="subtitle">frontend-internal API error responses every 5m by route</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> repo-updater: 2%+ frontend-internal API error responses every 5m by route for 5m0s

**Next steps**

- **Single-container deployments:** Check `docker logs $CONTAINER_ID` for logs starting with `repo-updater` that indicate requests to the frontend service are failing.
- **Kubernetes:**
	- Confirm that `kubectl get pods` shows the `frontend` pods are healthy.
	- Check `kubectl logs repo-updater` for logs indicate request failures to `frontend` or `frontend-internal`.
- **Docker Compose:**
	- Confirm that `docker ps` shows the `frontend-internal` container is healthy.
	- Check `docker logs repo-updater` for logs indicating request failures to `frontend` or `frontend-internal`.
- Learn more about the related dashboard panel in the [dashboards reference](./dashboards.md#repo-updater-frontend-internal-api-error-responses).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_repo-updater_frontend_internal_api_error_responses"
]
```

<sub>*Managed by the [Sourcegraph Repo Management team](https://handbook.sourcegraph.com/departments/engineering/teams/repo-management).*</sub>

<br />

## repo-updater: mean_blocked_seconds_per_conn_request

<p class="subtitle">mean blocked seconds per conn request</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> repo-updater: 0.05s+ mean blocked seconds per conn request for 10m0s
- <span class="badge badge-critical">critical</span> repo-updater: 0.1s+ mean blocked seconds per conn request for 15m0s

**Next steps**

- Increase SRC_PGSQL_MAX_OPEN together with giving more memory to the database if needed
- Scale up Postgres memory / cpus [See our scaling guide](https://docs.sourcegraph.com/admin/config/postgres-conf)
- Learn more about the related dashboard panel in the [dashboards reference](./dashboards.md#repo-updater-mean-blocked-seconds-per-conn-request).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_repo-updater_mean_blocked_seconds_per_conn_request",
  "critical_repo-updater_mean_blocked_seconds_per_conn_request"
]
```

<sub>*Managed by the [Sourcegraph Cloud DevOps team](https://handbook.sourcegraph.com/departments/engineering/teams/devops).*</sub>

<br />

## repo-updater: container_cpu_usage

<p class="subtitle">container cpu usage total (1m average) across all cores by instance</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> repo-updater: 99%+ container cpu usage total (1m average) across all cores by instance

**Next steps**

- **Kubernetes:** Consider increasing CPU limits in the the relevant `Deployment.yaml`.
- **Docker Compose:** Consider increasing `cpus:` of the repo-updater container in `docker-compose.yml`.
- Learn more about the related dashboard panel in the [dashboards reference](./dashboards.md#repo-updater-container-cpu-usage).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_repo-updater_container_cpu_usage"
]
```

<sub>*Managed by the [Sourcegraph Repo Management team](https://handbook.sourcegraph.com/departments/engineering/teams/repo-management).*</sub>

<br />

## repo-updater: container_memory_usage

<p class="subtitle">container memory usage by instance</p>

**Descriptions**

- <span class="badge badge-critical">critical</span> repo-updater: 90%+ container memory usage by instance for 10m0s

**Next steps**

- **Kubernetes:** Consider increasing memory limit in relevant `Deployment.yaml`.
- **Docker Compose:** Consider increasing `memory:` of repo-updater container in `docker-compose.yml`.
- Learn more about the related dashboard panel in the [dashboards reference](./dashboards.md#repo-updater-container-memory-usage).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "critical_repo-updater_container_memory_usage"
]
```

<sub>*Managed by the [Sourcegraph Repo Management team](https://handbook.sourcegraph.com/departments/engineering/teams/repo-management).*</sub>

<br />

## repo-updater: provisioning_container_cpu_usage_long_term

<p class="subtitle">container cpu usage total (90th percentile over 1d) across all cores by instance</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> repo-updater: 80%+ container cpu usage total (90th percentile over 1d) across all cores by instance for 336h0m0s

**Next steps**

- **Kubernetes:** Consider increasing CPU limits in the `Deployment.yaml` for the repo-updater service.
- **Docker Compose:** Consider increasing `cpus:` of the repo-updater container in `docker-compose.yml`.
- Learn more about the related dashboard panel in the [dashboards reference](./dashboards.md#repo-updater-provisioning-container-cpu-usage-long-term).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_repo-updater_provisioning_container_cpu_usage_long_term"
]
```

<sub>*Managed by the [Sourcegraph Repo Management team](https://handbook.sourcegraph.com/departments/engineering/teams/repo-management).*</sub>

<br />

## repo-updater: provisioning_container_memory_usage_long_term

<p class="subtitle">container memory usage (1d maximum) by instance</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> repo-updater: 80%+ container memory usage (1d maximum) by instance for 336h0m0s

**Next steps**

- **Kubernetes:** Consider increasing memory limits in the `Deployment.yaml` for the repo-updater service.
- **Docker Compose:** Consider increasing `memory:` of the repo-updater container in `docker-compose.yml`.
- Learn more about the related dashboard panel in the [dashboards reference](./dashboards.md#repo-updater-provisioning-container-memory-usage-long-term).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_repo-updater_provisioning_container_memory_usage_long_term"
]
```

<sub>*Managed by the [Sourcegraph Repo Management team](https://handbook.sourcegraph.com/departments/engineering/teams/repo-management).*</sub>

<br />

## repo-updater: provisioning_container_cpu_usage_short_term

<p class="subtitle">container cpu usage total (5m maximum) across all cores by instance</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> repo-updater: 90%+ container cpu usage total (5m maximum) across all cores by instance for 30m0s

**Next steps**

- **Kubernetes:** Consider increasing CPU limits in the the relevant `Deployment.yaml`.
- **Docker Compose:** Consider increasing `cpus:` of the repo-updater container in `docker-compose.yml`.
- Learn more about the related dashboard panel in the [dashboards reference](./dashboards.md#repo-updater-provisioning-container-cpu-usage-short-term).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_repo-updater_provisioning_container_cpu_usage_short_term"
]
```

<sub>*Managed by the [Sourcegraph Repo Management team](https://handbook.sourcegraph.com/departments/engineering/teams/repo-management).*</sub>

<br />

## repo-updater: provisioning_container_memory_usage_short_term

<p class="subtitle">container memory usage (5m maximum) by instance</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> repo-updater: 90%+ container memory usage (5m maximum) by instance

**Next steps**

- **Kubernetes:** Consider increasing memory limit in relevant `Deployment.yaml`.
- **Docker Compose:** Consider increasing `memory:` of repo-updater container in `docker-compose.yml`.
- Learn more about the related dashboard panel in the [dashboards reference](./dashboards.md#repo-updater-provisioning-container-memory-usage-short-term).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_repo-updater_provisioning_container_memory_usage_short_term"
]
```

<sub>*Managed by the [Sourcegraph Repo Management team](https://handbook.sourcegraph.com/departments/engineering/teams/repo-management).*</sub>

<br />

## repo-updater: container_oomkill_events_total

<p class="subtitle">container OOMKILL events total by instance</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> repo-updater: 1+ container OOMKILL events total by instance

**Next steps**

- **Kubernetes:** Consider increasing memory limit in relevant `Deployment.yaml`.
- **Docker Compose:** Consider increasing `memory:` of repo-updater container in `docker-compose.yml`.
- More help interpreting this metric is available in the [dashboards reference](./dashboards.md#repo-updater-container-oomkill-events-total).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_repo-updater_container_oomkill_events_total"
]
```

<sub>*Managed by the [Sourcegraph Repo Management team](https://handbook.sourcegraph.com/departments/engineering/teams/repo-management).*</sub>

<br />

## repo-updater: go_goroutines

<p class="subtitle">maximum active goroutines</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> repo-updater: 10000+ maximum active goroutines for 10m0s

**Next steps**

- More help interpreting this metric is available in the [dashboards reference](./dashboards.md#repo-updater-go-goroutines).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_repo-updater_go_goroutines"
]
```

<sub>*Managed by the [Sourcegraph Repo Management team](https://handbook.sourcegraph.com/departments/engineering/teams/repo-management).*</sub>

<br />

## repo-updater: go_gc_duration_seconds

<p class="subtitle">maximum go garbage collection duration</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> repo-updater: 2s+ maximum go garbage collection duration

**Next steps**

- Learn more about the related dashboard panel in the [dashboards reference](./dashboards.md#repo-updater-go-gc-duration-seconds).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_repo-updater_go_gc_duration_seconds"
]
```

<sub>*Managed by the [Sourcegraph Repo Management team](https://handbook.sourcegraph.com/departments/engineering/teams/repo-management).*</sub>

<br />

## repo-updater: pods_available_percentage

<p class="subtitle">percentage pods available</p>

**Descriptions**

- <span class="badge badge-critical">critical</span> repo-updater: less than 90% percentage pods available for 10m0s

**Next steps**

- Determine if the pod was OOM killed using `kubectl describe pod repo-updater` (look for `OOMKilled: true`) and, if so, consider increasing the memory limit in the relevant `Deployment.yaml`.
- Check the logs before the container restarted to see if there are `panic:` messages or similar using `kubectl logs -p repo-updater`.
- Learn more about the related dashboard panel in the [dashboards reference](./dashboards.md#repo-updater-pods-available-percentage).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "critical_repo-updater_pods_available_percentage"
]
```

<sub>*Managed by the [Sourcegraph Repo Management team](https://handbook.sourcegraph.com/departments/engineering/teams/repo-management).*</sub>

<br />

## searcher: unindexed_search_request_errors

<p class="subtitle">unindexed search request errors every 5m by code</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> searcher: 5%+ unindexed search request errors every 5m by code for 5m0s

**Next steps**

- Learn more about the related dashboard panel in the [dashboards reference](./dashboards.md#searcher-unindexed-search-request-errors).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_searcher_unindexed_search_request_errors"
]
```

<sub>*Managed by the [Sourcegraph Search Core team](https://handbook.sourcegraph.com/departments/engineering/teams/search/core).*</sub>

<br />

## searcher: replica_traffic

<p class="subtitle">requests per second over 10m</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> searcher: 5+ requests per second over 10m

**Next steps**

- Learn more about the related dashboard panel in the [dashboards reference](./dashboards.md#searcher-replica-traffic).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_searcher_replica_traffic"
]
```

<sub>*Managed by the [Sourcegraph Search Core team](https://handbook.sourcegraph.com/departments/engineering/teams/search/core).*</sub>

<br />

## searcher: mean_blocked_seconds_per_conn_request

<p class="subtitle">mean blocked seconds per conn request</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> searcher: 0.05s+ mean blocked seconds per conn request for 10m0s
- <span class="badge badge-critical">critical</span> searcher: 0.1s+ mean blocked seconds per conn request for 15m0s

**Next steps**

- Increase SRC_PGSQL_MAX_OPEN together with giving more memory to the database if needed
- Scale up Postgres memory / cpus [See our scaling guide](https://docs.sourcegraph.com/admin/config/postgres-conf)
- Learn more about the related dashboard panel in the [dashboards reference](./dashboards.md#searcher-mean-blocked-seconds-per-conn-request).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_searcher_mean_blocked_seconds_per_conn_request",
  "critical_searcher_mean_blocked_seconds_per_conn_request"
]
```

<sub>*Managed by the [Sourcegraph Cloud DevOps team](https://handbook.sourcegraph.com/departments/engineering/teams/devops).*</sub>

<br />

## searcher: frontend_internal_api_error_responses

<p class="subtitle">frontend-internal API error responses every 5m by route</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> searcher: 2%+ frontend-internal API error responses every 5m by route for 5m0s

**Next steps**

- **Single-container deployments:** Check `docker logs $CONTAINER_ID` for logs starting with `repo-updater` that indicate requests to the frontend service are failing.
- **Kubernetes:**
	- Confirm that `kubectl get pods` shows the `frontend` pods are healthy.
	- Check `kubectl logs searcher` for logs indicate request failures to `frontend` or `frontend-internal`.
- **Docker Compose:**
	- Confirm that `docker ps` shows the `frontend-internal` container is healthy.
	- Check `docker logs searcher` for logs indicating request failures to `frontend` or `frontend-internal`.
- Learn more about the related dashboard panel in the [dashboards reference](./dashboards.md#searcher-frontend-internal-api-error-responses).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_searcher_frontend_internal_api_error_responses"
]
```

<sub>*Managed by the [Sourcegraph Search Core team](https://handbook.sourcegraph.com/departments/engineering/teams/search/core).*</sub>

<br />

## searcher: container_cpu_usage

<p class="subtitle">container cpu usage total (1m average) across all cores by instance</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> searcher: 99%+ container cpu usage total (1m average) across all cores by instance

**Next steps**

- **Kubernetes:** Consider increasing CPU limits in the the relevant `Deployment.yaml`.
- **Docker Compose:** Consider increasing `cpus:` of the searcher container in `docker-compose.yml`.
- Learn more about the related dashboard panel in the [dashboards reference](./dashboards.md#searcher-container-cpu-usage).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_searcher_container_cpu_usage"
]
```

<sub>*Managed by the [Sourcegraph Search Core team](https://handbook.sourcegraph.com/departments/engineering/teams/search/core).*</sub>

<br />

## searcher: container_memory_usage

<p class="subtitle">container memory usage by instance</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> searcher: 99%+ container memory usage by instance

**Next steps**

- **Kubernetes:** Consider increasing memory limit in relevant `Deployment.yaml`.
- **Docker Compose:** Consider increasing `memory:` of searcher container in `docker-compose.yml`.
- Learn more about the related dashboard panel in the [dashboards reference](./dashboards.md#searcher-container-memory-usage).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_searcher_container_memory_usage"
]
```

<sub>*Managed by the [Sourcegraph Search Core team](https://handbook.sourcegraph.com/departments/engineering/teams/search/core).*</sub>

<br />

## searcher: provisioning_container_cpu_usage_long_term

<p class="subtitle">container cpu usage total (90th percentile over 1d) across all cores by instance</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> searcher: 80%+ container cpu usage total (90th percentile over 1d) across all cores by instance for 336h0m0s

**Next steps**

- **Kubernetes:** Consider increasing CPU limits in the `Deployment.yaml` for the searcher service.
- **Docker Compose:** Consider increasing `cpus:` of the searcher container in `docker-compose.yml`.
- Learn more about the related dashboard panel in the [dashboards reference](./dashboards.md#searcher-provisioning-container-cpu-usage-long-term).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_searcher_provisioning_container_cpu_usage_long_term"
]
```

<sub>*Managed by the [Sourcegraph Search Core team](https://handbook.sourcegraph.com/departments/engineering/teams/search/core).*</sub>

<br />

## searcher: provisioning_container_memory_usage_long_term

<p class="subtitle">container memory usage (1d maximum) by instance</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> searcher: 80%+ container memory usage (1d maximum) by instance for 336h0m0s

**Next steps**

- **Kubernetes:** Consider increasing memory limits in the `Deployment.yaml` for the searcher service.
- **Docker Compose:** Consider increasing `memory:` of the searcher container in `docker-compose.yml`.
- Learn more about the related dashboard panel in the [dashboards reference](./dashboards.md#searcher-provisioning-container-memory-usage-long-term).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_searcher_provisioning_container_memory_usage_long_term"
]
```

<sub>*Managed by the [Sourcegraph Search Core team](https://handbook.sourcegraph.com/departments/engineering/teams/search/core).*</sub>

<br />

## searcher: provisioning_container_cpu_usage_short_term

<p class="subtitle">container cpu usage total (5m maximum) across all cores by instance</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> searcher: 90%+ container cpu usage total (5m maximum) across all cores by instance for 30m0s

**Next steps**

- **Kubernetes:** Consider increasing CPU limits in the the relevant `Deployment.yaml`.
- **Docker Compose:** Consider increasing `cpus:` of the searcher container in `docker-compose.yml`.
- Learn more about the related dashboard panel in the [dashboards reference](./dashboards.md#searcher-provisioning-container-cpu-usage-short-term).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_searcher_provisioning_container_cpu_usage_short_term"
]
```

<sub>*Managed by the [Sourcegraph Search Core team](https://handbook.sourcegraph.com/departments/engineering/teams/search/core).*</sub>

<br />

## searcher: provisioning_container_memory_usage_short_term

<p class="subtitle">container memory usage (5m maximum) by instance</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> searcher: 90%+ container memory usage (5m maximum) by instance

**Next steps**

- **Kubernetes:** Consider increasing memory limit in relevant `Deployment.yaml`.
- **Docker Compose:** Consider increasing `memory:` of searcher container in `docker-compose.yml`.
- Learn more about the related dashboard panel in the [dashboards reference](./dashboards.md#searcher-provisioning-container-memory-usage-short-term).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_searcher_provisioning_container_memory_usage_short_term"
]
```

<sub>*Managed by the [Sourcegraph Search Core team](https://handbook.sourcegraph.com/departments/engineering/teams/search/core).*</sub>

<br />

## searcher: container_oomkill_events_total

<p class="subtitle">container OOMKILL events total by instance</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> searcher: 1+ container OOMKILL events total by instance

**Next steps**

- **Kubernetes:** Consider increasing memory limit in relevant `Deployment.yaml`.
- **Docker Compose:** Consider increasing `memory:` of searcher container in `docker-compose.yml`.
- More help interpreting this metric is available in the [dashboards reference](./dashboards.md#searcher-container-oomkill-events-total).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_searcher_container_oomkill_events_total"
]
```

<sub>*Managed by the [Sourcegraph Search Core team](https://handbook.sourcegraph.com/departments/engineering/teams/search/core).*</sub>

<br />

## searcher: go_goroutines

<p class="subtitle">maximum active goroutines</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> searcher: 10000+ maximum active goroutines for 10m0s

**Next steps**

- More help interpreting this metric is available in the [dashboards reference](./dashboards.md#searcher-go-goroutines).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_searcher_go_goroutines"
]
```

<sub>*Managed by the [Sourcegraph Search Core team](https://handbook.sourcegraph.com/departments/engineering/teams/search/core).*</sub>

<br />

## searcher: go_gc_duration_seconds

<p class="subtitle">maximum go garbage collection duration</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> searcher: 2s+ maximum go garbage collection duration

**Next steps**

- Learn more about the related dashboard panel in the [dashboards reference](./dashboards.md#searcher-go-gc-duration-seconds).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_searcher_go_gc_duration_seconds"
]
```

<sub>*Managed by the [Sourcegraph Search Core team](https://handbook.sourcegraph.com/departments/engineering/teams/search/core).*</sub>

<br />

## searcher: pods_available_percentage

<p class="subtitle">percentage pods available</p>

**Descriptions**

- <span class="badge badge-critical">critical</span> searcher: less than 90% percentage pods available for 10m0s

**Next steps**

- Determine if the pod was OOM killed using `kubectl describe pod searcher` (look for `OOMKilled: true`) and, if so, consider increasing the memory limit in the relevant `Deployment.yaml`.
- Check the logs before the container restarted to see if there are `panic:` messages or similar using `kubectl logs -p searcher`.
- Learn more about the related dashboard panel in the [dashboards reference](./dashboards.md#searcher-pods-available-percentage).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "critical_searcher_pods_available_percentage"
]
```

<sub>*Managed by the [Sourcegraph Search Core team](https://handbook.sourcegraph.com/departments/engineering/teams/search/core).*</sub>

<br />

## symbols: mean_blocked_seconds_per_conn_request

<p class="subtitle">mean blocked seconds per conn request</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> symbols: 0.05s+ mean blocked seconds per conn request for 10m0s
- <span class="badge badge-critical">critical</span> symbols: 0.1s+ mean blocked seconds per conn request for 15m0s

**Next steps**

- Increase SRC_PGSQL_MAX_OPEN together with giving more memory to the database if needed
- Scale up Postgres memory / cpus [See our scaling guide](https://docs.sourcegraph.com/admin/config/postgres-conf)
- Learn more about the related dashboard panel in the [dashboards reference](./dashboards.md#symbols-mean-blocked-seconds-per-conn-request).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_symbols_mean_blocked_seconds_per_conn_request",
  "critical_symbols_mean_blocked_seconds_per_conn_request"
]
```

<sub>*Managed by the [Sourcegraph Cloud DevOps team](https://handbook.sourcegraph.com/departments/engineering/teams/devops).*</sub>

<br />

## symbols: frontend_internal_api_error_responses

<p class="subtitle">frontend-internal API error responses every 5m by route</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> symbols: 2%+ frontend-internal API error responses every 5m by route for 5m0s

**Next steps**

- **Single-container deployments:** Check `docker logs $CONTAINER_ID` for logs starting with `repo-updater` that indicate requests to the frontend service are failing.
- **Kubernetes:**
	- Confirm that `kubectl get pods` shows the `frontend` pods are healthy.
	- Check `kubectl logs symbols` for logs indicate request failures to `frontend` or `frontend-internal`.
- **Docker Compose:**
	- Confirm that `docker ps` shows the `frontend-internal` container is healthy.
	- Check `docker logs symbols` for logs indicating request failures to `frontend` or `frontend-internal`.
- Learn more about the related dashboard panel in the [dashboards reference](./dashboards.md#symbols-frontend-internal-api-error-responses).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_symbols_frontend_internal_api_error_responses"
]
```

<sub>*Managed by the [Sourcegraph Code intelligence team](https://handbook.sourcegraph.com/departments/engineering/teams/code-intelligence).*</sub>

<br />

## symbols: container_cpu_usage

<p class="subtitle">container cpu usage total (1m average) across all cores by instance</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> symbols: 99%+ container cpu usage total (1m average) across all cores by instance

**Next steps**

- **Kubernetes:** Consider increasing CPU limits in the the relevant `Deployment.yaml`.
- **Docker Compose:** Consider increasing `cpus:` of the symbols container in `docker-compose.yml`.
- Learn more about the related dashboard panel in the [dashboards reference](./dashboards.md#symbols-container-cpu-usage).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_symbols_container_cpu_usage"
]
```

<sub>*Managed by the [Sourcegraph Code intelligence team](https://handbook.sourcegraph.com/departments/engineering/teams/code-intelligence).*</sub>

<br />

## symbols: container_memory_usage

<p class="subtitle">container memory usage by instance</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> symbols: 99%+ container memory usage by instance

**Next steps**

- **Kubernetes:** Consider increasing memory limit in relevant `Deployment.yaml`.
- **Docker Compose:** Consider increasing `memory:` of symbols container in `docker-compose.yml`.
- Learn more about the related dashboard panel in the [dashboards reference](./dashboards.md#symbols-container-memory-usage).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_symbols_container_memory_usage"
]
```

<sub>*Managed by the [Sourcegraph Code intelligence team](https://handbook.sourcegraph.com/departments/engineering/teams/code-intelligence).*</sub>

<br />

## symbols: provisioning_container_cpu_usage_long_term

<p class="subtitle">container cpu usage total (90th percentile over 1d) across all cores by instance</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> symbols: 80%+ container cpu usage total (90th percentile over 1d) across all cores by instance for 336h0m0s

**Next steps**

- **Kubernetes:** Consider increasing CPU limits in the `Deployment.yaml` for the symbols service.
- **Docker Compose:** Consider increasing `cpus:` of the symbols container in `docker-compose.yml`.
- Learn more about the related dashboard panel in the [dashboards reference](./dashboards.md#symbols-provisioning-container-cpu-usage-long-term).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_symbols_provisioning_container_cpu_usage_long_term"
]
```

<sub>*Managed by the [Sourcegraph Code intelligence team](https://handbook.sourcegraph.com/departments/engineering/teams/code-intelligence).*</sub>

<br />

## symbols: provisioning_container_memory_usage_long_term

<p class="subtitle">container memory usage (1d maximum) by instance</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> symbols: 80%+ container memory usage (1d maximum) by instance for 336h0m0s

**Next steps**

- **Kubernetes:** Consider increasing memory limits in the `Deployment.yaml` for the symbols service.
- **Docker Compose:** Consider increasing `memory:` of the symbols container in `docker-compose.yml`.
- Learn more about the related dashboard panel in the [dashboards reference](./dashboards.md#symbols-provisioning-container-memory-usage-long-term).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_symbols_provisioning_container_memory_usage_long_term"
]
```

<sub>*Managed by the [Sourcegraph Code intelligence team](https://handbook.sourcegraph.com/departments/engineering/teams/code-intelligence).*</sub>

<br />

## symbols: provisioning_container_cpu_usage_short_term

<p class="subtitle">container cpu usage total (5m maximum) across all cores by instance</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> symbols: 90%+ container cpu usage total (5m maximum) across all cores by instance for 30m0s

**Next steps**

- **Kubernetes:** Consider increasing CPU limits in the the relevant `Deployment.yaml`.
- **Docker Compose:** Consider increasing `cpus:` of the symbols container in `docker-compose.yml`.
- Learn more about the related dashboard panel in the [dashboards reference](./dashboards.md#symbols-provisioning-container-cpu-usage-short-term).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_symbols_provisioning_container_cpu_usage_short_term"
]
```

<sub>*Managed by the [Sourcegraph Code intelligence team](https://handbook.sourcegraph.com/departments/engineering/teams/code-intelligence).*</sub>

<br />

## symbols: provisioning_container_memory_usage_short_term

<p class="subtitle">container memory usage (5m maximum) by instance</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> symbols: 90%+ container memory usage (5m maximum) by instance

**Next steps**

- **Kubernetes:** Consider increasing memory limit in relevant `Deployment.yaml`.
- **Docker Compose:** Consider increasing `memory:` of symbols container in `docker-compose.yml`.
- Learn more about the related dashboard panel in the [dashboards reference](./dashboards.md#symbols-provisioning-container-memory-usage-short-term).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_symbols_provisioning_container_memory_usage_short_term"
]
```

<sub>*Managed by the [Sourcegraph Code intelligence team](https://handbook.sourcegraph.com/departments/engineering/teams/code-intelligence).*</sub>

<br />

## symbols: container_oomkill_events_total

<p class="subtitle">container OOMKILL events total by instance</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> symbols: 1+ container OOMKILL events total by instance

**Next steps**

- **Kubernetes:** Consider increasing memory limit in relevant `Deployment.yaml`.
- **Docker Compose:** Consider increasing `memory:` of symbols container in `docker-compose.yml`.
- More help interpreting this metric is available in the [dashboards reference](./dashboards.md#symbols-container-oomkill-events-total).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_symbols_container_oomkill_events_total"
]
```

<sub>*Managed by the [Sourcegraph Code intelligence team](https://handbook.sourcegraph.com/departments/engineering/teams/code-intelligence).*</sub>

<br />

## symbols: go_goroutines

<p class="subtitle">maximum active goroutines</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> symbols: 10000+ maximum active goroutines for 10m0s

**Next steps**

- More help interpreting this metric is available in the [dashboards reference](./dashboards.md#symbols-go-goroutines).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_symbols_go_goroutines"
]
```

<sub>*Managed by the [Sourcegraph Code intelligence team](https://handbook.sourcegraph.com/departments/engineering/teams/code-intelligence).*</sub>

<br />

## symbols: go_gc_duration_seconds

<p class="subtitle">maximum go garbage collection duration</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> symbols: 2s+ maximum go garbage collection duration

**Next steps**

- Learn more about the related dashboard panel in the [dashboards reference](./dashboards.md#symbols-go-gc-duration-seconds).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_symbols_go_gc_duration_seconds"
]
```

<sub>*Managed by the [Sourcegraph Code intelligence team](https://handbook.sourcegraph.com/departments/engineering/teams/code-intelligence).*</sub>

<br />

## symbols: pods_available_percentage

<p class="subtitle">percentage pods available</p>

**Descriptions**

- <span class="badge badge-critical">critical</span> symbols: less than 90% percentage pods available for 10m0s

**Next steps**

- Determine if the pod was OOM killed using `kubectl describe pod symbols` (look for `OOMKilled: true`) and, if so, consider increasing the memory limit in the relevant `Deployment.yaml`.
- Check the logs before the container restarted to see if there are `panic:` messages or similar using `kubectl logs -p symbols`.
- Learn more about the related dashboard panel in the [dashboards reference](./dashboards.md#symbols-pods-available-percentage).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "critical_symbols_pods_available_percentage"
]
```

<sub>*Managed by the [Sourcegraph Code intelligence team](https://handbook.sourcegraph.com/departments/engineering/teams/code-intelligence).*</sub>

<br />

## syntect-server: container_cpu_usage

<p class="subtitle">container cpu usage total (1m average) across all cores by instance</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> syntect-server: 99%+ container cpu usage total (1m average) across all cores by instance

**Next steps**

- **Kubernetes:** Consider increasing CPU limits in the the relevant `Deployment.yaml`.
- **Docker Compose:** Consider increasing `cpus:` of the syntect-server container in `docker-compose.yml`.
- Learn more about the related dashboard panel in the [dashboards reference](./dashboards.md#syntect-server-container-cpu-usage).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_syntect-server_container_cpu_usage"
]
```

<sub>*Managed by the [Sourcegraph Cloud DevOps team](https://handbook.sourcegraph.com/departments/engineering/teams/devops).*</sub>

<br />

## syntect-server: container_memory_usage

<p class="subtitle">container memory usage by instance</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> syntect-server: 99%+ container memory usage by instance

**Next steps**

- **Kubernetes:** Consider increasing memory limit in relevant `Deployment.yaml`.
- **Docker Compose:** Consider increasing `memory:` of syntect-server container in `docker-compose.yml`.
- Learn more about the related dashboard panel in the [dashboards reference](./dashboards.md#syntect-server-container-memory-usage).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_syntect-server_container_memory_usage"
]
```

<sub>*Managed by the [Sourcegraph Cloud DevOps team](https://handbook.sourcegraph.com/departments/engineering/teams/devops).*</sub>

<br />

## syntect-server: provisioning_container_cpu_usage_long_term

<p class="subtitle">container cpu usage total (90th percentile over 1d) across all cores by instance</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> syntect-server: 80%+ container cpu usage total (90th percentile over 1d) across all cores by instance for 336h0m0s

**Next steps**

- **Kubernetes:** Consider increasing CPU limits in the `Deployment.yaml` for the syntect-server service.
- **Docker Compose:** Consider increasing `cpus:` of the syntect-server container in `docker-compose.yml`.
- Learn more about the related dashboard panel in the [dashboards reference](./dashboards.md#syntect-server-provisioning-container-cpu-usage-long-term).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_syntect-server_provisioning_container_cpu_usage_long_term"
]
```

<sub>*Managed by the [Sourcegraph Cloud DevOps team](https://handbook.sourcegraph.com/departments/engineering/teams/devops).*</sub>

<br />

## syntect-server: provisioning_container_memory_usage_long_term

<p class="subtitle">container memory usage (1d maximum) by instance</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> syntect-server: 80%+ container memory usage (1d maximum) by instance for 336h0m0s

**Next steps**

- **Kubernetes:** Consider increasing memory limits in the `Deployment.yaml` for the syntect-server service.
- **Docker Compose:** Consider increasing `memory:` of the syntect-server container in `docker-compose.yml`.
- Learn more about the related dashboard panel in the [dashboards reference](./dashboards.md#syntect-server-provisioning-container-memory-usage-long-term).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_syntect-server_provisioning_container_memory_usage_long_term"
]
```

<sub>*Managed by the [Sourcegraph Cloud DevOps team](https://handbook.sourcegraph.com/departments/engineering/teams/devops).*</sub>

<br />

## syntect-server: provisioning_container_cpu_usage_short_term

<p class="subtitle">container cpu usage total (5m maximum) across all cores by instance</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> syntect-server: 90%+ container cpu usage total (5m maximum) across all cores by instance for 30m0s

**Next steps**

- **Kubernetes:** Consider increasing CPU limits in the the relevant `Deployment.yaml`.
- **Docker Compose:** Consider increasing `cpus:` of the syntect-server container in `docker-compose.yml`.
- Learn more about the related dashboard panel in the [dashboards reference](./dashboards.md#syntect-server-provisioning-container-cpu-usage-short-term).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_syntect-server_provisioning_container_cpu_usage_short_term"
]
```

<sub>*Managed by the [Sourcegraph Cloud DevOps team](https://handbook.sourcegraph.com/departments/engineering/teams/devops).*</sub>

<br />

## syntect-server: provisioning_container_memory_usage_short_term

<p class="subtitle">container memory usage (5m maximum) by instance</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> syntect-server: 90%+ container memory usage (5m maximum) by instance

**Next steps**

- **Kubernetes:** Consider increasing memory limit in relevant `Deployment.yaml`.
- **Docker Compose:** Consider increasing `memory:` of syntect-server container in `docker-compose.yml`.
- Learn more about the related dashboard panel in the [dashboards reference](./dashboards.md#syntect-server-provisioning-container-memory-usage-short-term).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_syntect-server_provisioning_container_memory_usage_short_term"
]
```

<sub>*Managed by the [Sourcegraph Cloud DevOps team](https://handbook.sourcegraph.com/departments/engineering/teams/devops).*</sub>

<br />

## syntect-server: container_oomkill_events_total

<p class="subtitle">container OOMKILL events total by instance</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> syntect-server: 1+ container OOMKILL events total by instance

**Next steps**

- **Kubernetes:** Consider increasing memory limit in relevant `Deployment.yaml`.
- **Docker Compose:** Consider increasing `memory:` of syntect-server container in `docker-compose.yml`.
- More help interpreting this metric is available in the [dashboards reference](./dashboards.md#syntect-server-container-oomkill-events-total).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_syntect-server_container_oomkill_events_total"
]
```

<sub>*Managed by the [Sourcegraph Cloud DevOps team](https://handbook.sourcegraph.com/departments/engineering/teams/devops).*</sub>

<br />

## syntect-server: pods_available_percentage

<p class="subtitle">percentage pods available</p>

**Descriptions**

- <span class="badge badge-critical">critical</span> syntect-server: less than 90% percentage pods available for 10m0s

**Next steps**

- Determine if the pod was OOM killed using `kubectl describe pod syntect-server` (look for `OOMKilled: true`) and, if so, consider increasing the memory limit in the relevant `Deployment.yaml`.
- Check the logs before the container restarted to see if there are `panic:` messages or similar using `kubectl logs -p syntect-server`.
- Learn more about the related dashboard panel in the [dashboards reference](./dashboards.md#syntect-server-pods-available-percentage).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "critical_syntect-server_pods_available_percentage"
]
```

<sub>*Managed by the [Sourcegraph Cloud DevOps team](https://handbook.sourcegraph.com/departments/engineering/teams/devops).*</sub>

<br />

## zoekt: average_resolve_revision_duration

<p class="subtitle">average resolve revision duration over 5m</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> zoekt: 15s+ average resolve revision duration over 5m

**Next steps**

- Learn more about the related dashboard panel in the [dashboards reference](./dashboards.md#zoekt-average-resolve-revision-duration).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_zoekt_average_resolve_revision_duration"
]
```

<sub>*Managed by the [Sourcegraph Search Core team](https://handbook.sourcegraph.com/departments/engineering/teams/search/core).*</sub>

<br />

## zoekt: get_index_options_error_increase

<p class="subtitle">the number of repositories we failed to get indexing options over 5m</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> zoekt: 100+ the number of repositories we failed to get indexing options over 5m for 5m0s
- <span class="badge badge-critical">critical</span> zoekt: 100+ the number of repositories we failed to get indexing options over 5m for 35m0s

**Next steps**

- View error rates on gitserver and frontend to identify root cause.
- Rollback frontend/gitserver deployment if due to a bad code change.
- View error logs for `getIndexOptions` via net/trace debug interface. For example click on a `indexed-search-indexer-` on https://sourcegraph.com/-/debug/. Then click on Traces. Replace sourcegraph.com with your instance address.
- More help interpreting this metric is available in the [dashboards reference](./dashboards.md#zoekt-get-index-options-error-increase).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_zoekt_get_index_options_error_increase",
  "critical_zoekt_get_index_options_error_increase"
]
```

<sub>*Managed by the [Sourcegraph Search Core team](https://handbook.sourcegraph.com/departments/engineering/teams/search/core).*</sub>

<br />

## zoekt: indexed_search_request_errors

<p class="subtitle">indexed search request errors every 5m by code</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> zoekt: 5%+ indexed search request errors every 5m by code for 5m0s

**Next steps**

- Learn more about the related dashboard panel in the [dashboards reference](./dashboards.md#zoekt-indexed-search-request-errors).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_zoekt_indexed_search_request_errors"
]
```

<sub>*Managed by the [Sourcegraph Search Core team](https://handbook.sourcegraph.com/departments/engineering/teams/search/core).*</sub>

<br />

## zoekt: container_cpu_usage

<p class="subtitle">container cpu usage total (1m average) across all cores by instance</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> zoekt: 99%+ container cpu usage total (1m average) across all cores by instance

**Next steps**

- **Kubernetes:** Consider increasing CPU limits in the the relevant `Deployment.yaml`.
- **Docker Compose:** Consider increasing `cpus:` of the zoekt-indexserver container in `docker-compose.yml`.
- Learn more about the related dashboard panel in the [dashboards reference](./dashboards.md#zoekt-container-cpu-usage).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_zoekt_container_cpu_usage"
]
```

<sub>*Managed by the [Sourcegraph Search Core team](https://handbook.sourcegraph.com/departments/engineering/teams/search/core).*</sub>

<br />

## zoekt: container_memory_usage

<p class="subtitle">container memory usage by instance</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> zoekt: 99%+ container memory usage by instance

**Next steps**

- **Kubernetes:** Consider increasing memory limit in relevant `Deployment.yaml`.
- **Docker Compose:** Consider increasing `memory:` of zoekt-indexserver container in `docker-compose.yml`.
- Learn more about the related dashboard panel in the [dashboards reference](./dashboards.md#zoekt-container-memory-usage).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_zoekt_container_memory_usage"
]
```

<sub>*Managed by the [Sourcegraph Search Core team](https://handbook.sourcegraph.com/departments/engineering/teams/search/core).*</sub>

<br />

## zoekt: container_cpu_usage

<p class="subtitle">container cpu usage total (1m average) across all cores by instance</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> zoekt: 99%+ container cpu usage total (1m average) across all cores by instance

**Next steps**

- **Kubernetes:** Consider increasing CPU limits in the the relevant `Deployment.yaml`.
- **Docker Compose:** Consider increasing `cpus:` of the zoekt-webserver container in `docker-compose.yml`.
- Learn more about the related dashboard panel in the [dashboards reference](./dashboards.md#zoekt-container-cpu-usage).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_zoekt_container_cpu_usage"
]
```

<sub>*Managed by the [Sourcegraph Search Core team](https://handbook.sourcegraph.com/departments/engineering/teams/search/core).*</sub>

<br />

## zoekt: container_memory_usage

<p class="subtitle">container memory usage by instance</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> zoekt: 99%+ container memory usage by instance

**Next steps**

- **Kubernetes:** Consider increasing memory limit in relevant `Deployment.yaml`.
- **Docker Compose:** Consider increasing `memory:` of zoekt-webserver container in `docker-compose.yml`.
- Learn more about the related dashboard panel in the [dashboards reference](./dashboards.md#zoekt-container-memory-usage).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_zoekt_container_memory_usage"
]
```

<sub>*Managed by the [Sourcegraph Search Core team](https://handbook.sourcegraph.com/departments/engineering/teams/search/core).*</sub>

<br />

## zoekt: provisioning_container_cpu_usage_long_term

<p class="subtitle">container cpu usage total (90th percentile over 1d) across all cores by instance</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> zoekt: 80%+ container cpu usage total (90th percentile over 1d) across all cores by instance for 336h0m0s

**Next steps**

- **Kubernetes:** Consider increasing CPU limits in the `Deployment.yaml` for the zoekt-indexserver service.
- **Docker Compose:** Consider increasing `cpus:` of the zoekt-indexserver container in `docker-compose.yml`.
- Learn more about the related dashboard panel in the [dashboards reference](./dashboards.md#zoekt-provisioning-container-cpu-usage-long-term).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_zoekt_provisioning_container_cpu_usage_long_term"
]
```

<sub>*Managed by the [Sourcegraph Search Core team](https://handbook.sourcegraph.com/departments/engineering/teams/search/core).*</sub>

<br />

## zoekt: provisioning_container_memory_usage_long_term

<p class="subtitle">container memory usage (1d maximum) by instance</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> zoekt: 80%+ container memory usage (1d maximum) by instance for 336h0m0s

**Next steps**

- **Kubernetes:** Consider increasing memory limits in the `Deployment.yaml` for the zoekt-indexserver service.
- **Docker Compose:** Consider increasing `memory:` of the zoekt-indexserver container in `docker-compose.yml`.
- Learn more about the related dashboard panel in the [dashboards reference](./dashboards.md#zoekt-provisioning-container-memory-usage-long-term).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_zoekt_provisioning_container_memory_usage_long_term"
]
```

<sub>*Managed by the [Sourcegraph Search Core team](https://handbook.sourcegraph.com/departments/engineering/teams/search/core).*</sub>

<br />

## zoekt: provisioning_container_cpu_usage_short_term

<p class="subtitle">container cpu usage total (5m maximum) across all cores by instance</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> zoekt: 90%+ container cpu usage total (5m maximum) across all cores by instance for 30m0s

**Next steps**

- **Kubernetes:** Consider increasing CPU limits in the the relevant `Deployment.yaml`.
- **Docker Compose:** Consider increasing `cpus:` of the zoekt-indexserver container in `docker-compose.yml`.
- Learn more about the related dashboard panel in the [dashboards reference](./dashboards.md#zoekt-provisioning-container-cpu-usage-short-term).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_zoekt_provisioning_container_cpu_usage_short_term"
]
```

<sub>*Managed by the [Sourcegraph Search Core team](https://handbook.sourcegraph.com/departments/engineering/teams/search/core).*</sub>

<br />

## zoekt: provisioning_container_memory_usage_short_term

<p class="subtitle">container memory usage (5m maximum) by instance</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> zoekt: 90%+ container memory usage (5m maximum) by instance

**Next steps**

- **Kubernetes:** Consider increasing memory limit in relevant `Deployment.yaml`.
- **Docker Compose:** Consider increasing `memory:` of zoekt-indexserver container in `docker-compose.yml`.
- Learn more about the related dashboard panel in the [dashboards reference](./dashboards.md#zoekt-provisioning-container-memory-usage-short-term).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_zoekt_provisioning_container_memory_usage_short_term"
]
```

<sub>*Managed by the [Sourcegraph Search Core team](https://handbook.sourcegraph.com/departments/engineering/teams/search/core).*</sub>

<br />

## zoekt: container_oomkill_events_total

<p class="subtitle">container OOMKILL events total by instance</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> zoekt: 1+ container OOMKILL events total by instance

**Next steps**

- **Kubernetes:** Consider increasing memory limit in relevant `Deployment.yaml`.
- **Docker Compose:** Consider increasing `memory:` of zoekt-indexserver container in `docker-compose.yml`.
- More help interpreting this metric is available in the [dashboards reference](./dashboards.md#zoekt-container-oomkill-events-total).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_zoekt_container_oomkill_events_total"
]
```

<sub>*Managed by the [Sourcegraph Search Core team](https://handbook.sourcegraph.com/departments/engineering/teams/search/core).*</sub>

<br />

## zoekt: provisioning_container_cpu_usage_long_term

<p class="subtitle">container cpu usage total (90th percentile over 1d) across all cores by instance</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> zoekt: 80%+ container cpu usage total (90th percentile over 1d) across all cores by instance for 336h0m0s

**Next steps**

- **Kubernetes:** Consider increasing CPU limits in the `Deployment.yaml` for the zoekt-webserver service.
- **Docker Compose:** Consider increasing `cpus:` of the zoekt-webserver container in `docker-compose.yml`.
- Learn more about the related dashboard panel in the [dashboards reference](./dashboards.md#zoekt-provisioning-container-cpu-usage-long-term).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_zoekt_provisioning_container_cpu_usage_long_term"
]
```

<sub>*Managed by the [Sourcegraph Search Core team](https://handbook.sourcegraph.com/departments/engineering/teams/search/core).*</sub>

<br />

## zoekt: provisioning_container_memory_usage_long_term

<p class="subtitle">container memory usage (1d maximum) by instance</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> zoekt: 80%+ container memory usage (1d maximum) by instance for 336h0m0s

**Next steps**

- **Kubernetes:** Consider increasing memory limits in the `Deployment.yaml` for the zoekt-webserver service.
- **Docker Compose:** Consider increasing `memory:` of the zoekt-webserver container in `docker-compose.yml`.
- Learn more about the related dashboard panel in the [dashboards reference](./dashboards.md#zoekt-provisioning-container-memory-usage-long-term).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_zoekt_provisioning_container_memory_usage_long_term"
]
```

<sub>*Managed by the [Sourcegraph Search Core team](https://handbook.sourcegraph.com/departments/engineering/teams/search/core).*</sub>

<br />

## zoekt: provisioning_container_cpu_usage_short_term

<p class="subtitle">container cpu usage total (5m maximum) across all cores by instance</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> zoekt: 90%+ container cpu usage total (5m maximum) across all cores by instance for 30m0s

**Next steps**

- **Kubernetes:** Consider increasing CPU limits in the the relevant `Deployment.yaml`.
- **Docker Compose:** Consider increasing `cpus:` of the zoekt-webserver container in `docker-compose.yml`.
- Learn more about the related dashboard panel in the [dashboards reference](./dashboards.md#zoekt-provisioning-container-cpu-usage-short-term).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_zoekt_provisioning_container_cpu_usage_short_term"
]
```

<sub>*Managed by the [Sourcegraph Search Core team](https://handbook.sourcegraph.com/departments/engineering/teams/search/core).*</sub>

<br />

## zoekt: provisioning_container_memory_usage_short_term

<p class="subtitle">container memory usage (5m maximum) by instance</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> zoekt: 90%+ container memory usage (5m maximum) by instance

**Next steps**

- **Kubernetes:** Consider increasing memory limit in relevant `Deployment.yaml`.
- **Docker Compose:** Consider increasing `memory:` of zoekt-webserver container in `docker-compose.yml`.
- Learn more about the related dashboard panel in the [dashboards reference](./dashboards.md#zoekt-provisioning-container-memory-usage-short-term).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_zoekt_provisioning_container_memory_usage_short_term"
]
```

<sub>*Managed by the [Sourcegraph Search Core team](https://handbook.sourcegraph.com/departments/engineering/teams/search/core).*</sub>

<br />

## zoekt: container_oomkill_events_total

<p class="subtitle">container OOMKILL events total by instance</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> zoekt: 1+ container OOMKILL events total by instance

**Next steps**

- **Kubernetes:** Consider increasing memory limit in relevant `Deployment.yaml`.
- **Docker Compose:** Consider increasing `memory:` of zoekt-webserver container in `docker-compose.yml`.
- More help interpreting this metric is available in the [dashboards reference](./dashboards.md#zoekt-container-oomkill-events-total).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_zoekt_container_oomkill_events_total"
]
```

<sub>*Managed by the [Sourcegraph Search Core team](https://handbook.sourcegraph.com/departments/engineering/teams/search/core).*</sub>

<br />

## zoekt: pods_available_percentage

<p class="subtitle">percentage pods available</p>

**Descriptions**

- <span class="badge badge-critical">critical</span> zoekt: less than 90% percentage pods available for 10m0s

**Next steps**

- Determine if the pod was OOM killed using `kubectl describe pod indexed-search` (look for `OOMKilled: true`) and, if so, consider increasing the memory limit in the relevant `Deployment.yaml`.
- Check the logs before the container restarted to see if there are `panic:` messages or similar using `kubectl logs -p indexed-search`.
- Learn more about the related dashboard panel in the [dashboards reference](./dashboards.md#zoekt-pods-available-percentage).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "critical_zoekt_pods_available_percentage"
]
```

<sub>*Managed by the [Sourcegraph Search Core team](https://handbook.sourcegraph.com/departments/engineering/teams/search/core).*</sub>

<br />

## prometheus: prometheus_rule_eval_duration

<p class="subtitle">average prometheus rule group evaluation duration over 10m by rule group</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> prometheus: 30s+ average prometheus rule group evaluation duration over 10m by rule group

**Next steps**

- Check the Container monitoring (not available on server) panels and try increasing resources for Prometheus if necessary.
- If the rule group taking a long time to evaluate belongs to `/sg_prometheus_addons`, try reducing the complexity of any custom Prometheus rules provided.
- If the rule group taking a long time to evaluate belongs to `/sg_config_prometheus`, please [open an issue](https://github.com/sourcegraph/sourcegraph/issues/new?assignees=&labels=&template=bug_report.md&title=).
- More help interpreting this metric is available in the [dashboards reference](./dashboards.md#prometheus-prometheus-rule-eval-duration).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_prometheus_prometheus_rule_eval_duration"
]
```

<sub>*Managed by the [Sourcegraph Cloud DevOps team](https://handbook.sourcegraph.com/departments/engineering/teams/devops).*</sub>

<br />

## prometheus: prometheus_rule_eval_failures

<p class="subtitle">failed prometheus rule evaluations over 5m by rule group</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> prometheus: 0+ failed prometheus rule evaluations over 5m by rule group

**Next steps**

- Check Prometheus logs for messages related to rule group evaluation (generally with log field `component="rule manager"`).
- If the rule group failing to evaluate belongs to `/sg_prometheus_addons`, ensure any custom Prometheus configuration provided is valid.
- If the rule group taking a long time to evaluate belongs to `/sg_config_prometheus`, please [open an issue](https://github.com/sourcegraph/sourcegraph/issues/new?assignees=&labels=&template=bug_report.md&title=).
- More help interpreting this metric is available in the [dashboards reference](./dashboards.md#prometheus-prometheus-rule-eval-failures).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_prometheus_prometheus_rule_eval_failures"
]
```

<sub>*Managed by the [Sourcegraph Cloud DevOps team](https://handbook.sourcegraph.com/departments/engineering/teams/devops).*</sub>

<br />

## prometheus: alertmanager_notification_latency

<p class="subtitle">alertmanager notification latency over 1m by integration</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> prometheus: 1s+ alertmanager notification latency over 1m by integration

**Next steps**

- Check the Container monitoring (not available on server) panels and try increasing resources for Prometheus if necessary.
- Ensure that your [`observability.alerts` configuration](https://docs.sourcegraph.com/admin/observability/alerting#setting-up-alerting) (in site configuration) is valid.
- Check if the relevant alert integration service is experiencing downtime or issues.
- Learn more about the related dashboard panel in the [dashboards reference](./dashboards.md#prometheus-alertmanager-notification-latency).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_prometheus_alertmanager_notification_latency"
]
```

<sub>*Managed by the [Sourcegraph Cloud DevOps team](https://handbook.sourcegraph.com/departments/engineering/teams/devops).*</sub>

<br />

## prometheus: alertmanager_notification_failures

<p class="subtitle">failed alertmanager notifications over 1m by integration</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> prometheus: 0+ failed alertmanager notifications over 1m by integration

**Next steps**

- Ensure that your [`observability.alerts` configuration](https://docs.sourcegraph.com/admin/observability/alerting#setting-up-alerting) (in site configuration) is valid.
- Check if the relevant alert integration service is experiencing downtime or issues.
- Learn more about the related dashboard panel in the [dashboards reference](./dashboards.md#prometheus-alertmanager-notification-failures).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_prometheus_alertmanager_notification_failures"
]
```

<sub>*Managed by the [Sourcegraph Cloud DevOps team](https://handbook.sourcegraph.com/departments/engineering/teams/devops).*</sub>

<br />

## prometheus: prometheus_config_status

<p class="subtitle">prometheus configuration reload status</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> prometheus: less than 1 prometheus configuration reload status

**Next steps**

- Check Prometheus logs for messages related to configuration loading.
- Ensure any [custom configuration you have provided Prometheus](https://docs.sourcegraph.com/admin/observability/metrics#prometheus-configuration) is valid.
- More help interpreting this metric is available in the [dashboards reference](./dashboards.md#prometheus-prometheus-config-status).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_prometheus_prometheus_config_status"
]
```

<sub>*Managed by the [Sourcegraph Cloud DevOps team](https://handbook.sourcegraph.com/departments/engineering/teams/devops).*</sub>

<br />

## prometheus: alertmanager_config_status

<p class="subtitle">alertmanager configuration reload status</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> prometheus: less than 1 alertmanager configuration reload status

**Next steps**

- Ensure that your [`observability.alerts` configuration](https://docs.sourcegraph.com/admin/observability/alerting#setting-up-alerting) (in site configuration) is valid.
- More help interpreting this metric is available in the [dashboards reference](./dashboards.md#prometheus-alertmanager-config-status).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_prometheus_alertmanager_config_status"
]
```

<sub>*Managed by the [Sourcegraph Cloud DevOps team](https://handbook.sourcegraph.com/departments/engineering/teams/devops).*</sub>

<br />

## prometheus: prometheus_tsdb_op_failure

<p class="subtitle">prometheus tsdb failures by operation over 1m by operation</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> prometheus: 0+ prometheus tsdb failures by operation over 1m by operation

**Next steps**

- Check Prometheus logs for messages related to the failing operation.
- Learn more about the related dashboard panel in the [dashboards reference](./dashboards.md#prometheus-prometheus-tsdb-op-failure).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_prometheus_prometheus_tsdb_op_failure"
]
```

<sub>*Managed by the [Sourcegraph Cloud DevOps team](https://handbook.sourcegraph.com/departments/engineering/teams/devops).*</sub>

<br />

## prometheus: prometheus_target_sample_exceeded

<p class="subtitle">prometheus scrapes that exceed the sample limit over 10m</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> prometheus: 0+ prometheus scrapes that exceed the sample limit over 10m

**Next steps**

- Check Prometheus logs for messages related to target scrape failures.
- Learn more about the related dashboard panel in the [dashboards reference](./dashboards.md#prometheus-prometheus-target-sample-exceeded).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_prometheus_prometheus_target_sample_exceeded"
]
```

<sub>*Managed by the [Sourcegraph Cloud DevOps team](https://handbook.sourcegraph.com/departments/engineering/teams/devops).*</sub>

<br />

## prometheus: prometheus_target_sample_duplicate

<p class="subtitle">prometheus scrapes rejected due to duplicate timestamps over 10m</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> prometheus: 0+ prometheus scrapes rejected due to duplicate timestamps over 10m

**Next steps**

- Check Prometheus logs for messages related to target scrape failures.
- Learn more about the related dashboard panel in the [dashboards reference](./dashboards.md#prometheus-prometheus-target-sample-duplicate).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_prometheus_prometheus_target_sample_duplicate"
]
```

<sub>*Managed by the [Sourcegraph Cloud DevOps team](https://handbook.sourcegraph.com/departments/engineering/teams/devops).*</sub>

<br />

## prometheus: container_cpu_usage

<p class="subtitle">container cpu usage total (1m average) across all cores by instance</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> prometheus: 99%+ container cpu usage total (1m average) across all cores by instance

**Next steps**

- **Kubernetes:** Consider increasing CPU limits in the the relevant `Deployment.yaml`.
- **Docker Compose:** Consider increasing `cpus:` of the prometheus container in `docker-compose.yml`.
- Learn more about the related dashboard panel in the [dashboards reference](./dashboards.md#prometheus-container-cpu-usage).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_prometheus_container_cpu_usage"
]
```

<sub>*Managed by the [Sourcegraph Cloud DevOps team](https://handbook.sourcegraph.com/departments/engineering/teams/devops).*</sub>

<br />

## prometheus: container_memory_usage

<p class="subtitle">container memory usage by instance</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> prometheus: 99%+ container memory usage by instance

**Next steps**

- **Kubernetes:** Consider increasing memory limit in relevant `Deployment.yaml`.
- **Docker Compose:** Consider increasing `memory:` of prometheus container in `docker-compose.yml`.
- Learn more about the related dashboard panel in the [dashboards reference](./dashboards.md#prometheus-container-memory-usage).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_prometheus_container_memory_usage"
]
```

<sub>*Managed by the [Sourcegraph Cloud DevOps team](https://handbook.sourcegraph.com/departments/engineering/teams/devops).*</sub>

<br />

## prometheus: provisioning_container_cpu_usage_long_term

<p class="subtitle">container cpu usage total (90th percentile over 1d) across all cores by instance</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> prometheus: 80%+ container cpu usage total (90th percentile over 1d) across all cores by instance for 336h0m0s

**Next steps**

- **Kubernetes:** Consider increasing CPU limits in the `Deployment.yaml` for the prometheus service.
- **Docker Compose:** Consider increasing `cpus:` of the prometheus container in `docker-compose.yml`.
- Learn more about the related dashboard panel in the [dashboards reference](./dashboards.md#prometheus-provisioning-container-cpu-usage-long-term).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_prometheus_provisioning_container_cpu_usage_long_term"
]
```

<sub>*Managed by the [Sourcegraph Cloud DevOps team](https://handbook.sourcegraph.com/departments/engineering/teams/devops).*</sub>

<br />

## prometheus: provisioning_container_memory_usage_long_term

<p class="subtitle">container memory usage (1d maximum) by instance</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> prometheus: 80%+ container memory usage (1d maximum) by instance for 336h0m0s

**Next steps**

- **Kubernetes:** Consider increasing memory limits in the `Deployment.yaml` for the prometheus service.
- **Docker Compose:** Consider increasing `memory:` of the prometheus container in `docker-compose.yml`.
- Learn more about the related dashboard panel in the [dashboards reference](./dashboards.md#prometheus-provisioning-container-memory-usage-long-term).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_prometheus_provisioning_container_memory_usage_long_term"
]
```

<sub>*Managed by the [Sourcegraph Cloud DevOps team](https://handbook.sourcegraph.com/departments/engineering/teams/devops).*</sub>

<br />

## prometheus: provisioning_container_cpu_usage_short_term

<p class="subtitle">container cpu usage total (5m maximum) across all cores by instance</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> prometheus: 90%+ container cpu usage total (5m maximum) across all cores by instance for 30m0s

**Next steps**

- **Kubernetes:** Consider increasing CPU limits in the the relevant `Deployment.yaml`.
- **Docker Compose:** Consider increasing `cpus:` of the prometheus container in `docker-compose.yml`.
- Learn more about the related dashboard panel in the [dashboards reference](./dashboards.md#prometheus-provisioning-container-cpu-usage-short-term).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_prometheus_provisioning_container_cpu_usage_short_term"
]
```

<sub>*Managed by the [Sourcegraph Cloud DevOps team](https://handbook.sourcegraph.com/departments/engineering/teams/devops).*</sub>

<br />

## prometheus: provisioning_container_memory_usage_short_term

<p class="subtitle">container memory usage (5m maximum) by instance</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> prometheus: 90%+ container memory usage (5m maximum) by instance

**Next steps**

- **Kubernetes:** Consider increasing memory limit in relevant `Deployment.yaml`.
- **Docker Compose:** Consider increasing `memory:` of prometheus container in `docker-compose.yml`.
- Learn more about the related dashboard panel in the [dashboards reference](./dashboards.md#prometheus-provisioning-container-memory-usage-short-term).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_prometheus_provisioning_container_memory_usage_short_term"
]
```

<sub>*Managed by the [Sourcegraph Cloud DevOps team](https://handbook.sourcegraph.com/departments/engineering/teams/devops).*</sub>

<br />

## prometheus: container_oomkill_events_total

<p class="subtitle">container OOMKILL events total by instance</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> prometheus: 1+ container OOMKILL events total by instance

**Next steps**

- **Kubernetes:** Consider increasing memory limit in relevant `Deployment.yaml`.
- **Docker Compose:** Consider increasing `memory:` of prometheus container in `docker-compose.yml`.
- More help interpreting this metric is available in the [dashboards reference](./dashboards.md#prometheus-container-oomkill-events-total).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_prometheus_container_oomkill_events_total"
]
```

<sub>*Managed by the [Sourcegraph Cloud DevOps team](https://handbook.sourcegraph.com/departments/engineering/teams/devops).*</sub>

<br />

## prometheus: pods_available_percentage

<p class="subtitle">percentage pods available</p>

**Descriptions**

- <span class="badge badge-critical">critical</span> prometheus: less than 90% percentage pods available for 10m0s

**Next steps**

- Determine if the pod was OOM killed using `kubectl describe pod prometheus` (look for `OOMKilled: true`) and, if so, consider increasing the memory limit in the relevant `Deployment.yaml`.
- Check the logs before the container restarted to see if there are `panic:` messages or similar using `kubectl logs -p prometheus`.
- Learn more about the related dashboard panel in the [dashboards reference](./dashboards.md#prometheus-pods-available-percentage).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "critical_prometheus_pods_available_percentage"
]
```

<sub>*Managed by the [Sourcegraph Cloud DevOps team](https://handbook.sourcegraph.com/departments/engineering/teams/devops).*</sub>

<br />

## executor: executor_handlers

<p class="subtitle">executor active handlers</p>

**Descriptions**

- <span class="badge badge-critical">critical</span> executor: 0 active executor handlers and > 0 queue size for 5m0s

<details>
<summary>Technical details</summary>

Custom alert query: `
		(sum(src_executor_processor_handlers{queue=~"${queue:regex}",sg_job=~"^sourcegraph-executors.*"}) OR vector(0)) == 0
			AND
		(sum by (queue)(src_executor_total{job=~"^sourcegraph-executors.*"})) > 0
	`

</details>

**Next steps**

- Check to see the state of any compute VMs, they may be taking longer than expected to boot.
- Make sure the executors appear under Site Admin > Executors.
- Check the Grafana dashboard section for APIClient, it should do frequent requests to Dequeue and Heartbeat and those must not fail.
- Learn more about the related dashboard panel in the [dashboards reference](./dashboards.md#executor-executor-handlers).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "critical_executor_executor_handlers"
]
```

<sub>*Managed by the [Sourcegraph Code intelligence team](https://handbook.sourcegraph.com/departments/engineering/teams/code-intelligence).*</sub>

<br />

## executor: executor_processor_error_rate

<p class="subtitle">executor operation error rate over 5m</p>

**Descriptions**

- <span class="badge badge-critical">critical</span> executor: 100%+ executor operation error rate over 5m for 1h0m0s

<details>
<summary>Technical details</summary>

Custom alert query: `last_over_time(sum(increase(src_executor_processor_errors_total{queue=~"${queue:regex}",sg_job=~"^sourcegraph-executors.*"}[5m]))[5h:]) / (last_over_time(sum(increase(src_executor_processor_total{queue=~"${queue:regex}",sg_job=~"^sourcegraph-executors.*"}[5m]))[5h:]) + last_over_time(sum(increase(src_executor_processor_errors_total{queue=~"${queue:regex}",sg_job=~"^sourcegraph-executors.*"}[5m]))[5h:])) * 100`

</details>

**Next steps**

- Determine the cause of failure from the auto-indexing job logs in the site-admin page.
- This alert fires if all executor jobs have been failing for the past hour. The alert will continue for up
to 5 hours until the error rate is no longer 100%, even if there are no running jobs in that time, as the
problem is not know to be resolved until jobs start succeeding again.
- Learn more about the related dashboard panel in the [dashboards reference](./dashboards.md#executor-executor-processor-error-rate).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "critical_executor_executor_processor_error_rate"
]
```

<sub>*Managed by the [Sourcegraph Code intelligence team](https://handbook.sourcegraph.com/departments/engineering/teams/code-intelligence).*</sub>

<br />

## executor: go_goroutines

<p class="subtitle">maximum active goroutines</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> executor: 10000+ maximum active goroutines for 10m0s

**Next steps**

- More help interpreting this metric is available in the [dashboards reference](./dashboards.md#executor-go-goroutines).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_executor_go_goroutines"
]
```

<sub>*Managed by the [Sourcegraph Code intelligence team](https://handbook.sourcegraph.com/departments/engineering/teams/code-intelligence).*</sub>

<br />

## executor: go_gc_duration_seconds

<p class="subtitle">maximum go garbage collection duration</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> executor: 2s+ maximum go garbage collection duration

**Next steps**

- Learn more about the related dashboard panel in the [dashboards reference](./dashboards.md#executor-go-gc-duration-seconds).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_executor_go_gc_duration_seconds"
]
```

<sub>*Managed by the [Sourcegraph Code intelligence team](https://handbook.sourcegraph.com/departments/engineering/teams/code-intelligence).*</sub>

<br />

## codeintel-uploads: codeintel_commit_graph_queued_max_age

<p class="subtitle">repository queue longest time in queue</p>

**Descriptions**

- <span class="badge badge-critical">critical</span> codeintel-uploads: 3600s+ repository queue longest time in queue

**Next steps**

- An alert here is generally indicative of either underprovisioned worker instance(s) and/or
an underprovisioned main postgres instance.
- Learn more about the related dashboard panel in the [dashboards reference](./dashboards.md#codeintel-uploads-codeintel-commit-graph-queued-max-age).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "critical_codeintel-uploads_codeintel_commit_graph_queued_max_age"
]
```

<sub>*Managed by the [Sourcegraph Code intelligence team](https://handbook.sourcegraph.com/departments/engineering/teams/code-intelligence).*</sub>

<br />

## telemetry: telemetry_job_error_rate

<p class="subtitle">usage data exporter operation error rate over 5m</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> telemetry: 0%+ usage data exporter operation error rate over 5m for 30m0s

**Next steps**

- Involved cloud team to inspect logs of the managed instance to determine error sources.
- Learn more about the related dashboard panel in the [dashboards reference](./dashboards.md#telemetry-telemetry-job-error-rate).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_telemetry_telemetry_job_error_rate"
]
```

<sub>*Managed by the [Sourcegraph Data & Analytics team](https://handbook.sourcegraph.com/departments/engineering/teams/data-analytics).*</sub>

<br />

## telemetry: telemetry_job_utilized_throughput

<p class="subtitle">utilized percentage of maximum throughput</p>

**Descriptions**

- <span class="badge badge-warning">warning</span> telemetry: 90%+ utilized percentage of maximum throughput for 30m0s

**Next steps**

- Throughput utilization is high. This could be a signal that this instance is producing too many events for the export job to keep up. Configure more throughput using the maxBatchSize option.
- Learn more about the related dashboard panel in the [dashboards reference](./dashboards.md#telemetry-telemetry-job-utilized-throughput).
- **Silence this alert:** If you are aware of this alert and want to silence notifications for it, add the following to your site configuration and set a reminder to re-evaluate the alert:

```json
"observability.silenceAlerts": [
  "warning_telemetry_telemetry_job_utilized_throughput"
]
```

<sub>*Managed by the [Sourcegraph Data & Analytics team](https://handbook.sourcegraph.com/departments/engineering/teams/data-analytics).*</sub>

<br />

