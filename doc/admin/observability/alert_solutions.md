# Alert solutions

This document contains possible solutions for when you find alerts are firing in Sourcegraph's monitoring.
If your alert isn't mentioned here, or if the solution doesn't help, [contact us](mailto:support@sourcegraph.com)
for assistance.

<!-- DO NOT EDIT: generated via: go generate ./monitoring -->

# frontend: 99th_percentile_search_request_duration

**Descriptions:**

- _frontend: 20s+ 99th percentile successful search request duration over 5m_ (`warning_frontend_99th_percentile_search_request_duration`)

**Possible solutions:**

- **Get details on the exact queries that are slow** by configuring `"observability.logSlowSearches": 20,` in the site configuration and looking for `frontend` warning logs prefixed with `slow search request` for additional details.
- **Check that most repositories are indexed** by visiting https://sourcegraph.example.com/site-admin/repositories?filter=needs-index (it should show few or no results.)
- **Kubernetes:** Check CPU usage of zoekt-webserver in the indexed-search pod, consider increasing CPU limits in the `indexed-search.Deployment.yaml` if regularly hitting max CPU utilization.
- **Docker Compose:** Check CPU usage on the Zoekt Web Server dashboard, consider increasing `cpus:` of the zoekt-webserver container in `docker-compose.yml` if regularly hitting max CPU utilization.

# frontend: 90th_percentile_search_request_duration

**Descriptions:**

- _frontend: 15s+ 90th percentile successful search request duration over 5m_ (`warning_frontend_90th_percentile_search_request_duration`)

**Possible solutions:**

- **Get details on the exact queries that are slow** by configuring `"observability.logSlowSearches": 15,` in the site configuration and looking for `frontend` warning logs prefixed with `slow search request` for additional details.
- **Check that most repositories are indexed** by visiting https://sourcegraph.example.com/site-admin/repositories?filter=needs-index (it should show few or no results.)
- **Kubernetes:** Check CPU usage of zoekt-webserver in the indexed-search pod, consider increasing CPU limits in the `indexed-search.Deployment.yaml` if regularly hitting max CPU utilization.
- **Docker Compose:** Check CPU usage on the Zoekt Web Server dashboard, consider increasing `cpus:` of the zoekt-webserver container in `docker-compose.yml` if regularly hitting max CPU utilization.

# frontend: search_alert_user_suggestions

**Descriptions:**

- _frontend: 50+ search alert user suggestions shown every 5m_ (`warning_frontend_search_alert_user_suggestions`)

**Possible solutions:**

- This indicates your user`s are making syntax errors or similar user errors.

# frontend: 99th_percentile_search_codeintel_request_duration

**Descriptions:**

- _frontend: 20s+ 99th percentile code-intel successful search request duration over 5m_ (`warning_frontend_99th_percentile_search_codeintel_request_duration`)

**Possible solutions:**

- **Get details on the exact queries that are slow** by configuring `"observability.logSlowSearches": 20,` in the site configuration and looking for `frontend` warning logs prefixed with `slow search request` for additional details.
- **Check that most repositories are indexed** by visiting https://sourcegraph.example.com/site-admin/repositories?filter=needs-index (it should show few or no results.)
- **Kubernetes:** Check CPU usage of zoekt-webserver in the indexed-search pod, consider increasing CPU limits in the `indexed-search.Deployment.yaml` if regularly hitting max CPU utilization.
- **Docker Compose:** Check CPU usage on the Zoekt Web Server dashboard, consider increasing `cpus:` of the zoekt-webserver container in `docker-compose.yml` if regularly hitting max CPU utilization.

# frontend: 90th_percentile_search_codeintel_request_duration

**Descriptions:**

- _frontend: 15s+ 90th percentile code-intel successful search request duration over 5m_ (`warning_frontend_90th_percentile_search_codeintel_request_duration`)

**Possible solutions:**

- **Get details on the exact queries that are slow** by configuring `"observability.logSlowSearches": 15,` in the site configuration and looking for `frontend` warning logs prefixed with `slow search request` for additional details.
- **Check that most repositories are indexed** by visiting https://sourcegraph.example.com/site-admin/repositories?filter=needs-index (it should show few or no results.)
- **Kubernetes:** Check CPU usage of zoekt-webserver in the indexed-search pod, consider increasing CPU limits in the `indexed-search.Deployment.yaml` if regularly hitting max CPU utilization.
- **Docker Compose:** Check CPU usage on the Zoekt Web Server dashboard, consider increasing `cpus:` of the zoekt-webserver container in `docker-compose.yml` if regularly hitting max CPU utilization.

# frontend: search_codeintel_alert_user_suggestions

**Descriptions:**

- _frontend: 50+ search code-intel alert user suggestions shown every 5m_ (`warning_frontend_search_codeintel_alert_user_suggestions`)

**Possible solutions:**

- This indicates a bug in Sourcegraph, please [open an issue](https://github.com/sourcegraph/sourcegraph/issues/new/choose).

# frontend: 99th_percentile_search_api_request_duration

**Descriptions:**

- _frontend: 50s+ 99th percentile successful search API request duration over 5m_ (`warning_frontend_99th_percentile_search_api_request_duration`)

**Possible solutions:**

- **Get details on the exact queries that are slow** by configuring `"observability.logSlowSearches": 20,` in the site configuration and looking for `frontend` warning logs prefixed with `slow search request` for additional details.
- **If your users are requesting many results** with a large `count:` parameter, consider using our [search pagination API](../../api/graphql/search.md).
- **Check that most repositories are indexed** by visiting https://sourcegraph.example.com/site-admin/repositories?filter=needs-index (it should show few or no results.)
- **Kubernetes:** Check CPU usage of zoekt-webserver in the indexed-search pod, consider increasing CPU limits in the `indexed-search.Deployment.yaml` if regularly hitting max CPU utilization.
- **Docker Compose:** Check CPU usage on the Zoekt Web Server dashboard, consider increasing `cpus:` of the zoekt-webserver container in `docker-compose.yml` if regularly hitting max CPU utilization.

# frontend: 90th_percentile_search_api_request_duration

**Descriptions:**

- _frontend: 40s+ 90th percentile successful search API request duration over 5m_ (`warning_frontend_90th_percentile_search_api_request_duration`)

**Possible solutions:**

- **Get details on the exact queries that are slow** by configuring `"observability.logSlowSearches": 15,` in the site configuration and looking for `frontend` warning logs prefixed with `slow search request` for additional details.
- **If your users are requesting many results** with a large `count:` parameter, consider using our [search pagination API](../../api/graphql/search.md).
- **Check that most repositories are indexed** by visiting https://sourcegraph.example.com/site-admin/repositories?filter=needs-index (it should show few or no results.)
- **Kubernetes:** Check CPU usage of zoekt-webserver in the indexed-search pod, consider increasing CPU limits in the `indexed-search.Deployment.yaml` if regularly hitting max CPU utilization.
- **Docker Compose:** Check CPU usage on the Zoekt Web Server dashboard, consider increasing `cpus:` of the zoekt-webserver container in `docker-compose.yml` if regularly hitting max CPU utilization.

# frontend: search_api_alert_user_suggestions

**Descriptions:**

- _frontend: 50+ search API alert user suggestions shown every 5m_ (`warning_frontend_search_api_alert_user_suggestions`)

**Possible solutions:**

- This indicates your user`s search API requests have syntax errors or a similar user error. Check the responses the API sends back for an explanation.

# frontend: internal_indexed_search_error_responses

**Descriptions:**

- _frontend: 5+ internal indexed search error responses every 5m_ (`warning_frontend_internal_indexed_search_error_responses`)

**Possible solutions:**

- Check the Zoekt Web Server dashboard for indications it might be unhealthy.

# frontend: internal_unindexed_search_error_responses

**Descriptions:**

- _frontend: 5+ internal unindexed search error responses every 5m_ (`warning_frontend_internal_unindexed_search_error_responses`)

**Possible solutions:**

- Check the Searcher dashboard for indications it might be unhealthy.

# frontend: internal_api_error_responses

**Descriptions:**

- _frontend: 25+ internal API error responses every 5m by route_ (`warning_frontend_internal_api_error_responses`)

**Possible solutions:**

- May not be a substantial issue, check the `frontend` logs for potential causes.

# frontend: container_restarts

**Descriptions:**

- _frontend: 1+ container restarts every 5m by instance_ (`warning_frontend_container_restarts`)

**Possible solutions:**

- **Kubernetes:**
	- Determine if the pod was OOM killed using `kubectl describe pod frontend` (look for `OOMKilled: true`) and, if so, consider increasing the memory limit in the relevant `Deployment.yaml`.
	- Check the logs before the container restarted to see if there are `panic:` messages or similar using `kubectl logs -p frontend`.
- **Docker Compose:**
	- Determine if the pod was OOM killed using `docker inspect -f '{{json .State}}' frontend` (look for `"OOMKilled":true`) and, if so, consider increasing the memory limit of the frontend container in `docker-compose.yml`.
	- Check the logs before the container restarted to see if there are `panic:` messages or similar using `docker logs frontend` (note this will include logs from the previous and currently running container).

