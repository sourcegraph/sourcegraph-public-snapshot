# Updating a Docker Compose Sourcegraph instance

This document describes the exact changes needed to update a Docker Compose Sourcegraph instance.

Each section comprehensively describes the steps needed to upgrade, and any manual migration steps you must perform.

## v3.13 -> 3.14

No manual migration is required.

Simply upgrade using the [`v3.14.0` tag of deploy-sourcegraph-docker](https://github.com/sourcegraph/deploy-sourcegraph-docker/tree/v3.14.0/docker-compose) by following the [standard upgrade procedure](#standard-upgrade-procedure).

## v3.12 -> v3.13

A manual migration is required. Please follow the [standard upgrade procedure](#standard-upgrade-procedure) to take down the current deployment, perform the manual migration, and then upgrade using the [`v3.13.2` tag of deploy-sourcegraph-docker](https://github.com/sourcegraph/deploy-sourcegraph-docker/tree/v3.13.2/docker-compose).

### Manual migration step: adjust file permissions

Please adjust the redis-store and redis-cache volume permissions by running the following on the host machine:

```
docker run --rm -it -v /var/lib/docker:/docker alpine:latest sh -c 'chown -R 999:1000 /docker/volumes/docker-compose_redis-store /docker/volumes/docker-compose_redis-cache'
```

### Standard upgrade procedure

In your fork of [the deploy-sourcegraph-docker](https://github.com/sourcegraph/deploy-sourcegraph-docker) repository, merge the new version into the `release` branch if you maintain any changes (see: [storing customizations in a fork](../install/docker-compose.md#optional-recommended-store-customizations-in-a-fork)):

```sh
cd docker-compose/
git fetch upstream
git merge upstream $NEW_VERSION
# Address any merge conflicts you may have.
```

Then on your server:

```sh
cd deploy-sourcegraph-docker/docker-compose/
docker-compose down
git pull
docker-compose up -d
```
