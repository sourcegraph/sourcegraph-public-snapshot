# Sourcegraph monitoring dashboards

This document contains details on how to intepret panels and metrics in Sourcegraph's monitoring dashboards.

To learn more about Sourcegraph's metrics and how to view these dashboards, see [our metrics documentation](https://docs.sourcegraph.com/admin/observability/metrics).

<!-- DO NOT EDIT: generated via: go generate ./monitoring -->

## Frontend

<p class="subtitle">Serves all end-user browser and API requests.</p>

### frontend: Search at a glance

#### frontend: 99th_percentile_search_request_duration

This panel indicates 99th percentile successful search request duration over 5m (search).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#frontend-99th-percentile-search-request-duration) for relevant alerts.

<br />

#### frontend: 90th_percentile_search_request_duration

This panel indicates 90th percentile successful search request duration over 5m (search).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#frontend-90th-percentile-search-request-duration) for relevant alerts.

<br />

#### frontend: hard_timeout_search_responses

This panel indicates hard timeout search responses every 5m (search).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#frontend-hard-timeout-search-responses) for relevant alerts.

<br />

#### frontend: hard_error_search_responses

This panel indicates hard error search responses every 5m (search).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#frontend-hard-error-search-responses) for relevant alerts.

<br />

#### frontend: partial_timeout_search_responses

This panel indicates partial timeout search responses every 5m (search).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#frontend-partial-timeout-search-responses) for relevant alerts.

<br />

#### frontend: search_alert_user_suggestions

This panel indicates search alert user suggestions shown every 5m (search).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#frontend-search-alert-user-suggestions) for relevant alerts.

<br />

#### frontend: page_load_latency

This panel indicates 90th percentile page load latency over all routes over 10m (cloud).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#frontend-page-load-latency) for relevant alerts.

<br />

#### frontend: blob_load_latency

This panel indicates 90th percentile blob load latency over 10m (cloud).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#frontend-blob-load-latency) for relevant alerts.

<br />

### frontend: Search-based code intelligence at a glance

#### frontend: 99th_percentile_search_codeintel_request_duration

This panel indicates 99th percentile code-intel successful search request duration over 5m (code-intel).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#frontend-99th-percentile-search-codeintel-request-duration) for relevant alerts.

<br />

#### frontend: 90th_percentile_search_codeintel_request_duration

This panel indicates 90th percentile code-intel successful search request duration over 5m (code-intel).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#frontend-90th-percentile-search-codeintel-request-duration) for relevant alerts.

<br />

#### frontend: hard_timeout_search_codeintel_responses

This panel indicates hard timeout search code-intel responses every 5m (code-intel).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#frontend-hard-timeout-search-codeintel-responses) for relevant alerts.

<br />

#### frontend: hard_error_search_codeintel_responses

This panel indicates hard error search code-intel responses every 5m (code-intel).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#frontend-hard-error-search-codeintel-responses) for relevant alerts.

<br />

#### frontend: partial_timeout_search_codeintel_responses

This panel indicates partial timeout search code-intel responses every 5m (code-intel).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#frontend-partial-timeout-search-codeintel-responses) for relevant alerts.

<br />

#### frontend: search_codeintel_alert_user_suggestions

This panel indicates search code-intel alert user suggestions shown every 5m (code-intel).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#frontend-search-codeintel-alert-user-suggestions) for relevant alerts.

<br />

### frontend: Search API usage at a glance

#### frontend: 99th_percentile_search_api_request_duration

This panel indicates 99th percentile successful search API request duration over 5m (search).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#frontend-99th-percentile-search-api-request-duration) for relevant alerts.

<br />

#### frontend: 90th_percentile_search_api_request_duration

This panel indicates 90th percentile successful search API request duration over 5m (search).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#frontend-90th-percentile-search-api-request-duration) for relevant alerts.

<br />

#### frontend: hard_timeout_search_api_responses

This panel indicates hard timeout search API responses every 5m (search).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#frontend-hard-timeout-search-api-responses) for relevant alerts.

<br />

#### frontend: hard_error_search_api_responses

This panel indicates hard error search API responses every 5m (search).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#frontend-hard-error-search-api-responses) for relevant alerts.

<br />

#### frontend: partial_timeout_search_api_responses

This panel indicates partial timeout search API responses every 5m (search).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#frontend-partial-timeout-search-api-responses) for relevant alerts.

<br />

#### frontend: search_api_alert_user_suggestions

This panel indicates search API alert user suggestions shown every 5m (search).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#frontend-search-api-alert-user-suggestions) for relevant alerts.

<br />

### frontend: Precise code intelligence usage at a glance

#### frontend: codeintel_resolvers_99th_percentile_duration

This panel indicates 99th percentile successful resolver duration over 5m (code-intel).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#frontend-codeintel-resolvers-99th-percentile-duration) for relevant alerts.

<br />

#### frontend: codeintel_resolvers_errors

This panel indicates resolver errors every 5m (code-intel).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#frontend-codeintel-resolvers-errors) for relevant alerts.

<br />

#### frontend: codeintel_api_99th_percentile_duration

This panel indicates 99th percentile successful codeintel API operation duration over 5m (code-intel).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#frontend-codeintel-api-99th-percentile-duration) for relevant alerts.

<br />

#### frontend: codeintel_api_errors

This panel indicates code intel API errors every 5m (code-intel).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#frontend-codeintel-api-errors) for relevant alerts.

<br />

### frontend: Precise code intelligence stores and clients

#### frontend: codeintel_dbstore_99th_percentile_duration

This panel indicates 99th percentile successful database store operation duration over 5m (code-intel).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#frontend-codeintel-dbstore-99th-percentile-duration) for relevant alerts.

<br />

#### frontend: codeintel_dbstore_errors

This panel indicates database store errors every 5m (code-intel).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#frontend-codeintel-dbstore-errors) for relevant alerts.

<br />

#### frontend: codeintel_upload_workerstore_99th_percentile_duration

This panel indicates 99th percentile successful upload worker store operation duration over 5m (code-intel).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#frontend-codeintel-upload-workerstore-99th-percentile-duration) for relevant alerts.

<br />

#### frontend: codeintel_upload_workerstore_errors

This panel indicates upload worker store errors every 5m (code-intel).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#frontend-codeintel-upload-workerstore-errors) for relevant alerts.

<br />

#### frontend: codeintel_index_workerstore_99th_percentile_duration

This panel indicates 99th percentile successful index worker store operation duration over 5m (code-intel).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#frontend-codeintel-index-workerstore-99th-percentile-duration) for relevant alerts.

<br />

#### frontend: codeintel_index_workerstore_errors

This panel indicates index worker store errors every 5m (code-intel).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#frontend-codeintel-index-workerstore-errors) for relevant alerts.

<br />

#### frontend: codeintel_lsifstore_99th_percentile_duration

This panel indicates 99th percentile successful LSIF store operation duration over 5m (code-intel).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#frontend-codeintel-lsifstore-99th-percentile-duration) for relevant alerts.

<br />

#### frontend: codeintel_lsifstore_errors

This panel indicates lSIF store errors every 5m (code-intel).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#frontend-codeintel-lsifstore-errors) for relevant alerts.

<br />

#### frontend: codeintel_uploadstore_99th_percentile_duration

This panel indicates 99th percentile successful upload store operation duration over 5m (code-intel).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#frontend-codeintel-uploadstore-99th-percentile-duration) for relevant alerts.

<br />

#### frontend: codeintel_uploadstore_errors

This panel indicates upload store errors every 5m (code-intel).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#frontend-codeintel-uploadstore-errors) for relevant alerts.

<br />

#### frontend: codeintel_gitserverclient_99th_percentile_duration

This panel indicates 99th percentile successful gitserver client operation duration over 5m (code-intel).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#frontend-codeintel-gitserverclient-99th-percentile-duration) for relevant alerts.

<br />

#### frontend: codeintel_gitserverclient_errors

This panel indicates gitserver client errors every 5m (code-intel).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#frontend-codeintel-gitserverclient-errors) for relevant alerts.

<br />

### frontend: Precise code intelligence commit graph updater

#### frontend: codeintel_commit_graph_queue_size

This panel indicates commit graph queue size (code-intel).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#frontend-codeintel-commit-graph-queue-size) for relevant alerts.