# frontend: container_memory_usage

**Descriptions:**

- _frontend: 99%+ container memory usage by instance_ (`warning_frontend_container_memory_usage`)

**Possible solutions:**

- **Kubernetes:** Consider increasing memory limit in relevant `Deployment.yaml`.
- **Docker Compose:** Consider increasing `memory:` of frontend container in `docker-compose.yml`.

# frontend: container_cpu_usage

**Descriptions:**

- _frontend: 99%+ container cpu usage total (1m average) across all cores by instance_ (`warning_frontend_container_cpu_usage`)

**Possible solutions:**

- **Kubernetes:** Consider increasing CPU limits in the the relevant `Deployment.yaml`.
- **Docker Compose:** Consider increasing `cpus:` of the frontend container in `docker-compose.yml`.

# frontend: provisioning_container_cpu_usage_7d

**Descriptions:**

- _frontend: 80%+ or less than 30% container cpu usage total (7d maximum) across all cores by instance_ (`warning_frontend_provisioning_container_cpu_usage_7d`)

**Possible solutions:**

- If usage is high:
	- **Kubernetes:** Consider decreasing CPU limits in the the relevant `Deployment.yaml`.
	- **Docker Compose:** Consider descreasing `cpus:` of the frontend container in `docker-compose.yml`.
- If usage is low, consider decreasing the above values.

# frontend: provisioning_container_memory_usage_7d

**Descriptions:**

- _frontend: 80%+ or less than 30% container memory usage (7d maximum) by instance_ (`warning_frontend_provisioning_container_memory_usage_7d`)

**Possible solutions:**

- If usage is high:
	- **Kubernetes:** Consider decreasing memory limit in relevant `Deployment.yaml`.
	- **Docker Compose:** Consider decreasing `memory:` of frontend container in `docker-compose.yml`.
- If usage is low, consider decreasing the above values.

# frontend: provisioning_container_cpu_usage_5m

**Descriptions:**

- _frontend: 90%+ container cpu usage total (5m maximum) across all cores by instance_ (`warning_frontend_provisioning_container_cpu_usage_5m`)

**Possible solutions:**

- **Kubernetes:** Consider increasing CPU limits in the the relevant `Deployment.yaml`.
- **Docker Compose:** Consider increasing `cpus:` of the frontend container in `docker-compose.yml`.

# frontend: provisioning_container_memory_usage_5m

**Descriptions:**

- _frontend: 90%+ container memory usage (5m maximum) by instance_ (`warning_frontend_provisioning_container_memory_usage_5m`)

**Possible solutions:**

- **Kubernetes:** Consider increasing memory limit in relevant `Deployment.yaml`.
- **Docker Compose:** Consider increasing `memory:` of frontend container in `docker-compose.yml`.

# gitserver: disk_space_remaining

**Descriptions:**

- _gitserver: less than 25% disk space remaining by instance_ (`warning_gitserver_disk_space_remaining`)


- _gitserver: less than 15% disk space remaining by instance_ (`critical_gitserver_disk_space_remaining`)

**Possible solutions:**

- **Provision more disk space:** Sourcegraph will begin deleting least-used repository clones at 10% disk space remaining which may result in decreased performance, users having to wait for repositories to clone, etc.

# gitserver: running_git_commands

**Descriptions:**

- _gitserver: 50+ running git commands (signals load)_ (`warning_gitserver_running_git_commands`)


- _gitserver: 100+ running git commands (signals load)_ (`critical_gitserver_running_git_commands`)

**Possible solutions:**

- **Check if the problem may be an intermittent and temporary peak** using the "Container monitoring" section at the bottom of the Git Server dashboard.
- **Single container deployments:** Consider upgrading to a [Docker Compose deployment](../install/docker-compose/migrate.md) which offers better scalability and resource isolation.
- **Kubernetes and Docker Compose:** Check that you are running a similar number of git server replicas and that their CPU/memory limits are allocated according to what is shown in the [Sourcegraph resource estimator](../install/resource_estimator.md).

# gitserver: repository_clone_queue_size

**Descriptions:**

- _gitserver: 25+ repository clone queue size_ (`warning_gitserver_repository_clone_queue_size`)

**Possible solutions:**

- **If you just added several repositories**, the warning may be expected.
- **Check which repositories need cloning**, by visiting e.g. https://sourcegraph.example.com/site-admin/repositories?filter=not-cloned

# gitserver: repository_existence_check_queue_size

**Descriptions:**

- _gitserver: 25+ repository existence check queue size_ (`warning_gitserver_repository_existence_check_queue_size`)

**Possible solutions:**

- **Check the code host status indicator for errors:** on the Sourcegraph app homepage, when signed in as an admin click the cloud icon in the top right corner of the page.
- **Check if the issue continues to happen after 30 minutes**, it may be temporary.
- **Check the gitserver logs for more information.**

# gitserver: echo_command_duration_test

**Descriptions:**

- _gitserver: 1s+ echo command duration test_ (`warning_gitserver_echo_command_duration_test`)


- _gitserver: 2s+ echo command duration test_ (`critical_gitserver_echo_command_duration_test`)

**Possible solutions:**

- **Check if the problem may be an intermittent and temporary peak** using the "Container monitoring" section at the bottom of the Git Server dashboard.
- **Single container deployments:** Consider upgrading to a [Docker Compose deployment](../install/docker-compose/migrate.md) which offers better scalability and resource isolation.
- **Kubernetes and Docker Compose:** Check that you are running a similar number of git server replicas and that their CPU/memory limits are allocated according to what is shown in the [Sourcegraph resource estimator](../install/resource_estimator.md).

# gitserver: frontend_internal_api_error_responses

**Descriptions:**

- _gitserver: 5+ frontend-internal API error responses every 5m by route_ (`warning_gitserver_frontend_internal_api_error_responses`)

**Possible solutions:**

- **Single-container deployments:** Check `docker logs $CONTAINER_ID` for logs starting with `repo-updater` that indicate requests to the frontend service are failing.
- **Kubernetes:**
	- Confirm that `kubectl get pods` shows the `frontend` pods are healthy.
	- Check `kubectl logs gitserver` for logs indicate request failures to `frontend` or `frontend-internal`.
- **Docker Compose:**
	- Confirm that `docker ps` shows the `frontend-internal` container is healthy.
	- Check `docker logs gitserver` for logs indicating request failures to `frontend` or `frontend-internal`.

# gitserver: container_restarts

**Descriptions:**

- _gitserver: 1+ container restarts every 5m by instance_ (`warning_gitserver_container_restarts`)

**Possible solutions:**

- **Kubernetes:**
	- Determine if the pod was OOM killed using `kubectl describe pod gitserver` (look for `OOMKilled: true`) and, if so, consider increasing the memory limit in the relevant `Deployment.yaml`.
	- Check the logs before the container restarted to see if there are `panic:` messages or similar using `kubectl logs -p gitserver`.
- **Docker Compose:**
	- Determine if the pod was OOM killed using `docker inspect -f '{{json .State}}' gitserver` (look for `"OOMKilled":true`) and, if so, consider increasing the memory limit of the gitserver container in `docker-compose.yml`.
	- Check the logs before the container restarted to see if there are `panic:` messages or similar using `docker logs gitserver` (note this will include logs from the previous and currently running container).

# gitserver: container_memory_usage

**Descriptions:**

- _gitserver: 99%+ container memory usage by instance_ (`warning_gitserver_container_memory_usage`)

**Possible solutions:**

- **Kubernetes:** Consider increasing memory limit in relevant `Deployment.yaml`.
- **Docker Compose:** Consider increasing `memory:` of gitserver container in `docker-compose.yml`.

# gitserver: container_cpu_usage

**Descriptions:**

- _gitserver: 99%+ container cpu usage total (1m average) across all cores by instance_ (`warning_gitserver_container_cpu_usage`)

**Possible solutions:**

- **Kubernetes:** Consider increasing CPU limits in the the relevant `Deployment.yaml`.
- **Docker Compose:** Consider increasing `cpus:` of the gitserver container in `docker-compose.yml`.

# gitserver: provisioning_container_cpu_usage_7d

**Descriptions:**

- _gitserver: 80%+ or less than 30% container cpu usage total (7d maximum) across all cores by instance_ (`warning_gitserver_provisioning_container_cpu_usage_7d`)

**Possible solutions:**

- If usage is high:
	- **Kubernetes:** Consider decreasing CPU limits in the the relevant `Deployment.yaml`.
	- **Docker Compose:** Consider descreasing `cpus:` of the gitserver container in `docker-compose.yml`.
- If usage is low, consider decreasing the above values.

# gitserver: provisioning_container_memory_usage_7d

**Descriptions:**

- _gitserver: 80%+ or less than 30% container memory usage (7d maximum) by instance_ (`warning_gitserver_provisioning_container_memory_usage_7d`)

**Possible solutions:**

- If usage is high:
	- **Kubernetes:** Consider decreasing memory limit in relevant `Deployment.yaml`.
	- **Docker Compose:** Consider decreasing `memory:` of gitserver container in `docker-compose.yml`.
