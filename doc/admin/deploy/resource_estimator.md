<style>
.estimator label {
    display: flex;
}

.estimator .radioInput label {
    display: inline-flex;
    align-items: center;
    margin-left: .5rem;
}

.estimator .radioInput label span {
    margin-left: .25rem;
    margin-right: .25rem;
}

.estimator input[type=range] {
    width: 15rem;
}

.estimator .post-label {
    font-size: 16px;
    margin-left: 0.5rem;
}

.estimator .copy-as-markdown {
    width: 100%;
    height: 8rem;
}

.estimator a[title]:hover:after {
  content: attr(title);
  background: red;
  position: relative;
  z-index: 1000;
  top: 16px;
  left: 0;
}

</style>

<script src="https://storage.googleapis.com/sourcegraph-resource-estimator/go_1_18_wasm_exec.js"></script>
<script src="https://storage.googleapis.com/sourcegraph-resource-estimator/launch_script.js?v2" version="0249096"></script>

# Sourcegraph resource estimator

Updating the form below will recalculate an estimate for the resources you can use to configure your Sourcegraph deployment.

The output is estimated based on existing data we collected from current running deployments.

<form id="root"></form>

## Additional information

#### How to apply these changes to your deployment?

In a docker-compose deployment, edit your docker-compose.yml file and set cpus and mem_limit to the limits shown above.

In Kubernetes deployments, edit the respective yaml file and update, limits, requests, and replicas according to the above.

#### What is the default deployment size?

Our default deployment should support up to ~1000 users and about 1500 repositories with one monorepo that is less than 5GB.

#### What is engagement rate?

Engagement rate refers to the percentage of users who use Sourcegraph regularly. It is generally used for existing deployments to estimate resources.


#### What is the recommended deployment type?

We recommend Kubernetes for any deployments requiring > 1 service replica, but docker-compose does support service replicas and can scale up with multiple replicas as long as you can provision a sufficiently large single machine.

#### If you plan to enforce repository permissions on Sourcegraph

Repository permissions on Sourcegraph can have a noticeable impact on search performance if you have a large number of users and/or repositories on your code host.

We suggest setting your authorization ttl values as high as you are comfortable setting it in order to reduce the chance of this (e.g. to 72h) in the repository permission configuration.

#### How does Sourcegraph scale?

[Click here to learn more about how Sourcegraph scales.](https://docs.sourcegraph.com/admin/install/kubernetes/scale)


## Service overview

| Service | Description |
|---------|------|
| cadvisor | A cAdvisor instance that exports container monitoring metrics scraped by Prometheus and visualized in Grafana |
| codeinsights-db | A PostgreSQL instance for storing code insights data |
| codeintel-db | A PostgreSQL instance for storing large-volume precise code intelligence data |
| frontend | Serves the web application, extensions, and graphQL services. Almost every service has a link back to the frontend, from which it gathers configuration updates. |
| github-proxy | Proxies all requests to github.com to keep track of rate limits and prevent triggering abuse mechanisms	|
| gitserver | Mirrors repositories from their code host. All other Sourcegraph services talk to gitserver when they need data from git  |
| grafana | A Grafana instance that displays data from Prometheus and Jaeger. It is shipped with customized dashboards for Sourcegraph services |
| jaeger | A Jaeger instance for end-to-end distributed tracing |
| minio | A MinIO instance that serves as a local S3-compatible object storage to hold user uploads for code-intel before they can be processed |
| pgsql | The main database. It is a PostgreSQL instance where things like repo lists, user data, site config files are stored (anything not related to code-intel and code-insights) |
| precise-code-intel | Converts LSIF upload file into Postgres data. The entire index must be read into memory to be correlated |
| prometheus | Collecting high-level, and low-cardinality, metrics across services. |
| redis-cache | A Redis instance for storing cache data. |
| redis-store  | A Redis instance for storing short-term information such as user sessions. |
| repo-updater | Repo-updater tracks the state of repositories, and is responsible for automatically scheduling updates using gitserver. Other apps which desire updates or fetches should be telling repo-updater, rather than using gitserver directly, so repo-updater can take their changes into account. |
| searcher | Provides on-demand unindexed search for repositories. It fetches archives from gitserver and searches them with regexp	|
| symbols | Indexes symbols in repositories using Ctags. By default, the symbols service saves SQLite DBs as files on disk, and copies an old one to a new file when a user visits a new commit. If Rockskip is enabled, the symbols are stored in the codeintel-db instead while the cache is stored on disk |
| syntect-server | An HTTP server that exposes the Rust Syntect syntax highlighting library for use by other services |
| worker |  Runs a collection of background jobs (for both Code-Intel and Code-Insight) periodically or in response to an external event. It is currently janitorial and commit based. |
| zoekt-indexserver | Indexes all enabled repositories on Sourcegraph, as well as keeping the indexes up to date |
| zoekt-webserver | Runs searches from in-memory indexes, but persists these indexes to disk to avoid re-indexing everything on startup |