<br />

#### frontend: codeintel_commit_graph_queue_growth_rate

This panel indicates commit graph queue growth rate over 30m (code-intel).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#frontend-codeintel-commit-graph-queue-growth-rate) for relevant alerts.

<br />

#### frontend: codeintel_commit_graph_updater_99th_percentile_duration

This panel indicates 99th percentile successful commit graph updater operation duration over 5m (code-intel).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#frontend-codeintel-commit-graph-updater-99th-percentile-duration) for relevant alerts.

<br />

#### frontend: codeintel_commit_graph_updater_errors

This panel indicates commit graph updater errors every 5m (code-intel).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#frontend-codeintel-commit-graph-updater-errors) for relevant alerts.

<br />

### frontend: Precise code intelligence janitor

#### frontend: codeintel_janitor_errors

This panel indicates janitor errors every 5m (code-intel).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#frontend-codeintel-janitor-errors) for relevant alerts.

<br />

#### frontend: codeintel_upload_records_removed

This panel indicates upload records expired or deleted every 5m (code-intel).

<br />

#### frontend: codeintel_index_records_removed

This panel indicates index records expired or deleted every 5m (code-intel).

<br />

#### frontend: codeintel_lsif_data_removed

This panel indicates data for unreferenced upload records removed every 5m (code-intel).

<br />

#### frontend: codeintel_background_upload_resets

This panel indicates upload records re-queued (due to unresponsive worker) every 5m (code-intel).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#frontend-codeintel-background-upload-resets) for relevant alerts.

<br />

#### frontend: codeintel_background_upload_reset_failures

This panel indicates upload records errored due to repeated reset every 5m (code-intel).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#frontend-codeintel-background-upload-reset-failures) for relevant alerts.

<br />

#### frontend: codeintel_background_index_resets

This panel indicates index records re-queued (due to unresponsive indexer) every 5m (code-intel).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#frontend-codeintel-background-index-resets) for relevant alerts.

<br />

#### frontend: codeintel_background_index_reset_failures

This panel indicates index records errored due to repeated reset every 5m (code-intel).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#frontend-codeintel-background-index-reset-failures) for relevant alerts.

<br />

### frontend: Auto-indexing

#### frontend: codeintel_indexing_99th_percentile_duration

This panel indicates 99th percentile successful indexing operation duration over 5m (code-intel).

<br />

#### frontend: codeintel_indexing_errors

This panel indicates indexing errors every 5m (code-intel).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#frontend-codeintel-indexing-errors) for relevant alerts.

<br />

### frontend: Internal service requests

#### frontend: internal_indexed_search_error_responses

This panel indicates internal indexed search error responses every 5m (search).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#frontend-internal-indexed-search-error-responses) for relevant alerts.

<br />

#### frontend: internal_unindexed_search_error_responses

This panel indicates internal unindexed search error responses every 5m (search).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#frontend-internal-unindexed-search-error-responses) for relevant alerts.

<br />

#### frontend: internal_api_error_responses

This panel indicates internal API error responses every 5m by route (cloud).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#frontend-internal-api-error-responses) for relevant alerts.

<br />

#### frontend: 99th_percentile_gitserver_duration

This panel indicates 99th percentile successful gitserver query duration over 5m (cloud).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#frontend-99th-percentile-gitserver-duration) for relevant alerts.

<br />

#### frontend: gitserver_error_responses

This panel indicates gitserver error responses every 5m (cloud).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#frontend-gitserver-error-responses) for relevant alerts.

<br />

#### frontend: observability_test_alert_warning

This panel indicates warning test alert metric (distribution).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#frontend-observability-test-alert-warning) for relevant alerts.

<br />

#### frontend: observability_test_alert_critical

This panel indicates critical test alert metric (distribution).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#frontend-observability-test-alert-critical) for relevant alerts.

<br />

### frontend: Container monitoring (not available on server)

#### frontend: container_cpu_usage

This panel indicates container cpu usage total (1m average) across all cores by instance (cloud).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#frontend-container-cpu-usage) for relevant alerts.

<br />

#### frontend: container_memory_usage

This panel indicates container memory usage by instance (cloud).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#frontend-container-memory-usage) for relevant alerts.

<br />

#### frontend: container_restarts

This panel indicates container restarts every 5m by instance (cloud).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#frontend-container-restarts) for relevant alerts.

<br />

#### frontend: fs_inodes_used

This panel indicates fs inodes in use by instance (cloud).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#frontend-fs-inodes-used) for relevant alerts.

<br />

### frontend: Provisioning indicators (not available on server)

#### frontend: provisioning_container_cpu_usage_long_term

This panel indicates container cpu usage total (90th percentile over 1d) across all cores by instance (cloud).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#frontend-provisioning-container-cpu-usage-long-term) for relevant alerts.

<br />

#### frontend: provisioning_container_memory_usage_long_term

This panel indicates container memory usage (1d maximum) by instance (cloud).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#frontend-provisioning-container-memory-usage-long-term) for relevant alerts.

<br />

#### frontend: provisioning_container_cpu_usage_short_term

This panel indicates container cpu usage total (5m maximum) across all cores by instance (cloud).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#frontend-provisioning-container-cpu-usage-short-term) for relevant alerts.

<br />

#### frontend: provisioning_container_memory_usage_short_term

This panel indicates container memory usage (5m maximum) by instance (cloud).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#frontend-provisioning-container-memory-usage-short-term) for relevant alerts.

<br />

### frontend: Golang runtime monitoring

#### frontend: go_goroutines

This panel indicates maximum active goroutines (cloud).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#frontend-go-goroutines) for relevant alerts.

<br />

#### frontend: go_gc_duration_seconds

This panel indicates maximum go garbage collection duration (cloud).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#frontend-go-gc-duration-seconds) for relevant alerts.

<br />

### frontend: Kubernetes monitoring (ignore if using Docker Compose or server)

#### frontend: pods_available_percentage

This panel indicates percentage pods available (cloud).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#frontend-pods-available-percentage) for relevant alerts.

<br />

## Git Server

<p class="subtitle">Stores, manages, and operates Git repositories.</p>

#### gitserver: disk_space_remaining

This panel indicates disk space remaining by instance (cloud).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#gitserver-disk-space-remaining) for relevant alerts.

<br />

#### gitserver: running_git_commands

This panel indicates running git commands (cloud).

A high value signals load.

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#gitserver-running-git-commands) for relevant alerts.

<br />

#### gitserver: repository_clone_queue_size

This panel indicates repository clone queue size (cloud).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#gitserver-repository-clone-queue-size) for relevant alerts.

<br />

#### gitserver: repository_existence_check_queue_size

This panel indicates repository existence check queue size (cloud).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#gitserver-repository-existence-check-queue-size) for relevant alerts.

<br />

#### gitserver: echo_command_duration_test

This panel indicates echo test command duration (cloud).

A high value here is likely to indicate a problem, especially if consistently high.
You can query for individual commands using `sum by (cmd)(src_gitserver_exec_running)` in Grafana (`/-/debug/grafana`) to see if a specific Git Server command might be spiking in frequency.

If this value is consistently high, consider the following:

- **Single container deployments:** Upgrade to a [Docker Compose deployment](../install/docker-compose/migrate.md) which offers better scalability and resource isolation.
- **Kubernetes and Docker Compose:** Check that you are running a similar number of git server replicas and that their CPU/memory limits are allocated according to what is shown in the [Sourcegraph resource estimator](../install/resource_estimator.md).

<br />

#### gitserver: frontend_internal_api_error_responses

This panel indicates frontend-internal API error responses every 5m by route (cloud).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#gitserver-frontend-internal-api-error-responses) for relevant alerts.

<br />

### gitserver: Container monitoring (not available on server)

#### gitserver: container_cpu_usage

This panel indicates container cpu usage total (1m average) across all cores by instance (cloud).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#gitserver-container-cpu-usage) for relevant alerts.

<br />

#### gitserver: container_memory_usage

This panel indicates container memory usage by instance (cloud).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#gitserver-container-memory-usage) for relevant alerts.

<br />

#### gitserver: container_restarts

This panel indicates container restarts every 5m by instance (cloud).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#gitserver-container-restarts) for relevant alerts.

<br />

#### gitserver: fs_inodes_used