- If usage is low, consider decreasing the above values.

# gitserver: provisioning_container_cpu_usage_5m

**Descriptions:**

- _gitserver: 90%+ container cpu usage total (5m maximum) across all cores by instance_ (`warning_gitserver_provisioning_container_cpu_usage_5m`)

**Possible solutions:**

- **Kubernetes:** Consider increasing CPU limits in the the relevant `Deployment.yaml`.
- **Docker Compose:** Consider increasing `cpus:` of the gitserver container in `docker-compose.yml`.

# gitserver: provisioning_container_memory_usage_5m

**Descriptions:**

- _gitserver: 90%+ container memory usage (5m maximum) by instance_ (`warning_gitserver_provisioning_container_memory_usage_5m`)

**Possible solutions:**

- **Kubernetes:** Consider increasing memory limit in relevant `Deployment.yaml`.
- **Docker Compose:** Consider increasing `memory:` of gitserver container in `docker-compose.yml`.

# github-proxy: container_restarts

**Descriptions:**

- _github-proxy: 1+ container restarts every 5m by instance_ (`warning_github-proxy_container_restarts`)

**Possible solutions:**

- **Kubernetes:**
	- Determine if the pod was OOM killed using `kubectl describe pod github-proxy` (look for `OOMKilled: true`) and, if so, consider increasing the memory limit in the relevant `Deployment.yaml`.
	- Check the logs before the container restarted to see if there are `panic:` messages or similar using `kubectl logs -p github-proxy`.
- **Docker Compose:**
	- Determine if the pod was OOM killed using `docker inspect -f '{{json .State}}' github-proxy` (look for `"OOMKilled":true`) and, if so, consider increasing the memory limit of the github-proxy container in `docker-compose.yml`.
	- Check the logs before the container restarted to see if there are `panic:` messages or similar using `docker logs github-proxy` (note this will include logs from the previous and currently running container).

# github-proxy: container_memory_usage

**Descriptions:**

- _github-proxy: 99%+ container memory usage by instance_ (`warning_github-proxy_container_memory_usage`)

**Possible solutions:**

- **Kubernetes:** Consider increasing memory limit in relevant `Deployment.yaml`.
- **Docker Compose:** Consider increasing `memory:` of github-proxy container in `docker-compose.yml`.

# github-proxy: container_cpu_usage

**Descriptions:**

- _github-proxy: 99%+ container cpu usage total (1m average) across all cores by instance_ (`warning_github-proxy_container_cpu_usage`)

**Possible solutions:**

- **Kubernetes:** Consider increasing CPU limits in the the relevant `Deployment.yaml`.
- **Docker Compose:** Consider increasing `cpus:` of the github-proxy container in `docker-compose.yml`.

# github-proxy: provisioning_container_cpu_usage_7d

**Descriptions:**

- _github-proxy: 80%+ or less than 30% container cpu usage total (7d maximum) across all cores by instance_ (`warning_github-proxy_provisioning_container_cpu_usage_7d`)

**Possible solutions:**

- If usage is high:
	- **Kubernetes:** Consider decreasing CPU limits in the the relevant `Deployment.yaml`.
	- **Docker Compose:** Consider descreasing `cpus:` of the github-proxy container in `docker-compose.yml`.
- If usage is low, consider decreasing the above values.

# github-proxy: provisioning_container_memory_usage_7d

**Descriptions:**

- _github-proxy: 80%+ or less than 30% container memory usage (7d maximum) by instance_ (`warning_github-proxy_provisioning_container_memory_usage_7d`)

**Possible solutions:**

- If usage is high:
	- **Kubernetes:** Consider decreasing memory limit in relevant `Deployment.yaml`.
	- **Docker Compose:** Consider decreasing `memory:` of github-proxy container in `docker-compose.yml`.
- If usage is low, consider decreasing the above values.

# github-proxy: provisioning_container_cpu_usage_5m

**Descriptions:**

- _github-proxy: 90%+ container cpu usage total (5m maximum) across all cores by instance_ (`warning_github-proxy_provisioning_container_cpu_usage_5m`)

**Possible solutions:**

- **Kubernetes:** Consider increasing CPU limits in the the relevant `Deployment.yaml`.
- **Docker Compose:** Consider increasing `cpus:` of the github-proxy container in `docker-compose.yml`.

# github-proxy: provisioning_container_memory_usage_5m

**Descriptions:**

- _github-proxy: 90%+ container memory usage (5m maximum) by instance_ (`warning_github-proxy_provisioning_container_memory_usage_5m`)

**Possible solutions:**

- **Kubernetes:** Consider increasing memory limit in relevant `Deployment.yaml`.
- **Docker Compose:** Consider increasing `memory:` of github-proxy container in `docker-compose.yml`.

# precise-code-intel-bundle-manager: disk_space_remaining

**Descriptions:**

- _precise-code-intel-bundle-manager: less than 25% disk space remaining by instance_ (`warning_precise-code-intel-bundle-manager_disk_space_remaining`)


- _precise-code-intel-bundle-manager: less than 15% disk space remaining by instance_ (`critical_precise-code-intel-bundle-manager_disk_space_remaining`)

**Possible solutions:**

- **Provision more disk space:** Sourcegraph will begin deleting the oldest uploaded bundle files at 10% disk space remaining.

# precise-code-intel-bundle-manager: frontend_internal_api_error_responses

**Descriptions:**

- _precise-code-intel-bundle-manager: 5+ frontend-internal API error responses every 5m by route_ (`warning_precise-code-intel-bundle-manager_frontend_internal_api_error_responses`)

**Possible solutions:**

- **Single-container deployments:** Check `docker logs $CONTAINER_ID` for logs starting with `repo-updater` that indicate requests to the frontend service are failing.
- **Kubernetes:**
	- Confirm that `kubectl get pods` shows the `frontend` pods are healthy.
	- Check `kubectl logs precise-code-intel-bundle-manager` for logs indicate request failures to `frontend` or `frontend-internal`.
- **Docker Compose:**
	- Confirm that `docker ps` shows the `frontend-internal` container is healthy.
	- Check `docker logs precise-code-intel-bundle-manager` for logs indicating request failures to `frontend` or `frontend-internal`.

# precise-code-intel-bundle-manager: container_restarts

**Descriptions:**

- _precise-code-intel-bundle-manager: 1+ container restarts every 5m by instance_ (`warning_precise-code-intel-bundle-manager_container_restarts`)

**Possible solutions:**

- **Kubernetes:**
	- Determine if the pod was OOM killed using `kubectl describe pod precise-code-intel-bundle-manager` (look for `OOMKilled: true`) and, if so, consider increasing the memory limit in the relevant `Deployment.yaml`.
	- Check the logs before the container restarted to see if there are `panic:` messages or similar using `kubectl logs -p precise-code-intel-bundle-manager`.
- **Docker Compose:**
	- Determine if the pod was OOM killed using `docker inspect -f '{{json .State}}' precise-code-intel-bundle-manager` (look for `"OOMKilled":true`) and, if so, consider increasing the memory limit of the precise-code-intel-bundle-manager container in `docker-compose.yml`.
	- Check the logs before the container restarted to see if there are `panic:` messages or similar using `docker logs precise-code-intel-bundle-manager` (note this will include logs from the previous and currently running container).

# precise-code-intel-bundle-manager: container_memory_usage

**Descriptions:**

- _precise-code-intel-bundle-manager: 99%+ container memory usage by instance_ (`warning_precise-code-intel-bundle-manager_container_memory_usage`)

**Possible solutions:**

- **Kubernetes:** Consider increasing memory limit in relevant `Deployment.yaml`.
- **Docker Compose:** Consider increasing `memory:` of precise-code-intel-bundle-manager container in `docker-compose.yml`.

# precise-code-intel-bundle-manager: container_cpu_usage

**Descriptions:**

- _precise-code-intel-bundle-manager: 99%+ container cpu usage total (1m average) across all cores by instance_ (`warning_precise-code-intel-bundle-manager_container_cpu_usage`)

**Possible solutions:**

- **Kubernetes:** Consider increasing CPU limits in the the relevant `Deployment.yaml`.
- **Docker Compose:** Consider increasing `cpus:` of the precise-code-intel-bundle-manager container in `docker-compose.yml`.

# precise-code-intel-bundle-manager: provisioning_container_cpu_usage_7d

**Descriptions:**

- _precise-code-intel-bundle-manager: 80%+ or less than 30% container cpu usage total (7d maximum) across all cores by instance_ (`warning_precise-code-intel-bundle-manager_provisioning_container_cpu_usage_7d`)

**Possible solutions:**

- If usage is high:
	- **Kubernetes:** Consider decreasing CPU limits in the the relevant `Deployment.yaml`.
	- **Docker Compose:** Consider descreasing `cpus:` of the precise-code-intel-bundle-manager container in `docker-compose.yml`.
- If usage is low, consider decreasing the above values.

# precise-code-intel-bundle-manager: provisioning_container_memory_usage_7d

**Descriptions:**

