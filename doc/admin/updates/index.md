# Updating Sourcegraph

## Update policy

A new version of Sourcegraph is released every month (with patch releases in between, released as needed). Check the [Sourcegraph blog](https://about.sourcegraph.com/blog) or the site admin updates page to learn about updates. We actively maintain the two most recent monthly releases of Sourcegraph.

### Standard upgrades

A **standard upgrade** moves an instance from one version to an adjacent minor version, for example the upgrade `v3.41 -> v3.42`. Note that patch releases do not have to be adopted when moving between minor versions. For example, upgrading from `v3.41.0 -> v3.41.1 -> v.3.42.0` has an unnecessary step.

> NOTE: Due to its compatibility with previous versions, we support the upgrade `v3.43 -> v4.0` as a one-minor-version "standard" upgrade.

This upgrade process involves updating only infrastructure: containers must reflect new version tags, additions and removal of services must be addressed, resource allocation may need to be readjusted, etc. For environments that support rolling updates, this process minimizes instance downtime, and eventually will become a zero-downtime process.

**We recommend following this upgrade process and keeping your instance up-to-date.** If your instance is currently more than one minor version behind the latest release, a single [multi-version upgrade](#multi-version-upgrades) can bring you up to speed.

To perform a standard upgrade, check the [update notes for your deployment type](#update-notes) for instructions and for any conditions that must be true or required manual actions that must be performed **prior to updating**.

Also [check the status out-of-band migrations](../how-to/unfinished_migration.md#checking-progress) prior to upgrading. If there is a warning displayed on this page, then an out-of-band migration is currently in progress and must finish prior to moving off of the current version. The newer version of Sourcegraph will detect that there is unmigrated data and will refuse to start, requiring an unnecessary rollback procedure. Note that multi-version upgrades will complete these migrations as part of the upgrade process, so checking this status prior to a multi-version upgrade is unnecessary.

### Multi-version upgrades

A **multi-version** upgrade moves an instance multiple minor versions ahead. We currently support jumping from `v3.20` to any future version (using a version of the `migrator` at least as new as the target version).

This upgrade process involves spinning down the active instance (incurring a definite downtime period), running a migration utility on the database, and spinning up the infrastructure for the new target instance. This migration utility performs both schema migrations, as well as data migrations that generally happen slowly in the background over several versions.

To perform a multi-version upgrade, check the [update notes for your deployment type](#update-notes) for instructions and for any conditions that must be true or required manual actions that must be performed **prior to updating**. Then check one of the following guides for environment-specific instructions.

- [Sourcegraph with Docker Compose](../deploy/docker-compose/upgrade.md#multi-version-upgrades)
- [Sourcegraph with Kubernetes](../deploy/kubernetes/update.md#multi-version-upgrades)
- [Sourcegraph with Kubernetes and Helm](../deploy/kubernetes/helm.md#multi-version-upgrades)
- [Single-container Sourcegraph with Docker](../deploy/docker-single-container.md#multi-version-upgrades)

## Update notes

Please see the instructions for your deployment type:

- [Sourcegraph with Docker Compose](docker_compose.md)
- [Sourcegraph with Kubernetes](kubernetes.md)
- [Pure-docker custom deployments](pure_docker.md)
- [Single-container Sourcegraph with Docker](server.md)

For product update notes, please refer to the [changelog](../../CHANGELOG.md).

## Migrating to a new deployment type

See [this page](../deploy/index.md) to get advice on which deployment type you should be running.

- [Migrate to Docker Compose](../deploy/docker-compose/migrate.md) for improved stability and performance if you are using a single-container `sourcegraph/server` deployment.
- [Migrate to a Kubernetes cluster](../deploy/kubernetes/index.md) if you exceed the limits of a single machine Docker Compose deployment.