This panel indicates fs inodes in use by instance (cloud).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#gitserver-fs-inodes-used) for relevant alerts.

<br />

#### gitserver: fs_io_operations

This panel indicates filesystem reads and writes rate by instance over 1h (cloud).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#gitserver-fs-io-operations) for relevant alerts.

<br />

### gitserver: Provisioning indicators (not available on server)

#### gitserver: provisioning_container_cpu_usage_long_term

This panel indicates container cpu usage total (90th percentile over 1d) across all cores by instance (cloud).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#gitserver-provisioning-container-cpu-usage-long-term) for relevant alerts.

<br />

#### gitserver: provisioning_container_memory_usage_long_term

This panel indicates container memory usage (1d maximum) by instance (cloud).

Git Server is expected to use up all the memory it is provided.

<br />

#### gitserver: provisioning_container_cpu_usage_short_term

This panel indicates container cpu usage total (5m maximum) across all cores by instance (cloud).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#gitserver-provisioning-container-cpu-usage-short-term) for relevant alerts.

<br />

#### gitserver: provisioning_container_memory_usage_short_term

This panel indicates container memory usage (5m maximum) by instance (cloud).

Git Server is expected to use up all the memory it is provided.

<br />

### gitserver: Golang runtime monitoring

#### gitserver: go_goroutines

This panel indicates maximum active goroutines (cloud).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#gitserver-go-goroutines) for relevant alerts.

<br />

#### gitserver: go_gc_duration_seconds

This panel indicates maximum go garbage collection duration (cloud).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#gitserver-go-gc-duration-seconds) for relevant alerts.

<br />

### gitserver: Kubernetes monitoring (ignore if using Docker Compose or server)

#### gitserver: pods_available_percentage

This panel indicates percentage pods available (cloud).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#gitserver-pods-available-percentage) for relevant alerts.

<br />

## GitHub Proxy

<p class="subtitle">Proxies all requests to github.com, keeping track of and managing rate limits.</p>

### github-proxy: GitHub API monitoring

#### github-proxy: github_proxy_waiting_requests

This panel indicates number of requests waiting on the global mutex (cloud).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#github-proxy-github-proxy-waiting-requests) for relevant alerts.

<br />

### github-proxy: Container monitoring (not available on server)

#### github-proxy: container_cpu_usage

This panel indicates container cpu usage total (1m average) across all cores by instance (cloud).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#github-proxy-container-cpu-usage) for relevant alerts.

<br />

#### github-proxy: container_memory_usage

This panel indicates container memory usage by instance (cloud).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#github-proxy-container-memory-usage) for relevant alerts.

<br />

#### github-proxy: container_restarts

This panel indicates container restarts every 5m by instance (cloud).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#github-proxy-container-restarts) for relevant alerts.

<br />

#### github-proxy: fs_inodes_used

This panel indicates fs inodes in use by instance (cloud).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#github-proxy-fs-inodes-used) for relevant alerts.

<br />

### github-proxy: Provisioning indicators (not available on server)

#### github-proxy: provisioning_container_cpu_usage_long_term

This panel indicates container cpu usage total (90th percentile over 1d) across all cores by instance (cloud).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#github-proxy-provisioning-container-cpu-usage-long-term) for relevant alerts.

<br />

#### github-proxy: provisioning_container_memory_usage_long_term

This panel indicates container memory usage (1d maximum) by instance (cloud).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#github-proxy-provisioning-container-memory-usage-long-term) for relevant alerts.

<br />

#### github-proxy: provisioning_container_cpu_usage_short_term

This panel indicates container cpu usage total (5m maximum) across all cores by instance (cloud).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#github-proxy-provisioning-container-cpu-usage-short-term) for relevant alerts.

<br />

#### github-proxy: provisioning_container_memory_usage_short_term

This panel indicates container memory usage (5m maximum) by instance (cloud).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#github-proxy-provisioning-container-memory-usage-short-term) for relevant alerts.

<br />

### github-proxy: Golang runtime monitoring

#### github-proxy: go_goroutines

This panel indicates maximum active goroutines (cloud).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#github-proxy-go-goroutines) for relevant alerts.

<br />

#### github-proxy: go_gc_duration_seconds

This panel indicates maximum go garbage collection duration (cloud).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#github-proxy-go-gc-duration-seconds) for relevant alerts.

<br />

### github-proxy: Kubernetes monitoring (ignore if using Docker Compose or server)

#### github-proxy: pods_available_percentage

This panel indicates percentage pods available (cloud).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#github-proxy-pods-available-percentage) for relevant alerts.

<br />

## Postgres

<p class="subtitle">Metrics from postgres_exporter.</p>

### postgres: Default postgres dashboard

#### postgres: connections

This panel indicates connections (cloud).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#postgres-connections) for relevant alerts.

<br />

#### postgres: transactions

This panel indicates transaction durations (cloud).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#postgres-transactions) for relevant alerts.

<br />

### postgres: Database and collector status

#### postgres: postgres_up

This panel indicates current db status (cloud).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#postgres-postgres-up) for relevant alerts.

<br />

#### postgres: pg_exporter_err

This panel indicates errors scraping postgres exporter (cloud).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#postgres-pg-exporter-err) for relevant alerts.

<br />

#### postgres: migration_in_progress

This panel indicates schema migration status (where 0 is no migration in progress) (cloud).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#postgres-migration-in-progress) for relevant alerts.

<br />

### postgres: Provisioning indicators (not available on server)

#### postgres: provisioning_container_cpu_usage_long_term

This panel indicates container cpu usage total (90th percentile over 1d) across all cores by instance (cloud).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#postgres-provisioning-container-cpu-usage-long-term) for relevant alerts.

<br />

#### postgres: provisioning_container_memory_usage_long_term

This panel indicates container memory usage (1d maximum) by instance (cloud).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#postgres-provisioning-container-memory-usage-long-term) for relevant alerts.

<br />

#### postgres: provisioning_container_cpu_usage_short_term

This panel indicates container cpu usage total (5m maximum) across all cores by instance (cloud).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#postgres-provisioning-container-cpu-usage-short-term) for relevant alerts.

<br />

#### postgres: provisioning_container_memory_usage_short_term

This panel indicates container memory usage (5m maximum) by instance (cloud).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#postgres-provisioning-container-memory-usage-short-term) for relevant alerts.

<br />

#### postgres: provisioning_container_cpu_usage_long_term

This panel indicates container cpu usage total (90th percentile over 1d) across all cores by instance (code-intel).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#postgres-provisioning-container-cpu-usage-long-term) for relevant alerts.

<br />

#### postgres: provisioning_container_memory_usage_long_term

This panel indicates container memory usage (1d maximum) by instance (code-intel).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#postgres-provisioning-container-memory-usage-long-term) for relevant alerts.

<br />

#### postgres: provisioning_container_cpu_usage_short_term

This panel indicates container cpu usage total (5m maximum) across all cores by instance (code-intel).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#postgres-provisioning-container-cpu-usage-short-term) for relevant alerts.

<br />

#### postgres: provisioning_container_memory_usage_short_term

This panel indicates container memory usage (5m maximum) by instance (code-intel).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#postgres-provisioning-container-memory-usage-short-term) for relevant alerts.

<br />

## Precise Code Intel Worker

<p class="subtitle">Handles conversion of uploaded precise code intelligence bundles.</p>

### precise-code-intel-worker: Upload queue

#### precise-code-intel-worker: upload_queue_size

This panel indicates queue size (code-intel).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#precise-code-intel-worker-upload-queue-size) for relevant alerts.

<br />

#### precise-code-intel-worker: upload_queue_growth_rate

This panel indicates queue growth rate over 30m (code-intel).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#precise-code-intel-worker-upload-queue-growth-rate) for relevant alerts.

<br />

#### precise-code-intel-worker: job_errors

This panel indicates job errors errors every 5m (code-intel).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#precise-code-intel-worker-job-errors) for relevant alerts.

<br />

#### precise-code-intel-worker: active_workers

This panel indicates active workers processing uploads (code-intel).

<br />

#### precise-code-intel-worker: active_jobs

This panel indicates active jobs (code-intel).

<br />

### precise-code-intel-worker: Workers

#### precise-code-intel-worker: job_99th_percentile_duration

This panel indicates 99th percentile successful job duration over 5m (code-intel).