- _precise-code-intel-bundle-manager: 80%+ or less than 30% container memory usage (7d maximum) by instance_ (`warning_precise-code-intel-bundle-manager_provisioning_container_memory_usage_7d`)

**Possible solutions:**

- If usage is high:
	- **Kubernetes:** Consider decreasing memory limit in relevant `Deployment.yaml`.
	- **Docker Compose:** Consider decreasing `memory:` of precise-code-intel-bundle-manager container in `docker-compose.yml`.
- If usage is low, consider decreasing the above values.

# precise-code-intel-bundle-manager: provisioning_container_cpu_usage_5m

**Descriptions:**

- _precise-code-intel-bundle-manager: 90%+ container cpu usage total (5m maximum) across all cores by instance_ (`warning_precise-code-intel-bundle-manager_provisioning_container_cpu_usage_5m`)

**Possible solutions:**

- **Kubernetes:** Consider increasing CPU limits in the the relevant `Deployment.yaml`.
- **Docker Compose:** Consider increasing `cpus:` of the precise-code-intel-bundle-manager container in `docker-compose.yml`.

# precise-code-intel-bundle-manager: provisioning_container_memory_usage_5m

**Descriptions:**

- _precise-code-intel-bundle-manager: 90%+ container memory usage (5m maximum) by instance_ (`warning_precise-code-intel-bundle-manager_provisioning_container_memory_usage_5m`)

**Possible solutions:**

- **Kubernetes:** Consider increasing memory limit in relevant `Deployment.yaml`.
- **Docker Compose:** Consider increasing `memory:` of precise-code-intel-bundle-manager container in `docker-compose.yml`.

# precise-code-intel-worker: frontend_internal_api_error_responses

**Descriptions:**

- _precise-code-intel-worker: 5+ frontend-internal API error responses every 5m by route_ (`warning_precise-code-intel-worker_frontend_internal_api_error_responses`)

**Possible solutions:**

- **Single-container deployments:** Check `docker logs $CONTAINER_ID` for logs starting with `repo-updater` that indicate requests to the frontend service are failing.
- **Kubernetes:**
	- Confirm that `kubectl get pods` shows the `frontend` pods are healthy.
	- Check `kubectl logs precise-code-intel-worker` for logs indicate request failures to `frontend` or `frontend-internal`.
- **Docker Compose:**
	- Confirm that `docker ps` shows the `frontend-internal` container is healthy.
	- Check `docker logs precise-code-intel-worker` for logs indicating request failures to `frontend` or `frontend-internal`.

# precise-code-intel-worker: container_restarts

**Descriptions:**

- _precise-code-intel-worker: 1+ container restarts every 5m by instance_ (`warning_precise-code-intel-worker_container_restarts`)

**Possible solutions:**

- **Kubernetes:**
	- Determine if the pod was OOM killed using `kubectl describe pod precise-code-intel-worker` (look for `OOMKilled: true`) and, if so, consider increasing the memory limit in the relevant `Deployment.yaml`.
	- Check the logs before the container restarted to see if there are `panic:` messages or similar using `kubectl logs -p precise-code-intel-worker`.
- **Docker Compose:**
	- Determine if the pod was OOM killed using `docker inspect -f '{{json .State}}' precise-code-intel-worker` (look for `"OOMKilled":true`) and, if so, consider increasing the memory limit of the precise-code-intel-worker container in `docker-compose.yml`.
	- Check the logs before the container restarted to see if there are `panic:` messages or similar using `docker logs precise-code-intel-worker` (note this will include logs from the previous and currently running container).

# precise-code-intel-worker: container_memory_usage

**Descriptions:**

- _precise-code-intel-worker: 99%+ container memory usage by instance_ (`warning_precise-code-intel-worker_container_memory_usage`)

**Possible solutions:**

- **Kubernetes:** Consider increasing memory limit in relevant `Deployment.yaml`.
- **Docker Compose:** Consider increasing `memory:` of precise-code-intel-worker container in `docker-compose.yml`.

# precise-code-intel-worker: container_cpu_usage

**Descriptions:**

- _precise-code-intel-worker: 99%+ container cpu usage total (1m average) across all cores by instance_ (`warning_precise-code-intel-worker_container_cpu_usage`)

**Possible solutions:**

- **Kubernetes:** Consider increasing CPU limits in the the relevant `Deployment.yaml`.
- **Docker Compose:** Consider increasing `cpus:` of the precise-code-intel-worker container in `docker-compose.yml`.

# precise-code-intel-worker: provisioning_container_cpu_usage_7d

**Descriptions:**

- _precise-code-intel-worker: 80%+ or less than 30% container cpu usage total (7d maximum) across all cores by instance_ (`warning_precise-code-intel-worker_provisioning_container_cpu_usage_7d`)

**Possible solutions:**

- If usage is high:
	- **Kubernetes:** Consider decreasing CPU limits in the the relevant `Deployment.yaml`.
	- **Docker Compose:** Consider descreasing `cpus:` of the precise-code-intel-worker container in `docker-compose.yml`.
- If usage is low, consider decreasing the above values.

# precise-code-intel-worker: provisioning_container_memory_usage_7d

**Descriptions:**

- _precise-code-intel-worker: 80%+ or less than 30% container memory usage (7d maximum) by instance_ (`warning_precise-code-intel-worker_provisioning_container_memory_usage_7d`)

**Possible solutions:**

- If usage is high:
	- **Kubernetes:** Consider decreasing memory limit in relevant `Deployment.yaml`.
	- **Docker Compose:** Consider decreasing `memory:` of precise-code-intel-worker container in `docker-compose.yml`.
- If usage is low, consider decreasing the above values.

# precise-code-intel-worker: provisioning_container_cpu_usage_5m

**Descriptions:**

- _precise-code-intel-worker: 90%+ container cpu usage total (5m maximum) across all cores by instance_ (`warning_precise-code-intel-worker_provisioning_container_cpu_usage_5m`)

**Possible solutions:**

- **Kubernetes:** Consider increasing CPU limits in the the relevant `Deployment.yaml`.
- **Docker Compose:** Consider increasing `cpus:` of the precise-code-intel-worker container in `docker-compose.yml`.

# precise-code-intel-worker: provisioning_container_memory_usage_5m

**Descriptions:**

- _precise-code-intel-worker: 90%+ container memory usage (5m maximum) by instance_ (`warning_precise-code-intel-worker_provisioning_container_memory_usage_5m`)

**Possible solutions:**

- **Kubernetes:** Consider increasing memory limit in relevant `Deployment.yaml`.
- **Docker Compose:** Consider increasing `memory:` of precise-code-intel-worker container in `docker-compose.yml`.

# precise-code-intel-indexer: frontend_internal_api_error_responses

**Descriptions:**

- _precise-code-intel-indexer: 5+ frontend-internal API error responses every 5m by route_ (`warning_precise-code-intel-indexer_frontend_internal_api_error_responses`)

**Possible solutions:**

- **Single-container deployments:** Check `docker logs $CONTAINER_ID` for logs starting with `repo-updater` that indicate requests to the frontend service are failing.
- **Kubernetes:**
	- Confirm that `kubectl get pods` shows the `frontend` pods are healthy.
	- Check `kubectl logs precise-code-intel-indexer` for logs indicate request failures to `frontend` or `frontend-internal`.
- **Docker Compose:**
	- Confirm that `docker ps` shows the `frontend-internal` container is healthy.
	- Check `docker logs precise-code-intel-indexer` for logs indicating request failures to `frontend` or `frontend-internal`.

# precise-code-intel-indexer: container_restarts

**Descriptions:**

- _precise-code-intel-indexer: 1+ container restarts every 5m by instance_ (`warning_precise-code-intel-indexer_container_restarts`)

**Possible solutions:**

- **Kubernetes:**
	- Determine if the pod was OOM killed using `kubectl describe pod precise-code-intel-indexer` (look for `OOMKilled: true`) and, if so, consider increasing the memory limit in the relevant `Deployment.yaml`.
	- Check the logs before the container restarted to see if there are `panic:` messages or similar using `kubectl logs -p precise-code-intel-indexer`.
- **Docker Compose:**
	- Determine if the pod was OOM killed using `docker inspect -f '{{json .State}}' precise-code-intel-indexer` (look for `"OOMKilled":true`) and, if so, consider increasing the memory limit of the precise-code-intel-indexer container in `docker-compose.yml`.
	- Check the logs before the container restarted to see if there are `panic:` messages or similar using `docker logs precise-code-intel-indexer` (note this will include logs from the previous and currently running container).

# precise-code-intel-indexer: container_memory_usage

**Descriptions:**

- _precise-code-intel-indexer: 99%+ container memory usage by instance_ (`warning_precise-code-intel-indexer_container_memory_usage`)

**Possible solutions:**

- **Kubernetes:** Consider increasing memory limit in relevant `Deployment.yaml`.
- **Docker Compose:** Consider increasing `memory:` of precise-code-intel-indexer container in `docker-compose.yml`.

# precise-code-intel-indexer: container_cpu_usage

**Descriptions:**

- _precise-code-intel-indexer: 99%+ container cpu usage total (1m average) across all cores by instance_ (`warning_precise-code-intel-indexer_container_cpu_usage`)

