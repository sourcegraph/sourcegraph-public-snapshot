# Updating a pure-Docker Sourcegraph cluster

This document describes the exact changes needed to update a [pure-Docker Sourcegraph cluster](https://github.com/sourcegraph/deploy-sourcegraph-docker).

Each section comprehensively describes the changes needed in Docker images, environment variables, and added/removed services.

## v3.12.5 -> v3.13.2 changes

### Confirm file permissions

Confirm that `redis-store-disk` has the correct file permissions:

```
sudo chown -R 999:1000 ~/sourcegraph-docker/redis-store-disk/ ~/sourcegraph-docker/redis-cache-disk/
```

### Update image tags

Change 3.12.5 → 3.13.2 for the following containers:

| Container         | New image                                       |
|-------------------|-------------------------------------------------|
| frontend          | index.docker.io/sourcegraph/frontend:3.13.2     |
| frontend-internal | index.docker.io/sourcegraph/frontend:3.13.2     |
| github-proxy      | index.docker.io/sourcegraph/github-proxy:3.13.2 |
| gitserver         | index.docker.io/sourcegraph/gitserver:3.13.2    |
| lsif-server       | index.docker.io/sourcegraph/lsif-server:3.13.2  |
| query-runner      | index.docker.io/sourcegraph/query-runner:3.13.2 |
| replacer          | index.docker.io/sourcegraph/replacer:3.13.2     |
| repo-updater      | index.docker.io/sourcegraph/repo-updater:3.13.2 |
| searcher          | index.docker.io/sourcegraph/searcher:3.13.2     |
| symbols           | index.docker.io/sourcegraph/symbols:3.13.2      |

Also change the follow which are not versioned alongside Sourcegraph currently:

| Container         | New image                                       |
|-------------------|-------------------------------------------------|
| grafana           | index.docker.io/sourcegraph/grafana:10.0.10@sha256:a6f9816346c3e38478f4b855eeee199fc91a4f69311f5dd57760bf74c3234715 |
| prometheus        | index.docker.io/sourcegraph/prometheus:10.0.8@sha256:75efaada5a335cda9895f337d8f31b6abb7a082ef3092b7bb24bf31fb78eafe6 |
| redis-cache       | index.docker.io/sourcegraph/redis-cache:20-02-03_da9d71ca@sha256:7820219195ab3e8fdae5875cd690fed1b2a01fd1063bd94210c0e9d529c38e56 |
| redis-store       | index.docker.io/sourcegraph/redis-store:20-01-30_c903717e@sha256:e8467a8279832207559bdfbc4a89b68916ecd5b44ab5cf7620c995461c005168 |
| syntect-server    | index.docker.io/sourcegraph/syntect_server:c0297a1@sha256:333abb45cfaae9c9d37e576c3853843b00eca33a40a7c71f6b93211ed96528df |
| zoekt-indexserver | index.docker.io/sourcegraph/zoekt-indexserver:0.0.20200302121716-13dbd22@sha256:91643d83223bb72f4aa2b5031ceb774c8e604a227c58ed54375bd341f25e2ef3 |
| zoekt-webserver   | index.docker.io/sourcegraph/zoekt-webserver:0.0.20200302121635-13dbd22@sha256:0183bd676fe1ba774edcca29f042d8d3594e833e08b6d603af98f74c575eaf69 |

## v3.12.2 → v3.12.5 changes

### Confirm file permissions

Confirm that the `replacer-disk` has the correct file permissions:

```
sudo chown 100:101 ~/sourcegraph-docker/replacer-disk
```

### Update image tags

Change 3.12.2 → 3.12.5 for the following containers:

| Container         | New image                                       |
|-------------------|-------------------------------------------------|
| frontend          | index.docker.io/sourcegraph/frontend:3.12.5     |
| frontend-internal | index.docker.io/sourcegraph/frontend:3.12.5     |
| github-proxy      | index.docker.io/sourcegraph/github-proxy:3.12.5 |
| gitserver         | index.docker.io/sourcegraph/gitserver:3.12.5    |
| lsif-server       | index.docker.io/sourcegraph/lsif-server:3.12.5  |
| query-runner      | index.docker.io/sourcegraph/query-runner:3.12.5 |
| replacer          | index.docker.io/sourcegraph/replacer:3.12.5     |
| repo-updater      | index.docker.io/sourcegraph/repo-updater:3.12.5 |
| searcher          | index.docker.io/sourcegraph/searcher:3.12.5     |
| symbols           | index.docker.io/sourcegraph/symbols:3.12.5      |

Also change the follow which are not versioned alongside Sourcegraph currently:

| Container         | New image                                       |
|-------------------|-------------------------------------------------|
| zoekt-indexserver | index.docker.io/sourcegraph/zoekt-indexserver:0.0.20200124185115-83b89a5@sha256:efd1fb37fc62bfab963f12e95f69778b0e2e6a253caed5be9025840072ea85b5 |
| zoekt-webserver   | index.docker.io/sourcegraph/zoekt-webserver:0.0.20200124185328-83b89a5@sha256:cde27ee7db0fe6c293a8c9df47b529fb01b5a898e6cbeea4c18d80fe218563db |

## v3.12.1 → v3.12.2 changes

### Update image tags

Change 3.12.1 → v3.12.2 for the following containers:

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

## v3.10.4 → v3.12.1 changes

### Management console removal

- Please remove the `management-console` container entirely from your deployment ([details](https://docs.sourcegraph.com/admin/migration/3_11#migration-notes-for-sourcegraph-3-11)).
- Please remove the `- management-console:6060` entry from your `prometheus/prometheus_targets.yml` file.

### If you use `SITE_CONFIG_FILE` or `CRITICAL_CONFIG_FILE`

If you are making use of `SITE_CONFIG_FILE` or `CRITICAL_CONFIG_FILE` environment variables please:

1. Simply copy all properties from `CRITICAL_CONFIG_FILE` and paste them into your `SITE_CONFIG_FILE`.
2. Delete and remove the `CRITICAL_CONFIG_FILE`, as it will no longer be used (the two are now just one).

### Update image tags

Change 3.10.4 → v3.12.1 for the following containers:

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

## v3.10.4 → v3.12.5 changes

### Confirm file permissions

Confirm that the `replacer-disk` has the correct file permissions:

```
sudo chown 100:101 ~/sourcegraph-docker/replacer-disk
```

### Management console removal

- Please remove the `management-console` container entirely from your deployment ([details](https://docs.sourcegraph.com/admin/migration/3_11#migration-notes-for-sourcegraph-3-11)).
- Please remove the `- management-console:6060` entry from your `prometheus/prometheus_targets.yml` file.

### If you use `SITE_CONFIG_FILE` or `CRITICAL_CONFIG_FILE`

If you are making use of `SITE_CONFIG_FILE` or `CRITICAL_CONFIG_FILE` environment variables please:

1. Simply copy all properties from `CRITICAL_CONFIG_FILE` and paste them into your `SITE_CONFIG_FILE`.
2. Delete and remove the `CRITICAL_CONFIG_FILE`, as it will no longer be used (the two are now just one).

### Update image tags

Change 3.10.4 → v3.12.5 for the following containers:

| Container         | New image                                       |
|-------------------|-------------------------------------------------|
| frontend          | index.docker.io/sourcegraph/frontend:3.12.5     |
| frontend-internal | index.docker.io/sourcegraph/frontend:3.12.5     |
| github-proxy      | index.docker.io/sourcegraph/github-proxy:3.12.5 |
| gitserver         | index.docker.io/sourcegraph/gitserver:3.12.5    |
| lsif-server       | index.docker.io/sourcegraph/lsif-server:3.12.5  |
| query-runner      | index.docker.io/sourcegraph/query-runner:3.12.5 |
| replacer          | index.docker.io/sourcegraph/replacer:3.12.5     |
| repo-updater      | index.docker.io/sourcegraph/repo-updater:3.12.5 |
| searcher          | index.docker.io/sourcegraph/searcher:3.12.5     |
| symbols           | index.docker.io/sourcegraph/symbols:3.12.5      |

Also change the follow which are not versioned alongside Sourcegraph currently:

| Container         | New image                                       |
|-------------------|-------------------------------------------------|
| grafana           | index.docker.io/sourcegraph/grafana:10.0.9@sha256:0132e5602030145803753468497a2d17640164b9c34df4ce2532dd93e4b1f6fc |
| prometheus        | index.docker.io/sourcegraph/prometheus:10.0.7@sha256:22d54f27c7df8733a06c7ae8c2e851b61b1ed42f1f5621d493ef58ebd8d815e0 |
| syntect-server    | index.docker.io/sourcegraph/syntect_server:2b5a3fb@sha256:ef5529cafdc68d5a21edea472ee8ad966878b173044aa5c3db93bc3d84765b1f |
| zoekt-indexserver | index.docker.io/sourcegraph/zoekt-indexserver:0.0.20200124185115-83b89a5@sha256:efd1fb37fc62bfab963f12e95f69778b0e2e6a253caed5be9025840072ea85b5 |
| zoekt-webserver   | index.docker.io/sourcegraph/zoekt-webserver:0.0.20200124185328-83b89a5@sha256:cde27ee7db0fe6c293a8c9df47b529fb01b5a898e6cbeea4c18d80fe218563db |
| pgsql (no change if using external Postgres) | index.docker.io/sourcegraph/postgres-11.4:19-11-14_b084311b@sha256:072481559d559cfd9a53ad77c3688b5cf583117457fd452ae238a20405923297 |
| redis-cache (no change if using external Redis) | index.docker.io/sourcegraph/redis-cache:19-04-16_6891de82@sha256:4cbfac8af0abb673899250d4fd859cc477d6426de519e9deb71e454e18322499 |
| redis-store (no change if using external Redis) | index.docker.io/sourcegraph/redis-store:19-04-16_6891de821@sha256:56426d601ce1f6d63088fea1cefa61f69a2e809c7d90fc1d157cca63cf81b277 |
