# Upgrade Sourcegraph on Docker Compose

This document describes the process to update a Docker Compose Sourcegraph instance.

## Process Overview 

1\. [Gather upgrade information](#1-gather-upgrade-information)  
2\. [Upgrade to the next minor version](#2-upgrade-to-the-next-minor-version)  
3\. [Restart Sourcegraph and allow migrations to finish](#3-restart-sourcegraph-and-allow-migrations-to-finish)  
4\. [Repeat steps 2 and 3 until Sourcegraph is up to date](#4-repeat-steps-2-and-3-until-sourcegraph-is-up-to-date)  

## 1. Gather upgrade information

To start, familiarise yourself with the details of the upgrade that you will be performing:

- The version number of your current instance.
- The version number of the latest Sourcegraph release.
- The sequence of minor releases required to upgrade from your version to the latest version (your upgrade sequence). This is required because Sourcegraph only supports upgrading one minor version at a time. 
- Details of the changes for each minor version in your upgrade sequence in the [product changelog](../../../CHANGELOG.md) and also in the [docker-compose changelog](../../updates/docker_compose.md). 

## 2. Upgrade to the next minor version

If you [configured Docker Compose with a release branch](index.md#configure-a-release-branch), you merge the upstream release tag for the next minor version into your `release` branch. In the following example the release branch is being upgraded to v3.40.2. 

```bash
# fetch updates
git fetch upstream
# merge the upstream release tag into your release branch
git checkout release
git merge v3.40.2
```

### Address any merge conflicts you might have

For each conflict, you need to reconcile any customizations you made with the updates from the new version. Use the information you gathered earlier from the change log and changes list to interpret the merge conflict and to ensure that it doesn't over-write your customizations. You may need to update your customizations to accommodate the new version. 

> NOTE: If you have made no changes or only very minimal changes to your configuration, you can also ask git to always select incoming changes in the event of merge conflicts. In the following example merges will be accepted from the upstream version v3.40.2:
>
> `git merge -X theirs v3.40.2`
>
> If you do this, make sure your configuration is correct before proceeding because it may have made changes to your docker-compose YAML file.

### Clone the updated release branch to your server

SSH into your instance and navigate to the appropriate folder:  
- AWS: `/home/ec2-user/deploy-sourcegraph-docker/docker-compose`  
- Digital Ocean: `/root/deploy-sourcegraph-docker/docker-compose`  
- Google Cloud: `/root/deploy-sourcegraph-docker/docker-compose`  

Download all the latest docker images to your local docker daemon:

```bash
docker-compose pull --include-deps
```
## 3. Restart Sourcegraph and allow migrations to finish

### Restart 

Restart Docker Compose using the new minor version along with your customizations:

```bash
docker-compose up -d ---remove-orphans
```
### Check on the status of migrations

Before upgrading to the next minor version in your upgrade sequence, you must allow the migrator service to finish any required database and out-of-band migrations associated with the upgrade. Check the migrator service and frontend service logs for information regarding the database migration status. Check the out of band migration status in Sourcegraph in the `Site Admin > Maintenance > Migrations` page to show the progress of all active migrations. This page will also display a prominent warning if an upgrade (or downgrade) would result in an instance that refuses to start due to an illegal migration state.

![Unfinished migration warning](https://storage.googleapis.com/sourcegraph-assets/oobmigration-warning.png)

In this situation, upgrading to the next version will not result in any data loss, but all new instances will detect the illegal migration state and refuse to start up with a fatal message (`Unfinished migrations`).

See [How to troubleshoot an unfinished migration](../../how-to/unfinished_migration.md) for more information.

## 4. Repeat steps 2 and 3 until Sourcegraph is up to date 

Now that Sourcegraph is stable after the minor version upgrade, you can continue to the next minor version in your upgrade sequence and repeat this process until you are at the latest version. 
