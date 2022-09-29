# Updating a pure-Docker Sourcegraph cluster

This document describes the exact changes needed to update a [pure-Docker Sourcegraph cluster](https://github.com/sourcegraph/deploy-sourcegraph-docker).
Each section comprehensively describes the changes needed in Docker images, environment variables, and added/removed services. **Always refer to this page before upgrading Sourcegraph,** as it comprehensively describes the steps needed to upgrade, and any manual migration steps you must perform.

1. Read our [update policy](index.md#update-policy) to learn about Sourcegraph updates.
2. Find the relevant entry for your update in the update notes on this page.

<!-- GENERATE UPGRADE GUIDE ON RELEASE (release tooling uses this to add entries) -->

## Unreleased

<!-- Add changes changes to this section before release. -->

TODO - replace me

## 3.43 -> 4.0.1

Follow the [steps](#upgrade-procedure) outlined at the top of this page to upgrade.

## 3.42 -> 3.43.2

To upgrade, please perform the changes in the following diff: https://github.com/sourcegraph/deploy-sourcegraph-docker/commit/a189e495813bc33d544b302eb98c197d70eacc87

## 3.41 -> 3.42.2

To upgrade, please perform the changes in the following diff: https://github.com/sourcegraph/deploy-sourcegraph-docker/commit/a189e495813bc33d544b302eb98c197d70eacc87

## 3.40.2 -> 3.41.0

To upgrade, please perform the changes in the following diff: https://github.com/sourcegraph/deploy-sourcegraph-docker/commit/8bfd70892c1bf56c5a88db0329826800c7a1097b

## 3.40.1 -> 3.40.2

To upgrade, please perform the changes in the following diff:
[https://github.com/sourcegraph/deploy-sourcegraph-docker/commit/312b9f8308148cf9403cc7868eee7b5c9611b121](https://github.com/sourcegraph/deploy-sourcegraph-docker/commit/312b9f8308148cf9403cc7868eee7b5c9611b121)

## 3.39 -> 3.40.1

- A fix that corrects the default behavior of the `migrator` service is included in this release. An attempt to standardize CLI packages in v3.39.0 unintentionally
broke the default behavior. In order to guard against this, all command line arguments are explicitly set in the deployment manifest.
- **CAUTION** Added the ability to customize postgres server configuration by mounting external configuration files. If you have customized the config in any way, you should copy your changes to the added `postgresql.conf` files [sourcegraph/deploy-sourcegraph-docker#806](https://github.com/sourcegraph/deploy-sourcegraph-docker/pull/806)

To upgrade, please perform the changes in the following diff:
[https://github.com/sourcegraph/deploy-sourcegraph-docker/commit/2c94a1fb5fa396759d4800a717af6658548943f7](https://github.com/sourcegraph/deploy-sourcegraph-docker/commit/2c94a1fb5fa396759d4800a717af6658548943f7)

## 3.39 -> 3.39.1

To upgrade, please perform the changes in the following diff:
[https://github.com/sourcegraph/deploy-sourcegraph-docker/commit/c2450311e385f077679d7666c09fd5a2aa7a6b6e](https://github.com/sourcegraph/deploy-sourcegraph-docker/commit/c2450311e385f077679d7666c09fd5a2aa7a6b6e)

## 3.38 -> 3.39

In this release we need to remove timescaledb from `shared_preload_libraries` configuration in `codeinsights-db`'s `postgresql.conf`. This step will be [performed automatically](https://github.com/sourcegraph/deploy-sourcegraph-docker/commit/b37367c738d28ef7e27c8b1f833eb9355bd9e8b1#diff-916162e35509bb582798c4306953fec9f43779d82420cb4435576e2873869f78R17). It can be performed manually instead of run as part of the deploy script.

To upgrade, please perform the changes in the following diff: https://github.com/sourcegraph/deploy-sourcegraph-docker/commit/b37367c738d28ef7e27c8b1f833eb9355bd9e8b1

## 3.38.0 -> 3.38.1

To upgrade, please perform the changes in the following diff:
[https://github.com/sourcegraph/deploy-sourcegraph-docker/commit/19735936834aab31134888c179bf07387f09a647](https://github.com/sourcegraph/deploy-sourcegraph-docker/commit/19735936834aab31134888c179bf07387f09a647)

## 3.37 -> 3.38

This release adds the requirement that the environment variables `SRC_GIT_SERVERS`, `SEARCHER_URL`, `SYMBOLS_URL`, and `INDEXED_SEARCH_SERVERS` are set for the worker process.

To upgrade, please perform the changes in the following diff:
[https://github.com/sourcegraph/deploy-sourcegraph-docker/commit/a66a74ce9a120a9da743eb44c6fea3a55f51842a](https://github.com/sourcegraph/deploy-sourcegraph-docker/commit/a66a74ce9a120a9da743eb44c6fea3a55f51842a)

## 3.36 -> 3.37

This release adds a new container that runs database migrations (`migrator`) independently of the frontend container. Confirm the environment variables on this new container match your database settings. [Read more about manual operation of the migrator](https://docs.sourcegraph.com/admin/how-to/manual_database_migrations)

To upgrade, please perform the changes in the following diff:
[https://github.com/sourcegraph/deploy-sourcegraph-docker/commit/9e369ec86cdef50b9e2a8350040d011cf2c7cd49](https://github.com/sourcegraph/deploy-sourcegraph-docker/commit/9e369ec86cdef50b9e2a8350040d011cf2c7cd49)

## 3.36.2 -> 3.36.3

To upgrade, please perform the changes in the following diff:
[https://github.com/sourcegraph/deploy-sourcegraph-docker/commit/3273d3c7da750ff15ba9d4f24d1e09e835bf11d9](https://github.com/sourcegraph/deploy-sourcegraph-docker/commit/3273d3c7da750ff15ba9d4f24d1e09e835bf11d9)

## 3.36.1 -> 3.36.2

To upgrade, please perform the changes in the following diff:
[https://github.com/sourcegraph/deploy-sourcegraph-docker/commit/45946fd69dd061cb39c85cfd06a037aeeaf74808](https://github.com/sourcegraph/deploy-sourcegraph-docker/commit/45946fd69dd061cb39c85cfd06a037aeeaf74808)

## 3.35 -> 3.36.1

To upgrade, please perform the changes in the following diff:
[https://github.com/sourcegraph/deploy-sourcegraph-docker/commit/25cdf1858de7fe3d0a3e3479a7e5620a02ac6a2c](https://github.com/sourcegraph/deploy-sourcegraph-docker/commit/25cdf1858de7fe3d0a3e3479a7e5620a02ac6a2c)

## 3.35.1 -> 3.35.2

To upgrade, please perform the changes in the following diff:
[https://github.com/sourcegraph/deploy-sourcegraph-docker/commit/abc948e60a489f559ebd5cc8f0affcd3c4371fa4](https://github.com/sourcegraph/deploy-sourcegraph-docker/commit/abc948e60a489f559ebd5cc8f0affcd3c4371fa4)

## 3.35.0 -> 3.35.1
**Due to issues related to Code Insights on the 3.35.0 release, users are advised to upgrade to 3.35.1 as soon as possible.**

There is a [known issue](../../code_insights/how-tos/Troubleshooting.md#oob-migration-has-made-progress-but-is-stuck-before-reaching-100) with the Code Insights out-of-band settings migration not reaching 100% complete when encountering deleted users or organizations.

To upgrade, please perform the changes in the following diff:
[https://github.com/sourcegraph/deploy-sourcegraph-docker/commit/ba0d94eb945fd3371ed888e4b7177828b33acd3d](https://github.com/sourcegraph/deploy-sourcegraph-docker/commit/ba0d94eb945fd3371ed888e4b7177828b33acd3d)

## 3.34 -> 3.35.1

**Due to issues related to Code Insights on the 3.35.0 release, users are advised to upgrade directly to 3.35.1.**

The `query-runner` service has been decomissioned in the 3.35.0 release. You can safely remove the `query-runner` service from your installation.

To upgrade, please perform the changes in the following diff:
[https://github.com/sourcegraph/deploy-sourcegraph-docker/commit/33b076a123c23930cc3339167bdd5502bebc5a3c](https://github.com/sourcegraph/deploy-sourcegraph-docker/commit/33b076a123c23930cc3339167bdd5502bebc5a3c)

## 3.34.x -> 3.34.2

To upgrade, please perform the changes in the following diff:
[https://github.com/sourcegraph/deploy-sourcegraph-docker/commit/e88d0f4615fc231576d37819b816576ac75b28d7](https://github.com/sourcegraph/deploy-sourcegraph-docker/commit/e88d0f4615fc231576d37819b816576ac75b28d7)

## 3.33 -> 3.34.2

To upgrade, please perform the changes in the following diff:
[https://github.com/sourcegraph/deploy-sourcegraph-docker/commit/d615dd5f63ec0984d60076aecf0bc598d9ffc1a8](https://github.com/sourcegraph/deploy-sourcegraph-docker/commit/d615dd5f63ec0984d60076aecf0bc598d9ffc1a8)

__Please upgrade directly to 3.34.2.__

A bug in our 3.34 and 3.34.1 release causes some repositories from older Sourcegraph versions to not appear in search results due to a database change.

## 3.32 -> 3.33

To upgrade, please perform the changes in the following diff:
[https://github.com/sourcegraph/deploy-sourcegraph-docker/commit/f6dc5c4a859b09faaea44a34e3ba8e85c92fcf58](https://github.com/sourcegraph/deploy-sourcegraph-docker/commit/f6dc5c4a859b09faaea44a34e3ba8e85c92fcf58)

## 3.31 -> 3.32

To upgrade, please perform the changes in the following diff:
[https://github.com/sourcegraph/deploy-sourcegraph-docker/commit/2c4c283ae9f89fa48232f0b99ed1982008034fee](https://github.com/sourcegraph/deploy-sourcegraph-docker/commit/2c4c283ae9f89fa48232f0b99ed1982008034fee)

## 3.30.3 -> 3.31

The **built-in** main Postgres (`pgsql`) and codeintel (`codeintel-db`) databases have switched to an alpine-based Docker image. Upon upgrading, Sourcegraph will need to re-index the entire database.

If you have already upgraded to 3.30.3, which uses the new alpine-based Docker images, all users that use our bundled (built-in) database instances should have already performed [the necessary re-indexing](../migration/3_31.md).

> NOTE: The above does not apply to users that use external databases (e.x: Amazon RDS, Google Cloud SQL, etc.).

## 3.30.x -> 3.31

The **built-in** main Postgres (`pgsql`) and codeintel (`codeintel-db`) databases have switched to an alpine-based Docker image. Upon upgrading, Sourcegraph will need to re-index the entire database.

All users that use our bundled (built-in) database instances **must** read through the [3.31 upgrade guide](../migration/3_31.md) _before_ upgrading.

> NOTE: The above does not apply to users that use external databases (e.x: Amazon RDS, Google Cloud SQL, etc.).

## 3.29 -> 3.30.3

> WARNING: **Users on 3.29.x are advised to upgrade directly to 3.30.3**. If you have already upgraded to 3.30.0, 3.30.1, or 3.30.2 please follow [this migration guide](../migration/3_30.md).

To upgrade, please perform the changes in the following diff:
[https://github.com/sourcegraph/deploy-sourcegraph-docker/commit/63802ca5966754162c2b3e077e64e60687138874](https://github.com/sourcegraph/deploy-sourcegraph-docker/commit/63802ca5966754162c2b3e077e64e60687138874)

## 3.28 -> 3.29

To upgrade, please perform the changes in the following diff:
[https://github.com/sourcegraph/deploy-sourcegraph-docker/commit/940100429fdd59f930436d47e226f5a7116bf6d9](https://github.com/sourcegraph/deploy-sourcegraph-docker/commit/940100429fdd59f930436d47e226f5a7116bf6d9)

This upgrade adds a new `worker` service that runs a number of background jobs that were previously run in the `frontend` service. See [notes on deploying workers](../workers.md#deploying-workers) for additional details. Good initial values for CPU and memory resources allocated to this new service should match the `frontend` service.

## 3.27 -> 3.28

To upgrade, please perform the changes in the following diff:
[https://github.com/sourcegraph/deploy-sourcegraph-docker/commit/fa9bd6b4749697e09a4a74537e180e8331d84a5b](https://github.com/sourcegraph/deploy-sourcegraph-docker/commit/fa9bd6b4749697e09a4a74537e180e8331d84a5b)

## 3.26 -> 3.27

To upgrade, please perform the changes in the following diff:

[https://github.com/sourcegraph/deploy-sourcegraph-docker/commit/1d01302a86d219a0f00f6dcbd27d4a511581ff27](https://github.com/sourcegraph/deploy-sourcegraph-docker/commit/1d01302a86d219a0f00f6dcbd27d4a511581ff27)

> Warning: ⚠️ Sourcegraph 3.27 now requires **Postgres 12+**.

If you are using an external database, [upgrade your database](https://docs.sourcegraph.com/admin/postgres#upgrading-external-postgresql-instances) to Postgres 12.5 or above prior to upgrading Sourcegraph. No action is required if you are using the supplied supplied database images.

## 3.26.0 -> 3.26.2

To upgrade, please perform the changes in the following diff:

https://github.com/sourcegraph/deploy-sourcegraph-docker/commit/de80a4af2ef2eeb23526e3ea560f7f72e1a71a5f

> NOTE: ⚠️ From **3.27** onwards we will only support PostgreSQL versions **starting from 12**.

## 3.25 -> 3.26.0

To upgrade, please perform the changes in the following diff:

https://github.com/sourcegraph/deploy-sourcegraph-docker/commit/68ffad230fc6f99477cedd303a55b91a8c4d9acb

> NOTE: ⚠️ From **3.27** onwards we will only support PostgreSQL versions **starting from 12**.

## 3.24 -> 3.25

Confirm that `codeinsights-db-disk` has the correct file permissions:

```
sudo chown -R 999:999 ~/sourcegraph-docker/codeinsights-db-disk/
```

- **If your are connecting to an external Postgres database using SSL/TLS:** Go `1.15` introduced changes to SSL/TLS connection validation which requires certificates to include a `SAN`. This field was not included in older certificates and clients relied on the `CN` field. You might see an error like `x509: certificate relies on legacy Common Name field`. We recommend that customers using Sourcegraph with an external database and connecting to it using SSL/TLS check whether the certificate is up to date.
  - AWS RDS customers please reference [AWS' documentation on updating the SSL/TLS certificate](https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/UsingWithRDS.SSL-certificate-rotation.html) for steps to rotate your certificate.

To upgrade, please perform the changes in the following diff:

https://github.com/sourcegraph/deploy-sourcegraph-docker/commit/ee2cbb59c80a382fb6cc649d4547b044d9a8b28d

## 3.23.0 -> 3.24.0

To upgrade, please perform the changes in the following diff:

https://github.com/sourcegraph/deploy-sourcegraph-docker/commit/4acc4c7ed5d49ce41b1f68d654a3f4e2f35bd622

## 3.22.0 -> 3.23.0

To upgrade, please perform the changes in the following diff:

https://github.com/sourcegraph/deploy-sourcegraph-docker/commit/10de1a4e34ab2c716bd63e52a68a6af896bd81b7

## 3.21.2 -> 3.22.0

To upgrade, please perform the changes in the following diff:

https://github.com/sourcegraph/deploy-sourcegraph-docker/commit/223c11dacffafb985c2d29b6c6a9b84bcc8255be

This upgrade removes the `code intel bundle manager`. This service has been deprecated and all references to it have been removed.

This upgrade also adds a MinIO container that doesn't require any custom configuration. You can find more detailed documentation in https://docs.sourcegraph.com/admin/external_services/object_storage.

## 3.20.1 -> 3.21.2

To upgrade, please perform the changes in the following diff:

https://github.com/sourcegraph/deploy-sourcegraph-docker/commit/4629ddfcdfd070b41801106199817ae903ead510

### Note new services

This upgrade includes a new code-intel DB (`deploy-codeintel-db.sh`) and a new service `minio` (`deploy-minio.sh`)
to store precise code intel indexes.
There is a new environment variable for frontend and frontend-internal called `CODEINTEL_PGHOST`.

(both of these changes are described exactly in the diff above)

## 3.19.1 -> 3.20.1

To upgrade, please perform the changes in the following diff:

https://github.com/sourcegraph/deploy-sourcegraph-docker/commit/7c57315a1ce05981d436860d79ca01553931e274

### Confirm file permissions

Confirm that `lsif-server-disk` has the correct file permissions:

```
sudo chown -R 100:101 ~/sourcegraph-docker/lsif-server-disk/ ~/sourcegraph-docker/lsif-server-disk/
```

## 3.18.0 -> 3.19.1

To upgrade, please perform the changes in the following diff:

https://github.com/sourcegraph/deploy-sourcegraph-docker/commit/a1648452c6f7c16637b0e069776df12604c27f73

### Confirm file permissions

Confirm that `lsif-server-disk` has the correct file permissions:

```
sudo chown -R 100:101 ~/sourcegraph-docker/lsif-server-disk/ ~/sourcegraph-docker/lsif-server-disk/
```

## 3.17.2 -> 3.18.0 changes

To upgrade, please perform the changes in the following diff:

https://github.com/sourcegraph/deploy-sourcegraph-docker/commit/7e6b23cdfead3be639048c5fa7fffe07441610f2

Note: `deploy-grafana.sh` and `deploy-prometheus.sh` had environment variables changed, otherwise only image tags have changed.

## v3.16.0 -> v3.17.2 changes

To upgrade, please perform the changes in the following diff:

https://github.com/sourcegraph/deploy-sourcegraph-docker/commit/2895236661de3ff633ee56fe0b87e9a0f530cc60

## v3.15.1 → v3.16.0 changes

This release involves two steps:

1. Change `3.15.1` image tags to `3.16.0`
2. Update `prometheus/prometheus_targets.yml` [as shown here](https://github.com/sourcegraph/deploy-sourcegraph-docker/compare/customer-replica-v3.15.1...customer-replica-v3.16.0#diff-1d4c5a677b37d150c65ea8356cad978a)

Exact diff of changes to make: https://github.com/sourcegraph/deploy-sourcegraph-docker/compare/customer-replica-v3.15.1...customer-replica-v3.16.0

## v3.14.2 → v3.15.1 changes

This release:

- Removes the 4-container Jaeger deployment, which had performance issues, in favor of a single-container one that works well.
- Removes the old `lsif-server` service deployment, replacing it with 3 `precise-code-intel` services.
- Changes the versions of all Sourcegraph images to be consistent (just `3.15.1` instead of some being inconsistent), which in some cases required changing the names of the images (but the containers / services / shell scripts remain the same for now).

### Update environment variables

- On `frontend` and `frontend-internal` containers: Remove the `LSIF_SERVER_URL` environment variable.
- On `frontend` and `frontend-internal` containers: Set `PRECISE_CODE_INTEL_API_SERVER_URL=http://precise-code-intel-api-server:3186`
- On all containers: Change `JAEGER_AGENT_HOST=jaeger-agent` to `JAEGER_AGENT_HOST=jaeger`

### Remove all old container deployments

- `jaeger-agent` container (`deploy-jaeger-agent.sh`).
- `jaeger-cassandra` container (`deploy-jaeger-cassandra.sh`).
- `jaeger-collector` container (`deploy-jaeger-collector.sh`).
- `jaeger-query` container (`deploy-jaeger-query.sh`).
- `lsif-server` container (`deploy-lsif-server.sh`).

### Add new container deployments

- Add a single `jaeger` container [following this spec](https://github.com/sourcegraph/deploy-sourcegraph-docker/blob/v3.15.1/deploy-jaeger.sh#L1).
- Add a single `precise-code-intel-api-server` container [following this spec](https://github.com/sourcegraph/deploy-sourcegraph-docker/blob/v3.15.1/deploy-precise-code-intel-api-server.sh)
- Add a single `precise-code-intel-bundle-manager` container [following this spec](https://github.com/sourcegraph/deploy-sourcegraph-docker/blob/v3.15.1/deploy-precise-code-intel-bundle-manager.sh)
- Add a single `precise-code-intel-worker` container [following this spec](https://github.com/sourcegraph/deploy-sourcegraph-docker/blob/v3.15.1/deploy-precise-code-intel-worker.sh)

### Update prometheus_targets.yml

```diff
-    - lsif-server:3186
-    - lsif-server:3187
+    - precise-code-intel-api-server:3186
+    - precise-code-intel-bundle-manager:3187
```

### Update image tags to 3.15.1

Please change *all sourcegraph/<service>* image tags to `3.15.1. This includes all images you previously had as `:3.14.2` AND all `sourcegraph/<service>` images:

```
index.docker.io/sourcegraph/grafana:3.15.1
index.docker.io/sourcegraph/prometheus:3.15.1
index.docker.io/sourcegraph/redis-cache:3.15.1
index.docker.io/sourcegraph/redis-store:3.15.1
index.docker.io/sourcegraph/pgsql:3.15.1
```

The following _images_ have been renamed AND use Sourcegraph versions now (their container names and shell script names remain the same for now):

```diff
- index.docker.io/sourcegraph/syntect_server:c0297a1@sha256:333abb45cfaae9c9d37e576c3853843b00eca33a40a7c71f6b93211ed96528df
+ index.docker.io/sourcegraph/syntax-highlighter:3.15.1

- index.docker.io/sourcegraph/zoekt-indexserver:0.0.20200318141948-0b140b7@sha256:b022fd7e4884a71786acae32e0ec8baf785c18350ebf5d574d52335a346364f9
+ index.docker.io/sourcegraph/search-indexer:3.15.1

- index.docker.io/sourcegraph/zoekt-webserver:0.0.20200318141342-0b140b7@sha256:0d0fbce55b51ec7bdd37927539f50459cd0f207b7cf219ca5122d07792012fb1
+ index.docker.io/sourcegraph/indexed-searcher:3.15.1
```