<br />

### precise-code-intel-worker: Stores and clients

#### precise-code-intel-worker: codeintel_dbstore_99th_percentile_duration

This panel indicates 99th percentile successful database store operation duration over 5m (code-intel).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#precise-code-intel-worker-codeintel-dbstore-99th-percentile-duration) for relevant alerts.

<br />

#### precise-code-intel-worker: codeintel_dbstore_errors

This panel indicates database store errors every 5m (code-intel).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#precise-code-intel-worker-codeintel-dbstore-errors) for relevant alerts.

<br />

#### precise-code-intel-worker: codeintel_workerstore_99th_percentile_duration

This panel indicates 99th percentile successful worker store operation duration over 5m (code-intel).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#precise-code-intel-worker-codeintel-workerstore-99th-percentile-duration) for relevant alerts.

<br />

#### precise-code-intel-worker: codeintel_workerstore_errors

This panel indicates worker store errors every 5m (code-intel).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#precise-code-intel-worker-codeintel-workerstore-errors) for relevant alerts.

<br />

#### precise-code-intel-worker: codeintel_lsifstore_99th_percentile_duration

This panel indicates 99th percentile successful LSIF store operation duration over 5m (code-intel).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#precise-code-intel-worker-codeintel-lsifstore-99th-percentile-duration) for relevant alerts.

<br />

#### precise-code-intel-worker: codeintel_lsifstore_errors

This panel indicates lSIF store errors every 5m (code-intel).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#precise-code-intel-worker-codeintel-lsifstore-errors) for relevant alerts.

<br />

#### precise-code-intel-worker: codeintel_uploadstore_99th_percentile_duration

This panel indicates 99th percentile successful upload store operation duration over 5m (code-intel).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#precise-code-intel-worker-codeintel-uploadstore-99th-percentile-duration) for relevant alerts.

<br />

#### precise-code-intel-worker: codeintel_uploadstore_errors

This panel indicates upload store errors every 5m (code-intel).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#precise-code-intel-worker-codeintel-uploadstore-errors) for relevant alerts.

<br />

#### precise-code-intel-worker: codeintel_gitserverclient_99th_percentile_duration

This panel indicates 99th percentile successful gitserver client operation duration over 5m (code-intel).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#precise-code-intel-worker-codeintel-gitserverclient-99th-percentile-duration) for relevant alerts.

<br />

#### precise-code-intel-worker: codeintel_gitserverclient_errors

This panel indicates gitserver client errors every 5m (code-intel).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#precise-code-intel-worker-codeintel-gitserverclient-errors) for relevant alerts.

<br />

### precise-code-intel-worker: Internal service requests

#### precise-code-intel-worker: frontend_internal_api_error_responses

This panel indicates frontend-internal API error responses every 5m by route (code-intel).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#precise-code-intel-worker-frontend-internal-api-error-responses) for relevant alerts.

<br />

### precise-code-intel-worker: Container monitoring (not available on server)

#### precise-code-intel-worker: container_cpu_usage

This panel indicates container cpu usage total (1m average) across all cores by instance (code-intel).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#precise-code-intel-worker-container-cpu-usage) for relevant alerts.

<br />

#### precise-code-intel-worker: container_memory_usage

This panel indicates container memory usage by instance (code-intel).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#precise-code-intel-worker-container-memory-usage) for relevant alerts.

<br />

#### precise-code-intel-worker: container_restarts

This panel indicates container restarts every 5m by instance (code-intel).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#precise-code-intel-worker-container-restarts) for relevant alerts.

<br />

#### precise-code-intel-worker: fs_inodes_used

This panel indicates fs inodes in use by instance (code-intel).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#precise-code-intel-worker-fs-inodes-used) for relevant alerts.

<br />

### precise-code-intel-worker: Provisioning indicators (not available on server)

#### precise-code-intel-worker: provisioning_container_cpu_usage_long_term

This panel indicates container cpu usage total (90th percentile over 1d) across all cores by instance (code-intel).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#precise-code-intel-worker-provisioning-container-cpu-usage-long-term) for relevant alerts.

<br />

#### precise-code-intel-worker: provisioning_container_memory_usage_long_term

This panel indicates container memory usage (1d maximum) by instance (code-intel).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#precise-code-intel-worker-provisioning-container-memory-usage-long-term) for relevant alerts.

<br />

#### precise-code-intel-worker: provisioning_container_cpu_usage_short_term

This panel indicates container cpu usage total (5m maximum) across all cores by instance (code-intel).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#precise-code-intel-worker-provisioning-container-cpu-usage-short-term) for relevant alerts.

<br />

#### precise-code-intel-worker: provisioning_container_memory_usage_short_term

This panel indicates container memory usage (5m maximum) by instance (code-intel).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#precise-code-intel-worker-provisioning-container-memory-usage-short-term) for relevant alerts.

<br />

### precise-code-intel-worker: Golang runtime monitoring

#### precise-code-intel-worker: go_goroutines

This panel indicates maximum active goroutines (code-intel).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#precise-code-intel-worker-go-goroutines) for relevant alerts.

<br />

#### precise-code-intel-worker: go_gc_duration_seconds

This panel indicates maximum go garbage collection duration (code-intel).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#precise-code-intel-worker-go-gc-duration-seconds) for relevant alerts.

<br />

### precise-code-intel-worker: Kubernetes monitoring (ignore if using Docker Compose or server)

#### precise-code-intel-worker: pods_available_percentage

This panel indicates percentage pods available (code-intel).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#precise-code-intel-worker-pods-available-percentage) for relevant alerts.

<br />

## Query Runner

<p class="subtitle">Periodically runs saved searches and instructs the frontend to send out notifications.</p>

#### query-runner: frontend_internal_api_error_responses

This panel indicates frontend-internal API error responses every 5m by route (search).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#query-runner-frontend-internal-api-error-responses) for relevant alerts.

<br />

### query-runner: Container monitoring (not available on server)

#### query-runner: container_memory_usage

This panel indicates container memory usage by instance (search).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#query-runner-container-memory-usage) for relevant alerts.

<br />

#### query-runner: container_cpu_usage

This panel indicates container cpu usage total (1m average) across all cores by instance (search).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#query-runner-container-cpu-usage) for relevant alerts.

<br />

#### query-runner: container_restarts

This panel indicates container restarts every 5m by instance (search).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#query-runner-container-restarts) for relevant alerts.

<br />

#### query-runner: fs_inodes_used

This panel indicates fs inodes in use by instance (search).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#query-runner-fs-inodes-used) for relevant alerts.

<br />

### query-runner: Provisioning indicators (not available on server)

#### query-runner: provisioning_container_cpu_usage_long_term

This panel indicates container cpu usage total (90th percentile over 1d) across all cores by instance (search).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#query-runner-provisioning-container-cpu-usage-long-term) for relevant alerts.

<br />

#### query-runner: provisioning_container_memory_usage_long_term

This panel indicates container memory usage (1d maximum) by instance (search).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#query-runner-provisioning-container-memory-usage-long-term) for relevant alerts.

<br />

#### query-runner: provisioning_container_cpu_usage_short_term

This panel indicates container cpu usage total (5m maximum) across all cores by instance (search).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#query-runner-provisioning-container-cpu-usage-short-term) for relevant alerts.

<br />

#### query-runner: provisioning_container_memory_usage_short_term

This panel indicates container memory usage (5m maximum) by instance (search).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#query-runner-provisioning-container-memory-usage-short-term) for relevant alerts.

<br />

### query-runner: Golang runtime monitoring

#### query-runner: go_goroutines

This panel indicates maximum active goroutines (search).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#query-runner-go-goroutines) for relevant alerts.

<br />

#### query-runner: go_gc_duration_seconds

This panel indicates maximum go garbage collection duration (search).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#query-runner-go-gc-duration-seconds) for relevant alerts.

<br />

### query-runner: Kubernetes monitoring (ignore if using Docker Compose or server)

#### query-runner: pods_available_percentage

This panel indicates percentage pods available (search).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#query-runner-pods-available-percentage) for relevant alerts.

<br />

## Repo Updater

<p class="subtitle">Manages interaction with code hosts, instructs Gitserver to update repositories.</p>

#### repo-updater: frontend_internal_api_error_responses