**Possible solutions:**

- **Kubernetes:** Consider increasing CPU limits in the the relevant `Deployment.yaml`.
- **Docker Compose:** Consider increasing `cpus:` of the precise-code-intel-indexer container in `docker-compose.yml`.

# precise-code-intel-indexer: provisioning_container_cpu_usage_7d

**Descriptions:**

- _precise-code-intel-indexer: 80%+ or less than 30% container cpu usage total (7d maximum) across all cores by instance_ (`warning_precise-code-intel-indexer_provisioning_container_cpu_usage_7d`)

**Possible solutions:**

- If usage is high:
	- **Kubernetes:** Consider decreasing CPU limits in the the relevant `Deployment.yaml`.
	- **Docker Compose:** Consider descreasing `cpus:` of the precise-code-intel-indexer container in `docker-compose.yml`.
- If usage is low, consider decreasing the above values.

# precise-code-intel-indexer: provisioning_container_memory_usage_7d

**Descriptions:**

- _precise-code-intel-indexer: 80%+ or less than 30% container memory usage (7d maximum) by instance_ (`warning_precise-code-intel-indexer_provisioning_container_memory_usage_7d`)

**Possible solutions:**

- If usage is high:
	- **Kubernetes:** Consider decreasing memory limit in relevant `Deployment.yaml`.
	- **Docker Compose:** Consider decreasing `memory:` of precise-code-intel-indexer container in `docker-compose.yml`.
- If usage is low, consider decreasing the above values.

# precise-code-intel-indexer: provisioning_container_cpu_usage_5m

**Descriptions:**

- _precise-code-intel-indexer: 90%+ container cpu usage total (5m maximum) across all cores by instance_ (`warning_precise-code-intel-indexer_provisioning_container_cpu_usage_5m`)

**Possible solutions:**

- **Kubernetes:** Consider increasing CPU limits in the the relevant `Deployment.yaml`.
- **Docker Compose:** Consider increasing `cpus:` of the precise-code-intel-indexer container in `docker-compose.yml`.

# precise-code-intel-indexer: provisioning_container_memory_usage_5m

**Descriptions:**

- _precise-code-intel-indexer: 90%+ container memory usage (5m maximum) by instance_ (`warning_precise-code-intel-indexer_provisioning_container_memory_usage_5m`)

**Possible solutions:**

- **Kubernetes:** Consider increasing memory limit in relevant `Deployment.yaml`.
- **Docker Compose:** Consider increasing `memory:` of precise-code-intel-indexer container in `docker-compose.yml`.

# query-runner: frontend_internal_api_error_responses

**Descriptions:**

- _query-runner: 5+ frontend-internal API error responses every 5m by route_ (`warning_query-runner_frontend_internal_api_error_responses`)

**Possible solutions:**

- **Single-container deployments:** Check `docker logs $CONTAINER_ID` for logs starting with `repo-updater` that indicate requests to the frontend service are failing.
- **Kubernetes:**
	- Confirm that `kubectl get pods` shows the `frontend` pods are healthy.
	- Check `kubectl logs query-runner` for logs indicate request failures to `frontend` or `frontend-internal`.
- **Docker Compose:**
	- Confirm that `docker ps` shows the `frontend-internal` container is healthy.
	- Check `docker logs query-runner` for logs indicating request failures to `frontend` or `frontend-internal`.

# query-runner: container_restarts

**Descriptions:**

- _query-runner: 1+ container restarts every 5m by instance_ (`warning_query-runner_container_restarts`)

**Possible solutions:**

- **Kubernetes:**
	- Determine if the pod was OOM killed using `kubectl describe pod query-runner` (look for `OOMKilled: true`) and, if so, consider increasing the memory limit in the relevant `Deployment.yaml`.
	- Check the logs before the container restarted to see if there are `panic:` messages or similar using `kubectl logs -p query-runner`.
- **Docker Compose:**
	- Determine if the pod was OOM killed using `docker inspect -f '{{json .State}}' query-runner` (look for `"OOMKilled":true`) and, if so, consider increasing the memory limit of the query-runner container in `docker-compose.yml`.
	- Check the logs before the container restarted to see if there are `panic:` messages or similar using `docker logs query-runner` (note this will include logs from the previous and currently running container).

# query-runner: container_memory_usage

**Descriptions:**

- _query-runner: 99%+ container memory usage by instance_ (`warning_query-runner_container_memory_usage`)

**Possible solutions:**

- **Kubernetes:** Consider increasing memory limit in relevant `Deployment.yaml`.
- **Docker Compose:** Consider increasing `memory:` of query-runner container in `docker-compose.yml`.

# query-runner: container_cpu_usage

**Descriptions:**

- _query-runner: 99%+ container cpu usage total (1m average) across all cores by instance_ (`warning_query-runner_container_cpu_usage`)

**Possible solutions:**

- **Kubernetes:** Consider increasing CPU limits in the the relevant `Deployment.yaml`.
- **Docker Compose:** Consider increasing `cpus:` of the query-runner container in `docker-compose.yml`.

# query-runner: provisioning_container_cpu_usage_7d

**Descriptions:**

- _query-runner: 80%+ or less than 30% container cpu usage total (7d maximum) across all cores by instance_ (`warning_query-runner_provisioning_container_cpu_usage_7d`)

**Possible solutions:**

- If usage is high:
	- **Kubernetes:** Consider decreasing CPU limits in the the relevant `Deployment.yaml`.
	- **Docker Compose:** Consider descreasing `cpus:` of the query-runner container in `docker-compose.yml`.
- If usage is low, consider decreasing the above values.

# query-runner: provisioning_container_memory_usage_7d

**Descriptions:**

- _query-runner: 80%+ or less than 30% container memory usage (7d maximum) by instance_ (`warning_query-runner_provisioning_container_memory_usage_7d`)

**Possible solutions:**

- If usage is high:
	- **Kubernetes:** Consider decreasing memory limit in relevant `Deployment.yaml`.
	- **Docker Compose:** Consider decreasing `memory:` of query-runner container in `docker-compose.yml`.
- If usage is low, consider decreasing the above values.

# query-runner: provisioning_container_cpu_usage_5m

**Descriptions:**

- _query-runner: 90%+ container cpu usage total (5m maximum) across all cores by instance_ (`warning_query-runner_provisioning_container_cpu_usage_5m`)

**Possible solutions:**

- **Kubernetes:** Consider increasing CPU limits in the the relevant `Deployment.yaml`.
- **Docker Compose:** Consider increasing `cpus:` of the query-runner container in `docker-compose.yml`.

# query-runner: provisioning_container_memory_usage_5m

**Descriptions:**

- _query-runner: 90%+ container memory usage (5m maximum) by instance_ (`warning_query-runner_provisioning_container_memory_usage_5m`)

**Possible solutions:**

- **Kubernetes:** Consider increasing memory limit in relevant `Deployment.yaml`.
- **Docker Compose:** Consider increasing `memory:` of query-runner container in `docker-compose.yml`.

# replacer: frontend_internal_api_error_responses

**Descriptions:**

- _replacer: 5+ frontend-internal API error responses every 5m by route_ (`warning_replacer_frontend_internal_api_error_responses`)

**Possible solutions:**

- **Single-container deployments:** Check `docker logs $CONTAINER_ID` for logs starting with `repo-updater` that indicate requests to the frontend service are failing.
- **Kubernetes:**
	- Confirm that `kubectl get pods` shows the `frontend` pods are healthy.
	- Check `kubectl logs replacer` for logs indicate request failures to `frontend` or `frontend-internal`.
- **Docker Compose:**
	- Confirm that `docker ps` shows the `frontend-internal` container is healthy.
	- Check `docker logs replacer` for logs indicating request failures to `frontend` or `frontend-internal`.

# replacer: container_restarts

**Descriptions:**

- _replacer: 1+ container restarts every 5m by instance_ (`warning_replacer_container_restarts`)

**Possible solutions:**

- **Kubernetes:**
	- Determine if the pod was OOM killed using `kubectl describe pod replacer` (look for `OOMKilled: true`) and, if so, consider increasing the memory limit in the relevant `Deployment.yaml`.
	- Check the logs before the container restarted to see if there are `panic:` messages or similar using `kubectl logs -p replacer`.
- **Docker Compose:**
	- Determine if the pod was OOM killed using `docker inspect -f '{{json .State}}' replacer` (look for `"OOMKilled":true`) and, if so, consider increasing the memory limit of the replacer container in `docker-compose.yml`.
	- Check the logs before the container restarted to see if there are `panic:` messages or similar using `docker logs replacer` (note this will include logs from the previous and currently running container).

# replacer: container_memory_usage

**Descriptions:**

- _replacer: 99%+ container memory usage by instance_ (`warning_replacer_container_memory_usage`)

**Possible solutions:**

- **Kubernetes:** Consider increasing memory limit in relevant `Deployment.yaml`.
- **Docker Compose:** Consider increasing `memory:` of replacer container in `docker-compose.yml`.

# replacer: container_cpu_usage

**Descriptions:**

