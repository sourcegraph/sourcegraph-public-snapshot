# Upgrades and Migration

A new version of Sourcegraph is released every month (with patch releases in between as needed). We actively maintain the two most recent, monthly releases of Sourcegraph. The [changelog](../../CHANGELOG.md) provides all information related to any changes that are/were in a release.
 
This, in combination with [Upgrade notes](#upgrade-notes) will help you prepare your instance for an upgrade. 

In parallel or independent of your upgrade activities, migration of data may be required as well. This can be true when upgrading to a specific version, or when you are looking to migrate from one deployment typeAs part of this activity you should review the [Migration](#migration) section below as well as the Migration instructions specific to your deployment type.

## Upgrades

### Upgrade policy

A new version of Sourcegraph is released every month (with patch releases in between, released as needed). Check the [Sourcegraph blog](https://about.sourcegraph.com/blog) or the site admin updates page to learn about updates. We actively maintain the two most recent monthly releases of Sourcegraph.

**Regardless of your deployment type**, the following rules apply:

- **Upgrade one minor version at a time**, e.g. v3.26 --> v3.27 --> v3.28.
  - Patches (e.g. vX.X.4 vs. vX.X.5) do not have to be adopted when moving between vX.X versions.
- **Check the [update notes for your deployment type](#update-notes) for any required manual actions** before updating.
- Check your [out of band migration status](../migration/index.md) prior to upgrade to avoid a necessary rollback while the migration finishes.

### Upgrade notes

Please see the instructions for your deployment type:

- [Sourcegraph with Docker Compose](docker_compose.md)
- [Sourcegraph with Kubernetes](kubernetes.md)
- [Single-container Sourcegraph with Docker](server.md)
- [Pure-Docker custom deployments](pure_docker.md)

For product update notes, please refer to the [changelog](../../CHANGELOG.md).

## Migration

For Sourcegraph versions 3.37 and later the migrator service runs as an initial step of the upgrade process for Kubernetes and Docker-compose instance deployments. This service is also designed to be invokable directly by a site administrator to perform common tasks dealing with database state. For more information on the service, view our [migration docs](TBD).


## Migrating to a new deployment type

See [this page](../install/index.md) to get advice on which deployment type you should be running.

- [Migrate to Docker Compose](../install/docker-compose/migrate.md) for improved stability and performance if you are using a single-container `sourcegraph/server` deployment.
- [Migrate to a Kubernetes cluster](../install/kubernetes/index.md) if you exceed the limits of a single machine Docker Compose deployment.

Sourcegraph runs data migrations in the background while the instance is active instead of requiring a blocking migration during startup or manual migrations requiring downtime.

Migrations are introduced at a particular version with an expected lifetime (a course of several versions). At the end of this lifetime, the migration will be marked as deprecated and the instance will no longer be able to read the old data. This requires that migrations finish prior to an upgrade to a version that no longer understands your instance's data.

### Out of band migrations

The `Site Admin > Maintenance > Migrations` page shows the progress of all active migrations. This page will also display a prominent warning if when upgrade (or downgrade) would result in an instance that refuses to start due to an illegal migration state.

![Unfinished migration warning](https://storage.googleapis.com/sourcegraph-assets/oobmigration-warning.png)

In this situation, upgrading to the next version will not result in any data loss, but all new instances will detect the illegal migration state and refuse to start up with a fatal message (`Unfinished migrations`).

### Migration guides

- [Migrating from Oracle OpenGrok to Sourcegraph for code search](migration/opengrok.md)
- [Back up or migrate to a new Sourcegraph instance](migrate-backup.md)
- [How to troubleshoot an unfinished migration](../how-to/unfinished_migration.md)
- [Migrate from the single Docker image to Docker Compose](migrate-to-docker-compose.md))
- [Migrating from Sourcegraph 3.30.0, 3.30.1, and 3.30.2](3_30.md)
- [Migrating to Sourcegraph 3.31.x](3_31.md)
- **Deprecated** [Migrating from Sourcegraph 2.13 to 3.0.0](3_0.md)
- **Deprecated** [Migrating from Sourcegraph 3.x to 3.7.2+](3_7.md)
- **Deprecated** [Migrating from Sourcegraph 3.x to 3.11](3_11.md)