This panel indicates frontend-internal API error responses every 5m by route (cloud).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#repo-updater-frontend-internal-api-error-responses) for relevant alerts.

<br />

### repo-updater: Repositories

#### repo-updater: syncer_sync_last_time

This panel indicates time since last sync (cloud).

A high value here indicates issues synchronizing repository permissions.
If the value is persistently high, make sure all external services have valid tokens.

<br />

#### repo-updater: src_repoupdater_max_sync_backoff

This panel indicates time since oldest sync (cloud).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#repo-updater-src-repoupdater-max-sync-backoff) for relevant alerts.

<br />

#### repo-updater: src_repoupdater_syncer_sync_errors_total

This panel indicates sync error rate (cloud).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#repo-updater-src-repoupdater-syncer-sync-errors-total) for relevant alerts.

<br />

#### repo-updater: syncer_sync_start

This panel indicates sync was started (cloud).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#repo-updater-syncer-sync-start) for relevant alerts.

<br />

#### repo-updater: syncer_sync_duration

This panel indicates 95th repositories sync duration (cloud).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#repo-updater-syncer-sync-duration) for relevant alerts.

<br />

#### repo-updater: source_duration

This panel indicates 95th repositories source duration (cloud).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#repo-updater-source-duration) for relevant alerts.

<br />

#### repo-updater: syncer_synced_repos

This panel indicates repositories synced (cloud).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#repo-updater-syncer-synced-repos) for relevant alerts.

<br />

#### repo-updater: sourced_repos

This panel indicates repositories sourced (cloud).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#repo-updater-sourced-repos) for relevant alerts.

<br />

#### repo-updater: user_added_repos

This panel indicates total number of user added repos (cloud).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#repo-updater-user-added-repos) for relevant alerts.

<br />

#### repo-updater: purge_failed

This panel indicates repositories purge failed (cloud).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#repo-updater-purge-failed) for relevant alerts.

<br />

#### repo-updater: sched_auto_fetch

This panel indicates repositories scheduled due to hitting a deadline (cloud).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#repo-updater-sched-auto-fetch) for relevant alerts.

<br />

#### repo-updater: sched_manual_fetch

This panel indicates repositories scheduled due to user traffic (cloud).

Check repo-updater logs if this value is persistently high.
This does not indicate anything if there are no user added code hosts.

<br />

#### repo-updater: sched_known_repos

This panel indicates repositories managed by the scheduler (cloud).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#repo-updater-sched-known-repos) for relevant alerts.

<br />

#### repo-updater: sched_update_queue_length

This panel indicates rate of growth of update queue length over 5 minutes (cloud).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#repo-updater-sched-update-queue-length) for relevant alerts.

<br />

#### repo-updater: sched_loops

This panel indicates scheduler loops (cloud).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#repo-updater-sched-loops) for relevant alerts.

<br />

#### repo-updater: sched_error

This panel indicates repositories schedule error rate (cloud).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#repo-updater-sched-error) for relevant alerts.

<br />

### repo-updater: Permissions

#### repo-updater: perms_syncer_perms

This panel indicates time gap between least and most up to date permissions (cloud).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#repo-updater-perms-syncer-perms) for relevant alerts.

<br />

#### repo-updater: perms_syncer_stale_perms

This panel indicates number of entities with stale permissions (cloud).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#repo-updater-perms-syncer-stale-perms) for relevant alerts.

<br />

#### repo-updater: perms_syncer_no_perms

This panel indicates number of entities with no permissions (cloud).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#repo-updater-perms-syncer-no-perms) for relevant alerts.

<br />

#### repo-updater: perms_syncer_sync_duration

This panel indicates 95th permissions sync duration (cloud).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#repo-updater-perms-syncer-sync-duration) for relevant alerts.

<br />

#### repo-updater: perms_syncer_queue_size

This panel indicates permissions sync queued items (cloud).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#repo-updater-perms-syncer-queue-size) for relevant alerts.

<br />

#### repo-updater: perms_syncer_sync_errors

This panel indicates permissions sync error rate (cloud).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#repo-updater-perms-syncer-sync-errors) for relevant alerts.

<br />

### repo-updater: External services

#### repo-updater: src_repoupdater_external_services_total

This panel indicates the total number of external services (cloud).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#repo-updater-src-repoupdater-external-services-total) for relevant alerts.

<br />

#### repo-updater: src_repoupdater_user_external_services_total

This panel indicates the total number of user added external services (cloud).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#repo-updater-src-repoupdater-user-external-services-total) for relevant alerts.

<br />

#### repo-updater: repoupdater_queued_sync_jobs_total

This panel indicates the total number of queued sync jobs (cloud).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#repo-updater-repoupdater-queued-sync-jobs-total) for relevant alerts.

<br />

#### repo-updater: repoupdater_completed_sync_jobs_total

This panel indicates the total number of completed sync jobs (cloud).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#repo-updater-repoupdater-completed-sync-jobs-total) for relevant alerts.

<br />

#### repo-updater: repoupdater_errored_sync_jobs_total

This panel indicates the total number of errored sync jobs (cloud).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#repo-updater-repoupdater-errored-sync-jobs-total) for relevant alerts.

<br />

#### repo-updater: github_graphql_rate_limit_remaining

This panel indicates remaining calls to GitHub graphql API before hitting the rate limit (cloud).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#repo-updater-github-graphql-rate-limit-remaining) for relevant alerts.

<br />

#### repo-updater: github_rest_rate_limit_remaining

This panel indicates remaining calls to GitHub rest API before hitting the rate limit (cloud).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#repo-updater-github-rest-rate-limit-remaining) for relevant alerts.

<br />

#### repo-updater: github_search_rate_limit_remaining

This panel indicates remaining calls to GitHub search API before hitting the rate limit (cloud).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#repo-updater-github-search-rate-limit-remaining) for relevant alerts.

<br />

#### repo-updater: gitlab_rest_rate_limit_remaining

This panel indicates remaining calls to GitLab rest API before hitting the rate limit (cloud).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#repo-updater-gitlab-rest-rate-limit-remaining) for relevant alerts.

<br />

### repo-updater: Container monitoring (not available on server)

#### repo-updater: container_cpu_usage

This panel indicates container cpu usage total (1m average) across all cores by instance (cloud).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#repo-updater-container-cpu-usage) for relevant alerts.

<br />

#### repo-updater: container_memory_usage

This panel indicates container memory usage by instance (cloud).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#repo-updater-container-memory-usage) for relevant alerts.

<br />

#### repo-updater: container_restarts

This panel indicates container restarts every 5m by instance (cloud).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#repo-updater-container-restarts) for relevant alerts.

<br />

#### repo-updater: fs_inodes_used

This panel indicates fs inodes in use by instance (cloud).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#repo-updater-fs-inodes-used) for relevant alerts.

<br />

### repo-updater: Provisioning indicators (not available on server)

#### repo-updater: provisioning_container_cpu_usage_long_term

This panel indicates container cpu usage total (90th percentile over 1d) across all cores by instance (cloud).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#repo-updater-provisioning-container-cpu-usage-long-term) for relevant alerts.

<br />

#### repo-updater: provisioning_container_memory_usage_long_term

This panel indicates container memory usage (1d maximum) by instance (cloud).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#repo-updater-provisioning-container-memory-usage-long-term) for relevant alerts.

<br />

#### repo-updater: provisioning_container_cpu_usage_short_term

This panel indicates container cpu usage total (5m maximum) across all cores by instance (cloud).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#repo-updater-provisioning-container-cpu-usage-short-term) for relevant alerts.

<br />

#### repo-updater: provisioning_container_memory_usage_short_term

This panel indicates container memory usage (5m maximum) by instance (cloud).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#repo-updater-provisioning-container-memory-usage-short-term) for relevant alerts.

<br />

### repo-updater: Golang runtime monitoring

#### repo-updater: go_goroutines

This panel indicates maximum active goroutines (cloud).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#repo-updater-go-goroutines) for relevant alerts.

<br />

#### repo-updater: go_gc_duration_seconds

This panel indicates maximum go garbage collection duration (cloud).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#repo-updater-go-gc-duration-seconds) for relevant alerts.

<br />

### repo-updater: Kubernetes monitoring (ignore if using Docker Compose or server)

#### repo-updater: pods_available_percentage

