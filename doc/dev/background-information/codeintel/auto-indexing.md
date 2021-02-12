# How Sourcegraph auto-indexes source code

_Auto-indexing is enabled only in the Cloud environment and are written to work well for the usage patterns found there. Once we have proven that auto-indexing would also be beneficial in private instances, we will consider making the feature available there as well._

## Scheduling

The [IndexabilityUpdater](https://sourcegraph.com/github.com/sourcegraph/sourcegraph@fc5a827bfaf5bcd5c9fb2fb05148ef687dd56a2e/-/blob/enterprise/cmd/frontend/internal/codeintel/background/indexability_updater.go#L52:31) periodically [updates a database table](https://sourcegraph.com/github.com/sourcegraph/sourcegraph@fc5a827/-/blob/enterprise/internal/codeintel/stores/dbstore/repo_usage.go#L47:17) that aggregates code intelligence events into a list of repositories orderable by their popularity (or a close proxy thereof).

The [IndexScheduler](https://sourcegraph.com/github.com/sourcegraph/sourcegraph@fc5a827bfaf5bcd5c9fb2fb05148ef687dd56a2e/-/blob/enterprise/cmd/frontend/internal/codeintel/background/index_scheduler.go#L52:26) will periodically [query the table](https://sourcegraph.com/github.com/sourcegraph/sourcegraph@fc5a827/-/blob/enterprise/internal/codeintel/stores/dbstore/indexable_repos.go#L71:17) maintained by the indexability updater for a batch of repositories to index. The ordering expression for this query takes several parameters into account:

- The time since the last index task was enqueued for this repository
- The number of precise code intel results for this repository in the last week
- The number of search-based code intel results for this repository in the last week
- The ratio of precise code intel results over total code intel results for this repository in the last week

Once the set of repositories to index have been determined, the set of steps required to index the repository are determined.

If a user has explicitly configured indexing steps for this repository, the configuration may be found in the [database](https://sourcegraph.com/github.com/sourcegraph/sourcegraph@fc5a827bfaf5bcd5c9fb2fb05148ef687dd56a2e/-/blob/enterprise/cmd/frontend/internal/codeintel/background/index_scheduler.go#L169:26) (configured via the UI), or in the [sourcegraph.yaml](https://sourcegraph.com/github.com/sourcegraph/sourcegraph@fc5a827bfaf5bcd5c9fb2fb05148ef687dd56a2e/-/blob/enterprise/cmd/frontend/internal/codeintel/background/index_scheduler.go#L190:26) configuration file in the root of the repository.

If no explicit configuration exists, the steps are [inferred from the repository structure](https://sourcegraph.com/github.com/sourcegraph/sourcegraph@fc5a827bfaf5bcd5c9fb2fb05148ef687dd56a2e/-/blob/enterprise/cmd/frontend/internal/codeintel/background/index_scheduler.go#L216:26). We currently support detection of projects in the following languages:

- [Go](https://sourcegraph.com/github.com/sourcegraph/sourcegraph@fc5a827bfaf5bcd5c9fb2fb05148ef687dd56a2e/-/blob/enterprise/cmd/frontend/internal/codeintel/autoindex/inference/go.go#L27:28)
- [TypeScript](https://sourcegraph.com/github.com/sourcegraph/sourcegraph@fc5a827bfaf5bcd5c9fb2fb05148ef687dd56a2e/-/blob/enterprise/cmd/frontend/internal/codeintel/autoindex/inference/typescript.go#L27:29)

The steps to index the repository are serialized into an index record and [inserted into a task queue](https://sourcegraph.com/github.com/sourcegraph/sourcegraph@fc5a827/-/blob/enterprise/internal/codeintel/stores/dbstore/indexes.go#L280:17) to be processed asynchronously by a pool of task executors.

## Processing

Because indexing an arbitrary code base may require arbitrary commands to be run (e.g., dependency gathering, compilation steps, code generation, etc), we process each index job in a [Firecracker](https://firecracker-microvm.github.io/) virtual machine managed by [Weave Ignite](https://ignite.readthedocs.io/en/stable/). These virtual machines are coordinated by the [executor](https://github.com/sourcegraph/sourcegraph/tree/main/enterprise/cmd/executor) service which is [deployed directly on GCP compute nodes](./deployment.md).

<a href="diagrams/executor.svg" target="_blank">
  <img src="diagrams/executor.svg">
</a>

The executor, deployed externally to the rest of the cluster, makes requests to the [executor-queue](https://github.com/sourcegraph/sourcegraph/tree/main/enterprise/cmd/executor-queue) and [gitserver](https://github.com/sourcegraph/sourcegraph/tree/main/cmd/gitserver) services via a proxy routes in the public API which are protected by a shared token.

When idle, the executor process will periodically poll the executor-queue asking for an index job. If one exists, the executor-queue will open a long-running database transaction in order to lock the record during processing. A periodic heartbeat request between the executor and the executor-queue will ensure that transactions do not stay permanently open if the executor crashes or becomes partitioned from the Sourcegraph instance.

On dequeue, a queued but unlocked row in the `lsif_indexes` table is locked and the record is transformed into a generic (non-code-intel-specific) task to be sent back to the executor. This payload consists of a sequence of docker and src-cli commands to run.

Once the executor receives a job, it will clone the target repository and checkout a target commit. A Firecracker virtual machine is started and the local git clone is copied into it. The commands determined by the executor-queue task translation layer are invoked inside of the virtual machine. Once done, the virtual machine is removed and a request is made ot the executor-queue to mark the index as successfully processed.

### Code appendix

- Executor: [Handle](https://sourcegraph.com/github.com/sourcegraph/sourcegraph@fc5a827bfaf5bcd5c9fb2fb05148ef687dd56a2e/-/blob/enterprise/internal/apiworker/handler.go#L31:19), [setupFirecracker](https://sourcegraph.com/github.com/sourcegraph/sourcegraph@fc5a827bfaf5bcd5c9fb2fb05148ef687dd56a2e/-/blob/enterprise/internal/apiworker/command/firecracker.go#L65:6), [formatFirecrackerCommand](https://sourcegraph.com/github.com/sourcegraph/sourcegraph@fc5a827bfaf5bcd5c9fb2fb05148ef687dd56a2e/-/blob/enterprise/internal/apiworker/command/firecracker.go#L37:6), [teardownFirecracker](https://sourcegraph.com/github.com/sourcegraph/sourcegraph@fc5a827bfaf5bcd5c9fb2fb05148ef687dd56a2e/-/blob/enterprise/internal/apiworker/command/firecracker.go#L131:6)
- Executor Proxy: [newInternalProxyHandler](https://sourcegraph.com/github.com/sourcegraph/sourcegraph@fc5a827bfaf5bcd5c9fb2fb05148ef687dd56a2e/-/blob/enterprise/cmd/frontend/internal/executor/internal_proxy_handler.go#L24:6)
- Executor Queue: [handleDequeue](https://sourcegraph.com/github.com/sourcegraph/sourcegraph@fc5a827bfaf5bcd5c9fb2fb05148ef687dd56a2e/-/blob/enterprise/internal/apiworker/apiserver/routes.go#L37:19), [setLogContents](https://sourcegraph.com/github.com/sourcegraph/sourcegraph@fc5a827bfaf5bcd5c9fb2fb05148ef687dd56a2e/-/blob/enterprise/internal/apiworker/apiserver/routes.go#L51:19), [handleMarkComplete](https://sourcegraph.com/github.com/sourcegraph/sourcegraph@fc5a827bfaf5bcd5c9fb2fb05148ef687dd56a2e/-/blob/enterprise/internal/apiworker/apiserver/routes.go#L61:19), [handleHeartbeat](https://sourcegraph.com/github.com/sourcegraph/sourcegraph@fc5a827bfaf5bcd5c9fb2fb05148ef687dd56a2e/-/blob/enterprise/internal/apiworker/apiserver/routes.go#L89:19), [transformRecord](https://sourcegraph.com/github.com/sourcegraph/sourcegraph@fc5a827bfaf5bcd5c9fb2fb05148ef687dd56a2e/-/blob/enterprise/cmd/executor-queue/internal/queues/codeintel/transform.go#L14:6)
