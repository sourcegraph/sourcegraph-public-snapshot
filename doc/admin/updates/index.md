# Updating sourcegraph

This page is intended as an entry point into Sourcegraph versioning and the `migrator` service which manages our database schemas. Here we'll cover general concepts and direct you toward relevant operations pages.

> *Note: For product update notes, please refer to the [changelog](../../CHANGELOG.md). You can also [skip general concepts for technical operations](#upgrade-operations), if you know what kind of upgrade you need.*

## General concepts

### Upgrading

A new Sourcegraph release consists of updated images that incorporate code changes from our [main repository](https://github.com/sourcegraph/sourcegraph). It also includes modifications to deployment manifests in our supporting deployment repositories, such as our [k8s helm repo](https://github.com/sourcegraph/deploy-sourcegraph-helm). 

When upgrading Sourcegraph, it is necessary to update and apply the deployment manifests. However, in addition to the manifest updates, it is crucial to ensure that all underlying database schemas of Sourcegraph are also updated.

### Release Schedule and versioning
Sourcegraph releases use semantic versioning, for example: `v5.0.3`
| Description   | Version |
|---------------|---------|
| major         | `5`     |
| minor         | `0`     |
| patch         | `3`     |

To learn more about our release schedule see our [handbook](https://handbook.sourcegraph.com/departments/engineering/dev/process/releases/#sourcegraph-releases). In general a new minor version of Sourcegraph is released every month, accompanied by weekly patch releases. Major releases are less frequent and represent significant changes in Sourcegraph.You can also check the [Sourcegraph blog](https://about.sourcegraph.com/blog) for more information about the latest release.

### Upgrade types

Sourcegraph has two upgrade types. **Standard** upgrades and **Multiversion** upgrades. We generally recommend standard upgrades.

**Standard** 
- Moves Sourcegraph one version forward (`v5.0.0` to `v5.1.0`), *this is usually one minor version unless the next version released is a major version*.
- Requires no downtime

**Multiversion**
- Moves Sourcegraph multiple versions forward (`v5.0.0` to `v5.2.0`).
- Requires manual `migrator` operations
- Requires downtime
- We currently support jumping from version `v3.20` or later to any future version.

> *Note: Patch versions don't determine upgrade type -- you should always upgrade to the latest patch!*

| From Version | To Version | Upgrade Type | Notes                                |
|--------------|------------|--------------|--------------------------------------|
| `v5.0.0`     | `v5.1.0`   | Standard     | A minor version                       |
| `v5.0.0`     | `v5.2.0`   | Multiversion  | Two minor versions                    |
| `v3.41.0`    | `v3.42.2`  | Standard     | A minor version and patch version     |
| `v3.43.0`    | `v4.0.0`   | Standard     | A major version change but only one absolute version |
| `v5.0.0`     | `v5.2.2`   | Multiversion  | Two minor versions and patch          |
| `v3.33.0`    | `v5.2.0`   | Multiversion  | Major and minor                       |
| `v4.5.0`     | `v5.0.0`   | Standard     | This is a major version but only one version change |
| `v4.4.2`     | `v5.0.3`   | Multiversion  | Major, minor, and patch               |

> *Note: Our major releases don't occur on a consistent interval, so make sure to check our changelog if you aren't certain about wether a major version is multiple minor versions away from your current version. You can also reach out to our support team [support@sourcegraph.com](mailto:support@sourcegraph.com)*

### Sourcegraph databases & migrator

To facilitate the management of Sourcegraph's databases, we have created the `migrator` service. `migrator` is usually triggered automatically on Sourcegraph startup but can also be interacted with like a cli tool. Migrator's primary purpose is to manage and apply schema migrations. 
- To learn more about migrations see our [developer docs](https://docs.sourcegraph.com/dev/background-information/sql/migrations_overview). 
- For a full listing of migrator's command arguments see its [usage docs](https://docs.sourcegraph.com/admin/how-to/manual_database_migrations).

## Upgrade operations

### General upgrade procedure

Sourcegraph upgrades take the following general form:
1. Determine if your instance is ready to Upgrade
2. Merge the latest Sourcegraph release into your deployment manifests
3. Run migrator by reapplying your manifests in a **standard upgrade** (migrator by default uses the `up` command), or by disabling services connected to your databases and running `migrator` with the `upgrade` argument.

**Determine if your release is ready for upgrade**

Starting 5.0.0, as an admin you are able to check instance upgrade readiness by navigating to the `Site admin > Updates` page. Here you'll be notified if your instance has any **schema drift** or unfinished **out of band migrations**.

![Screenshot 2023-05-17 at 1 37 12 PM](https://github.com/sourcegraph/sourcegraph/assets/13024338/185fc3e8-0706-4a23-b9fe-e262f9a9e4b3)

### Standard upgrades index

To perform a standard upgrade, check the notes and follow the guide for your specific environment:

- [Sourcegraph with Docker Compose](docker_compose.md#upgrade-procedure)
- [Sourcegraph with Kubernetes](kubernetes.md#upgrade-procedure)
- [Single-container Sourcegraph with Docker](server.md#upgrade-procedure)
- [Pure-docker custom deployments](pure_docker.md)

### Multi-version upgrades index

- [Sourcegraph with Docker Compose](docker_compose.md#multi-version-upgrade-procedure)
- [Sourcegraph with Kubernetes](kubernetes.md#multi-version-upgrade-procedure)
- [Single-container Sourcegraph with Docker](server.md#multi-version-upgrade-procedure)
- [Pure-docker custom deployments](pure_docker.md)