This panel indicates percentage pods available (cloud).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#repo-updater-pods-available-percentage) for relevant alerts.

<br />

## Searcher

<p class="subtitle">Performs unindexed searches (diff and commit search, text search for unindexed branches).</p>

#### searcher: unindexed_search_request_errors

This panel indicates unindexed search request errors every 5m by code (search).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#searcher-unindexed-search-request-errors) for relevant alerts.

<br />

#### searcher: replica_traffic

This panel indicates requests per second over 10m (search).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#searcher-replica-traffic) for relevant alerts.

<br />

#### searcher: frontend_internal_api_error_responses

This panel indicates frontend-internal API error responses every 5m by route (search).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#searcher-frontend-internal-api-error-responses) for relevant alerts.

<br />

### searcher: Container monitoring (not available on server)

#### searcher: container_cpu_usage

This panel indicates container cpu usage total (1m average) across all cores by instance (search).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#searcher-container-cpu-usage) for relevant alerts.

<br />

#### searcher: container_memory_usage

This panel indicates container memory usage by instance (search).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#searcher-container-memory-usage) for relevant alerts.

<br />

#### searcher: container_restarts

This panel indicates container restarts every 5m by instance (search).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#searcher-container-restarts) for relevant alerts.

<br />

#### searcher: fs_inodes_used

This panel indicates fs inodes in use by instance (search).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#searcher-fs-inodes-used) for relevant alerts.

<br />

### searcher: Provisioning indicators (not available on server)

#### searcher: provisioning_container_cpu_usage_long_term

This panel indicates container cpu usage total (90th percentile over 1d) across all cores by instance (search).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#searcher-provisioning-container-cpu-usage-long-term) for relevant alerts.

<br />

#### searcher: provisioning_container_memory_usage_long_term

This panel indicates container memory usage (1d maximum) by instance (search).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#searcher-provisioning-container-memory-usage-long-term) for relevant alerts.

<br />

#### searcher: provisioning_container_cpu_usage_short_term

This panel indicates container cpu usage total (5m maximum) across all cores by instance (search).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#searcher-provisioning-container-cpu-usage-short-term) for relevant alerts.

<br />

#### searcher: provisioning_container_memory_usage_short_term

This panel indicates container memory usage (5m maximum) by instance (search).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#searcher-provisioning-container-memory-usage-short-term) for relevant alerts.

<br />

### searcher: Golang runtime monitoring

#### searcher: go_goroutines

This panel indicates maximum active goroutines (search).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#searcher-go-goroutines) for relevant alerts.

<br />

#### searcher: go_gc_duration_seconds

This panel indicates maximum go garbage collection duration (search).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#searcher-go-gc-duration-seconds) for relevant alerts.

<br />

### searcher: Kubernetes monitoring (ignore if using Docker Compose or server)

#### searcher: pods_available_percentage

This panel indicates percentage pods available (search).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#searcher-pods-available-percentage) for relevant alerts.

<br />

## Symbols

<p class="subtitle">Handles symbol searches for unindexed branches.</p>

#### symbols: store_fetch_failures

This panel indicates store fetch failures every 5m (code-intel).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#symbols-store-fetch-failures) for relevant alerts.

<br />

#### symbols: current_fetch_queue_size

This panel indicates current fetch queue size (code-intel).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#symbols-current-fetch-queue-size) for relevant alerts.

<br />

#### symbols: frontend_internal_api_error_responses

This panel indicates frontend-internal API error responses every 5m by route (code-intel).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#symbols-frontend-internal-api-error-responses) for relevant alerts.

<br />

### symbols: Container monitoring (not available on server)

#### symbols: container_cpu_usage

This panel indicates container cpu usage total (1m average) across all cores by instance (code-intel).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#symbols-container-cpu-usage) for relevant alerts.

<br />

#### symbols: container_memory_usage

This panel indicates container memory usage by instance (code-intel).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#symbols-container-memory-usage) for relevant alerts.

<br />

#### symbols: container_restarts

This panel indicates container restarts every 5m by instance (code-intel).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#symbols-container-restarts) for relevant alerts.

<br />

#### symbols: fs_inodes_used

This panel indicates fs inodes in use by instance (code-intel).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#symbols-fs-inodes-used) for relevant alerts.

<br />

### symbols: Provisioning indicators (not available on server)

#### symbols: provisioning_container_cpu_usage_long_term

This panel indicates container cpu usage total (90th percentile over 1d) across all cores by instance (code-intel).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#symbols-provisioning-container-cpu-usage-long-term) for relevant alerts.

<br />

#### symbols: provisioning_container_memory_usage_long_term

This panel indicates container memory usage (1d maximum) by instance (code-intel).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#symbols-provisioning-container-memory-usage-long-term) for relevant alerts.

<br />

#### symbols: provisioning_container_cpu_usage_short_term

This panel indicates container cpu usage total (5m maximum) across all cores by instance (code-intel).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#symbols-provisioning-container-cpu-usage-short-term) for relevant alerts.

<br />

#### symbols: provisioning_container_memory_usage_short_term

This panel indicates container memory usage (5m maximum) by instance (code-intel).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#symbols-provisioning-container-memory-usage-short-term) for relevant alerts.

<br />

### symbols: Golang runtime monitoring

#### symbols: go_goroutines

This panel indicates maximum active goroutines (code-intel).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#symbols-go-goroutines) for relevant alerts.

<br />

#### symbols: go_gc_duration_seconds

This panel indicates maximum go garbage collection duration (code-intel).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#symbols-go-gc-duration-seconds) for relevant alerts.

<br />

### symbols: Kubernetes monitoring (ignore if using Docker Compose or server)

#### symbols: pods_available_percentage

This panel indicates percentage pods available (code-intel).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#symbols-pods-available-percentage) for relevant alerts.

<br />

## Syntect Server

<p class="subtitle">Handles syntax highlighting for code files.</p>

#### syntect-server: syntax_highlighting_errors

This panel indicates syntax highlighting errors every 5m (cloud).

<br />

#### syntect-server: syntax_highlighting_timeouts

This panel indicates syntax highlighting timeouts every 5m (cloud).

<br />

#### syntect-server: syntax_highlighting_panics

This panel indicates syntax highlighting panics every 5m (cloud).

<br />

#### syntect-server: syntax_highlighting_worker_deaths

This panel indicates syntax highlighter worker deaths every 5m (cloud).

<br />

### syntect-server: Container monitoring (not available on server)

#### syntect-server: container_cpu_usage

This panel indicates container cpu usage total (1m average) across all cores by instance (cloud).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#syntect-server-container-cpu-usage) for relevant alerts.

<br />

#### syntect-server: container_memory_usage

This panel indicates container memory usage by instance (cloud).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#syntect-server-container-memory-usage) for relevant alerts.

<br />

#### syntect-server: container_restarts

This panel indicates container restarts every 5m by instance (cloud).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#syntect-server-container-restarts) for relevant alerts.

<br />

#### syntect-server: fs_inodes_used

This panel indicates fs inodes in use by instance (cloud).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#syntect-server-fs-inodes-used) for relevant alerts.

<br />

### syntect-server: Provisioning indicators (not available on server)

#### syntect-server: provisioning_container_cpu_usage_long_term

This panel indicates container cpu usage total (90th percentile over 1d) across all cores by instance (cloud).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#syntect-server-provisioning-container-cpu-usage-long-term) for relevant alerts.

<br />

#### syntect-server: provisioning_container_memory_usage_long_term

This panel indicates container memory usage (1d maximum) by instance (cloud).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#syntect-server-provisioning-container-memory-usage-long-term) for relevant alerts.

<br />

#### syntect-server: provisioning_container_cpu_usage_short_term

This panel indicates container cpu usage total (5m maximum) across all cores by instance (cloud).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#syntect-server-provisioning-container-cpu-usage-short-term) for relevant alerts.

<br />

#### syntect-server: provisioning_container_memory_usage_short_term

This panel indicates container memory usage (5m maximum) by instance (cloud).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#syntect-server-provisioning-container-memory-usage-short-term) for relevant alerts.

<br />

### syntect-server: Kubernetes monitoring (ignore if using Docker Compose or server)

#### syntect-server: pods_available_percentage

This panel indicates percentage pods available (cloud).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#syntect-server-pods-available-percentage) for relevant alerts.

