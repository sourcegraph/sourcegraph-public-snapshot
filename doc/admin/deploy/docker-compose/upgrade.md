# Upgrade Sourcegraph on Docker Compose

This document describes the process to update a Docker Compose Sourcegraph instance.

### Standard upgrades

A [standard upgrade](../../updates/index.md#standard-upgrades) occurs between two minor versions of Sourcegraph. If you are looking to jump forward several versions, you must perform a [multi-version upgrade](#multi-version-upgrades) instead.

If you've [configured Docker Compose with a release branch](index.md#step-1-prepare-the-deployment-repository), please merge the upstream release tag for the next minor version into your `release` branch. In the following example, the release branch is being upgraded to v3.43.2.

```bash
# first, checkout the release branch
git checkout release
# fetch updates
git fetch upstream
# merge the upstream release tag into your release branch
git checkout release
git merge v3.43.2
```

#### Address any merge conflicts you might have

For each conflict, you need to reconcile any customizations you made with the updates from the new version. Use the information you gathered earlier from the change log and changes list to interpret the merge conflict and to ensure that it doesn't over-write your customizations. You may need to update your customizations to accommodate the new version. 

> NOTE: If you have made no changes or only very minimal changes to your configuration, you can also ask git to always select incoming changes in the event of merge conflicts. In the following example merges will be accepted from the upstream version v3.43.2:
>
> `git merge -X theirs v3.43.2`
>
> If you do this, make sure your configuration is correct before proceeding because it may have made changes to your docker-compose YAML file.

#### Clone the updated release branch to your server

SSH into your instance and navigate to the appropriate folder:  

```bash
# AWS
cd /home/ec2-user/deploy-sourcegraph-docker/docker-compose
# Azure, Digital Ocean, Google Cloud
cd /root/deploy-sourcegraph-docker/docker-compose
```

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

> NOTE: It is highly recommended to **take an up-to-date snapshot of your databases** prior to starting a multi-version upgrade. The upgrade process aggressively mutates the shape and contents of your database, and undiscovered errors in the migration process or unexpected environmental differences may cause an unusable instance or data loss.
>
> We recommend performing the entire upgrade procedure on an idle clone of the production instance and switch traffic over on success, if possible. This may be low-effort for installations with a canary environment or a blue/green deployment strategy.
>
> **If you do not feel confident running this process solo**, contact customer support team to help guide you thorough the process.

**Before performing a multi-version upgrade**:
 
- Read our [update policy](../../updates/index.md#update-policy) to learn about Sourcegraph updates.
- Find the entries that apply to the version range you're passing through in the [update notes for Sourcegraph with Docker Compose](../../updates/docker_compose.md#multi-version-upgrade-procedure).

To perform a multi-version upgrade on a Sourcegraph instance running on Docker compose:

1. Spin down the instance:
  - `docker-compose stop`
1. Spin back up each in-use local database so that the `migrator` can access them. Any [externalized database](../../external_services/postgres.md) is already accessible from the `migrator` so no action is needed for them.
  - `docker-compose up -d pgsql`
  - `docker-compose up -d codeintel-db`
  - `docker-compose up -d codeinsights-db`
1. Run the `migrator upgrade` command targetting the same databases as your instance. See the [command documentation](./../../how-to/manual_database_migrations.md#upgrade) for additional details. In short, the [migrator](https://sourcegraph.com/search?q=context:global+repo:%5Egithub%5C.com/sourcegraph/deploy-sourcegraph-docker%24+file:docker-compose%5C.yaml+container_name:+migrator) is invoked with a `docker run` command using the same compose network and using environment variables indicating the instance's databases.
1. Now that the data has been prepared to run against a new version of Sourcegraph, the infrastructure can be updated. The remaining steps follow the [standard upgrade for Docker Compose](#standard-upgrades).