- _replacer: 99%+ container cpu usage total (1m average) across all cores by instance_ (`warning_replacer_container_cpu_usage`)

**Possible solutions:**

- **Kubernetes:** Consider increasing CPU limits in the the relevant `Deployment.yaml`.
- **Docker Compose:** Consider increasing `cpus:` of the replacer container in `docker-compose.yml`.

# replacer: provisioning_container_cpu_usage_7d

**Descriptions:**

- _replacer: 80%+ or less than 30% container cpu usage total (7d maximum) across all cores by instance_ (`warning_replacer_provisioning_container_cpu_usage_7d`)

**Possible solutions:**

- If usage is high:
	- **Kubernetes:** Consider decreasing CPU limits in the the relevant `Deployment.yaml`.
	- **Docker Compose:** Consider descreasing `cpus:` of the replacer container in `docker-compose.yml`.
- If usage is low, consider decreasing the above values.

# replacer: provisioning_container_memory_usage_7d

**Descriptions:**

- _replacer: 80%+ or less than 30% container memory usage (7d maximum) by instance_ (`warning_replacer_provisioning_container_memory_usage_7d`)

**Possible solutions:**

- If usage is high:
	- **Kubernetes:** Consider decreasing memory limit in relevant `Deployment.yaml`.
	- **Docker Compose:** Consider decreasing `memory:` of replacer container in `docker-compose.yml`.
- If usage is low, consider decreasing the above values.

# replacer: provisioning_container_cpu_usage_5m

**Descriptions:**

- _replacer: 90%+ container cpu usage total (5m maximum) across all cores by instance_ (`warning_replacer_provisioning_container_cpu_usage_5m`)

**Possible solutions:**

- **Kubernetes:** Consider increasing CPU limits in the the relevant `Deployment.yaml`.
- **Docker Compose:** Consider increasing `cpus:` of the replacer container in `docker-compose.yml`.

# replacer: provisioning_container_memory_usage_5m

**Descriptions:**

- _replacer: 90%+ container memory usage (5m maximum) by instance_ (`warning_replacer_provisioning_container_memory_usage_5m`)

**Possible solutions:**

- **Kubernetes:** Consider increasing memory limit in relevant `Deployment.yaml`.
- **Docker Compose:** Consider increasing `memory:` of replacer container in `docker-compose.yml`.

# repo-updater: frontend_internal_api_error_responses

**Descriptions:**

- _repo-updater: 5+ frontend-internal API error responses every 5m by route_ (`warning_repo-updater_frontend_internal_api_error_responses`)

**Possible solutions:**

- **Single-container deployments:** Check `docker logs $CONTAINER_ID` for logs starting with `repo-updater` that indicate requests to the frontend service are failing.
- **Kubernetes:**
	- Confirm that `kubectl get pods` shows the `frontend` pods are healthy.
	- Check `kubectl logs repo-updater` for logs indicate request failures to `frontend` or `frontend-internal`.
- **Docker Compose:**
	- Confirm that `docker ps` shows the `frontend-internal` container is healthy.
	- Check `docker logs repo-updater` for logs indicating request failures to `frontend` or `frontend-internal`.

# repo-updater: container_restarts

**Descriptions:**

- _repo-updater: 1+ container restarts every 5m by instance_ (`warning_repo-updater_container_restarts`)

**Possible solutions:**

- **Kubernetes:**
	- Determine if the pod was OOM killed using `kubectl describe pod repo-updater` (look for `OOMKilled: true`) and, if so, consider increasing the memory limit in the relevant `Deployment.yaml`.
	- Check the logs before the container restarted to see if there are `panic:` messages or similar using `kubectl logs -p repo-updater`.
- **Docker Compose:**
	- Determine if the pod was OOM killed using `docker inspect -f '{{json .State}}' repo-updater` (look for `"OOMKilled":true`) and, if so, consider increasing the memory limit of the repo-updater container in `docker-compose.yml`.
	- Check the logs before the container restarted to see if there are `panic:` messages or similar using `docker logs repo-updater` (note this will include logs from the previous and currently running container).

# repo-updater: container_memory_usage

**Descriptions:**

- _repo-updater: 99%+ container memory usage by instance_ (`warning_repo-updater_container_memory_usage`)

**Possible solutions:**

- **Kubernetes:** Consider increasing memory limit in relevant `Deployment.yaml`.
- **Docker Compose:** Consider increasing `memory:` of repo-updater container in `docker-compose.yml`.

# repo-updater: container_cpu_usage

**Descriptions:**

- _repo-updater: 99%+ container cpu usage total (1m average) across all cores by instance_ (`warning_repo-updater_container_cpu_usage`)

**Possible solutions:**

- **Kubernetes:** Consider increasing CPU limits in the the relevant `Deployment.yaml`.
- **Docker Compose:** Consider increasing `cpus:` of the repo-updater container in `docker-compose.yml`.

# repo-updater: provisioning_container_cpu_usage_7d

**Descriptions:**

- _repo-updater: 80%+ or less than 30% container cpu usage total (7d maximum) across all cores by instance_ (`warning_repo-updater_provisioning_container_cpu_usage_7d`)

**Possible solutions:**

- If usage is high:
	- **Kubernetes:** Consider decreasing CPU limits in the the relevant `Deployment.yaml`.
	- **Docker Compose:** Consider descreasing `cpus:` of the repo-updater container in `docker-compose.yml`.
- If usage is low, consider decreasing the above values.

# repo-updater: provisioning_container_memory_usage_7d

**Descriptions:**

- _repo-updater: 80%+ or less than 30% container memory usage (7d maximum) by instance_ (`warning_repo-updater_provisioning_container_memory_usage_7d`)

**Possible solutions:**

- If usage is high:
	- **Kubernetes:** Consider decreasing memory limit in relevant `Deployment.yaml`.
	- **Docker Compose:** Consider decreasing `memory:` of repo-updater container in `docker-compose.yml`.
- If usage is low, consider decreasing the above values.

# repo-updater: provisioning_container_cpu_usage_5m

**Descriptions:**

- _repo-updater: 90%+ container cpu usage total (5m maximum) across all cores by instance_ (`warning_repo-updater_provisioning_container_cpu_usage_5m`)

**Possible solutions:**

- **Kubernetes:** Consider increasing CPU limits in the the relevant `Deployment.yaml`.
- **Docker Compose:** Consider increasing `cpus:` of the repo-updater container in `docker-compose.yml`.

# repo-updater: provisioning_container_memory_usage_5m

**Descriptions:**

- _repo-updater: 90%+ container memory usage (5m maximum) by instance_ (`warning_repo-updater_provisioning_container_memory_usage_5m`)

**Possible solutions:**

- **Kubernetes:** Consider increasing memory limit in relevant `Deployment.yaml`.
- **Docker Compose:** Consider increasing `memory:` of repo-updater container in `docker-compose.yml`.

# searcher: frontend_internal_api_error_responses

**Descriptions:**

- _searcher: 5+ frontend-internal API error responses every 5m by route_ (`warning_searcher_frontend_internal_api_error_responses`)

**Possible solutions:**

- **Single-container deployments:** Check `docker logs $CONTAINER_ID` for logs starting with `repo-updater` that indicate requests to the frontend service are failing.
- **Kubernetes:**
	- Confirm that `kubectl get pods` shows the `frontend` pods are healthy.
	- Check `kubectl logs searcher` for logs indicate request failures to `frontend` or `frontend-internal`.
- **Docker Compose:**
	- Confirm that `docker ps` shows the `frontend-internal` container is healthy.
	- Check `docker logs searcher` for logs indicating request failures to `frontend` or `frontend-internal`.

# searcher: container_restarts

**Descriptions:**

- _searcher: 1+ container restarts every 5m by instance_ (`warning_searcher_container_restarts`)

**Possible solutions:**

- **Kubernetes:**
	- Determine if the pod was OOM killed using `kubectl describe pod searcher` (look for `OOMKilled: true`) and, if so, consider increasing the memory limit in the relevant `Deployment.yaml`.
	- Check the logs before the container restarted to see if there are `panic:` messages or similar using `kubectl logs -p searcher`.
- **Docker Compose:**
	- Determine if the pod was OOM killed using `docker inspect -f '{{json .State}}' searcher` (look for `"OOMKilled":true`) and, if so, consider increasing the memory limit of the searcher container in `docker-compose.yml`.
	- Check the logs before the container restarted to see if there are `panic:` messages or similar using `docker logs searcher` (note this will include logs from the previous and currently running container).

# searcher: container_memory_usage

**Descriptions:**

- _searcher: 99%+ container memory usage by instance_ (`warning_searcher_container_memory_usage`)

**Possible solutions:**

- **Kubernetes:** Consider increasing memory limit in relevant `Deployment.yaml`.
- **Docker Compose:** Consider increasing `memory:` of searcher container in `docker-compose.yml`.

# searcher: container_cpu_usage

**Descriptions:**

- _searcher: 99%+ container cpu usage total (1m average) across all cores by instance_ (`warning_searcher_container_cpu_usage`)

**Possible solutions:**

