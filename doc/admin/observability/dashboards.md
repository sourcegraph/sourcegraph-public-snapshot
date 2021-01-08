# Dashboards reference

<!-- DO NOT EDIT: generated via: go generate ./monitoring -->

This document contains a complete reference on Sourcegraph's available dashboards, as well as details on how to interpret the panels and metrics.

To learn more about Sourcegraph's metrics and how to view these dashboards, see [our metrics guide](https://docs.sourcegraph.com/admin/observability/metrics).

## Frontend

<p class="subtitle">Serves all end-user browser and API requests.</p>

### Frontend: Search at a glance

#### frontend: 99th_percentile_search_request_duration

This search panel indicates 99th percentile successful search request duration over 5m.

Refer to the [alert solutions reference](./alert_solutions.md#frontend-99th-percentile-search-request-duration) for relevant alerts.

<br />

#### frontend: 90th_percentile_search_request_duration

This search panel indicates 90th percentile successful search request duration over 5m.

Refer to the [alert solutions reference](./alert_solutions.md#frontend-90th-percentile-search-request-duration) for relevant alerts.

<br />

#### frontend: hard_timeout_search_responses

This search panel indicates hard timeout search responses every 5m.

Refer to the [alert solutions reference](./alert_solutions.md#frontend-hard-timeout-search-responses) for relevant alerts.

<br />

#### frontend: hard_error_search_responses

This search panel indicates hard error search responses every 5m.

Refer to the [alert solutions reference](./alert_solutions.md#frontend-hard-error-search-responses) for relevant alerts.

<br />

#### frontend: partial_timeout_search_responses

This search panel indicates partial timeout search responses every 5m.

Refer to the [alert solutions reference](./alert_solutions.md#frontend-partial-timeout-search-responses) for relevant alerts.

<br />

#### frontend: search_alert_user_suggestions

This search panel indicates search alert user suggestions shown every 5m.

Refer to the [alert solutions reference](./alert_solutions.md#frontend-search-alert-user-suggestions) for relevant alerts.

<br />

#### frontend: page_load_latency

This cloud panel indicates 90th percentile page load latency over all routes over 10m.

Refer to the [alert solutions reference](./alert_solutions.md#frontend-page-load-latency) for relevant alerts.

<br />

#### frontend: blob_load_latency

This cloud panel indicates 90th percentile blob load latency over 10m.

Refer to the [alert solutions reference](./alert_solutions.md#frontend-blob-load-latency) for relevant alerts.

<br />

### Frontend: Search-based code intelligence at a glance

#### frontend: 99th_percentile_search_codeintel_request_duration

This code-intel panel indicates 99th percentile code-intel successful search request duration over 5m.

Refer to the [alert solutions reference](./alert_solutions.md#frontend-99th-percentile-search-codeintel-request-duration) for relevant alerts.

<br />

#### frontend: 90th_percentile_search_codeintel_request_duration

This code-intel panel indicates 90th percentile code-intel successful search request duration over 5m.

Refer to the [alert solutions reference](./alert_solutions.md#frontend-90th-percentile-search-codeintel-request-duration) for relevant alerts.

<br />

#### frontend: hard_timeout_search_codeintel_responses

This code-intel panel indicates hard timeout search code-intel responses every 5m.

Refer to the [alert solutions reference](./alert_solutions.md#frontend-hard-timeout-search-codeintel-responses) for relevant alerts.

<br />

#### frontend: hard_error_search_codeintel_responses

This code-intel panel indicates hard error search code-intel responses every 5m.

Refer to the [alert solutions reference](./alert_solutions.md#frontend-hard-error-search-codeintel-responses) for relevant alerts.

<br />

#### frontend: partial_timeout_search_codeintel_responses

This code-intel panel indicates partial timeout search code-intel responses every 5m.

Refer to the [alert solutions reference](./alert_solutions.md#frontend-partial-timeout-search-codeintel-responses) for relevant alerts.

<br />

#### frontend: search_codeintel_alert_user_suggestions

This code-intel panel indicates search code-intel alert user suggestions shown every 5m.

Refer to the [alert solutions reference](./alert_solutions.md#frontend-search-codeintel-alert-user-suggestions) for relevant alerts.

<br />

### Frontend: Search API usage at a glance

#### frontend: 99th_percentile_search_api_request_duration

This search panel indicates 99th percentile successful search API request duration over 5m.

Refer to the [alert solutions reference](./alert_solutions.md#frontend-99th-percentile-search-api-request-duration) for relevant alerts.

<br />

#### frontend: 90th_percentile_search_api_request_duration

This search panel indicates 90th percentile successful search API request duration over 5m.

Refer to the [alert solutions reference](./alert_solutions.md#frontend-90th-percentile-search-api-request-duration) for relevant alerts.

<br />

#### frontend: hard_timeout_search_api_responses

This search panel indicates hard timeout search API responses every 5m.

Refer to the [alert solutions reference](./alert_solutions.md#frontend-hard-timeout-search-api-responses) for relevant alerts.

<br />

#### frontend: hard_error_search_api_responses

This search panel indicates hard error search API responses every 5m.

Refer to the [alert solutions reference](./alert_solutions.md#frontend-hard-error-search-api-responses) for relevant alerts.

<br />

#### frontend: partial_timeout_search_api_responses

This search panel indicates partial timeout search API responses every 5m.

Refer to the [alert solutions reference](./alert_solutions.md#frontend-partial-timeout-search-api-responses) for relevant alerts.

<br />

#### frontend: search_api_alert_user_suggestions

This search panel indicates search API alert user suggestions shown every 5m.

Refer to the [alert solutions reference](./alert_solutions.md#frontend-search-api-alert-user-suggestions) for relevant alerts.

<br />

### Frontend: Precise code intelligence usage at a glance

#### frontend: codeintel_resolvers_99th_percentile_duration

This code-intel panel indicates 99th percentile successful resolver duration over 5m.

Refer to the [alert solutions reference](./alert_solutions.md#frontend-codeintel-resolvers-99th-percentile-duration) for relevant alerts.

<br />

#### frontend: codeintel_resolvers_errors

This code-intel panel indicates resolver errors every 5m.

Refer to the [alert solutions reference](./alert_solutions.md#frontend-codeintel-resolvers-errors) for relevant alerts.

<br />

#### frontend: codeintel_api_99th_percentile_duration

This code-intel panel indicates 99th percentile successful codeintel API operation duration over 5m.

Refer to the [alert solutions reference](./alert_solutions.md#frontend-codeintel-api-99th-percentile-duration) for relevant alerts.

<br />

#### frontend: codeintel_api_errors

This code-intel panel indicates code intel API errors every 5m.

Refer to the [alert solutions reference](./alert_solutions.md#frontend-codeintel-api-errors) for relevant alerts.

<br />

### Frontend: Precise code intelligence stores and clients

#### frontend: codeintel_dbstore_99th_percentile_duration

This code-intel panel indicates 99th percentile successful database store operation duration over 5m.

Refer to the [alert solutions reference](./alert_solutions.md#frontend-codeintel-dbstore-99th-percentile-duration) for relevant alerts.

<br />

#### frontend: codeintel_dbstore_errors

This code-intel panel indicates database store errors every 5m.

Refer to the [alert solutions reference](./alert_solutions.md#frontend-codeintel-dbstore-errors) for relevant alerts.

<br />

#### frontend: codeintel_upload_workerstore_99th_percentile_duration

This code-intel panel indicates 99th percentile successful upload worker store operation duration over 5m.

Refer to the [alert solutions reference](./alert_solutions.md#frontend-codeintel-upload-workerstore-99th-percentile-duration) for relevant alerts.

<br />

#### frontend: codeintel_upload_workerstore_errors

This code-intel panel indicates upload worker store errors every 5m.

Refer to the [alert solutions reference](./alert_solutions.md#frontend-codeintel-upload-workerstore-errors) for relevant alerts.

<br />

#### frontend: codeintel_index_workerstore_99th_percentile_duration

This code-intel panel indicates 99th percentile successful index worker store operation duration over 5m.

Refer to the [alert solutions reference](./alert_solutions.md#frontend-codeintel-index-workerstore-99th-percentile-duration) for relevant alerts.

<br />

#### frontend: codeintel_index_workerstore_errors

This code-intel panel indicates index worker store errors every 5m.

Refer to the [alert solutions reference](./alert_solutions.md#frontend-codeintel-index-workerstore-errors) for relevant alerts.

<br />

#### frontend: codeintel_lsifstore_99th_percentile_duration

This code-intel panel indicates 99th percentile successful LSIF store operation duration over 5m.

Refer to the [alert solutions reference](./alert_solutions.md#frontend-codeintel-lsifstore-99th-percentile-duration) for relevant alerts.

<br />

#### frontend: codeintel_lsifstore_errors

This code-intel panel indicates lSIF store errors every 5m.

Refer to the [alert solutions reference](./alert_solutions.md#frontend-codeintel-lsifstore-errors) for relevant alerts.

<br />

#### frontend: codeintel_uploadstore_99th_percentile_duration

This code-intel panel indicates 99th percentile successful upload store operation duration over 5m.

Refer to the [alert solutions reference](./alert_solutions.md#frontend-codeintel-uploadstore-99th-percentile-duration) for relevant alerts.

<br />

#### frontend: codeintel_uploadstore_errors

This code-intel panel indicates upload store errors every 5m.

Refer to the [alert solutions reference](./alert_solutions.md#frontend-codeintel-uploadstore-errors) for relevant alerts.

<br />

#### frontend: codeintel_gitserverclient_99th_percentile_duration

This code-intel panel indicates 99th percentile successful gitserver client operation duration over 5m.

Refer to the [alert solutions reference](./alert_solutions.md#frontend-codeintel-gitserverclient-99th-percentile-duration) for relevant alerts.

<br />

#### frontend: codeintel_gitserverclient_errors

This code-intel panel indicates gitserver client errors every 5m.

Refer to the [alert solutions reference](./alert_solutions.md#frontend-codeintel-gitserverclient-errors) for relevant alerts.

<br />

### Frontend: Precise code intelligence commit graph updater

#### frontend: codeintel_commit_graph_queue_size

This code-intel panel indicates commit graph queue size.

Refer to the [alert solutions reference](./alert_solutions.md#frontend-codeintel-commit-graph-queue-size) for relevant alerts.

<br />

#### frontend: codeintel_commit_graph_queue_growth_rate

This code-intel panel indicates commit graph queue growth rate over 30m.

Refer to the [alert solutions reference](./alert_solutions.md#frontend-codeintel-commit-graph-queue-growth-rate) for relevant alerts.

<br />

#### frontend: codeintel_commit_graph_updater_99th_percentile_duration

This code-intel panel indicates 99th percentile successful commit graph updater operation duration over 5m.

Refer to the [alert solutions reference](./alert_solutions.md#frontend-codeintel-commit-graph-updater-99th-percentile-duration) for relevant alerts.

<br />

#### frontend: codeintel_commit_graph_updater_errors

This code-intel panel indicates commit graph updater errors every 5m.

Refer to the [alert solutions reference](./alert_solutions.md#frontend-codeintel-commit-graph-updater-errors) for relevant alerts.

<br />

### Frontend: Precise code intelligence janitor

#### frontend: codeintel_janitor_errors

This code-intel panel indicates janitor errors every 5m.

Refer to the [alert solutions reference](./alert_solutions.md#frontend-codeintel-janitor-errors) for relevant alerts.

<br />

#### frontend: codeintel_upload_records_removed

This code-intel panel indicates upload records expired or deleted every 5m.


<br />

#### frontend: codeintel_index_records_removed

This code-intel panel indicates index records expired or deleted every 5m.


<br />

#### frontend: codeintel_lsif_data_removed

This code-intel panel indicates data for unreferenced upload records removed every 5m.


<br />

#### frontend: codeintel_background_upload_resets

This code-intel panel indicates upload records re-queued (due to unresponsive worker) every 5m.

Refer to the [alert solutions reference](./alert_solutions.md#frontend-codeintel-background-upload-resets) for relevant alerts.

<br />

#### frontend: codeintel_background_upload_reset_failures

This code-intel panel indicates upload records errored due to repeated reset every 5m.

Refer to the [alert solutions reference](./alert_solutions.md#frontend-codeintel-background-upload-reset-failures) for relevant alerts.

<br />

#### frontend: codeintel_background_index_resets

This code-intel panel indicates index records re-queued (due to unresponsive indexer) every 5m.

Refer to the [alert solutions reference](./alert_solutions.md#frontend-codeintel-background-index-resets) for relevant alerts.

<br />

#### frontend: codeintel_background_index_reset_failures

This code-intel panel indicates index records errored due to repeated reset every 5m.

Refer to the [alert solutions reference](./alert_solutions.md#frontend-codeintel-background-index-reset-failures) for relevant alerts.

<br />

### Frontend: Auto-indexing

#### frontend: codeintel_indexing_99th_percentile_duration

This code-intel panel indicates 99th percentile successful indexing operation duration over 5m.


<br />

#### frontend: codeintel_indexing_errors

This code-intel panel indicates indexing errors every 5m.

Refer to the [alert solutions reference](./alert_solutions.md#frontend-codeintel-indexing-errors) for relevant alerts.

<br />

### Frontend: Internal service requests

#### frontend: internal_indexed_search_error_responses

This search panel indicates internal indexed search error responses every 5m.

Refer to the [alert solutions reference](./alert_solutions.md#frontend-internal-indexed-search-error-responses) for relevant alerts.

<br />

#### frontend: internal_unindexed_search_error_responses

This search panel indicates internal unindexed search error responses every 5m.

Refer to the [alert solutions reference](./alert_solutions.md#frontend-internal-unindexed-search-error-responses) for relevant alerts.

<br />

#### frontend: internal_api_error_responses

This cloud panel indicates internal API error responses every 5m by route.

Refer to the [alert solutions reference](./alert_solutions.md#frontend-internal-api-error-responses) for relevant alerts.

<br />

#### frontend: 99th_percentile_gitserver_duration

This cloud panel indicates 99th percentile successful gitserver query duration over 5m.

Refer to the [alert solutions reference](./alert_solutions.md#frontend-99th-percentile-gitserver-duration) for relevant alerts.

<br />

#### frontend: gitserver_error_responses

This cloud panel indicates gitserver error responses every 5m.

Refer to the [alert solutions reference](./alert_solutions.md#frontend-gitserver-error-responses) for relevant alerts.

<br />

#### frontend: observability_test_alert_warning

This distribution panel indicates warning test alert metric.

Refer to the [alert solutions reference](./alert_solutions.md#frontend-observability-test-alert-warning) for relevant alerts.

<br />

#### frontend: observability_test_alert_critical

This distribution panel indicates critical test alert metric.

Refer to the [alert solutions reference](./alert_solutions.md#frontend-observability-test-alert-critical) for relevant alerts.

<br />

### Frontend: Container monitoring (not available on server)

#### frontend: container_cpu_usage

This cloud panel indicates container cpu usage total (1m average) across all cores by instance.

Refer to the [alert solutions reference](./alert_solutions.md#frontend-container-cpu-usage) for relevant alerts.

<br />

#### frontend: container_memory_usage

This cloud panel indicates container memory usage by instance.

Refer to the [alert solutions reference](./alert_solutions.md#frontend-container-memory-usage) for relevant alerts.

<br />

#### frontend: container_restarts

This cloud panel indicates container restarts every 5m by instance.

Refer to the [alert solutions reference](./alert_solutions.md#frontend-container-restarts) for relevant alerts.

<br />

#### frontend: fs_inodes_used

This cloud panel indicates fs inodes in use by instance.

Refer to the [alert solutions reference](./alert_solutions.md#frontend-fs-inodes-used) for relevant alerts.

<br />

### Frontend: Provisioning indicators (not available on server)

#### frontend: provisioning_container_cpu_usage_long_term

This cloud panel indicates container cpu usage total (90th percentile over 1d) across all cores by instance.

Refer to the [alert solutions reference](./alert_solutions.md#frontend-provisioning-container-cpu-usage-long-term) for relevant alerts.

<br />

#### frontend: provisioning_container_memory_usage_long_term

This cloud panel indicates container memory usage (1d maximum) by instance.

Refer to the [alert solutions reference](./alert_solutions.md#frontend-provisioning-container-memory-usage-long-term) for relevant alerts.

<br />

#### frontend: provisioning_container_cpu_usage_short_term

This cloud panel indicates container cpu usage total (5m maximum) across all cores by instance.

Refer to the [alert solutions reference](./alert_solutions.md#frontend-provisioning-container-cpu-usage-short-term) for relevant alerts.

<br />

#### frontend: provisioning_container_memory_usage_short_term

This cloud panel indicates container memory usage (5m maximum) by instance.

Refer to the [alert solutions reference](./alert_solutions.md#frontend-provisioning-container-memory-usage-short-term) for relevant alerts.

<br />

### Frontend: Golang runtime monitoring

#### frontend: go_goroutines

This cloud panel indicates maximum active goroutines.

Refer to the [alert solutions reference](./alert_solutions.md#frontend-go-goroutines) for relevant alerts.

<br />

#### frontend: go_gc_duration_seconds

This cloud panel indicates maximum go garbage collection duration.

Refer to the [alert solutions reference](./alert_solutions.md#frontend-go-gc-duration-seconds) for relevant alerts.

<br />

### Frontend: Kubernetes monitoring (ignore if using Docker Compose or server)

#### frontend: pods_available_percentage

This cloud panel indicates percentage pods available.

Refer to the [alert solutions reference](./alert_solutions.md#frontend-pods-available-percentage) for relevant alerts.

<br />

## Git Server

<p class="subtitle">Stores, manages, and operates Git repositories.</p>

#### gitserver: disk_space_remaining

This cloud panel indicates disk space remaining by instance.

Refer to the [alert solutions reference](./alert_solutions.md#gitserver-disk-space-remaining) for relevant alerts.

<br />

#### gitserver: running_git_commands

This cloud panel indicates running git commands.

A high value signals load.

Refer to the [alert solutions reference](./alert_solutions.md#gitserver-running-git-commands) for relevant alerts.

<br />

#### gitserver: repository_clone_queue_size

This cloud panel indicates repository clone queue size.

Refer to the [alert solutions reference](./alert_solutions.md#gitserver-repository-clone-queue-size) for relevant alerts.

<br />

#### gitserver: repository_existence_check_queue_size

This cloud panel indicates repository existence check queue size.

Refer to the [alert solutions reference](./alert_solutions.md#gitserver-repository-existence-check-queue-size) for relevant alerts.

<br />

#### gitserver: echo_command_duration_test

This cloud panel indicates echo test command duration.

A high value here likely indicates a problem, especially if consistently high.
You can query for individual commands using `sum by (cmd)(src_gitserver_exec_running)` in Grafana (`/-/debug/grafana`) to see if a specific Git Server command might be spiking in frequency.

If this value is consistently high, consider the following:

- **Single container deployments:** Upgrade to a [Docker Compose deployment](../install/docker-compose/migrate.md) which offers better scalability and resource isolation.
- **Kubernetes and Docker Compose:** Check that you are running a similar number of git server replicas and that their CPU/memory limits are allocated according to what is shown in the [Sourcegraph resource estimator](../install/resource_estimator.md).


<br />

#### gitserver: frontend_internal_api_error_responses

This cloud panel indicates frontend-internal API error responses every 5m by route.

Refer to the [alert solutions reference](./alert_solutions.md#gitserver-frontend-internal-api-error-responses) for relevant alerts.

<br />

### Git Server: Container monitoring (not available on server)

#### gitserver: container_cpu_usage

This cloud panel indicates container cpu usage total (1m average) across all cores by instance.

Refer to the [alert solutions reference](./alert_solutions.md#gitserver-container-cpu-usage) for relevant alerts.

<br />

#### gitserver: container_memory_usage

This cloud panel indicates container memory usage by instance.

Refer to the [alert solutions reference](./alert_solutions.md#gitserver-container-memory-usage) for relevant alerts.

<br />

#### gitserver: container_restarts

This cloud panel indicates container restarts every 5m by instance.

Refer to the [alert solutions reference](./alert_solutions.md#gitserver-container-restarts) for relevant alerts.

<br />

#### gitserver: fs_inodes_used

This cloud panel indicates fs inodes in use by instance.

Refer to the [alert solutions reference](./alert_solutions.md#gitserver-fs-inodes-used) for relevant alerts.

<br />

#### gitserver: fs_io_operations

This cloud panel indicates filesystem reads and writes rate by instance over 1h.

Refer to the [alert solutions reference](./alert_solutions.md#gitserver-fs-io-operations) for relevant alerts.

<br />

### Git Server: Provisioning indicators (not available on server)

#### gitserver: provisioning_container_cpu_usage_long_term

This cloud panel indicates container cpu usage total (90th percentile over 1d) across all cores by instance.

Refer to the [alert solutions reference](./alert_solutions.md#gitserver-provisioning-container-cpu-usage-long-term) for relevant alerts.

<br />

#### gitserver: provisioning_container_memory_usage_long_term

This cloud panel indicates container memory usage (1d maximum) by instance.

Git Server is expected to use up all the memory it is provided.


<br />

#### gitserver: provisioning_container_cpu_usage_short_term

This cloud panel indicates container cpu usage total (5m maximum) across all cores by instance.

Refer to the [alert solutions reference](./alert_solutions.md#gitserver-provisioning-container-cpu-usage-short-term) for relevant alerts.

<br />

#### gitserver: provisioning_container_memory_usage_short_term

This cloud panel indicates container memory usage (5m maximum) by instance.

Git Server is expected to use up all the memory it is provided.


<br />

### Git Server: Golang runtime monitoring

#### gitserver: go_goroutines

This cloud panel indicates maximum active goroutines.

Refer to the [alert solutions reference](./alert_solutions.md#gitserver-go-goroutines) for relevant alerts.

<br />

#### gitserver: go_gc_duration_seconds

This cloud panel indicates maximum go garbage collection duration.

Refer to the [alert solutions reference](./alert_solutions.md#gitserver-go-gc-duration-seconds) for relevant alerts.

<br />

### Git Server: Kubernetes monitoring (ignore if using Docker Compose or server)

#### gitserver: pods_available_percentage

This cloud panel indicates percentage pods available.

Refer to the [alert solutions reference](./alert_solutions.md#gitserver-pods-available-percentage) for relevant alerts.

<br />

## GitHub Proxy

<p class="subtitle">Proxies all requests to github.com, keeping track of and managing rate limits.</p>

### GitHub Proxy: GitHub API monitoring

#### github-proxy: github_proxy_waiting_requests

This cloud panel indicates number of requests waiting on the global mutex.

Refer to the [alert solutions reference](./alert_solutions.md#github-proxy-github-proxy-waiting-requests) for relevant alerts.

<br />

### GitHub Proxy: Container monitoring (not available on server)

#### github-proxy: container_cpu_usage

This cloud panel indicates container cpu usage total (1m average) across all cores by instance.

Refer to the [alert solutions reference](./alert_solutions.md#github-proxy-container-cpu-usage) for relevant alerts.

<br />

#### github-proxy: container_memory_usage

This cloud panel indicates container memory usage by instance.

Refer to the [alert solutions reference](./alert_solutions.md#github-proxy-container-memory-usage) for relevant alerts.

<br />

#### github-proxy: container_restarts

This cloud panel indicates container restarts every 5m by instance.

Refer to the [alert solutions reference](./alert_solutions.md#github-proxy-container-restarts) for relevant alerts.

<br />

#### github-proxy: fs_inodes_used

This cloud panel indicates fs inodes in use by instance.

Refer to the [alert solutions reference](./alert_solutions.md#github-proxy-fs-inodes-used) for relevant alerts.

<br />

### GitHub Proxy: Provisioning indicators (not available on server)

#### github-proxy: provisioning_container_cpu_usage_long_term

This cloud panel indicates container cpu usage total (90th percentile over 1d) across all cores by instance.

Refer to the [alert solutions reference](./alert_solutions.md#github-proxy-provisioning-container-cpu-usage-long-term) for relevant alerts.

<br />

#### github-proxy: provisioning_container_memory_usage_long_term

This cloud panel indicates container memory usage (1d maximum) by instance.

Refer to the [alert solutions reference](./alert_solutions.md#github-proxy-provisioning-container-memory-usage-long-term) for relevant alerts.

<br />

#### github-proxy: provisioning_container_cpu_usage_short_term

This cloud panel indicates container cpu usage total (5m maximum) across all cores by instance.

Refer to the [alert solutions reference](./alert_solutions.md#github-proxy-provisioning-container-cpu-usage-short-term) for relevant alerts.

<br />

#### github-proxy: provisioning_container_memory_usage_short_term

This cloud panel indicates container memory usage (5m maximum) by instance.

Refer to the [alert solutions reference](./alert_solutions.md#github-proxy-provisioning-container-memory-usage-short-term) for relevant alerts.

<br />

### GitHub Proxy: Golang runtime monitoring

#### github-proxy: go_goroutines

This cloud panel indicates maximum active goroutines.

Refer to the [alert solutions reference](./alert_solutions.md#github-proxy-go-goroutines) for relevant alerts.

<br />

#### github-proxy: go_gc_duration_seconds

This cloud panel indicates maximum go garbage collection duration.

Refer to the [alert solutions reference](./alert_solutions.md#github-proxy-go-gc-duration-seconds) for relevant alerts.

<br />

### GitHub Proxy: Kubernetes monitoring (ignore if using Docker Compose or server)

#### github-proxy: pods_available_percentage

This cloud panel indicates percentage pods available.

Refer to the [alert solutions reference](./alert_solutions.md#github-proxy-pods-available-percentage) for relevant alerts.

<br />

## Postgres

<p class="subtitle">Postgres metrics, exported from postgres_exporter (only available on Kubernetes).</p>

#### postgres: connections

This cloud panel indicates active connections.

Refer to the [alert solutions reference](./alert_solutions.md#postgres-connections) for relevant alerts.

<br />

#### postgres: transaction_durations

This cloud panel indicates maximum transaction durations.

Refer to the [alert solutions reference](./alert_solutions.md#postgres-transaction-durations) for relevant alerts.

<br />

### Postgres: Database and collector status

#### postgres: postgres_up

This cloud panel indicates database availability.

A non-zero value indicates the database is online.

Refer to the [alert solutions reference](./alert_solutions.md#postgres-postgres-up) for relevant alerts.

<br />

#### postgres: pg_exporter_err

This cloud panel indicates errors scraping postgres exporter.

This value indicates issues retrieving metrics from postgres_exporter.

Refer to the [alert solutions reference](./alert_solutions.md#postgres-pg-exporter-err) for relevant alerts.

<br />

#### postgres: migration_in_progress

This cloud panel indicates active schema migration.

A 0 value indicates that no migration is in progress.

Refer to the [alert solutions reference](./alert_solutions.md#postgres-migration-in-progress) for relevant alerts.

<br />

### Postgres: Provisioning indicators (not available on server)

#### postgres: provisioning_container_cpu_usage_long_term

This cloud panel indicates container cpu usage total (90th percentile over 1d) across all cores by instance.

Refer to the [alert solutions reference](./alert_solutions.md#postgres-provisioning-container-cpu-usage-long-term) for relevant alerts.

<br />

#### postgres: provisioning_container_memory_usage_long_term

This cloud panel indicates container memory usage (1d maximum) by instance.

Refer to the [alert solutions reference](./alert_solutions.md#postgres-provisioning-container-memory-usage-long-term) for relevant alerts.

<br />

#### postgres: provisioning_container_cpu_usage_short_term

This cloud panel indicates container cpu usage total (5m maximum) across all cores by instance.

Refer to the [alert solutions reference](./alert_solutions.md#postgres-provisioning-container-cpu-usage-short-term) for relevant alerts.

<br />

#### postgres: provisioning_container_memory_usage_short_term

This cloud panel indicates container memory usage (5m maximum) by instance.

Refer to the [alert solutions reference](./alert_solutions.md#postgres-provisioning-container-memory-usage-short-term) for relevant alerts.

<br />

### Postgres: Kubernetes monitoring (ignore if using Docker Compose or server)

#### postgres: pods_available_percentage

This cloud panel indicates percentage pods available.

Refer to the [alert solutions reference](./alert_solutions.md#postgres-pods-available-percentage) for relevant alerts.

<br />

## Precise Code Intel Worker

<p class="subtitle">Handles conversion of uploaded precise code intelligence bundles.</p>

### Precise Code Intel Worker: Upload queue

#### precise-code-intel-worker: upload_queue_size

This code-intel panel indicates queue size.

Refer to the [alert solutions reference](./alert_solutions.md#precise-code-intel-worker-upload-queue-size) for relevant alerts.

<br />

#### precise-code-intel-worker: upload_queue_growth_rate

This code-intel panel indicates queue growth rate over 30m.

Refer to the [alert solutions reference](./alert_solutions.md#precise-code-intel-worker-upload-queue-growth-rate) for relevant alerts.

<br />

#### precise-code-intel-worker: job_errors

This code-intel panel indicates job errors errors every 5m.

Refer to the [alert solutions reference](./alert_solutions.md#precise-code-intel-worker-job-errors) for relevant alerts.

<br />

#### precise-code-intel-worker: active_workers

This code-intel panel indicates active workers processing uploads.


<br />

#### precise-code-intel-worker: active_jobs

This code-intel panel indicates active jobs.


<br />

### Precise Code Intel Worker: Workers

#### precise-code-intel-worker: job_99th_percentile_duration

This code-intel panel indicates 99th percentile successful job duration over 5m.


<br />

### Precise Code Intel Worker: Stores and clients

#### precise-code-intel-worker: codeintel_dbstore_99th_percentile_duration

This code-intel panel indicates 99th percentile successful database store operation duration over 5m.

Refer to the [alert solutions reference](./alert_solutions.md#precise-code-intel-worker-codeintel-dbstore-99th-percentile-duration) for relevant alerts.

<br />

#### precise-code-intel-worker: codeintel_dbstore_errors

This code-intel panel indicates database store errors every 5m.

Refer to the [alert solutions reference](./alert_solutions.md#precise-code-intel-worker-codeintel-dbstore-errors) for relevant alerts.

<br />

#### precise-code-intel-worker: codeintel_workerstore_99th_percentile_duration

This code-intel panel indicates 99th percentile successful worker store operation duration over 5m.

Refer to the [alert solutions reference](./alert_solutions.md#precise-code-intel-worker-codeintel-workerstore-99th-percentile-duration) for relevant alerts.

<br />

#### precise-code-intel-worker: codeintel_workerstore_errors

This code-intel panel indicates worker store errors every 5m.

Refer to the [alert solutions reference](./alert_solutions.md#precise-code-intel-worker-codeintel-workerstore-errors) for relevant alerts.

<br />

#### precise-code-intel-worker: codeintel_lsifstore_99th_percentile_duration

This code-intel panel indicates 99th percentile successful LSIF store operation duration over 5m.

Refer to the [alert solutions reference](./alert_solutions.md#precise-code-intel-worker-codeintel-lsifstore-99th-percentile-duration) for relevant alerts.

<br />

#### precise-code-intel-worker: codeintel_lsifstore_errors

This code-intel panel indicates lSIF store errors every 5m.

Refer to the [alert solutions reference](./alert_solutions.md#precise-code-intel-worker-codeintel-lsifstore-errors) for relevant alerts.

<br />

#### precise-code-intel-worker: codeintel_uploadstore_99th_percentile_duration

This code-intel panel indicates 99th percentile successful upload store operation duration over 5m.

Refer to the [alert solutions reference](./alert_solutions.md#precise-code-intel-worker-codeintel-uploadstore-99th-percentile-duration) for relevant alerts.

<br />

#### precise-code-intel-worker: codeintel_uploadstore_errors

This code-intel panel indicates upload store errors every 5m.

Refer to the [alert solutions reference](./alert_solutions.md#precise-code-intel-worker-codeintel-uploadstore-errors) for relevant alerts.

<br />

#### precise-code-intel-worker: codeintel_gitserverclient_99th_percentile_duration

This code-intel panel indicates 99th percentile successful gitserver client operation duration over 5m.

Refer to the [alert solutions reference](./alert_solutions.md#precise-code-intel-worker-codeintel-gitserverclient-99th-percentile-duration) for relevant alerts.

<br />

#### precise-code-intel-worker: codeintel_gitserverclient_errors

This code-intel panel indicates gitserver client errors every 5m.

Refer to the [alert solutions reference](./alert_solutions.md#precise-code-intel-worker-codeintel-gitserverclient-errors) for relevant alerts.

<br />

### Precise Code Intel Worker: Internal service requests

#### precise-code-intel-worker: frontend_internal_api_error_responses

This code-intel panel indicates frontend-internal API error responses every 5m by route.

Refer to the [alert solutions reference](./alert_solutions.md#precise-code-intel-worker-frontend-internal-api-error-responses) for relevant alerts.

<br />

### Precise Code Intel Worker: Container monitoring (not available on server)

#### precise-code-intel-worker: container_cpu_usage

This code-intel panel indicates container cpu usage total (1m average) across all cores by instance.

Refer to the [alert solutions reference](./alert_solutions.md#precise-code-intel-worker-container-cpu-usage) for relevant alerts.

<br />

#### precise-code-intel-worker: container_memory_usage

This code-intel panel indicates container memory usage by instance.

Refer to the [alert solutions reference](./alert_solutions.md#precise-code-intel-worker-container-memory-usage) for relevant alerts.

<br />

#### precise-code-intel-worker: container_restarts

This code-intel panel indicates container restarts every 5m by instance.

Refer to the [alert solutions reference](./alert_solutions.md#precise-code-intel-worker-container-restarts) for relevant alerts.

<br />

#### precise-code-intel-worker: fs_inodes_used

This code-intel panel indicates fs inodes in use by instance.

Refer to the [alert solutions reference](./alert_solutions.md#precise-code-intel-worker-fs-inodes-used) for relevant alerts.

<br />

### Precise Code Intel Worker: Provisioning indicators (not available on server)

#### precise-code-intel-worker: provisioning_container_cpu_usage_long_term

This code-intel panel indicates container cpu usage total (90th percentile over 1d) across all cores by instance.

Refer to the [alert solutions reference](./alert_solutions.md#precise-code-intel-worker-provisioning-container-cpu-usage-long-term) for relevant alerts.

<br />

#### precise-code-intel-worker: provisioning_container_memory_usage_long_term

This code-intel panel indicates container memory usage (1d maximum) by instance.

Refer to the [alert solutions reference](./alert_solutions.md#precise-code-intel-worker-provisioning-container-memory-usage-long-term) for relevant alerts.

<br />

#### precise-code-intel-worker: provisioning_container_cpu_usage_short_term

This code-intel panel indicates container cpu usage total (5m maximum) across all cores by instance.

Refer to the [alert solutions reference](./alert_solutions.md#precise-code-intel-worker-provisioning-container-cpu-usage-short-term) for relevant alerts.

<br />

#### precise-code-intel-worker: provisioning_container_memory_usage_short_term

This code-intel panel indicates container memory usage (5m maximum) by instance.

Refer to the [alert solutions reference](./alert_solutions.md#precise-code-intel-worker-provisioning-container-memory-usage-short-term) for relevant alerts.

<br />

### Precise Code Intel Worker: Golang runtime monitoring

#### precise-code-intel-worker: go_goroutines

This code-intel panel indicates maximum active goroutines.

Refer to the [alert solutions reference](./alert_solutions.md#precise-code-intel-worker-go-goroutines) for relevant alerts.

<br />

#### precise-code-intel-worker: go_gc_duration_seconds

This code-intel panel indicates maximum go garbage collection duration.

Refer to the [alert solutions reference](./alert_solutions.md#precise-code-intel-worker-go-gc-duration-seconds) for relevant alerts.

<br />

### Precise Code Intel Worker: Kubernetes monitoring (ignore if using Docker Compose or server)

#### precise-code-intel-worker: pods_available_percentage

This code-intel panel indicates percentage pods available.

Refer to the [alert solutions reference](./alert_solutions.md#precise-code-intel-worker-pods-available-percentage) for relevant alerts.

<br />

## Query Runner

<p class="subtitle">Periodically runs saved searches and instructs the frontend to send out notifications.</p>

#### query-runner: frontend_internal_api_error_responses

This search panel indicates frontend-internal API error responses every 5m by route.

Refer to the [alert solutions reference](./alert_solutions.md#query-runner-frontend-internal-api-error-responses) for relevant alerts.

<br />

### Query Runner: Container monitoring (not available on server)

#### query-runner: container_memory_usage

This search panel indicates container memory usage by instance.

Refer to the [alert solutions reference](./alert_solutions.md#query-runner-container-memory-usage) for relevant alerts.

<br />

#### query-runner: container_cpu_usage

This search panel indicates container cpu usage total (1m average) across all cores by instance.

Refer to the [alert solutions reference](./alert_solutions.md#query-runner-container-cpu-usage) for relevant alerts.

<br />

#### query-runner: container_restarts

This search panel indicates container restarts every 5m by instance.

Refer to the [alert solutions reference](./alert_solutions.md#query-runner-container-restarts) for relevant alerts.

<br />

#### query-runner: fs_inodes_used

This search panel indicates fs inodes in use by instance.

Refer to the [alert solutions reference](./alert_solutions.md#query-runner-fs-inodes-used) for relevant alerts.

<br />

### Query Runner: Provisioning indicators (not available on server)

#### query-runner: provisioning_container_cpu_usage_long_term

This search panel indicates container cpu usage total (90th percentile over 1d) across all cores by instance.

Refer to the [alert solutions reference](./alert_solutions.md#query-runner-provisioning-container-cpu-usage-long-term) for relevant alerts.

<br />

#### query-runner: provisioning_container_memory_usage_long_term

This search panel indicates container memory usage (1d maximum) by instance.

Refer to the [alert solutions reference](./alert_solutions.md#query-runner-provisioning-container-memory-usage-long-term) for relevant alerts.

<br />

#### query-runner: provisioning_container_cpu_usage_short_term

This search panel indicates container cpu usage total (5m maximum) across all cores by instance.

Refer to the [alert solutions reference](./alert_solutions.md#query-runner-provisioning-container-cpu-usage-short-term) for relevant alerts.

<br />

#### query-runner: provisioning_container_memory_usage_short_term

This search panel indicates container memory usage (5m maximum) by instance.

Refer to the [alert solutions reference](./alert_solutions.md#query-runner-provisioning-container-memory-usage-short-term) for relevant alerts.

<br />

### Query Runner: Golang runtime monitoring

#### query-runner: go_goroutines

This search panel indicates maximum active goroutines.

Refer to the [alert solutions reference](./alert_solutions.md#query-runner-go-goroutines) for relevant alerts.

<br />

#### query-runner: go_gc_duration_seconds

This search panel indicates maximum go garbage collection duration.

Refer to the [alert solutions reference](./alert_solutions.md#query-runner-go-gc-duration-seconds) for relevant alerts.

<br />

### Query Runner: Kubernetes monitoring (ignore if using Docker Compose or server)

#### query-runner: pods_available_percentage

This search panel indicates percentage pods available.

Refer to the [alert solutions reference](./alert_solutions.md#query-runner-pods-available-percentage) for relevant alerts.

<br />

## Repo Updater

<p class="subtitle">Manages interaction with code hosts, instructs Gitserver to update repositories.</p>

#### repo-updater: frontend_internal_api_error_responses

This cloud panel indicates frontend-internal API error responses every 5m by route.

Refer to the [alert solutions reference](./alert_solutions.md#repo-updater-frontend-internal-api-error-responses) for relevant alerts.

<br />

### Repo Updater: Repositories

#### repo-updater: syncer_sync_last_time

This cloud panel indicates time since last sync.

A high value here indicates issues synchronizing repository permissions.
If the value is persistently high, make sure all external services have valid tokens.


<br />

#### repo-updater: src_repoupdater_max_sync_backoff

This cloud panel indicates time since oldest sync.

Refer to the [alert solutions reference](./alert_solutions.md#repo-updater-src-repoupdater-max-sync-backoff) for relevant alerts.

<br />

#### repo-updater: src_repoupdater_syncer_sync_errors_total

This cloud panel indicates sync error rate.

Refer to the [alert solutions reference](./alert_solutions.md#repo-updater-src-repoupdater-syncer-sync-errors-total) for relevant alerts.

<br />

#### repo-updater: syncer_sync_start

This cloud panel indicates sync was started.

Refer to the [alert solutions reference](./alert_solutions.md#repo-updater-syncer-sync-start) for relevant alerts.

<br />

#### repo-updater: syncer_sync_duration

This cloud panel indicates 95th repositories sync duration.

Refer to the [alert solutions reference](./alert_solutions.md#repo-updater-syncer-sync-duration) for relevant alerts.

<br />

#### repo-updater: source_duration

This cloud panel indicates 95th repositories source duration.

Refer to the [alert solutions reference](./alert_solutions.md#repo-updater-source-duration) for relevant alerts.

<br />

#### repo-updater: syncer_synced_repos

This cloud panel indicates repositories synced.

Refer to the [alert solutions reference](./alert_solutions.md#repo-updater-syncer-synced-repos) for relevant alerts.

<br />

#### repo-updater: sourced_repos

This cloud panel indicates repositories sourced.

Refer to the [alert solutions reference](./alert_solutions.md#repo-updater-sourced-repos) for relevant alerts.

<br />

#### repo-updater: user_added_repos

This cloud panel indicates total number of user added repos.

Refer to the [alert solutions reference](./alert_solutions.md#repo-updater-user-added-repos) for relevant alerts.

<br />

#### repo-updater: purge_failed

This cloud panel indicates repositories purge failed.

Refer to the [alert solutions reference](./alert_solutions.md#repo-updater-purge-failed) for relevant alerts.

<br />

#### repo-updater: sched_auto_fetch

This cloud panel indicates repositories scheduled due to hitting a deadline.

Refer to the [alert solutions reference](./alert_solutions.md#repo-updater-sched-auto-fetch) for relevant alerts.

<br />

#### repo-updater: sched_manual_fetch

This cloud panel indicates repositories scheduled due to user traffic.

Check repo-updater logs if this value is persistently high.
This does not indicate anything if there are no user added code hosts.


<br />

#### repo-updater: sched_known_repos

This cloud panel indicates repositories managed by the scheduler.

Refer to the [alert solutions reference](./alert_solutions.md#repo-updater-sched-known-repos) for relevant alerts.

<br />

#### repo-updater: sched_update_queue_length

This cloud panel indicates rate of growth of update queue length over 5 minutes.

Refer to the [alert solutions reference](./alert_solutions.md#repo-updater-sched-update-queue-length) for relevant alerts.

<br />

#### repo-updater: sched_loops

This cloud panel indicates scheduler loops.

Refer to the [alert solutions reference](./alert_solutions.md#repo-updater-sched-loops) for relevant alerts.

<br />

#### repo-updater: sched_error

This cloud panel indicates repositories schedule error rate.

Refer to the [alert solutions reference](./alert_solutions.md#repo-updater-sched-error) for relevant alerts.

<br />

### Repo Updater: Permissions

#### repo-updater: perms_syncer_perms

This cloud panel indicates time gap between least and most up to date permissions.

Refer to the [alert solutions reference](./alert_solutions.md#repo-updater-perms-syncer-perms) for relevant alerts.

<br />

#### repo-updater: perms_syncer_stale_perms

This cloud panel indicates number of entities with stale permissions.

Refer to the [alert solutions reference](./alert_solutions.md#repo-updater-perms-syncer-stale-perms) for relevant alerts.

<br />

#### repo-updater: perms_syncer_no_perms

This cloud panel indicates number of entities with no permissions.

Refer to the [alert solutions reference](./alert_solutions.md#repo-updater-perms-syncer-no-perms) for relevant alerts.

<br />

#### repo-updater: perms_syncer_sync_duration

This cloud panel indicates 95th permissions sync duration.

Refer to the [alert solutions reference](./alert_solutions.md#repo-updater-perms-syncer-sync-duration) for relevant alerts.

<br />

#### repo-updater: perms_syncer_queue_size

This cloud panel indicates permissions sync queued items.

Refer to the [alert solutions reference](./alert_solutions.md#repo-updater-perms-syncer-queue-size) for relevant alerts.

<br />

#### repo-updater: perms_syncer_sync_errors

This cloud panel indicates permissions sync error rate.

Refer to the [alert solutions reference](./alert_solutions.md#repo-updater-perms-syncer-sync-errors) for relevant alerts.

<br />

### Repo Updater: External services

#### repo-updater: src_repoupdater_external_services_total

This cloud panel indicates the total number of external services.

Refer to the [alert solutions reference](./alert_solutions.md#repo-updater-src-repoupdater-external-services-total) for relevant alerts.

<br />

#### repo-updater: src_repoupdater_user_external_services_total

This cloud panel indicates the total number of user added external services.

Refer to the [alert solutions reference](./alert_solutions.md#repo-updater-src-repoupdater-user-external-services-total) for relevant alerts.

<br />

#### repo-updater: repoupdater_queued_sync_jobs_total

This cloud panel indicates the total number of queued sync jobs.

Refer to the [alert solutions reference](./alert_solutions.md#repo-updater-repoupdater-queued-sync-jobs-total) for relevant alerts.

<br />

#### repo-updater: repoupdater_completed_sync_jobs_total

This cloud panel indicates the total number of completed sync jobs.

Refer to the [alert solutions reference](./alert_solutions.md#repo-updater-repoupdater-completed-sync-jobs-total) for relevant alerts.

<br />

#### repo-updater: repoupdater_errored_sync_jobs_total

This cloud panel indicates the total number of errored sync jobs.

Refer to the [alert solutions reference](./alert_solutions.md#repo-updater-repoupdater-errored-sync-jobs-total) for relevant alerts.

<br />

#### repo-updater: github_graphql_rate_limit_remaining

This cloud panel indicates remaining calls to GitHub graphql API before hitting the rate limit.

Refer to the [alert solutions reference](./alert_solutions.md#repo-updater-github-graphql-rate-limit-remaining) for relevant alerts.

<br />

#### repo-updater: github_rest_rate_limit_remaining

This cloud panel indicates remaining calls to GitHub rest API before hitting the rate limit.

Refer to the [alert solutions reference](./alert_solutions.md#repo-updater-github-rest-rate-limit-remaining) for relevant alerts.

<br />

#### repo-updater: github_search_rate_limit_remaining

This cloud panel indicates remaining calls to GitHub search API before hitting the rate limit.

Refer to the [alert solutions reference](./alert_solutions.md#repo-updater-github-search-rate-limit-remaining) for relevant alerts.

<br />

#### repo-updater: gitlab_rest_rate_limit_remaining

This cloud panel indicates remaining calls to GitLab rest API before hitting the rate limit.

Refer to the [alert solutions reference](./alert_solutions.md#repo-updater-gitlab-rest-rate-limit-remaining) for relevant alerts.

<br />

### Repo Updater: Container monitoring (not available on server)

#### repo-updater: container_cpu_usage

This cloud panel indicates container cpu usage total (1m average) across all cores by instance.

Refer to the [alert solutions reference](./alert_solutions.md#repo-updater-container-cpu-usage) for relevant alerts.

<br />

#### repo-updater: container_memory_usage

This cloud panel indicates container memory usage by instance.

Refer to the [alert solutions reference](./alert_solutions.md#repo-updater-container-memory-usage) for relevant alerts.

<br />

#### repo-updater: container_restarts

This cloud panel indicates container restarts every 5m by instance.

Refer to the [alert solutions reference](./alert_solutions.md#repo-updater-container-restarts) for relevant alerts.

<br />

#### repo-updater: fs_inodes_used

This cloud panel indicates fs inodes in use by instance.

Refer to the [alert solutions reference](./alert_solutions.md#repo-updater-fs-inodes-used) for relevant alerts.

<br />

### Repo Updater: Provisioning indicators (not available on server)

#### repo-updater: provisioning_container_cpu_usage_long_term

This cloud panel indicates container cpu usage total (90th percentile over 1d) across all cores by instance.

Refer to the [alert solutions reference](./alert_solutions.md#repo-updater-provisioning-container-cpu-usage-long-term) for relevant alerts.

<br />

#### repo-updater: provisioning_container_memory_usage_long_term

This cloud panel indicates container memory usage (1d maximum) by instance.

Refer to the [alert solutions reference](./alert_solutions.md#repo-updater-provisioning-container-memory-usage-long-term) for relevant alerts.

<br />

#### repo-updater: provisioning_container_cpu_usage_short_term

This cloud panel indicates container cpu usage total (5m maximum) across all cores by instance.

Refer to the [alert solutions reference](./alert_solutions.md#repo-updater-provisioning-container-cpu-usage-short-term) for relevant alerts.

<br />

#### repo-updater: provisioning_container_memory_usage_short_term

This cloud panel indicates container memory usage (5m maximum) by instance.

Refer to the [alert solutions reference](./alert_solutions.md#repo-updater-provisioning-container-memory-usage-short-term) for relevant alerts.

<br />

### Repo Updater: Golang runtime monitoring

#### repo-updater: go_goroutines

This cloud panel indicates maximum active goroutines.

Refer to the [alert solutions reference](./alert_solutions.md#repo-updater-go-goroutines) for relevant alerts.

<br />

#### repo-updater: go_gc_duration_seconds

This cloud panel indicates maximum go garbage collection duration.

Refer to the [alert solutions reference](./alert_solutions.md#repo-updater-go-gc-duration-seconds) for relevant alerts.

<br />

### Repo Updater: Kubernetes monitoring (ignore if using Docker Compose or server)

#### repo-updater: pods_available_percentage

This cloud panel indicates percentage pods available.

Refer to the [alert solutions reference](./alert_solutions.md#repo-updater-pods-available-percentage) for relevant alerts.

<br />

## Searcher

<p class="subtitle">Performs unindexed searches (diff and commit search, text search for unindexed branches).</p>

#### searcher: unindexed_search_request_errors

This search panel indicates unindexed search request errors every 5m by code.

Refer to the [alert solutions reference](./alert_solutions.md#searcher-unindexed-search-request-errors) for relevant alerts.

<br />

#### searcher: replica_traffic

This search panel indicates requests per second over 10m.

Refer to the [alert solutions reference](./alert_solutions.md#searcher-replica-traffic) for relevant alerts.

<br />

#### searcher: frontend_internal_api_error_responses

This search panel indicates frontend-internal API error responses every 5m by route.

Refer to the [alert solutions reference](./alert_solutions.md#searcher-frontend-internal-api-error-responses) for relevant alerts.

<br />

### Searcher: Container monitoring (not available on server)

#### searcher: container_cpu_usage

This search panel indicates container cpu usage total (1m average) across all cores by instance.

Refer to the [alert solutions reference](./alert_solutions.md#searcher-container-cpu-usage) for relevant alerts.

<br />

#### searcher: container_memory_usage

This search panel indicates container memory usage by instance.

Refer to the [alert solutions reference](./alert_solutions.md#searcher-container-memory-usage) for relevant alerts.

<br />

#### searcher: container_restarts

This search panel indicates container restarts every 5m by instance.

Refer to the [alert solutions reference](./alert_solutions.md#searcher-container-restarts) for relevant alerts.

<br />

#### searcher: fs_inodes_used

This search panel indicates fs inodes in use by instance.

Refer to the [alert solutions reference](./alert_solutions.md#searcher-fs-inodes-used) for relevant alerts.

<br />

### Searcher: Provisioning indicators (not available on server)

#### searcher: provisioning_container_cpu_usage_long_term

This search panel indicates container cpu usage total (90th percentile over 1d) across all cores by instance.

Refer to the [alert solutions reference](./alert_solutions.md#searcher-provisioning-container-cpu-usage-long-term) for relevant alerts.

<br />

#### searcher: provisioning_container_memory_usage_long_term

This search panel indicates container memory usage (1d maximum) by instance.

Refer to the [alert solutions reference](./alert_solutions.md#searcher-provisioning-container-memory-usage-long-term) for relevant alerts.

<br />

#### searcher: provisioning_container_cpu_usage_short_term

This search panel indicates container cpu usage total (5m maximum) across all cores by instance.

Refer to the [alert solutions reference](./alert_solutions.md#searcher-provisioning-container-cpu-usage-short-term) for relevant alerts.

<br />

#### searcher: provisioning_container_memory_usage_short_term

This search panel indicates container memory usage (5m maximum) by instance.

Refer to the [alert solutions reference](./alert_solutions.md#searcher-provisioning-container-memory-usage-short-term) for relevant alerts.

<br />

### Searcher: Golang runtime monitoring

#### searcher: go_goroutines

This search panel indicates maximum active goroutines.

Refer to the [alert solutions reference](./alert_solutions.md#searcher-go-goroutines) for relevant alerts.

<br />

#### searcher: go_gc_duration_seconds

This search panel indicates maximum go garbage collection duration.

Refer to the [alert solutions reference](./alert_solutions.md#searcher-go-gc-duration-seconds) for relevant alerts.

<br />

### Searcher: Kubernetes monitoring (ignore if using Docker Compose or server)

#### searcher: pods_available_percentage

This search panel indicates percentage pods available.

Refer to the [alert solutions reference](./alert_solutions.md#searcher-pods-available-percentage) for relevant alerts.

<br />

## Symbols

<p class="subtitle">Handles symbol searches for unindexed branches.</p>

#### symbols: store_fetch_failures

This code-intel panel indicates store fetch failures every 5m.

Refer to the [alert solutions reference](./alert_solutions.md#symbols-store-fetch-failures) for relevant alerts.

<br />

#### symbols: current_fetch_queue_size

This code-intel panel indicates current fetch queue size.

Refer to the [alert solutions reference](./alert_solutions.md#symbols-current-fetch-queue-size) for relevant alerts.

<br />

#### symbols: frontend_internal_api_error_responses

This code-intel panel indicates frontend-internal API error responses every 5m by route.

Refer to the [alert solutions reference](./alert_solutions.md#symbols-frontend-internal-api-error-responses) for relevant alerts.

<br />

### Symbols: Container monitoring (not available on server)

#### symbols: container_cpu_usage

This code-intel panel indicates container cpu usage total (1m average) across all cores by instance.

Refer to the [alert solutions reference](./alert_solutions.md#symbols-container-cpu-usage) for relevant alerts.

<br />

#### symbols: container_memory_usage

This code-intel panel indicates container memory usage by instance.

Refer to the [alert solutions reference](./alert_solutions.md#symbols-container-memory-usage) for relevant alerts.

<br />

#### symbols: container_restarts

This code-intel panel indicates container restarts every 5m by instance.

Refer to the [alert solutions reference](./alert_solutions.md#symbols-container-restarts) for relevant alerts.

<br />

#### symbols: fs_inodes_used

This code-intel panel indicates fs inodes in use by instance.

Refer to the [alert solutions reference](./alert_solutions.md#symbols-fs-inodes-used) for relevant alerts.

<br />

### Symbols: Provisioning indicators (not available on server)

#### symbols: provisioning_container_cpu_usage_long_term

This code-intel panel indicates container cpu usage total (90th percentile over 1d) across all cores by instance.

Refer to the [alert solutions reference](./alert_solutions.md#symbols-provisioning-container-cpu-usage-long-term) for relevant alerts.

<br />

#### symbols: provisioning_container_memory_usage_long_term

This code-intel panel indicates container memory usage (1d maximum) by instance.

Refer to the [alert solutions reference](./alert_solutions.md#symbols-provisioning-container-memory-usage-long-term) for relevant alerts.

<br />

#### symbols: provisioning_container_cpu_usage_short_term

This code-intel panel indicates container cpu usage total (5m maximum) across all cores by instance.

Refer to the [alert solutions reference](./alert_solutions.md#symbols-provisioning-container-cpu-usage-short-term) for relevant alerts.

<br />

#### symbols: provisioning_container_memory_usage_short_term

This code-intel panel indicates container memory usage (5m maximum) by instance.

Refer to the [alert solutions reference](./alert_solutions.md#symbols-provisioning-container-memory-usage-short-term) for relevant alerts.

<br />

### Symbols: Golang runtime monitoring

#### symbols: go_goroutines

This code-intel panel indicates maximum active goroutines.

Refer to the [alert solutions reference](./alert_solutions.md#symbols-go-goroutines) for relevant alerts.

<br />

#### symbols: go_gc_duration_seconds

This code-intel panel indicates maximum go garbage collection duration.

Refer to the [alert solutions reference](./alert_solutions.md#symbols-go-gc-duration-seconds) for relevant alerts.

<br />

### Symbols: Kubernetes monitoring (ignore if using Docker Compose or server)

#### symbols: pods_available_percentage

This code-intel panel indicates percentage pods available.

Refer to the [alert solutions reference](./alert_solutions.md#symbols-pods-available-percentage) for relevant alerts.

<br />

## Syntect Server

<p class="subtitle">Handles syntax highlighting for code files.</p>

#### syntect-server: syntax_highlighting_errors

This cloud panel indicates syntax highlighting errors every 5m.


<br />

#### syntect-server: syntax_highlighting_timeouts

This cloud panel indicates syntax highlighting timeouts every 5m.


<br />

#### syntect-server: syntax_highlighting_panics

This cloud panel indicates syntax highlighting panics every 5m.


<br />

#### syntect-server: syntax_highlighting_worker_deaths

This cloud panel indicates syntax highlighter worker deaths every 5m.


<br />

### Syntect Server: Container monitoring (not available on server)

#### syntect-server: container_cpu_usage

This cloud panel indicates container cpu usage total (1m average) across all cores by instance.

Refer to the [alert solutions reference](./alert_solutions.md#syntect-server-container-cpu-usage) for relevant alerts.

<br />

#### syntect-server: container_memory_usage

This cloud panel indicates container memory usage by instance.

Refer to the [alert solutions reference](./alert_solutions.md#syntect-server-container-memory-usage) for relevant alerts.

<br />

#### syntect-server: container_restarts

This cloud panel indicates container restarts every 5m by instance.

Refer to the [alert solutions reference](./alert_solutions.md#syntect-server-container-restarts) for relevant alerts.

<br />

#### syntect-server: fs_inodes_used

This cloud panel indicates fs inodes in use by instance.

Refer to the [alert solutions reference](./alert_solutions.md#syntect-server-fs-inodes-used) for relevant alerts.

<br />

### Syntect Server: Provisioning indicators (not available on server)

#### syntect-server: provisioning_container_cpu_usage_long_term

This cloud panel indicates container cpu usage total (90th percentile over 1d) across all cores by instance.

Refer to the [alert solutions reference](./alert_solutions.md#syntect-server-provisioning-container-cpu-usage-long-term) for relevant alerts.

<br />

#### syntect-server: provisioning_container_memory_usage_long_term

This cloud panel indicates container memory usage (1d maximum) by instance.

Refer to the [alert solutions reference](./alert_solutions.md#syntect-server-provisioning-container-memory-usage-long-term) for relevant alerts.

<br />

#### syntect-server: provisioning_container_cpu_usage_short_term

This cloud panel indicates container cpu usage total (5m maximum) across all cores by instance.

Refer to the [alert solutions reference](./alert_solutions.md#syntect-server-provisioning-container-cpu-usage-short-term) for relevant alerts.

<br />

#### syntect-server: provisioning_container_memory_usage_short_term

This cloud panel indicates container memory usage (5m maximum) by instance.

Refer to the [alert solutions reference](./alert_solutions.md#syntect-server-provisioning-container-memory-usage-short-term) for relevant alerts.

<br />

### Syntect Server: Kubernetes monitoring (ignore if using Docker Compose or server)

#### syntect-server: pods_available_percentage

This cloud panel indicates percentage pods available.

Refer to the [alert solutions reference](./alert_solutions.md#syntect-server-pods-available-percentage) for relevant alerts.

<br />

## Zoekt Index Server

<p class="subtitle">Indexes repositories and populates the search index.</p>

#### zoekt-indexserver: average_resolve_revision_duration

This search panel indicates average resolve revision duration over 5m.

Refer to the [alert solutions reference](./alert_solutions.md#zoekt-indexserver-average-resolve-revision-duration) for relevant alerts.

<br />

### Zoekt Index Server: Container monitoring (not available on server)

#### zoekt-indexserver: container_cpu_usage

This search panel indicates container cpu usage total (1m average) across all cores by instance.

Refer to the [alert solutions reference](./alert_solutions.md#zoekt-indexserver-container-cpu-usage) for relevant alerts.

<br />

#### zoekt-indexserver: container_memory_usage

This search panel indicates container memory usage by instance.

Refer to the [alert solutions reference](./alert_solutions.md#zoekt-indexserver-container-memory-usage) for relevant alerts.

<br />

#### zoekt-indexserver: container_restarts

This search panel indicates container restarts every 5m by instance.

Refer to the [alert solutions reference](./alert_solutions.md#zoekt-indexserver-container-restarts) for relevant alerts.

<br />

#### zoekt-indexserver: fs_inodes_used

This search panel indicates fs inodes in use by instance.

Refer to the [alert solutions reference](./alert_solutions.md#zoekt-indexserver-fs-inodes-used) for relevant alerts.

<br />

#### zoekt-indexserver: fs_io_operations

This search panel indicates filesystem reads and writes rate by instance over 1h.

Refer to the [alert solutions reference](./alert_solutions.md#zoekt-indexserver-fs-io-operations) for relevant alerts.

<br />

### Zoekt Index Server: Provisioning indicators (not available on server)

#### zoekt-indexserver: provisioning_container_cpu_usage_long_term

This search panel indicates container cpu usage total (90th percentile over 1d) across all cores by instance.

Refer to the [alert solutions reference](./alert_solutions.md#zoekt-indexserver-provisioning-container-cpu-usage-long-term) for relevant alerts.

<br />

#### zoekt-indexserver: provisioning_container_memory_usage_long_term

This search panel indicates container memory usage (1d maximum) by instance.

Refer to the [alert solutions reference](./alert_solutions.md#zoekt-indexserver-provisioning-container-memory-usage-long-term) for relevant alerts.

<br />

#### zoekt-indexserver: provisioning_container_cpu_usage_short_term

This search panel indicates container cpu usage total (5m maximum) across all cores by instance.

Refer to the [alert solutions reference](./alert_solutions.md#zoekt-indexserver-provisioning-container-cpu-usage-short-term) for relevant alerts.

<br />

#### zoekt-indexserver: provisioning_container_memory_usage_short_term

This search panel indicates container memory usage (5m maximum) by instance.

Refer to the [alert solutions reference](./alert_solutions.md#zoekt-indexserver-provisioning-container-memory-usage-short-term) for relevant alerts.

<br />

### Zoekt Index Server: Kubernetes monitoring (ignore if using Docker Compose or server)

#### zoekt-indexserver: pods_available_percentage

This search panel indicates percentage pods available.

Refer to the [alert solutions reference](./alert_solutions.md#zoekt-indexserver-pods-available-percentage) for relevant alerts.

<br />

## Zoekt Web Server

<p class="subtitle">Serves indexed search requests using the search index.</p>

#### zoekt-webserver: indexed_search_request_errors

This search panel indicates indexed search request errors every 5m by code.

Refer to the [alert solutions reference](./alert_solutions.md#zoekt-webserver-indexed-search-request-errors) for relevant alerts.

<br />

### Zoekt Web Server: Container monitoring (not available on server)

#### zoekt-webserver: container_cpu_usage

This search panel indicates container cpu usage total (1m average) across all cores by instance.

Refer to the [alert solutions reference](./alert_solutions.md#zoekt-webserver-container-cpu-usage) for relevant alerts.

<br />

#### zoekt-webserver: container_memory_usage

This search panel indicates container memory usage by instance.

Refer to the [alert solutions reference](./alert_solutions.md#zoekt-webserver-container-memory-usage) for relevant alerts.

<br />

#### zoekt-webserver: container_restarts

This search panel indicates container restarts every 5m by instance.

Refer to the [alert solutions reference](./alert_solutions.md#zoekt-webserver-container-restarts) for relevant alerts.

<br />

#### zoekt-webserver: fs_inodes_used

This search panel indicates fs inodes in use by instance.

Refer to the [alert solutions reference](./alert_solutions.md#zoekt-webserver-fs-inodes-used) for relevant alerts.

<br />

#### zoekt-webserver: fs_io_operations

This search panel indicates filesystem reads and writes by instance rate over 1h.

Refer to the [alert solutions reference](./alert_solutions.md#zoekt-webserver-fs-io-operations) for relevant alerts.

<br />

### Zoekt Web Server: Provisioning indicators (not available on server)

#### zoekt-webserver: provisioning_container_cpu_usage_long_term

This search panel indicates container cpu usage total (90th percentile over 1d) across all cores by instance.

Refer to the [alert solutions reference](./alert_solutions.md#zoekt-webserver-provisioning-container-cpu-usage-long-term) for relevant alerts.

<br />

#### zoekt-webserver: provisioning_container_memory_usage_long_term

This search panel indicates container memory usage (1d maximum) by instance.

Refer to the [alert solutions reference](./alert_solutions.md#zoekt-webserver-provisioning-container-memory-usage-long-term) for relevant alerts.

<br />

#### zoekt-webserver: provisioning_container_cpu_usage_short_term

This search panel indicates container cpu usage total (5m maximum) across all cores by instance.

Refer to the [alert solutions reference](./alert_solutions.md#zoekt-webserver-provisioning-container-cpu-usage-short-term) for relevant alerts.

<br />

#### zoekt-webserver: provisioning_container_memory_usage_short_term

This search panel indicates container memory usage (5m maximum) by instance.

Refer to the [alert solutions reference](./alert_solutions.md#zoekt-webserver-provisioning-container-memory-usage-short-term) for relevant alerts.

<br />

## Prometheus

<p class="subtitle">Sourcegraph's all-in-one Prometheus and Alertmanager service.</p>

### Prometheus: Metrics

#### prometheus: prometheus_metrics_bloat

This distribution panel indicates prometheus metrics payload size.

Refer to the [alert solutions reference](./alert_solutions.md#prometheus-prometheus-metrics-bloat) for relevant alerts.

<br />

### Prometheus: Alerts

#### prometheus: alertmanager_notifications_failed_total

This distribution panel indicates failed alertmanager notifications over 1m.

Refer to the [alert solutions reference](./alert_solutions.md#prometheus-alertmanager-notifications-failed-total) for relevant alerts.

<br />

### Prometheus: Container monitoring (not available on server)

#### prometheus: container_cpu_usage

This distribution panel indicates container cpu usage total (1m average) across all cores by instance.

Refer to the [alert solutions reference](./alert_solutions.md#prometheus-container-cpu-usage) for relevant alerts.

<br />

#### prometheus: container_memory_usage

This distribution panel indicates container memory usage by instance.

Refer to the [alert solutions reference](./alert_solutions.md#prometheus-container-memory-usage) for relevant alerts.

<br />

#### prometheus: container_restarts

This distribution panel indicates container restarts every 5m by instance.

Refer to the [alert solutions reference](./alert_solutions.md#prometheus-container-restarts) for relevant alerts.

<br />

#### prometheus: fs_inodes_used

This distribution panel indicates fs inodes in use by instance.

Refer to the [alert solutions reference](./alert_solutions.md#prometheus-fs-inodes-used) for relevant alerts.

<br />

### Prometheus: Provisioning indicators (not available on server)

#### prometheus: provisioning_container_cpu_usage_long_term

This distribution panel indicates container cpu usage total (90th percentile over 1d) across all cores by instance.

Refer to the [alert solutions reference](./alert_solutions.md#prometheus-provisioning-container-cpu-usage-long-term) for relevant alerts.

<br />

#### prometheus: provisioning_container_memory_usage_long_term

This distribution panel indicates container memory usage (1d maximum) by instance.

Refer to the [alert solutions reference](./alert_solutions.md#prometheus-provisioning-container-memory-usage-long-term) for relevant alerts.

<br />

#### prometheus: provisioning_container_cpu_usage_short_term

This distribution panel indicates container cpu usage total (5m maximum) across all cores by instance.

Refer to the [alert solutions reference](./alert_solutions.md#prometheus-provisioning-container-cpu-usage-short-term) for relevant alerts.

<br />

#### prometheus: provisioning_container_memory_usage_short_term

This distribution panel indicates container memory usage (5m maximum) by instance.

Refer to the [alert solutions reference](./alert_solutions.md#prometheus-provisioning-container-memory-usage-short-term) for relevant alerts.

<br />

### Prometheus: Kubernetes monitoring (ignore if using Docker Compose or server)

#### prometheus: pods_available_percentage

This distribution panel indicates percentage pods available.

Refer to the [alert solutions reference](./alert_solutions.md#prometheus-pods-available-percentage) for relevant alerts.

<br />

## Executor Queue

<p class="subtitle">Coordinates the executor work queues.</p>

### Executor Queue: Code intelligence queue

#### executor-queue: codeintel_queue_size

This code-intel panel indicates queue size.

Refer to the [alert solutions reference](./alert_solutions.md#executor-queue-codeintel-queue-size) for relevant alerts.

<br />

#### executor-queue: codeintel_queue_growth_rate

This code-intel panel indicates queue growth rate over 30m.

Refer to the [alert solutions reference](./alert_solutions.md#executor-queue-codeintel-queue-growth-rate) for relevant alerts.

<br />

#### executor-queue: codeintel_job_errors

This code-intel panel indicates job errors every 5m.

Refer to the [alert solutions reference](./alert_solutions.md#executor-queue-codeintel-job-errors) for relevant alerts.

<br />

#### executor-queue: codeintel_active_executors

This code-intel panel indicates active executors processing codeintel jobs.


<br />

#### executor-queue: codeintel_active_jobs

This code-intel panel indicates active jobs.


<br />

### Executor Queue: Stores and clients

#### executor-queue: codeintel_workerstore_99th_percentile_duration

This code-intel panel indicates 99th percentile successful worker store operation duration over 5m.

Refer to the [alert solutions reference](./alert_solutions.md#executor-queue-codeintel-workerstore-99th-percentile-duration) for relevant alerts.

<br />

#### executor-queue: codeintel_workerstore_errors

This code-intel panel indicates worker store errors every 5m.

Refer to the [alert solutions reference](./alert_solutions.md#executor-queue-codeintel-workerstore-errors) for relevant alerts.

<br />

### Executor Queue: Internal service requests

#### executor-queue: frontend_internal_api_error_responses

This code-intel panel indicates frontend-internal API error responses every 5m by route.

Refer to the [alert solutions reference](./alert_solutions.md#executor-queue-frontend-internal-api-error-responses) for relevant alerts.

<br />

### Executor Queue: Container monitoring (not available on server)

#### executor-queue: container_cpu_usage

This code-intel panel indicates container cpu usage total (1m average) across all cores by instance.

Refer to the [alert solutions reference](./alert_solutions.md#executor-queue-container-cpu-usage) for relevant alerts.

<br />

#### executor-queue: container_memory_usage

This code-intel panel indicates container memory usage by instance.

Refer to the [alert solutions reference](./alert_solutions.md#executor-queue-container-memory-usage) for relevant alerts.

<br />

#### executor-queue: container_restarts

This code-intel panel indicates container restarts every 5m by instance.

Refer to the [alert solutions reference](./alert_solutions.md#executor-queue-container-restarts) for relevant alerts.

<br />

#### executor-queue: fs_inodes_used

This code-intel panel indicates fs inodes in use by instance.

Refer to the [alert solutions reference](./alert_solutions.md#executor-queue-fs-inodes-used) for relevant alerts.

<br />

### Executor Queue: Provisioning indicators (not available on server)

#### executor-queue: provisioning_container_cpu_usage_long_term

This code-intel panel indicates container cpu usage total (90th percentile over 1d) across all cores by instance.

Refer to the [alert solutions reference](./alert_solutions.md#executor-queue-provisioning-container-cpu-usage-long-term) for relevant alerts.

<br />

#### executor-queue: provisioning_container_memory_usage_long_term

This code-intel panel indicates container memory usage (1d maximum) by instance.

Refer to the [alert solutions reference](./alert_solutions.md#executor-queue-provisioning-container-memory-usage-long-term) for relevant alerts.

<br />

#### executor-queue: provisioning_container_cpu_usage_short_term

This code-intel panel indicates container cpu usage total (5m maximum) across all cores by instance.

Refer to the [alert solutions reference](./alert_solutions.md#executor-queue-provisioning-container-cpu-usage-short-term) for relevant alerts.

<br />

#### executor-queue: provisioning_container_memory_usage_short_term

This code-intel panel indicates container memory usage (5m maximum) by instance.

Refer to the [alert solutions reference](./alert_solutions.md#executor-queue-provisioning-container-memory-usage-short-term) for relevant alerts.

<br />

### Executor Queue: Golang runtime monitoring

#### executor-queue: go_goroutines

This code-intel panel indicates maximum active goroutines.

Refer to the [alert solutions reference](./alert_solutions.md#executor-queue-go-goroutines) for relevant alerts.

<br />

#### executor-queue: go_gc_duration_seconds

This code-intel panel indicates maximum go garbage collection duration.

Refer to the [alert solutions reference](./alert_solutions.md#executor-queue-go-gc-duration-seconds) for relevant alerts.

<br />

### Executor Queue: Kubernetes monitoring (ignore if using Docker Compose or server)

#### executor-queue: pods_available_percentage

This code-intel panel indicates percentage pods available.

Refer to the [alert solutions reference](./alert_solutions.md#executor-queue-pods-available-percentage) for relevant alerts.

<br />

## Precise Code Intel Indexer

<p class="subtitle">Executes jobs from the "codeintel" work queue.</p>

### Precise Code Intel Indexer: Executor

#### precise-code-intel-indexer: codeintel_job_99th_percentile_duration

This code-intel panel indicates 99th percentile successful job duration over 5m.


<br />

#### precise-code-intel-indexer: codeintel_active_handlers

This code-intel panel indicates active handlers processing jobs.


<br />

#### precise-code-intel-indexer: codeintel_job_errors

This code-intel panel indicates job errors every 5m.

Refer to the [alert solutions reference](./alert_solutions.md#precise-code-intel-indexer-codeintel-job-errors) for relevant alerts.

<br />

### Precise Code Intel Indexer: Stores and clients

#### precise-code-intel-indexer: executor_apiclient_99th_percentile_duration

This code-intel panel indicates 99th percentile successful API request duration over 5m.

Refer to the [alert solutions reference](./alert_solutions.md#precise-code-intel-indexer-executor-apiclient-99th-percentile-duration) for relevant alerts.

<br />

#### precise-code-intel-indexer: executor_apiclient_errors

This code-intel panel indicates aPI errors every 5m.

Refer to the [alert solutions reference](./alert_solutions.md#precise-code-intel-indexer-executor-apiclient-errors) for relevant alerts.

<br />

### Precise Code Intel Indexer: Commands

#### precise-code-intel-indexer: executor_setup_command_99th_percentile_duration

This code-intel panel indicates 99th percentile successful setup command duration over 5m.


<br />

#### precise-code-intel-indexer: executor_setup_command_errors

This code-intel panel indicates setup command errors every 5m.

Refer to the [alert solutions reference](./alert_solutions.md#precise-code-intel-indexer-executor-setup-command-errors) for relevant alerts.

<br />

#### precise-code-intel-indexer: executor_exec_command_99th_percentile_duration

This code-intel panel indicates 99th percentile successful exec command duration over 5m.


<br />

#### precise-code-intel-indexer: executor_exec_command_errors

This code-intel panel indicates exec command errors every 5m.

Refer to the [alert solutions reference](./alert_solutions.md#precise-code-intel-indexer-executor-exec-command-errors) for relevant alerts.

<br />

#### precise-code-intel-indexer: executor_teardown_command_99th_percentile_duration

This code-intel panel indicates 99th percentile successful teardown command duration over 5m.


<br />

#### precise-code-intel-indexer: executor_teardown_command_errors

This code-intel panel indicates teardown command errors every 5m.

Refer to the [alert solutions reference](./alert_solutions.md#precise-code-intel-indexer-executor-teardown-command-errors) for relevant alerts.

<br />

### Precise Code Intel Indexer: Container monitoring (not available on server)

#### precise-code-intel-indexer: container_cpu_usage

This code-intel panel indicates container cpu usage total (1m average) across all cores by instance.

Refer to the [alert solutions reference](./alert_solutions.md#precise-code-intel-indexer-container-cpu-usage) for relevant alerts.

<br />

#### precise-code-intel-indexer: container_memory_usage

This code-intel panel indicates container memory usage by instance.

Refer to the [alert solutions reference](./alert_solutions.md#precise-code-intel-indexer-container-memory-usage) for relevant alerts.

<br />

#### precise-code-intel-indexer: container_restarts

This code-intel panel indicates container restarts every 5m by instance.

Refer to the [alert solutions reference](./alert_solutions.md#precise-code-intel-indexer-container-restarts) for relevant alerts.

<br />

#### precise-code-intel-indexer: fs_inodes_used

This code-intel panel indicates fs inodes in use by instance.

Refer to the [alert solutions reference](./alert_solutions.md#precise-code-intel-indexer-fs-inodes-used) for relevant alerts.

<br />

### Precise Code Intel Indexer: Provisioning indicators (not available on server)

#### precise-code-intel-indexer: provisioning_container_cpu_usage_long_term

This code-intel panel indicates container cpu usage total (90th percentile over 1d) across all cores by instance.

Refer to the [alert solutions reference](./alert_solutions.md#precise-code-intel-indexer-provisioning-container-cpu-usage-long-term) for relevant alerts.

<br />

#### precise-code-intel-indexer: provisioning_container_memory_usage_long_term

This code-intel panel indicates container memory usage (1d maximum) by instance.

Refer to the [alert solutions reference](./alert_solutions.md#precise-code-intel-indexer-provisioning-container-memory-usage-long-term) for relevant alerts.

<br />

#### precise-code-intel-indexer: provisioning_container_cpu_usage_short_term

This code-intel panel indicates container cpu usage total (5m maximum) across all cores by instance.

Refer to the [alert solutions reference](./alert_solutions.md#precise-code-intel-indexer-provisioning-container-cpu-usage-short-term) for relevant alerts.

<br />

#### precise-code-intel-indexer: provisioning_container_memory_usage_short_term

This code-intel panel indicates container memory usage (5m maximum) by instance.

Refer to the [alert solutions reference](./alert_solutions.md#precise-code-intel-indexer-provisioning-container-memory-usage-short-term) for relevant alerts.

<br />

### Precise Code Intel Indexer: Golang runtime monitoring

#### precise-code-intel-indexer: go_goroutines

This code-intel panel indicates maximum active goroutines.

Refer to the [alert solutions reference](./alert_solutions.md#precise-code-intel-indexer-go-goroutines) for relevant alerts.

<br />

#### precise-code-intel-indexer: go_gc_duration_seconds

This code-intel panel indicates maximum go garbage collection duration.

Refer to the [alert solutions reference](./alert_solutions.md#precise-code-intel-indexer-go-gc-duration-seconds) for relevant alerts.

<br />

### Precise Code Intel Indexer: Kubernetes monitoring (ignore if using Docker Compose or server)

#### precise-code-intel-indexer: pods_available_percentage

This code-intel panel indicates percentage pods available.

Refer to the [alert solutions reference](./alert_solutions.md#precise-code-intel-indexer-pods-available-percentage) for relevant alerts.

<br />

