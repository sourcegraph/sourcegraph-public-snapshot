# Upgrade Sourcegraph on Docker Compose

This document describes the process to update a Docker Compose Sourcegraph instance.

> ***Attention: Always consult the [release notes](../../updates/docker_compose.md) for the version you plan to upgrade to before an upgrade.***

### Standard upgrades

A [standard upgrade](../../updates/index.md#upgrade-types) occurs between a Sourcegraph version and the minor or major version released immediately after it. If you are looking to jump forward several versions, you must perform a [multi-version upgrade](#multi-version-upgrades) instead.

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

1. Spin down any pods that access the database. The easiest way to do this is to shut down the instance entirely:
  - Run `docker-compose stop` in the directory with the `docker-compose.yaml` file.
  - > Note here about 3.26 or before [upgrading really old stuff](../../updates/upgrading-early-versions.md)
2. Pull the upstream changes for the target instance version and resolve any git merge conflicts. The [standard upgrade procedure](#standard-upgrades) describes this step in more detail.
3. If using local database instances, start the containers now via `docker-compose up -d pgsql codeintel-db codeinsights-db`. The following migrator command will start these containers on-demand if this step is skipped, but running them separately will make startup errors more apparent.
4. Follow the instructions on [how to run the migrator job in Docker Compose](../../how-to/manual_database_migrations.md#docker--docker-compose) to perform the upgrade migration. For specific documentation on the `upgrade` command, see the [command documentation](../../how-to/manual_database_migrations.md#upgrade). The following specific steps are an easy way to run the upgrade command:
  1. Edit the definition of the `migrator` container in the `docker-compose.yaml` so that the value of the `command` key is set to `['upgrade', '--from=<old version>', '--to=<new version>']`. It is recommended to also add the `--dry-run` flag on a trial invocation to detect if there are any issues with database connection, schema drift, or mismatched versions that need to be addressed.
  2. Run the upgrade via `docker-compose up migrator` and wait for it to complete.
  3. Reset the `command` key altered in the previous steps to `['up']` so that the container initialization process will work as expected.
5. The remaining infrastructure can now be updated. The [standard upgrade procedure](#standard-upgrades) describes this step in more detail.
  - Run `docker-compose pull --include-deps` to pull new images.
  - Run `docker-compose up -d --remove-orphans` to start the containers of the updated instance.