- **Kubernetes:** Consider increasing CPU limits in the the relevant `Deployment.yaml`.
- **Docker Compose:** Consider increasing `cpus:` of the searcher container in `docker-compose.yml`.

# searcher: provisioning_container_cpu_usage_7d

**Descriptions:**

- _searcher: 80%+ or less than 30% container cpu usage total (7d maximum) across all cores by instance_ (`warning_searcher_provisioning_container_cpu_usage_7d`)

**Possible solutions:**

- If usage is high:
	- **Kubernetes:** Consider decreasing CPU limits in the the relevant `Deployment.yaml`.
	- **Docker Compose:** Consider descreasing `cpus:` of the searcher container in `docker-compose.yml`.
- If usage is low, consider decreasing the above values.

# searcher: provisioning_container_memory_usage_7d

**Descriptions:**

- _searcher: 80%+ or less than 30% container memory usage (7d maximum) by instance_ (`warning_searcher_provisioning_container_memory_usage_7d`)

**Possible solutions:**

- If usage is high:
	- **Kubernetes:** Consider decreasing memory limit in relevant `Deployment.yaml`.
	- **Docker Compose:** Consider decreasing `memory:` of searcher container in `docker-compose.yml`.
- If usage is low, consider decreasing the above values.

# searcher: provisioning_container_cpu_usage_5m

**Descriptions:**

- _searcher: 90%+ container cpu usage total (5m maximum) across all cores by instance_ (`warning_searcher_provisioning_container_cpu_usage_5m`)

**Possible solutions:**

- **Kubernetes:** Consider increasing CPU limits in the the relevant `Deployment.yaml`.
- **Docker Compose:** Consider increasing `cpus:` of the searcher container in `docker-compose.yml`.

# searcher: provisioning_container_memory_usage_5m

**Descriptions:**

- _searcher: 90%+ container memory usage (5m maximum) by instance_ (`warning_searcher_provisioning_container_memory_usage_5m`)

**Possible solutions:**

- **Kubernetes:** Consider increasing memory limit in relevant `Deployment.yaml`.
- **Docker Compose:** Consider increasing `memory:` of searcher container in `docker-compose.yml`.

# symbols: frontend_internal_api_error_responses

**Descriptions:**

- _symbols: 5+ frontend-internal API error responses every 5m by route_ (`warning_symbols_frontend_internal_api_error_responses`)

**Possible solutions:**

- **Single-container deployments:** Check `docker logs $CONTAINER_ID` for logs starting with `repo-updater` that indicate requests to the frontend service are failing.
- **Kubernetes:**
	- Confirm that `kubectl get pods` shows the `frontend` pods are healthy.
	- Check `kubectl logs symbols` for logs indicate request failures to `frontend` or `frontend-internal`.
- **Docker Compose:**
	- Confirm that `docker ps` shows the `frontend-internal` container is healthy.
	- Check `docker logs symbols` for logs indicating request failures to `frontend` or `frontend-internal`.

# symbols: container_restarts

**Descriptions:**

- _symbols: 1+ container restarts every 5m by instance_ (`warning_symbols_container_restarts`)

**Possible solutions:**

- **Kubernetes:**
	- Determine if the pod was OOM killed using `kubectl describe pod symbols` (look for `OOMKilled: true`) and, if so, consider increasing the memory limit in the relevant `Deployment.yaml`.
	- Check the logs before the container restarted to see if there are `panic:` messages or similar using `kubectl logs -p symbols`.
- **Docker Compose:**
	- Determine if the pod was OOM killed using `docker inspect -f '{{json .State}}' symbols` (look for `"OOMKilled":true`) and, if so, consider increasing the memory limit of the symbols container in `docker-compose.yml`.
	- Check the logs before the container restarted to see if there are `panic:` messages or similar using `docker logs symbols` (note this will include logs from the previous and currently running container).

# symbols: container_memory_usage

**Descriptions:**

- _symbols: 99%+ container memory usage by instance_ (`warning_symbols_container_memory_usage`)

**Possible solutions:**

- **Kubernetes:** Consider increasing memory limit in relevant `Deployment.yaml`.
- **Docker Compose:** Consider increasing `memory:` of symbols container in `docker-compose.yml`.

# symbols: container_cpu_usage

**Descriptions:**

- _symbols: 99%+ container cpu usage total (1m average) across all cores by instance_ (`warning_symbols_container_cpu_usage`)

**Possible solutions:**

- **Kubernetes:** Consider increasing CPU limits in the the relevant `Deployment.yaml`.
- **Docker Compose:** Consider increasing `cpus:` of the symbols container in `docker-compose.yml`.

# symbols: provisioning_container_cpu_usage_7d

**Descriptions:**

- _symbols: 80%+ or less than 30% container cpu usage total (7d maximum) across all cores by instance_ (`warning_symbols_provisioning_container_cpu_usage_7d`)

**Possible solutions:**

- If usage is high:
	- **Kubernetes:** Consider decreasing CPU limits in the the relevant `Deployment.yaml`.
	- **Docker Compose:** Consider descreasing `cpus:` of the symbols container in `docker-compose.yml`.
- If usage is low, consider decreasing the above values.

# symbols: provisioning_container_memory_usage_7d

**Descriptions:**

- _symbols: 80%+ or less than 30% container memory usage (7d maximum) by instance_ (`warning_symbols_provisioning_container_memory_usage_7d`)

**Possible solutions:**

- If usage is high:
	- **Kubernetes:** Consider decreasing memory limit in relevant `Deployment.yaml`.
	- **Docker Compose:** Consider decreasing `memory:` of symbols container in `docker-compose.yml`.
- If usage is low, consider decreasing the above values.

# symbols: provisioning_container_cpu_usage_5m

**Descriptions:**

- _symbols: 90%+ container cpu usage total (5m maximum) across all cores by instance_ (`warning_symbols_provisioning_container_cpu_usage_5m`)

**Possible solutions:**

- **Kubernetes:** Consider increasing CPU limits in the the relevant `Deployment.yaml`.
- **Docker Compose:** Consider increasing `cpus:` of the symbols container in `docker-compose.yml`.

# symbols: provisioning_container_memory_usage_5m

**Descriptions:**

- _symbols: 90%+ container memory usage (5m maximum) by instance_ (`warning_symbols_provisioning_container_memory_usage_5m`)

**Possible solutions:**

- **Kubernetes:** Consider increasing memory limit in relevant `Deployment.yaml`.
- **Docker Compose:** Consider increasing `memory:` of symbols container in `docker-compose.yml`.

# syntect-server: container_restarts

**Descriptions:**

- _syntect-server: 1+ container restarts every 5m by instance_ (`warning_syntect-server_container_restarts`)

**Possible solutions:**

- **Kubernetes:**
	- Determine if the pod was OOM killed using `kubectl describe pod syntect-server` (look for `OOMKilled: true`) and, if so, consider increasing the memory limit in the relevant `Deployment.yaml`.
	- Check the logs before the container restarted to see if there are `panic:` messages or similar using `kubectl logs -p syntect-server`.
- **Docker Compose:**
	- Determine if the pod was OOM killed using `docker inspect -f '{{json .State}}' syntect-server` (look for `"OOMKilled":true`) and, if so, consider increasing the memory limit of the syntect-server container in `docker-compose.yml`.
	- Check the logs before the container restarted to see if there are `panic:` messages or similar using `docker logs syntect-server` (note this will include logs from the previous and currently running container).

# syntect-server: container_memory_usage

**Descriptions:**

- _syntect-server: 99%+ container memory usage by instance_ (`warning_syntect-server_container_memory_usage`)

**Possible solutions:**

- **Kubernetes:** Consider increasing memory limit in relevant `Deployment.yaml`.
- **Docker Compose:** Consider increasing `memory:` of syntect-server container in `docker-compose.yml`.

# syntect-server: container_cpu_usage

**Descriptions:**

- _syntect-server: 99%+ container cpu usage total (1m average) across all cores by instance_ (`warning_syntect-server_container_cpu_usage`)

**Possible solutions:**

- **Kubernetes:** Consider increasing CPU limits in the the relevant `Deployment.yaml`.
- **Docker Compose:** Consider increasing `cpus:` of the syntect-server container in `docker-compose.yml`.

# syntect-server: provisioning_container_cpu_usage_7d

**Descriptions:**

- _syntect-server: 80%+ or less than 30% container cpu usage total (7d maximum) across all cores by instance_ (`warning_syntect-server_provisioning_container_cpu_usage_7d`)

**Possible solutions:**

- If usage is high:
	- **Kubernetes:** Consider decreasing CPU limits in the the relevant `Deployment.yaml`.
	- **Docker Compose:** Consider descreasing `cpus:` of the syntect-server container in `docker-compose.yml`.
- If usage is low, consider decreasing the above values.

# syntect-server: provisioning_container_memory_usage_7d

**Descriptions:**

- _syntect-server: 80%+ or less than 30% container memory usage (7d maximum) by instance_ (`warning_syntect-server_provisioning_container_memory_usage_7d`)

**Possible solutions:**

- If usage is high:
	- **Kubernetes:** Consider decreasing memory limit in relevant `Deployment.yaml`.
	- **Docker Compose:** Consider decreasing `memory:` of syntect-server container in `docker-compose.yml`.