<br />

## Zoekt Index Server

<p class="subtitle">Indexes repositories and populates the search index.</p>

#### zoekt-indexserver: average_resolve_revision_duration

This panel indicates average resolve revision duration over 5m (search).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#zoekt-indexserver-average-resolve-revision-duration) for relevant alerts.

<br />

### zoekt-indexserver: Container monitoring (not available on server)

#### zoekt-indexserver: container_cpu_usage

This panel indicates container cpu usage total (1m average) across all cores by instance (search).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#zoekt-indexserver-container-cpu-usage) for relevant alerts.

<br />

#### zoekt-indexserver: container_memory_usage

This panel indicates container memory usage by instance (search).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#zoekt-indexserver-container-memory-usage) for relevant alerts.

<br />

#### zoekt-indexserver: container_restarts

This panel indicates container restarts every 5m by instance (search).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#zoekt-indexserver-container-restarts) for relevant alerts.

<br />

#### zoekt-indexserver: fs_inodes_used

This panel indicates fs inodes in use by instance (search).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#zoekt-indexserver-fs-inodes-used) for relevant alerts.

<br />

#### zoekt-indexserver: fs_io_operations

This panel indicates filesystem reads and writes rate by instance over 1h (search).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#zoekt-indexserver-fs-io-operations) for relevant alerts.

<br />

### zoekt-indexserver: Provisioning indicators (not available on server)

#### zoekt-indexserver: provisioning_container_cpu_usage_long_term

This panel indicates container cpu usage total (90th percentile over 1d) across all cores by instance (search).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#zoekt-indexserver-provisioning-container-cpu-usage-long-term) for relevant alerts.

<br />

#### zoekt-indexserver: provisioning_container_memory_usage_long_term

This panel indicates container memory usage (1d maximum) by instance (search).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#zoekt-indexserver-provisioning-container-memory-usage-long-term) for relevant alerts.

<br />

#### zoekt-indexserver: provisioning_container_cpu_usage_short_term

This panel indicates container cpu usage total (5m maximum) across all cores by instance (search).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#zoekt-indexserver-provisioning-container-cpu-usage-short-term) for relevant alerts.

<br />

#### zoekt-indexserver: provisioning_container_memory_usage_short_term

This panel indicates container memory usage (5m maximum) by instance (search).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#zoekt-indexserver-provisioning-container-memory-usage-short-term) for relevant alerts.

<br />

### zoekt-indexserver: Kubernetes monitoring (ignore if using Docker Compose or server)

#### zoekt-indexserver: pods_available_percentage

This panel indicates percentage pods available (search).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#zoekt-indexserver-pods-available-percentage) for relevant alerts.

<br />

## Zoekt Web Server

<p class="subtitle">Serves indexed search requests using the search index.</p>

#### zoekt-webserver: indexed_search_request_errors

This panel indicates indexed search request errors every 5m by code (search).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#zoekt-webserver-indexed-search-request-errors) for relevant alerts.

<br />

### zoekt-webserver: Container monitoring (not available on server)

#### zoekt-webserver: container_cpu_usage

This panel indicates container cpu usage total (1m average) across all cores by instance (search).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#zoekt-webserver-container-cpu-usage) for relevant alerts.

<br />

#### zoekt-webserver: container_memory_usage

This panel indicates container memory usage by instance (search).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#zoekt-webserver-container-memory-usage) for relevant alerts.

<br />

#### zoekt-webserver: container_restarts

This panel indicates container restarts every 5m by instance (search).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#zoekt-webserver-container-restarts) for relevant alerts.

<br />

#### zoekt-webserver: fs_inodes_used

This panel indicates fs inodes in use by instance (search).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#zoekt-webserver-fs-inodes-used) for relevant alerts.

<br />

#### zoekt-webserver: fs_io_operations

This panel indicates filesystem reads and writes by instance rate over 1h (search).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#zoekt-webserver-fs-io-operations) for relevant alerts.

<br />

### zoekt-webserver: Provisioning indicators (not available on server)

#### zoekt-webserver: provisioning_container_cpu_usage_long_term

This panel indicates container cpu usage total (90th percentile over 1d) across all cores by instance (search).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#zoekt-webserver-provisioning-container-cpu-usage-long-term) for relevant alerts.

<br />

#### zoekt-webserver: provisioning_container_memory_usage_long_term

This panel indicates container memory usage (1d maximum) by instance (search).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#zoekt-webserver-provisioning-container-memory-usage-long-term) for relevant alerts.

<br />

#### zoekt-webserver: provisioning_container_cpu_usage_short_term

This panel indicates container cpu usage total (5m maximum) across all cores by instance (search).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#zoekt-webserver-provisioning-container-cpu-usage-short-term) for relevant alerts.

<br />

#### zoekt-webserver: provisioning_container_memory_usage_short_term

This panel indicates container memory usage (5m maximum) by instance (search).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#zoekt-webserver-provisioning-container-memory-usage-short-term) for relevant alerts.

<br />

## Prometheus

<p class="subtitle">Sourcegraph's all-in-one Prometheus and Alertmanager service.</p>

### prometheus: Metrics

#### prometheus: prometheus_metrics_bloat

This panel indicates prometheus metrics payload size (distribution).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#prometheus-prometheus-metrics-bloat) for relevant alerts.

<br />

### prometheus: Alerts

#### prometheus: alertmanager_notifications_failed_total

This panel indicates failed alertmanager notifications over 1m (distribution).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#prometheus-alertmanager-notifications-failed-total) for relevant alerts.

<br />

### prometheus: Container monitoring (not available on server)

#### prometheus: container_cpu_usage

This panel indicates container cpu usage total (1m average) across all cores by instance (distribution).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#prometheus-container-cpu-usage) for relevant alerts.

<br />

#### prometheus: container_memory_usage

This panel indicates container memory usage by instance (distribution).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#prometheus-container-memory-usage) for relevant alerts.

<br />

#### prometheus: container_restarts

This panel indicates container restarts every 5m by instance (distribution).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#prometheus-container-restarts) for relevant alerts.

<br />

#### prometheus: fs_inodes_used

This panel indicates fs inodes in use by instance (distribution).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#prometheus-fs-inodes-used) for relevant alerts.

<br />

### prometheus: Provisioning indicators (not available on server)

#### prometheus: provisioning_container_cpu_usage_long_term

This panel indicates container cpu usage total (90th percentile over 1d) across all cores by instance (distribution).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#prometheus-provisioning-container-cpu-usage-long-term) for relevant alerts.

<br />

#### prometheus: provisioning_container_memory_usage_long_term

This panel indicates container memory usage (1d maximum) by instance (distribution).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#prometheus-provisioning-container-memory-usage-long-term) for relevant alerts.

<br />

#### prometheus: provisioning_container_cpu_usage_short_term

This panel indicates container cpu usage total (5m maximum) across all cores by instance (distribution).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#prometheus-provisioning-container-cpu-usage-short-term) for relevant alerts.

<br />

#### prometheus: provisioning_container_memory_usage_short_term

This panel indicates container memory usage (5m maximum) by instance (distribution).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#prometheus-provisioning-container-memory-usage-short-term) for relevant alerts.

<br />

### prometheus: Kubernetes monitoring (ignore if using Docker Compose or server)

#### prometheus: pods_available_percentage

This panel indicates percentage pods available (distribution).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#prometheus-pods-available-percentage) for relevant alerts.

<br />

## Executor Queue

<p class="subtitle">Coordinates the executor work queues.</p>

### executor-queue: Code intelligence queue

#### executor-queue: codeintel_queue_size

This panel indicates queue size (code-intel).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#executor-queue-codeintel-queue-size) for relevant alerts.

<br />

#### executor-queue: codeintel_queue_growth_rate

This panel indicates queue growth rate over 30m (code-intel).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#executor-queue-codeintel-queue-growth-rate) for relevant alerts.

<br />

#### executor-queue: codeintel_job_errors

This panel indicates job errors every 5m (code-intel).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#executor-queue-codeintel-job-errors) for relevant alerts.

<br />

#### executor-queue: codeintel_active_executors

This panel indicates active executors processing codeintel jobs (code-intel).

<br />

#### executor-queue: codeintel_active_jobs

This panel indicates active jobs (code-intel).

<br />

