# Migrate from the single Docker image to Docker Compose

Sourcegraph's user data can be migrated from the single Docker image (`sourcegraph/server`) to the Docker Compose deployment by dumping and restoring the Postgres database.

## Notes

### Version requirements

* This migration can only be done with Sourcegraph `3.13.0` and above (e.g. `sourcegraph/server:3.13.0` and [v3.13.0 (TODO FILL IN RELEASE HERE) Docker Compose](TODO) ).
* Sourcegraph's user data can only be transferred between deployments that are running the same Sourcegraph verion (e.g. `sourcegraph/server:3.13.0` can only transfer its data to `v3.13.0` of the Docker Compose definition). If you're running a version of Sourcegraph server that's older than the Docker Compose deployment version, you **must** upgrade to a newer `sourcegraph/server` version before continuing.

### Storage location

Note that after this process, Sourcegraph's data will be stored in Docker volumes instead of `~/.sourcegraph/`. For more information, see the cloud-provider documentation referred to in [Create the new Docker Compose instance](#create-the-new-docker-compose-instance).

### Only user data will be migrated

While this process will migrate your user data, the new Docker Compose deployment will need to regenerate all the other ephemeral data:

* repositories will need to be re-cloned
* search indexes will need to be recreated
* etc.

## Backup Postgres database

* `ssh` into the instance running the `sourcegraph/server` container
* find the `CONTAINER_ID` of the `sourcegraph/server` image from the `docker ps` output:

```bash
> docker ps
CONTAINER ID        IMAGE
...                 sourcegraph/server
```

* Generate Postgres dump inside `sourcegraph/server` container:

```bash
# Open a shell inside sourcegraph/server using the CONTAINER_ID found in the previous step
> docker exec -it "$CONTAINER_ID" /bin/sh

# Dump Postgres database to db.out file 
> pg_dumpall --verbose --username=postgres > /tmp/db.out

# Exit container shell session
> exit
```

* Copy Postgres dump from the `sourcegraph/server` container onto the host machine:

```bash
> docker cp "$CONTAINER_ID":/tmp/db.out ~/db.out

# You can run "less ~/db.out" to verify that it has the contents that you expect
```

* TODO: Copy `~db.out` to local laptop since you'll be spinning down the `sourcegraph/server` machine
  * `scp`? `gcloud compute scp`?


## Create the new Docker Compose instance

Follow the installation guide for your cloud provider to create the new Docker Compose instance:

* [Install Sourcegraph with Docker Compose on AWS](../../install/docker-compose/aws.md)
* [Install Sourcegraph with Docker Compose on Google Cloud](../../install/docker-compose/google_cloud.md)
* [Install Sourcegraph with Docker Compose on DigitalOcean](../../install/docker-compose/digitalocean.md)

## Bring up the Postgres database on its own

```bash
> cd "$DEPLOY_SOURCEGRAPH_DOCKER_CHECKOUT"/docker-compose # refer to cloud provider script for DEPLOY_SOURCEGRAPH_DOCKER_CHECKOUT value

# Tear down existing instance (including volumes) so that we don't encounter conflicting transactions
> docker-compose down --volumes

# Bring up the Postgres database on its own
> docker-compose -f pgsql-only-migrate.docker-compose.yaml up -d
```

## TODO: copy database dump from local laptop to Docker Compose instance

## Restore database dump

```bash

# Copy database dump from host to Postgres container
> docker cp ~/db.out pgsql:/tmp/db.out

# Open up a shell session inside the Postgres container
> docker exec -it pgsql /bin/sh

# Restore the database dump
> psql --username=sg -f /tmp/db.out postgres

# Open up a psql session inside the Postgres container
> psql --username=sg

# Apply tweaks to transform sourcegraph/server's DB schema into Docker Compose's
> DROP DATABASE sg;
> ALTER DATABASE sourcegraph RENAME TO sg;
> ALTER DATABASE sg OWNER TO sg;
```

## Start the rest of the sourcegraph containers

```bash
> docker-compose -f docker-compose.yaml up -d
```

The migration process is now complete. You should be able to log into your instance and verify that all your users and configuration are still present. It is now safe to tear down the `sourcegraph/server` instance.
