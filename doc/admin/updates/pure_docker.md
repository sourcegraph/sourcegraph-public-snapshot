# Updating a pure-Docker Sourcegraph cluster

This document describes the exact changes needed to update a [pure-Docker Sourcegraph cluster](https://github.com/sourcegraph/deploy-sourcegraph-docker).

Each section comprehensively describes the changes needed in Docker images, environment variables, and added/removed services.

# v3.10.4 → v3.12.1 changes

### Management console removal

- Please remove the `management-console` container entirely from your deployment ([details](https://docs.sourcegraph.com/admin/migration/3_11#migration-notes-for-sourcegraph-3-11)).
- Please remove the `- management-console:6060` entry from your `prometheus/prometheus_targets.yml` file.

### If you use `SITE_CONFIG_FILE` or `CRITICAL_CONFIG_FILE`

If you are making use of `SITE_CONFIG_FILE` or `CRITICAL_CONFIG_FILE` environment variables please:

1. Simply copy all properties from `CRITICAL_CONFIG_FILE` and paste them into your `SITE_CONFIG_FILE`.
2. Delete and remove the `CRITICAL_CONFIG_FILE`, as it will no longer be used (the two are now just one).

### Updated image tags

| Container | New image |
|-----------|-----------|
| frontend | index.docker.io/sourcegraph/frontend:3.12.1@sha256:55b2e66ed8303cdaedd9998e6e8dfd93eb9b69c92e00cecad810c6eede2271fb |
| frontend-internal | index.docker.io/sourcegraph/frontend:3.12.1@sha256:55b2e66ed8303cdaedd9998e6e8dfd93eb9b69c92e00cecad810c6eede2271fb |
| github-proxy | index.docker.io/sourcegraph/github-proxy:3.12.1@sha256:f3692d0d229f1504610d7d849251dfe0396ea94edb7325389c9ff7bf3cbecc52 |
| gitserver | index.docker.io/sourcegraph/gitserver:3.12.1@sha256:693b975d131baf57941a3b551cf9fb7241ede7853d290fddb30e26212c62b0e3 |
| grafana | index.docker.io/sourcegraph/grafana:10.0.9@sha256:0132e5602030145803753468497a2d17640164b9c34df4ce2532dd93e4b1f6fc |
| lsif-server | index.docker.io/sourcegraph/lsif-server:3.12.1@sha256:9afda69da68ea606c1aaa9cd18c2d621b5b4088a48133167d1b969f7d13c4014 |
| prometheus | index.docker.io/sourcegraph/prometheus:10.0.6@sha256:f681ceb9400f0d546601cbf827ac9c3db16acd37e810da0860cf05d4f42305d1 |
| query-runner | index.docker.io/sourcegraph/query-runner:3.12.1@sha256:5df160cdd942db3544a126f8f7fe38760a9d5921b2089c24a3f9e5ff0a177851 |
| replacer | index.docker.io/sourcegraph/replacer:3.12.1@sha256:a0de687dc8f7590bcfa0f791e452371fb4d8c542749194bde2aa0830f640b3a2 |
| repo-updater | index.docker.io/sourcegraph/repo-updater:3.12.1@sha256:32f50b9af12a1d6554e7e99b5304685f0ad7111613d9865b25a9fd9c1ab3c5ed |
| searcher | index.docker.io/sourcegraph/searcher:3.12.1@sha256:82b8533f60bfc6df6b7ac86bf5a21ed1611da4ad2dde73c52b7559629c501d47 |
| symbols | index.docker.io/sourcegraph/symbols:3.12.1@sha256:12944d2e3b304921ae5c69a91f95c4966dcfad2f2a1f6abfc5d561814030be31 |
| syntect-server | index.docker.io/sourcegraph/syntect_server:2b5a3fb@sha256:ef5529cafdc68d5a21edea472ee8ad966878b173044aa5c3db93bc3d84765b1f |
| zoekt-indexserver | index.docker.io/sourcegraph/zoekt-indexserver:0.0.20191204145522-b470e5f@sha256:84e9de8be269277d6e4711a61d0c9675a44d409a4bf7f7dd1b90a22175095fa7 |
| zoekt-webserver | index.docker.io/sourcegraph/zoekt-webserver:0.0.20191204145231-b470e5f@sha256:fc3bfa69fc60b7a049a6646b71e45896cfae8adf3484602d140965c3781463a0 |
| pgsql (no change if using external Postgres) | index.docker.io/sourcegraph/postgres-11.4:19-11-14_b084311b@sha256:072481559d559cfd9a53ad77c3688b5cf583117457fd452ae238a20405923297 |
| redis-cache (no change if using external Redis) | index.docker.io/sourcegraph/redis-cache:19-04-16_6891de82@sha256:4cbfac8af0abb673899250d4fd859cc477d6426de519e9deb71e454e18322499 |
| redis-store (no change if using external Redis) | index.docker.io/sourcegraph/redis-store:19-04-16_6891de821@sha256:56426d601ce1f6d63088fea1cefa61f69a2e809c7d90fc1d157cca63cf81b277 |
