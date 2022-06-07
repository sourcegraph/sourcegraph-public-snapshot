# Updating a Docker Compose Sourcegraph instance

This document describes the exact changes needed to update a [Docker Compose Sourcegraph instance](../deploy/docker-compose.md).
Each section comprehensively describes the steps needed to upgrade, and any manual migration steps you must perform.
**Always refer to this page before upgrading Sourcegraph**, as it comprehensively describes the steps needed to upgrade, and any manual migration steps you must perform.

1. Read our [update policy](index.md#update-policy) to learn about Sourcegraph updates.
2. Find the relevant entry for your update in the update notes on this page.
3. After checking the relevant update notes, refer to the [Sourcegraph with Docker Compose upgrade guide](../deploy/docker-compose/index.md#upgrade) to upgrade your instance.

<!-- GENERATE UPGRADE GUIDE ON RELEASE (release tooling uses this to add entries) -->

## Unreleased

- The Postgres DBs `frontend` and `codeintel-db` are now given 1 hour to begin accepting connections before Kubernetes restarts the containers. [#4136](https://github.com/sourcegraph/deploy-sourcegraph/pull/4136)

## 3.39 -> 3.40.1

- `cadvisor` now defaults to run in `privileged` mode. This allows `cadvisor` to collect out of memory events happening to containers which can be used to discover underprovisoned resources. [#804](https://github.com/sourcegraph/deploy-sourcegraph-docker/pull/804)

Follow the [standard upgrade procedure](../deploy/docker-compose/index.md#upgrade) to upgrade your deployment.

*How smooth was this upgrade process for you? You can give us your feedback on this upgrade by filling out [this feedback form](https://share.hsforms.com/1aGeG7ALQQEGO6zyfauIiCA1n7ku?update_version=3.40).*

## 3.39.0 -> 3.39.1

Follow the [standard upgrade procedure](../deploy/docker-compose/index.md#upgrade) to upgrade your deployment.

*How smooth was this upgrade process for you? You can give us your feedback on this upgrade by filling out [this feedback form](https://share.hsforms.com/1aGeG7ALQQEGO6zyfauIiCA1n7ku?update_version=3.38).*

## 3.38 -> 3.39

We made a number of changes to our built-in postgres databases (the `pgsql`, `codeintel-db`, and `codeinsights-db` container)

- **CAUTION** Added the ability to customize postgres server configuration by mounting external configuration files. If you have customized the config in any way, you should copy your changes to the added `postgresql.conf` files [sourcegraph/deploy-sourcegraph-docker#792](https://github.com/sourcegraph/deploy-sourcegraph-docker/pull/792)
- Increased the minimal memory requirement of `pgsql` and `codeintel-db` from `2GB` to `4GB`
-`codeinsights-db` container no longer uses TimescaleDB and is now based on the standard Postgres image [sourcegraph/deploy-sourcegraph-docker#780](https://github.com/sourcegraph/deploy-sourcegraph-docker/pull/780). Metrics scraping is also enabled.

Follow the [standard upgrade procedure](../deploy/docker-compose/index.md#upgrade) to upgrade your deployment.

*How smooth was this upgrade process for you? You can give us your feedback on this upgrade by filling out [this feedback form](https://share.hsforms.com/1aGeG7ALQQEGO6zyfauIiCA1n7ku?update_version=3.39).*

## 3.38.0 -> 3.38.1

Follow the [standard upgrade procedure](../deploy/docker-compose/index.md#upgrade) to upgrade your deployment.

*How smooth was this upgrade process for you? You can give us your feedback on this upgrade by filling out [this feedback form](https://share.hsforms.com/1aGeG7ALQQEGO6zyfauIiCA1n7ku?update_version=3.38).*

## 3.37 -> 3.38

**Minimum version of 1.29 for docker compose is required for this update**

This release adds the requirement that the environment variables `SRC_GIT_SERVERS`, `SEARCHER_URL`, `SYMBOLS_URL`, and `INDEXED_SEARCH_SERVERS` are set for the worker process.

Follow the [standard upgrade procedure](../deploy/docker-compose/index.md#upgrade) to upgrade your deployment.

*How smooth was this upgrade process for you? You can give us your feedback on this upgrade by filling out [this feedback form](https://share.hsforms.com/1aGeG7ALQQEGO6zyfauIiCA1n7ku?update_version=3.37).*

## 3.36 -> 3.37

This release adds a new container that runs database migrations (`migrator`) independently of the frontend container. Confirm the environment variables on this new container match your database settings. [Docs](../deploy/docker-compose/index.md#database-migrations)

Follow the [standard upgrade procedure](../deploy/docker-compose/index.md#upgrade) to upgrade your deployment.

*How smooth was this upgrade process for you? You can give us your feedback on this upgrade by filling out [this feedback form](https://share.hsforms.com/1aGeG7ALQQEGO6zyfauIiCA1n7ku?update_version=3.36).*

## 3.35 -> 3.36

No manual migration is required - follow the [standard upgrade procedure](../deploy/docker-compose/index.md#upgrade) to upgrade your deployment.

*How smooth was this upgrade process for you? You can give us your feedback on this upgrade by filling out [this feedback form](https://share.hsforms.com/1aGeG7ALQQEGO6zyfauIiCA1n7ku?update_version=3.35).*

## 3.35.0 -> 3.35.1

**Due to issues related to Code Insights on the 3.35.0 release, users are advised to upgrade to 3.35.1 as soon as possible.**

There is a [known issue](../../code_insights/how-tos/Troubleshooting.md#oob-migration-has-made-progress-but-is-stuck-before-reaching-100) with the Code Insights out-of-band settings migration not reaching 100% complete when encountering deleted users or organizations.

## 3.34 -> 3.35.1

**Due to issues related to Code Insights on the 3.35.0 release, users are advised to upgrade directly to 3.35.1.**

The `query-runner` service has been decommissioned in the 3.35 release and will be removed during the upgrade.

Follow the [standard upgrade procedure](../deploy/docker-compose/index.md#upgrade) to upgrade your deployment.
To delete the `query-runner` service, specify `--remove-orphans` to your `docker-compose` command.

*How smooth was this upgrade process for you? You can give us your feedback on this upgrade by filling out [this feedback form](https://share.hsforms.com/1aGeG7ALQQEGO6zyfauIiCA1n7ku?update_version=3.34).*

## 3.33 -> 3.34

No manual migration is required - follow the [standard upgrade procedure](../deploy/docker-compose/index.md#upgrade) to upgrade your deployment.

*How smooth was this upgrade process for you? You can give us your feedback on this upgrade by filling out [this feedback form](https://share.hsforms.com/1aGeG7ALQQEGO6zyfauIiCA1n7ku?update_version=3.33).*

## 3.32 -> 3.33

No manual migration is required - follow the [standard upgrade procedure](../deploy/docker-compose/index.md#upgrade) to upgrade your deployment.

*How smooth was this upgrade process for you? You can give us your feedback on this upgrade by filling out [this feedback form](https://share.hsforms.com/1aGeG7ALQQEGO6zyfauIiCA1n7ku?update_version=3.32).*

## 3.31 -> 3.32

No manual migration is required - follow the [standard upgrade procedure](../deploy/docker-compose/index.md#upgrade) to upgrade your deployment.

*How smooth was this upgrade process for you? You can give us your feedback on this upgrade by filling out [this feedback form](https://share.hsforms.com/1aGeG7ALQQEGO6zyfauIiCA1n7ku?update_version=3.31).*

## 3.30.3 -> 3.31

The **built-in** main Postgres (`pgsql`) and codeintel (`codeintel-db`) databases have switched to an alpine-based Docker image. Upon upgrading, Sourcegraph will need to re-index the entire database.

If you have already upgraded to 3.30.3, which uses the new alpine-based Docker images, all users that use our bundled (built-in) database instances should have already performed [the necessary re-indexing](../migration/3_31.md).

> NOTE: The above does not apply to users that use external databases (e.x: Amazon RDS, Google Cloud SQL, etc.).

*How smooth was this upgrade process for you? You can give us your feedback on this upgrade by filling out [this feedback form](https://share.hsforms.com/1aGeG7ALQQEGO6zyfauIiCA1n7ku?update_version=3.30).*

## 3.30.x -> 3.31.2

The **built-in** main Postgres (`pgsql`) and codeintel (`codeintel-db`) databases have switched to an alpine-based Docker image. Upon upgrading, Sourcegraph will need to re-index the entire database.

All users that use our bundled (built-in) database instances **must** read through the [3.31 upgrade guide](../migration/3_31.md) _before_ upgrading.

> NOTE: The above does not apply to users that use external databases (e.x: Amazon RDS, Google Cloud SQL, etc.).

## 3.29 -> 3.30.3

> WARNING: **Users on 3.29.x are advised to upgrade directly to 3.30.3**. If you have already upgraded to 3.30.0, 3.30.1, or 3.30.2 please follow [this migration guide](../migration/3_30.md).

No manual migration required.

Please upgrade to the [`v3.30.2` tag of deploy-sourcegraph-docker](https://github.com/sourcegraph/deploy-sourcegraph-docker/tree/v3.30.2/docker-compose) by following the [standard upgrade procedure](#standard-upgrade-procedure).

## 3.28 -> 3.29

This upgrade adds a new `worker` service that runs a number of background jobs that were previously run in the `frontend` service. See [notes on deploying workers](../workers.md#deploying-workers) for additional details. Good initial values for CPU and memory resources allocated to this new service should match the `frontend` service.

*How smooth was this upgrade process for you? You can give us your feedback on this upgrade by filling out [this feedback form](https://share.hsforms.com/1aGeG7ALQQEGO6zyfauIiCA1n7ku?update_version=3.28).*

## 3.27 -> 3.28

- The memory requirements for `redis-cache` and `redis-store` have been increased by 1GB. See https://github.com/sourcegraph/deploy-sourcegraph-docker/pull/373 for more context.

*How smooth was this upgrade process for you? You can give us your feedback on this upgrade by filling out [this feedback form](https://share.hsforms.com/1aGeG7ALQQEGO6zyfauIiCA1n7ku?update_version=3.27).*

## 3.26 -> 3.27

> Warning: ⚠️ Sourcegraph 3.27 now requires **Postgres 12+**.

If you are using an external database, [upgrade your database](https://docs.sourcegraph.com/admin/postgres#upgrading-external-postgresql-instances) to Postgres 12 or above prior to upgrading Sourcegraph. No action is required if you are using the supplied supplied database images.

Afterwards, please upgrade to the [`v3.27.0` tag of deploy-sourcegraph-docker](https://github.com/sourcegraph/deploy-sourcegraph-docker/tree/v3.27.0/docker-compose) by following the [standard upgrade procedure](#standard-upgrade-procedure).

*How smooth was this upgrade process for you? You can give us your feedback on this upgrade by filling out [this feedback form](https://share.hsforms.com/1aGeG7ALQQEGO6zyfauIiCA1n7ku?update_version=3.26).*

## 3.25 -> 3.26

No manual migration required.

Please upgrade to the [`v3.26.0` tag of deploy-sourcegraph-docker](https://github.com/sourcegraph/deploy-sourcegraph-docker/tree/v3.26.0/docker-compose) by following the [standard upgrade procedure](#standard-upgrade-procedure).

> NOTE: ⚠️ From **3.27** onwards we will only support PostgreSQL versions **starting from 12**.

## 3.24 -> 3.25

- Go `1.15` introduced changes to SSL/TLS connection validation which requires certificates to include a `SAN`. This field was not included in older certificates and clients relied on the `CN` field. You might see an error like `x509: certificate relies on legacy Common Name field`. We recommend that customers using Sourcegraph with an external database and and connecting to it using SSL/TLS check whether the certificate is up to date.
    - AWS RDS customers please reference [AWS' documentation on updating the SSL/TLS certificate](https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/UsingWithRDS.SSL-certificate-rotation.html) for steps to rotate your certificate.

## 3.23 -> 3.24

Please upgrade to the [`v3.24.0` tag of deploy-sourcegraph-docker](https://github.com/sourcegraph/deploy-sourcegraph-docker/tree/v3.24.0/docker-compose) by following the [standard upgrade procedure](#standard-upgrade-procedure).

## 3.22 -> 3.23

No manual migration required.

Please upgrade to the [`v3.23.0` tag of deploy-sourcegraph-docker](https://github.com/sourcegraph/deploy-sourcegraph-docker/tree/v3.23.0/docker-compose) by following the [standard upgrade procedure](#standard-upgrade-procedure).

## 3.21 -> 3.22

No manual migration required.

Please upgrade to the [`v3.22.0` tag of deploy-sourcegraph-docker](https://github.com/sourcegraph/deploy-sourcegraph-docker/tree/v3.22.0/docker-compose) by following the [standard upgrade procedure](#standard-upgrade-procedure).

This upgrade removes the `code intel bundle manager`. This service has been deprecated and all references to it have been removed.

This upgrade also adds a MinIO container that doesn't require any custom configuration. You can find more detailed documentation in https://docs.sourcegraph.com/admin/external_services/object_storage.

## 3.21.0 -> 3.21.1

No manual migration required.

Please upgrade to the [`v3.20.1` tag of deploy-sourcegraph-docker](https://github.com/sourcegraph/deploy-sourcegraph-docker/tree/v3.20.1/docker-compose) by following the [standard upgrade procedure](#standard-upgrade-procedure).

## 3.20.1 -> 3.21.0

Please upgrade to the [`v3.21.0` tag of deploy-sourcegraph-docker](https://github.com/sourcegraph/deploy-sourcegraph-docker/tree/v3.21.0/docker-compose) by following the [standard upgrade procedure](#standard-upgrade-procedure).

This release introduces a second database instance, `codeintel-db`. If you have configured Sourcegraph with an external database, then update the `CODEINTEL_PG*` environment variables to point to a new external database as described in the [external database documentation](../external_services/postgres.md). Again, these must not point to the same database or the Sourcegraph instance will refuse to start.

### If you wish to keep existing LSIF data

> Warning: **Do not upgrade out of the 3.21.x release branch** until you have seen the log message indicating the completion of the LSIF data migration, or verified that the `/lsif-storage/dbs` directory on the precise-code-intel-bundle-manager volume is empty. Otherwise, you risk data loss for precise code intelligence.

If you had LSIF data uploaded prior to upgrading to 3.21.0, there is a background migration that moves all existing LSIF data into the `codeintel-db` upon upgrade. Once this process completes, the `/lsif-storage/dbs` directory on the precise-code-intel-bundle-manager volume should be empty, and the bundle manager should print the following log message:

> Migration to Postgres has completed. All existing LSIF bundles have moved to the path /lsif-storage/db-backups and can be removed from the filesystem to reclaim space.

**Wait for the above message to be printed in `docker logs precise-code-intel-bundle-manager` before upgrading to the next Sourcegraph version**, then if everything is working you can free up disk space by deleting the backup bundle files using this command:

```sh
docker exec -it precise-code-intel-bundle-manager sh -c 'rm -rf /lsif-storage/db-backups'
```

## 3.19.2 -> 3.20.1

No manual migration required.

Please upgrade to the [`v3.20.1` tag of deploy-sourcegraph-docker](https://github.com/sourcegraph/deploy-sourcegraph-docker/tree/v3.20.1/docker-compose) by following the [standard upgrade procedure](#standard-upgrade-procedure).

## 3.19.1 -> 3.19.2

No manual migration required.

Please upgrade to the [`v3.19.2` tag of deploy-sourcegraph-docker](https://github.com/sourcegraph/deploy-sourcegraph-docker/tree/v3.19.2/docker-compose) by following the [standard upgrade procedure](#standard-upgrade-procedure).

## 3.18.0-1 -> 3.19.1

No manual migration required.

Please upgrade to the [`v3.19.1` tag of deploy-sourcegraph-docker](https://github.com/sourcegraph/deploy-sourcegraph-docker/tree/v3.19.1/docker-compose) by following the [standard upgrade procedure](#standard-upgrade-procedure).

## 3.18.0 -> 3.18.0-1

This release fixes `observability.alerts` in the site configuration. No manual migration required.

Please upgrade to the [`v3.18.0-1` tag of deploy-sourcegraph-docker](https://github.com/sourcegraph/deploy-sourcegraph-docker/tree/3.18/docker-compose) by following the [standard upgrade procedure](#standard-upgrade-procedure).

## v3.17.2 -> 3.18.0

No manual migration required.

Please upgrade to the [`v3.18.0` tag of deploy-sourcegraph-docker](https://github.com/sourcegraph/deploy-sourcegraph-docker/tree/3.18/docker-compose) by following the [standard upgrade procedure](#standard-upgrade-procedure).

## v3.16.0 -> v3.17.2

No manual migration is required.

Please upgrade to the [`v3.17.2` tag of deploy-sourcegraph-docker](https://github.com/sourcegraph/deploy-sourcegraph-docker/tree/v3.16.0/docker-compose) by following the [standard upgrade procedure](#standard-upgrade-procedure).

## v3.15.1 -> v3.16.0

No manual migration is required.

Please upgrade to the [`v3.16.0` tag of deploy-sourcegraph-docker](https://github.com/sourcegraph/deploy-sourcegraph-docker/tree/v3.16.0/docker-compose) by following the [standard upgrade procedure](#standard-upgrade-procedure).

## (v3.14.2, v3.14.4) -> v3.15.1

No manual migration is required.

Please upgrade to the [`v3.15.1` tag of deploy-sourcegraph-docker](https://github.com/sourcegraph/deploy-sourcegraph-docker/tree/v3.15.1/docker-compose) by following the [standard upgrade procedure](#standard-upgrade-procedure).

### (Optional) Keeping LSIF data

If your users have uploaded LSIF precise code intelligence data, you may keep it by running the following command after you have ran `docker-compose up` with the new v3.15.1 version:

```
docker run --rm -it -v /var/lib/docker:/docker alpine:latest sh -c 'cp -R /docker/volumes/docker-compose_lsif-server/_data/* /docker/volumes/docker-compose_precise-code-intel-bundle-manager/_data/'
```

Followed by:

```sh
docker run --rm -it -v /var/lib/docker:/docker alpine:latest sh -c 'chown -R 100:101 /docker/volumes/docker-compose_precise-code-intel-bundle-manager'
docker restart precise-code-intel-bundle-manager
```

## v3.14.2 -> v3.14.4

No manual migration is required.

Please upgrade to the [`v3.14.4` tag of deploy-sourcegraph-docker](https://github.com/sourcegraph/deploy-sourcegraph-docker/tree/v3.14.4/docker-compose) by following the [standard upgrade procedure](#standard-upgrade-procedure).

## v3.14.0 -> v3.14.2

No manual migration is required.

Please upgrade to the [`v3.14.2` tag of deploy-sourcegraph-docker](https://github.com/sourcegraph/deploy-sourcegraph-docker/tree/v3.14.2/docker-compose) by following the [standard upgrade procedure](#standard-upgrade-procedure).

## v3.13 -> 3.14

No manual migration is required.

Please be sure to upgrade to the [`v3.14.0-1` tag of deploy-sourcegraph-docker](https://github.com/sourcegraph/deploy-sourcegraph-docker/tree/v3.14.0-1/docker-compose) by following the [standard upgrade procedure](#standard-upgrade-procedure).

If you have upgrade to `v3.14.0` already (not the `v3.14.0-1` version) and are experiencing restarts of lsif-server, please run the following on the host machine to correct it:

```sh
docker run --rm -it -v /var/lib/docker:/docker alpine:latest sh -c 'chown -R 100:101 /docker/volumes/docker-compose_lsif-server'
docker restart lsif-server
```

## v3.12 -> v3.13

A manual migration is required. Please follow the [standard upgrade procedure](#standard-upgrade-procedure) to take down the current deployment, perform the manual migration, and then upgrade using the [`v3.13.2` tag of deploy-sourcegraph-docker](https://github.com/sourcegraph/deploy-sourcegraph-docker/tree/v3.13.2/docker-compose).

### Manual migration step: adjust file permissions

Please adjust the redis-store and redis-cache volume permissions by running the following on the host machine:

```
docker run --rm -it -v /var/lib/docker:/docker alpine:latest sh -c 'chown -R 999:1000 /docker/volumes/docker-compose_redis-store /docker/volumes/docker-compose_redis-cache'
```

### Standard upgrade procedure

Refer to the [upgrade guide](../deploy/docker-compose/index.md#upgrade).
