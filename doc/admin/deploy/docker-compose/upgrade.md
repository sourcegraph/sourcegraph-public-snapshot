# Upgrade Sourcegraph on Docker Compose

This document describes the process to update a Docker Compose Sourcegraph instance.

**Before upgrading**:

1. Read our [update policy](../../updates/index.md#update-policy) to learn about Sourcegraph updates.
2. Find the relevant entry for your update in the [update notes for Sourcegraph with Docker Compose](../../updates/docker_compose.md).

### Standard upgrades

A [standard upgrade](../../updates/index.md#standard-upgrades) occurs between two minor versions of Sourcegraph. If you are looking to jump forward several versions, you must perform a [multi-version upgrade](#multi-version-upgrades) instead.

If you [configured Docker Compose with a release branch](index.md#step-3-configure-the-release-branch), please merge the upstream release tag for the next minor version into your `release` branch. In the following example, the release branch is being upgraded to v3.40.2.

```bash
# first, checkout the release branch
git checkout release
# fetch updates
git fetch upstream
# merge the upstream release tag into your release branch
git checkout release
git merge v3.40.2
```

#### Address any merge conflicts you might have

For each conflict, you need to reconcile any customizations you made with the updates from the new version. Use the information you gathered earlier from the change log and changes list to interpret the merge conflict and to ensure that it doesn't over-write your customizations. You may need to update your customizations to accommodate the new version. 

> NOTE: If you have made no changes or only very minimal changes to your configuration, you can also ask git to always select incoming changes in the event of merge conflicts. In the following example merges will be accepted from the upstream version v3.40.2:
>
> `git merge -X theirs v3.40.2`
>
> If you do this, make sure your configuration is correct before proceeding because it may have made changes to your docker-compose YAML file.

#### Clone the updated release branch to your server

SSH into your instance and navigate to the appropriate folder:  
- AWS: `/home/ec2-user/deploy-sourcegraph-docker/docker-compose`  
- Digital Ocean: `/root/deploy-sourcegraph-docker/docker-compose`  
- Google Cloud: `/root/deploy-sourcegraph-docker/docker-compose`  

Download all the latest docker images to your local docker daemon:

```bash
docker-compose pull --include-deps
```

Restart Docker Compose using the new minor version along with your customizations:

```bash
docker-compose up -d --remove-orphans
```
### Multi-version upgrades

A [multi-version upgrade](../../updates/index.md#multi-version-upgrades) is a downtime-incurring upgrade from version 3.20 or later to any future version. Multi-version upgrades will run both schema and data migrations to ensure the data available from the instance remains available post-upgrade.

**Before performing a multi-version upgrade**:

- **Take an up-to-date snapshot of your databases.** We are unable to exhaustively test all upgrade paths or catch all possible edge cases in a customer environment. The upgrade process aggressively mutates the shape and contents of your database, and uncaught errors in the migration process or unexpected environmental differences may cause data loss. **If you do not feel confident running this process solo**, contact customer support team to help walk you thorough the process.
- If possible, upgrade an idle clone of the production instance and switch traffic over on success. This may be low-effort for installations with a canary environment or a blue/green deployment strategy.
- Run the `migrator` drift detection on your current version to detect and repair any database schema discrepencies. Running with an unexpected schema may cause a painful upgrade process that may require engineering support. See the [command documentation](./../../how-to/manual_database_migrations.md#drift) for additional details.

To perform a multi-version upgrade on a Sourcegraph instance running on Docker compose:

1. Spin down the instance:
  - `docker-compose stop`
1. Spin back up each in-use local database so that the `migrator` can access them. Any [externalized database](../../external_services/postgres.md) is already accessible from the `migrator` so no action is needed for them.
  - `docker-compose up -d pgsql`
  - `docker-compose up -d codeintel-db`
  - `docker-compose up -d codeinsights-db`
1. Run the `migrator upgrade` command targetting the same databases as your instance. See the [command documentation](./../../how-to/manual_database_migrations.md#upgrade) for additional details.
1. Now that the data has been prepared to run against a new version of Sourcegraph, the infrastructure can be updated. The remaining steps follow the [standard upgrade for Docker Compose](#standard-upgrades).
