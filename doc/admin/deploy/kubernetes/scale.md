# Scaling Sourcegraph with Kubernetes

Sourcegraph can be configured to scale to very large codebases and large numbers of
users. If you notice latency for search or code navigation is higher than desired, changing these
parameters can yield a drastic improvement in performance.

See [Scaling Overview for Services](../scale.md) for more information about scaling.

> NOTE: For assistance when scaling and tuning Sourcegraph, [contact us](https://about.sourcegraph.com/contact/). We're happy to help!

---

## Cluster resource allocation guidelines

For production environments, we recommend allocate resources for your instance based on your [instance size](../instance-size.md). You can also refer to our [resource estimator](../resource_estimator.md) for more information regarding resources allocation for your Sourcegraph deployment.

---

## Improving performance with a large number of repositories

When you're using Sourcegraph with many repositories (100s-10,000s), the most important parameters to tune are:

- `sourcegraph-frontend` CPU/memory resource allocations
- `searcher` replica count
- `indexedSearch` replica count and CPU/memory resource allocations
- `gitserver` replica count
- `symbols` replica count and CPU/memory resource allocations
- `gitMaxConcurrentClones`, because `git clone` and `git fetch` operations are IO- and CPU-intensive
- `repoListUpdateInterval` (in minutes), because each interval triggers `git fetch` operations for all repositories

Notes:

- If your change requires `gitserver` pods to be restarted and they are scheduled on another node
  when they restart, they may go offline for 60-90 seconds (and temporarily show a `Multi-Attach`
  error). This delay is caused by Kubernetes detaching and reattaching the volume. Mitigation
  steps depend on your cloud provider; [contact us](https://about.sourcegraph.com/contact/) for
  advice.

- For context on what each service does, see [Sourcegraph Architecture Overview](https://docs.sourcegraph.com/dev/architecture) and [Scaling Overview for Services](../scale.md).

---

## Improving performance with large monorepos

When you're using Sourcegraph with a large monorepo (or several large monorepos), the most important parameters to tune
are:

- `sourcegraph-frontend` CPU/memory resource allocations
- `searcher` CPU/memory resource allocations (allocate enough memory to hold all non-binary files in your repositories)
- `indexedSearch` CPU/memory resource allocations (for the `zoekt-indexserver` pod, allocate enough memory to hold all non-binary files in your largest repository; for the `zoekt-webserver` pod, allocate enough memory to hold ~2.7x the size of all non-binary files in your repositories)
- `symbols` CPU/memory resource allocations
- `gitserver` CPU/memory resource allocations (allocate enough memory to hold your Git packed bare repositories)

---

## Configuring faster disk I/O for caches

Many parts of Sourcegraph's infrastructure benefit from using SSDs for caches. This is especially
important for search performance. By default, disk caches will use the
Kubernetes [hostPath](https://kubernetes.io/docs/concepts/storage/volumes/#hostpath) and will be the
same IO speed as the underlying node's disk. Even if the node's default disk is a SSD, however, it
is likely network-mounted rather than local.

See [configure/ssd/README.md](https://github.com/sourcegraph/deploy-sourcegraph/blob/master/configure/ssd/README.md) for instructions about configuring SSDs.