- If usage is low, consider decreasing the above values.

# syntect-server: provisioning_container_cpu_usage_5m

**Descriptions:**

- _syntect-server: 90%+ container cpu usage total (5m maximum) across all cores by instance_ (`warning_syntect-server_provisioning_container_cpu_usage_5m`)

**Possible solutions:**

- **Kubernetes:** Consider increasing CPU limits in the the relevant `Deployment.yaml`.
- **Docker Compose:** Consider increasing `cpus:` of the syntect-server container in `docker-compose.yml`.

# syntect-server: provisioning_container_memory_usage_5m

**Descriptions:**

- _syntect-server: 90%+ container memory usage (5m maximum) by instance_ (`warning_syntect-server_provisioning_container_memory_usage_5m`)

**Possible solutions:**

- **Kubernetes:** Consider increasing memory limit in relevant `Deployment.yaml`.
- **Docker Compose:** Consider increasing `memory:` of syntect-server container in `docker-compose.yml`.

# zoekt-indexserver: container_restarts

**Descriptions:**

- _zoekt-indexserver: 1+ container restarts every 5m by instance_ (`warning_zoekt-indexserver_container_restarts`)

**Possible solutions:**

- **Kubernetes:**
	- Determine if the pod was OOM killed using `kubectl describe pod zoekt-indexserver` (look for `OOMKilled: true`) and, if so, consider increasing the memory limit in the relevant `Deployment.yaml`.
	- Check the logs before the container restarted to see if there are `panic:` messages or similar using `kubectl logs -p zoekt-indexserver`.
- **Docker Compose:**
	- Determine if the pod was OOM killed using `docker inspect -f '{{json .State}}' zoekt-indexserver` (look for `"OOMKilled":true`) and, if so, consider increasing the memory limit of the zoekt-indexserver container in `docker-compose.yml`.
	- Check the logs before the container restarted to see if there are `panic:` messages or similar using `docker logs zoekt-indexserver` (note this will include logs from the previous and currently running container).

# zoekt-indexserver: container_memory_usage

**Descriptions:**

- _zoekt-indexserver: 99%+ container memory usage by instance_ (`warning_zoekt-indexserver_container_memory_usage`)

**Possible solutions:**

- **Kubernetes:** Consider increasing memory limit in relevant `Deployment.yaml`.
- **Docker Compose:** Consider increasing `memory:` of zoekt-indexserver container in `docker-compose.yml`.

# zoekt-indexserver: container_cpu_usage

**Descriptions:**

- _zoekt-indexserver: 99%+ container cpu usage total (1m average) across all cores by instance_ (`warning_zoekt-indexserver_container_cpu_usage`)

**Possible solutions:**

- **Kubernetes:** Consider increasing CPU limits in the the relevant `Deployment.yaml`.
- **Docker Compose:** Consider increasing `cpus:` of the zoekt-indexserver container in `docker-compose.yml`.

# zoekt-indexserver: provisioning_container_cpu_usage_7d

**Descriptions:**

- _zoekt-indexserver: 80%+ or less than 30% container cpu usage total (7d maximum) across all cores by instance_ (`warning_zoekt-indexserver_provisioning_container_cpu_usage_7d`)

**Possible solutions:**

- If usage is high:
	- **Kubernetes:** Consider decreasing CPU limits in the the relevant `Deployment.yaml`.
	- **Docker Compose:** Consider descreasing `cpus:` of the zoekt-indexserver container in `docker-compose.yml`.
- If usage is low, consider decreasing the above values.

# zoekt-indexserver: provisioning_container_memory_usage_7d

**Descriptions:**

- _zoekt-indexserver: 80%+ or less than 30% container memory usage (7d maximum) by instance_ (`warning_zoekt-indexserver_provisioning_container_memory_usage_7d`)

**Possible solutions:**

- If usage is high:
	- **Kubernetes:** Consider decreasing memory limit in relevant `Deployment.yaml`.
	- **Docker Compose:** Consider decreasing `memory:` of zoekt-indexserver container in `docker-compose.yml`.
- If usage is low, consider decreasing the above values.

# zoekt-indexserver: provisioning_container_cpu_usage_5m

**Descriptions:**

- _zoekt-indexserver: 90%+ container cpu usage total (5m maximum) across all cores by instance_ (`warning_zoekt-indexserver_provisioning_container_cpu_usage_5m`)

**Possible solutions:**

- **Kubernetes:** Consider increasing CPU limits in the the relevant `Deployment.yaml`.
- **Docker Compose:** Consider increasing `cpus:` of the zoekt-indexserver container in `docker-compose.yml`.

# zoekt-indexserver: provisioning_container_memory_usage_5m

**Descriptions:**

- _zoekt-indexserver: 90%+ container memory usage (5m maximum) by instance_ (`warning_zoekt-indexserver_provisioning_container_memory_usage_5m`)

**Possible solutions:**

- **Kubernetes:** Consider increasing memory limit in relevant `Deployment.yaml`.
- **Docker Compose:** Consider increasing `memory:` of zoekt-indexserver container in `docker-compose.yml`.

# zoekt-webserver: container_restarts

**Descriptions:**

- _zoekt-webserver: 1+ container restarts every 5m by instance_ (`warning_zoekt-webserver_container_restarts`)

**Possible solutions:**

- **Kubernetes:**
	- Determine if the pod was OOM killed using `kubectl describe pod zoekt-webserver` (look for `OOMKilled: true`) and, if so, consider increasing the memory limit in the relevant `Deployment.yaml`.
	- Check the logs before the container restarted to see if there are `panic:` messages or similar using `kubectl logs -p zoekt-webserver`.
- **Docker Compose:**
	- Determine if the pod was OOM killed using `docker inspect -f '{{json .State}}' zoekt-webserver` (look for `"OOMKilled":true`) and, if so, consider increasing the memory limit of the zoekt-webserver container in `docker-compose.yml`.
	- Check the logs before the container restarted to see if there are `panic:` messages or similar using `docker logs zoekt-webserver` (note this will include logs from the previous and currently running container).

# zoekt-webserver: container_memory_usage

**Descriptions:**

- _zoekt-webserver: 99%+ container memory usage by instance_ (`warning_zoekt-webserver_container_memory_usage`)

**Possible solutions:**

- **Kubernetes:** Consider increasing memory limit in relevant `Deployment.yaml`.
- **Docker Compose:** Consider increasing `memory:` of zoekt-webserver container in `docker-compose.yml`.

# zoekt-webserver: container_cpu_usage

**Descriptions:**

- _zoekt-webserver: 99%+ container cpu usage total (1m average) across all cores by instance_ (`warning_zoekt-webserver_container_cpu_usage`)

**Possible solutions:**

- **Kubernetes:** Consider increasing CPU limits in the the relevant `Deployment.yaml`.
- **Docker Compose:** Consider increasing `cpus:` of the zoekt-webserver container in `docker-compose.yml`.

# zoekt-webserver: provisioning_container_cpu_usage_7d

**Descriptions:**

- _zoekt-webserver: 80%+ or less than 30% container cpu usage total (7d maximum) across all cores by instance_ (`warning_zoekt-webserver_provisioning_container_cpu_usage_7d`)

**Possible solutions:**

- If usage is high:
	- **Kubernetes:** Consider decreasing CPU limits in the the relevant `Deployment.yaml`.
	- **Docker Compose:** Consider descreasing `cpus:` of the zoekt-webserver container in `docker-compose.yml`.
- If usage is low, consider decreasing the above values.

# zoekt-webserver: provisioning_container_memory_usage_7d

**Descriptions:**

- _zoekt-webserver: 80%+ or less than 30% container memory usage (7d maximum) by instance_ (`warning_zoekt-webserver_provisioning_container_memory_usage_7d`)

**Possible solutions:**

- If usage is high:
	- **Kubernetes:** Consider decreasing memory limit in relevant `Deployment.yaml`.
	- **Docker Compose:** Consider decreasing `memory:` of zoekt-webserver container in `docker-compose.yml`.
- If usage is low, consider decreasing the above values.

# zoekt-webserver: provisioning_container_cpu_usage_5m

**Descriptions:**

- _zoekt-webserver: 90%+ container cpu usage total (5m maximum) across all cores by instance_ (`warning_zoekt-webserver_provisioning_container_cpu_usage_5m`)

**Possible solutions:**

- **Kubernetes:** Consider increasing CPU limits in the the relevant `Deployment.yaml`.
- **Docker Compose:** Consider increasing `cpus:` of the zoekt-webserver container in `docker-compose.yml`.

# zoekt-webserver: provisioning_container_memory_usage_5m

**Descriptions:**

- _zoekt-webserver: 90%+ container memory usage (5m maximum) by instance_ (`warning_zoekt-webserver_provisioning_container_memory_usage_5m`)

**Possible solutions:**

- **Kubernetes:** Consider increasing memory limit in relevant `Deployment.yaml`.
- **Docker Compose:** Consider increasing `memory:` of zoekt-webserver container in `docker-compose.yml`.