### executor-queue: Stores and clients

#### executor-queue: codeintel_workerstore_99th_percentile_duration

This panel indicates 99th percentile successful worker store operation duration over 5m (code-intel).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#executor-queue-codeintel-workerstore-99th-percentile-duration) for relevant alerts.

<br />

#### executor-queue: codeintel_workerstore_errors

This panel indicates worker store errors every 5m (code-intel).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#executor-queue-codeintel-workerstore-errors) for relevant alerts.

<br />

### executor-queue: Internal service requests

#### executor-queue: frontend_internal_api_error_responses

This panel indicates frontend-internal API error responses every 5m by route (code-intel).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#executor-queue-frontend-internal-api-error-responses) for relevant alerts.

<br />

### executor-queue: Container monitoring (not available on server)

#### executor-queue: container_cpu_usage

This panel indicates container cpu usage total (1m average) across all cores by instance (code-intel).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#executor-queue-container-cpu-usage) for relevant alerts.

<br />

#### executor-queue: container_memory_usage

This panel indicates container memory usage by instance (code-intel).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#executor-queue-container-memory-usage) for relevant alerts.

<br />

#### executor-queue: container_restarts

This panel indicates container restarts every 5m by instance (code-intel).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#executor-queue-container-restarts) for relevant alerts.

<br />

#### executor-queue: fs_inodes_used

This panel indicates fs inodes in use by instance (code-intel).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#executor-queue-fs-inodes-used) for relevant alerts.

<br />

### executor-queue: Provisioning indicators (not available on server)

#### executor-queue: provisioning_container_cpu_usage_long_term

This panel indicates container cpu usage total (90th percentile over 1d) across all cores by instance (code-intel).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#executor-queue-provisioning-container-cpu-usage-long-term) for relevant alerts.

<br />

#### executor-queue: provisioning_container_memory_usage_long_term

This panel indicates container memory usage (1d maximum) by instance (code-intel).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#executor-queue-provisioning-container-memory-usage-long-term) for relevant alerts.

<br />

#### executor-queue: provisioning_container_cpu_usage_short_term

This panel indicates container cpu usage total (5m maximum) across all cores by instance (code-intel).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#executor-queue-provisioning-container-cpu-usage-short-term) for relevant alerts.

<br />

#### executor-queue: provisioning_container_memory_usage_short_term

This panel indicates container memory usage (5m maximum) by instance (code-intel).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#executor-queue-provisioning-container-memory-usage-short-term) for relevant alerts.

<br />

### executor-queue: Golang runtime monitoring

#### executor-queue: go_goroutines

This panel indicates maximum active goroutines (code-intel).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#executor-queue-go-goroutines) for relevant alerts.

<br />

#### executor-queue: go_gc_duration_seconds

This panel indicates maximum go garbage collection duration (code-intel).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#executor-queue-go-gc-duration-seconds) for relevant alerts.

<br />

### executor-queue: Kubernetes monitoring (ignore if using Docker Compose or server)

#### executor-queue: pods_available_percentage

This panel indicates percentage pods available (code-intel).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#executor-queue-pods-available-percentage) for relevant alerts.

<br />

## Precise Code Intel Indexer

<p class="subtitle">Executes jobs from the "codeintel" work queue.</p>

### precise-code-intel-indexer: Executor

#### precise-code-intel-indexer: codeintel_job_99th_percentile_duration

This panel indicates 99th percentile successful job duration over 5m (code-intel).

<br />

#### precise-code-intel-indexer: codeintel_active_handlers

This panel indicates active handlers processing jobs (code-intel).

<br />

#### precise-code-intel-indexer: codeintel_job_errors

This panel indicates job errors every 5m (code-intel).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#precise-code-intel-indexer-codeintel-job-errors) for relevant alerts.

<br />

### precise-code-intel-indexer: Stores and clients

#### precise-code-intel-indexer: executor_apiclient_99th_percentile_duration

This panel indicates 99th percentile successful API request duration over 5m (code-intel).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#precise-code-intel-indexer-executor-apiclient-99th-percentile-duration) for relevant alerts.

<br />

#### precise-code-intel-indexer: executor_apiclient_errors

This panel indicates aPI errors every 5m (code-intel).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#precise-code-intel-indexer-executor-apiclient-errors) for relevant alerts.

<br />

### precise-code-intel-indexer: Commands

#### precise-code-intel-indexer: executor_setup_command_99th_percentile_duration

This panel indicates 99th percentile successful setup command duration over 5m (code-intel).

<br />

#### precise-code-intel-indexer: executor_setup_command_errors

This panel indicates setup command errors every 5m (code-intel).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#precise-code-intel-indexer-executor-setup-command-errors) for relevant alerts.

<br />

#### precise-code-intel-indexer: executor_exec_command_99th_percentile_duration

This panel indicates 99th percentile successful exec command duration over 5m (code-intel).

<br />

#### precise-code-intel-indexer: executor_exec_command_errors

This panel indicates exec command errors every 5m (code-intel).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#precise-code-intel-indexer-executor-exec-command-errors) for relevant alerts.

<br />

#### precise-code-intel-indexer: executor_teardown_command_99th_percentile_duration

This panel indicates 99th percentile successful teardown command duration over 5m (code-intel).

<br />

#### precise-code-intel-indexer: executor_teardown_command_errors

This panel indicates teardown command errors every 5m (code-intel).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#precise-code-intel-indexer-executor-teardown-command-errors) for relevant alerts.

<br />

### precise-code-intel-indexer: Container monitoring (not available on server)

#### precise-code-intel-indexer: container_cpu_usage

This panel indicates container cpu usage total (1m average) across all cores by instance (code-intel).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#precise-code-intel-indexer-container-cpu-usage) for relevant alerts.

<br />

#### precise-code-intel-indexer: container_memory_usage

This panel indicates container memory usage by instance (code-intel).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#precise-code-intel-indexer-container-memory-usage) for relevant alerts.

<br />

#### precise-code-intel-indexer: container_restarts

This panel indicates container restarts every 5m by instance (code-intel).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#precise-code-intel-indexer-container-restarts) for relevant alerts.

<br />

#### precise-code-intel-indexer: fs_inodes_used

This panel indicates fs inodes in use by instance (code-intel).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#precise-code-intel-indexer-fs-inodes-used) for relevant alerts.

<br />

### precise-code-intel-indexer: Provisioning indicators (not available on server)

#### precise-code-intel-indexer: provisioning_container_cpu_usage_long_term

This panel indicates container cpu usage total (90th percentile over 1d) across all cores by instance (code-intel).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#precise-code-intel-indexer-provisioning-container-cpu-usage-long-term) for relevant alerts.

<br />

#### precise-code-intel-indexer: provisioning_container_memory_usage_long_term

This panel indicates container memory usage (1d maximum) by instance (code-intel).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#precise-code-intel-indexer-provisioning-container-memory-usage-long-term) for relevant alerts.

<br />

#### precise-code-intel-indexer: provisioning_container_cpu_usage_short_term

This panel indicates container cpu usage total (5m maximum) across all cores by instance (code-intel).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#precise-code-intel-indexer-provisioning-container-cpu-usage-short-term) for relevant alerts.

<br />

#### precise-code-intel-indexer: provisioning_container_memory_usage_short_term

This panel indicates container memory usage (5m maximum) by instance (code-intel).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#precise-code-intel-indexer-provisioning-container-memory-usage-short-term) for relevant alerts.

<br />

### precise-code-intel-indexer: Golang runtime monitoring

#### precise-code-intel-indexer: go_goroutines

This panel indicates maximum active goroutines (code-intel).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#precise-code-intel-indexer-go-goroutines) for relevant alerts.

<br />

#### precise-code-intel-indexer: go_gc_duration_seconds

This panel indicates maximum go garbage collection duration (code-intel).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#precise-code-intel-indexer-go-gc-duration-seconds) for relevant alerts.

<br />

### precise-code-intel-indexer: Kubernetes monitoring (ignore if using Docker Compose or server)

#### precise-code-intel-indexer: pods_available_percentage

This panel indicates percentage pods available (code-intel).

Refer to the [alert solutions reference](https://docs.sourcegraph.com/admin/observability/alert_solutions#precise-code-intel-indexer-pods-available-percentage) for relevant alerts.

<br />

