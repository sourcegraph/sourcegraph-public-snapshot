# Updating a Docker Compose Sourcegraph instance

This document describes the exact changes needed to update a Docker Compose Sourcegraph instance.

Each section comprehensively describes the steps needed to upgrade, and any manual migration steps you must perform.

## v3.12 -> v3.13

### Manual migration step: adjust file permissions

Please adjust the redis-store and redis-cache volume permissions by running the following on the host machine:

```
docker run --rm -it -v /var/lib/docker:/docker alpine:latest sh -c 'chown -R 999:1000 /docker/volumes/docker-compose_redis-store /docker/volumes/docker-compose_redis-cache'
```

### Upgrade

In your fork of [the deploy-sourcegraph-docker](https://github.com/sourcegraph/deploy-sourcegraph-docker) repository:

```sh
cd docker-compose/
git fetch upstream
git merge upstream v3.13.2
# Address any merge conflicts you may have.
```

Then on your server:

```sh
cd deploy-sourcegraph-docker/docker-compose/
docker-compose down
git pull
docker-compose up -d
```
