# Updating a pure-Docker Sourcegraph cluster

This document describes the exact changes needed to update a [pure-Docker Sourcegraph cluster](https://github.com/sourcegraph/deploy-sourcegraph-docker).

Each section comprehensively describes the changes needed in Docker images, environment variables, and added/removed services.

# v3.12.1 → v3.12.2 changes

### Update image tags

Change 3.12.1 → v3.12.2:

| Container         | New image                                       |
|-------------------|-------------------------------------------------|
| frontend          | index.docker.io/sourcegraph/frontend:3.12.2     |
| frontend-internal | index.docker.io/sourcegraph/frontend:3.12.2     |
| github-proxy      | index.docker.io/sourcegraph/github-proxy:3.12.2 |
| gitserver         | index.docker.io/sourcegraph/gitserver:3.12.2    |
| lsif-server       | index.docker.io/sourcegraph/lsif-server:3.12.2  |
| query-runner      | index.docker.io/sourcegraph/query-runner:3.12.2 |
| replacer          | index.docker.io/sourcegraph/replacer:3.12.2     |
| repo-updater      | index.docker.io/sourcegraph/repo-updater:3.12.2 |
| searcher          | index.docker.io/sourcegraph/searcher:3.12.2     |
| symbols           | index.docker.io/sourcegraph/symbols:3.12.2      |

Also change the follow which are not versioned alongside Sourcegraph currently:

| Container         | New image                                       |
|-------------------|-------------------------------------------------|
| prometheus        | index.docker.io/sourcegraph/prometheus:10.0.7@sha256:22d54f27c7df8733a06c7ae8c2e851b61b1ed42f1f5621d493ef58ebd8d815e0 |

# v3.10.4 → v3.12.1 changes

### Management console removal

- Please remove the `management-console` container entirely from your deployment ([details](https://docs.sourcegraph.com/admin/migration/3_11#migration-notes-for-sourcegraph-3-11)).
- Please remove the `- management-console:6060` entry from your `prometheus/prometheus_targets.yml` file.

### If you use `SITE_CONFIG_FILE` or `CRITICAL_CONFIG_FILE`

If you are making use of `SITE_CONFIG_FILE` or `CRITICAL_CONFIG_FILE` environment variables please:

1. Simply copy all properties from `CRITICAL_CONFIG_FILE` and paste them into your `SITE_CONFIG_FILE`.
2. Delete and remove the `CRITICAL_CONFIG_FILE`, as it will no longer be used (the two are now just one).

### Update image tags

Change 3.10.4 → v3.12.1:

| Container         | New image                                       |
|-------------------|-------------------------------------------------|
| frontend          | index.docker.io/sourcegraph/frontend:3.12.1     |
| frontend-internal | index.docker.io/sourcegraph/frontend:3.12.1     |
| github-proxy      | index.docker.io/sourcegraph/github-proxy:3.12.1 |
| gitserver         | index.docker.io/sourcegraph/gitserver:3.12.1    |
| lsif-server       | index.docker.io/sourcegraph/lsif-server:3.12.1  |
| query-runner      | index.docker.io/sourcegraph/query-runner:3.12.1 |
| replacer          | index.docker.io/sourcegraph/replacer:3.12.1     |
| repo-updater      | index.docker.io/sourcegraph/repo-updater:3.12.1 |
| searcher          | index.docker.io/sourcegraph/searcher:3.12.1     |
| symbols           | index.docker.io/sourcegraph/symbols:3.12.1      |

Also change the follow which are not versioned alongside Sourcegraph currently:

| Container         | New image                                       |
|-------------------|-------------------------------------------------|
| grafana           | index.docker.io/sourcegraph/grafana:10.0.9@sha256:0132e5602030145803753468497a2d17640164b9c34df4ce2532dd93e4b1f6fc |
| prometheus        | index.docker.io/sourcegraph/prometheus:10.0.6@sha256:f681ceb9400f0d546601cbf827ac9c3db16acd37e810da0860cf05d4f42305d1 |
| syntect-server    | index.docker.io/sourcegraph/syntect_server:2b5a3fb@sha256:ef5529cafdc68d5a21edea472ee8ad966878b173044aa5c3db93bc3d84765b1f |
| zoekt-indexserver | index.docker.io/sourcegraph/zoekt-indexserver:0.0.20191204145522-b470e5f@sha256:84e9de8be269277d6e4711a61d0c9675a44d409a4bf7f7dd1b90a22175095fa7 |
| zoekt-webserver   | index.docker.io/sourcegraph/zoekt-webserver:0.0.20191204145231-b470e5f@sha256:fc3bfa69fc60b7a049a6646b71e45896cfae8adf3484602d140965c3781463a0 |
| pgsql (no change if using external Postgres) | index.docker.io/sourcegraph/postgres-11.4:19-11-14_b084311b@sha256:072481559d559cfd9a53ad77c3688b5cf583117457fd452ae238a20405923297 |
| redis-cache (no change if using external Redis) | index.docker.io/sourcegraph/redis-cache:19-04-16_6891de82@sha256:4cbfac8af0abb673899250d4fd859cc477d6426de519e9deb71e454e18322499 |
| redis-store (no change if using external Redis) | index.docker.io/sourcegraph/redis-store:19-04-16_6891de821@sha256:56426d601ce1f6d63088fea1cefa61f69a2e809c7d90fc1d157cca63cf81b277 |
