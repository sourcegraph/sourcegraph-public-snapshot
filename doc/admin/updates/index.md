# Updating Sourcegraph

For product update notes, please refer to the [changelog](../../CHANGELOG.md).

## Update policy

A new version of Sourcegraph is released every month (with patch releases in between, released as needed). Check the [Sourcegraph blog](https://about.sourcegraph.com/blog) or the site admin updates page to learn about updates. We actively maintain the two most recent monthly releases of Sourcegraph.

## Choosing the Correct Upgrade Path
We support two upgrade paths: moving one minor version ahead, i.e. `3.42` to `3.43` (standard upgrade), moving many minor versions ahead, i.e `3.36` to `3.43` (multi-version upgrades). It is vital that you choose the correct upgrade path when upgrading your instance. If you attempt to upgrade multiple versions using the standard upgrade process, it will fail. 

### Standard upgrades

A **standard upgrade** moves an instance from *one version to an adjacent minor version*, for example the upgrade `v3.41 -> v3.42`. Note that patch releases do not have to be adopted when moving between minor versions. For example, upgrading from `v3.41.0 -> v3.41.1 -> v3.42.0` has an unnecessary step.

> NOTE: Due to its compatibility with previous versions, we support the upgrade `v3.43 -> v4.0.1` as a one-minor-version "standard" upgrade.

This upgrade process involves updating only infrastructure: containers must reflect new version tags, additions and removal of services must be addressed, resource allocation may need to be readjusted, etc. For environments that support rolling updates (Kubernetes with or without Helm), a standard upgrade minimizes service disruption as new and old versions of the instance can work with the same database schema for a short period.

**We recommend following this upgrade process and keeping your instance up-to-date**, once it is on the most recent version. If your instance is currently more than one minor version behind the latest release, a single [multi-version upgrade](#multi-version-upgrades) can bring you up to speed.

Also [check the status out-of-band migrations](../how-to/unfinished_migration.md#checking-progress) prior to upgrading. If there is a warning displayed on this page, then an out-of-band migration is currently in progress and must finish prior to moving off of the current version. The newer version of Sourcegraph will detect that there is unmigrated data and will refuse to start, requiring an unnecessary rollback procedure. Note that multi-version upgrades will complete these migrations as part of the upgrade process, so checking this status prior to a multi-version upgrade is unnecessary.

To perform a standard upgrade, check the notes and follow the guide for your specific environment:

- [Sourcegraph with Docker Compose](docker_compose.md#upgrade-procedure)
- [Sourcegraph with Kubernetes](kubernetes.md#upgrade-procedure)
- [Single-container Sourcegraph with Docker](server.md#upgrade-procedure)
- [Pure-docker custom deployments](pure_docker.md)

### Multi-version upgrades

A **multi-version** upgrade moves an instance *multiple minor versions ahead*. We currently support jumping from `v3.20` to any future version (using a version of the `migrator` at least as new as the target version).

This upgrade process involves spinning down the active instance (incurring a definite downtime period), running a migration utility on the database, and spinning up the infrastructure for the new target instance. This migration utility performs both schema migrations, as well as data migrations that generally happen slowly in the background over several versions.

To perform a multi-version upgrade, check the notes and follow the guide for your specific environment:

- [Sourcegraph with Docker Compose](docker_compose.md#multi-version-upgrade-procedure)
- [Sourcegraph with Kubernetes](kubernetes.md#multi-version-upgrade-procedure)
- [Single-container Sourcegraph with Docker](server.md#multi-version-upgrade-procedure)
- [Pure-docker custom deployments](pure_docker.md)

## Migrating to a new deployment type

See [this page](../deploy/index.md) to get advice on which deployment type you should be running.

- [Migrate to Docker Compose](../deploy/docker-compose/migrate.md) for improved stability and performance if you are using a single-container `sourcegraph/server` deployment.
- [Migrate to a Kubernetes cluster](../deploy/kubernetes/index.md) if you exceed the limits of a single machine Docker Compose deployment.
