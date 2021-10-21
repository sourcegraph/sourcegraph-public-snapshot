# Operations guides for Sourcegraph with Docker Compose

Operations guides specific to managing [Sourcegraph with Docker Compose](./index.md) installations.

Trying to deploy Sourcegraph with Docker Compose? Refer to our [installation guide](./index.md#installation).

## Featured guides

<div class="getting-started">
  <a href="#configure" class="btn btn-primary" alt="Configure">
   <span>Configure</span>
   </br>
   Configure your Sourcegraph deployment with our deployment reference.
  </a>

  <a href="#upgrade" class="btn" alt="Upgrade">
   <span>Upgrade</span>
   </br>
   Upgrade your deployment to the latest Sourcegraph release.
  </a>

  <a href="#backup-and-restore" class="btn" alt="Backup and restore">
   <span>Backup and restore</span>
   </br>
   Back up your Sourcegraph instance and restore from a previous backup.
  </a>
</div>

## Deploy

Refer to our [installation guide](./index.md#installation) for more details on how to deploy Sourcegraph.

Migrating from another [deployment type](../index.md)? Refer to our [migrating to Docker Compose guides](./migrate.md).

## Configure

We **strongly** recommend that you create and run Sourcegraph from your own fork of the [reference repository](./index.md#reference-repository) to track customizations to the [Sourcegraph Docker Compose YAML](https://github.com/sourcegraph/deploy-sourcegraph-docker/blob/master/docker-compose/docker-compose.yaml). **This will make [upgrades](#upgrade) far easier.**

- [Create a fork](https://docs.github.com/en/get-started/quickstart/fork-a-repo#forking-a-repository) of the [sourcegraph/deploy-sourcegraph-docker](https://github.com/sourcegraph/deploy-sourcegraph-docker/) [reference repository](./index.md#reference-repository).

    > WARNING: Set your fork to **private** if you plan to store secrets (SSL certificates, external Postgres credentials, etc.) within the repository.

- Clone your fork using the repository's URL.

    > NOTE: The `docker-compose.yaml` file currently depends on configuration files which live in the repository, so you must have the entire repository cloned onto your server.

  ```bash
  git clone $FORK_URL
  ```

- Add the [reference repository](./index.md#reference-repository) as an `upstream` remote so that you can [get updates](#upgrade).

  ```bash
  git remote add upstream https://github.com/sourcegraph/deploy-sourcegraph-docker
  ```

- Create a `release` branch to track all of your customizations to Sourcegraph. This branch will be used to [upgrade Sourcegraph](#upgrade) and [install your Sourcegraph instance](./index.md#installation).

  ```bash
  # Specify the version you want to install
  export SOURCEGRAPH_VERSION="v3.33.0"
  # Check out the selected version for use, in a new branch called "release"
  git checkout $SOURCEGRAPH_VERSION -b release
  ```

- Make and [commit](https://git-scm.com/docs/git-commit) customizations to the [Sourcegraph Docker Compose YAML](https://github.com/sourcegraph/deploy-sourcegraph-docker/blob/master/docker-compose/docker-compose.yaml) to your `release` branch.

### Configuration best practices

- The version argument in the YAML file must be the same as in the standard deployment
- Users should only alter the YAML file to adjust resource limits, or duplicate container entries to add more container replicas

### Enable tracing

To enable [tracing](../../observability/tracing.md), add `SAMPLING_STRATEGIES_FILE=/etc/jaeger/sampling_strategies.json` to the `jaeger` container:

```yaml
jaeger:
  container_name: jaeger
  # ...
  environment:
    - 'SAMPLING_STRATEGIES_FILE=/etc/jaeger/sampling_strategies.json'
```

### Git configuration and authentication

Learn more about Git [configuration](../../repo/git_config.md) and [authentication](../../repo/auth.md).

#### SSH configuration

Provide your `gitserver` instance with your SSH / Git configuration (e.g. `.ssh/config`, `.ssh/id_rsa`, `.ssh/id_rsa.pub`, and `.ssh/known_hosts`--but you can also provide other files like `.netrc`, `.gitconfig`, etc. if needed) by mounting a directory that contains this configuration into the `gitserver` container.

For example, in the `gitserver-0` container configuration in your docker-compose.yaml file, add the second volume listed below, replacing `~/path/on/host/` with the path on the host machine to the `.ssh` directory:

```yaml
gitserver-0:
  container_name: gitserver-0
  ...
  volumes:
    - 'gitserver-0:/data/repos'
    - '~/path/on/host/.ssh:/home/sourcegraph/.ssh'
  ...
```

> WARNING: The permission of your SSH / Git configuration must be set to be readable by the user in the `gitserver` container. For example, run `chmod -v -R 600 ~/path/to/.ssh` in the folder on the host machine.

#### HTTP(S) authentication

The easiest way to specify HTTP(S) authentication for repositories is to include the username and password in the clone URL itself, such as `https://user:password@example.com/my/repo`. These credentials won't be displayed to non-admin users.

Otherwise, follow the steps above for mounting SSH configuration to mount a host directory containing the desired `.netrc` file to `/home/sourcegraph/` in the `gitserver` container.

### Expose debug port

To [generate pprof profiling data](../../pprof.md), you must [configure your deployment](#configure) to expose port 6060 on one of your frontend containers, for example:

```diff
  sourcegraph-frontend-0:
    container_name: sourcegraph-frontend-0
    # ...
+   ports:
+     - '0.0.0.0:6060:6060'
```

Also see [debug ports](../../pprof.md#debug-ports) for specific ports that can be exposed.

### Use an external database

The Docker Compose configuration has its own internal PostgreSQL and Redis databases. To preserve this data when you kill and recreate the containers, you can [use external services](../../external_services/index.md) for persistence.

### Set environment variables

Add/modify the environment variables to all of the sourcegraph-frontend-* services and the sourcegraph-frontend-internal service in the [Docker Compose YAML](https://github.com/sourcegraph/deploy-sourcegraph-docker/blob/master/docker-compose/docker-compose.yaml):

```yaml
sourcegraph-frontend-0:
  # ...
  environment:
    # ...
    - (YOUR CODE)
    # ...
```

See ["Environment variables in Compose"](https://docs.docker.com/compose/environment-variables/) for other ways to pass these environment variables to the relevant services (including from the command line, a .env file, etc.).

## Upgrade

This requires you to have [set up configuration for Docker Compose](#configure).

When you upgrade, merge the corresponding upstream release tag into your `release` branch.

```bash
# fetch updates
git fetch upstream
# merge the upstream release tag into your release branch
git checkout release
git merge v$SOURCEGRAPH_VERSION
```

Address any merge conflicts you might have.

> NOTE: If you have made no changes or only very minimal changes to your configuration, you can also ask git to always select incoming changes in the event of merge conflicts:
>
> `git merge -X theirs v$SOURCEGRAPH_VERSION`
>
> If you do this, make sure to validate your configuration is correct before proceeding.

If you are upgrading a live deployment, make sure to check the [release upgrade notes](../../updates/docker_compose.md) for any additional actions you need to take **before proceeding**.

Download all the latest docker images to your local docker daemon:

```bash
docker-compose pull --include-deps
```
Then, ensure that the current Sourcegraph instance is completely stopped:

```bash
docker-compose down --remove-orphans
```

**Once the instance has fully stopped**, you can then start Docker Compose again, now using the latest contents of the Sourcegraph configuration:

```bash
docker-compose up -d
```

You can see what's changed in the [Sourcegraph changelog](../../../CHANGELOG.md).

## Monitoring

You can monitor the health of a deployment in several ways:

- Using [Sourcegraph's built-in observability suite](../../observability/index.md), which includes dashboards and alerting for Sourcegraph services.
- Using [`docker ps`](https://docs.docker.com/engine/reference/commandline/ps/) to check on the status of containers within the deployment (any tooling designed to work with Docker containers and/or Docker Compose will work too).
  - This requires direct access to your instance's host machine.

## Manage storage

The [Sourcegraph Docker Compose definition](https://github.com/sourcegraph/deploy-sourcegraph-docker/blob/master/docker-compose/docker-compose.yaml) uses [Docker volumes](https://docs.docker.com/storage/volumes/) to store its data. These volumes are stored at `/var/lib/docker/volumes` by [default on Linux](https://docs.docker.com/storage/#choose-the-right-type-of-mount).

Guides for managing cloud storage and backups are available in our [cloud-specific installation guides](./index.md#cloud-installation):

- [Storage and backups for Amazon Web Services](./aws.md#storage-and-backups)
- [Storage and backups for Google Cloud](./google_cloud.md#storage-and-backups)
- [Storage and backups for Digital Ocean](./digitalocean.md#storage-and-backups)

## Access the database

The following command allows a user to shell into a Sourcegraph database container and run `psql` to interact with the container's postgres database:

```bash
docker exec -it pgsql psql -U sg #access pgsql container and run psql
docker exec -it codeintel-db psql -U sg #access codeintel-db container and run psql
```

## Backup and restore

The following instructions are specific to backing up and restoring the sourcegraph databases in a Docker Compose deployment. These do not apply to other deployment types.

> WARNING: **Only core data will be backed up**.
>
> These instructions will only back up core data including user accounts, configuration, repository-metadata, etc. Other data will be regenerated automatically:
>
> - Repositories will be re-cloned
> - Search indexes will be rebuilt from scratch
>
> The above may take a while if you have a lot of repositories. In the meantime, searches may be slow or return incomplete results. This process rarely takes longer than 6 hours and is usually **much** faster.

### Back up sourcegraph databases

These instructions will back up the primary `sourcegraph` database and the [codeintel](../../../code_intelligence/index.md) database.

1. `ssh` from your local machine into the machine hosting the `sourcegraph` deployment
2. `cd` to the `deploy-sourcegraph-docker/docker-compose` directory on the host
3. Verify the deployment is running:

```bash
docker-compose ps
          Name                         Command                       State                                                           Ports
-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------
caddy                       caddy run --config /etc/ca ...   Up                      2019/tcp, 0.0.0.0:443->443/tcp, 0.0.0.0:80->80/tcp
cadvisor                    /usr/bin/cadvisor -logtost ...   Up (health: starting)   8080/tcp
codeinsights-db             docker-entrypoint.sh postgres    Up                      5432/tcp
codeintel-db                /postgres.sh                     Up (healthy)            5432/tcp
github-proxy                /sbin/tini -- /usr/local/b ...   Up
gitserver-0                 /sbin/tini -- /usr/local/b ...   Up
grafana                     /entry.sh                        Up                      3000/tcp, 0.0.0.0:3370->3370/tcp
jaeger                      /go/bin/all-in-one-linux - ...   Up                      0.0.0.0:14250->14250/tcp, 14268/tcp, 0.0.0.0:16686->16686/tcp, 5775/udp, 0.0.0.0:5778->5778/tcp,
                                                                                     0.0.0.0:6831->6831/tcp, 6831/udp, 0.0.0.0:6832->6832/tcp, 6832/udp
minio                       /usr/bin/docker-entrypoint ...   Up (healthy)            9000/tcp
pgsql                       /postgres.sh                     Up (healthy)            5432/tcp
precise-code-intel-worker   /sbin/tini -- /usr/local/b ...   Up (health: starting)   3188/tcp
prometheus                  /bin/prom-wrapper                Up                      0.0.0.0:9090->9090/tcp
query-runner                /sbin/tini -- /usr/local/b ...   Up
redis-cache                 /sbin/tini -- redis-server ...   Up                      6379/tcp
redis-store                 /sbin/tini -- redis-server ...   Up                      6379/tcp
repo-updater                /sbin/tini -- /usr/local/b ...   Up
searcher-0                  /sbin/tini -- /usr/local/b ...   Up (healthy)
symbols-0                   /sbin/tini -- /usr/local/b ...   Up (healthy)            3184/tcp
syntect-server              sh -c /http-server-stabili ...   Up (healthy)            9238/tcp
worker                      /sbin/tini -- /usr/local/b ...   Up                      3189/tcp
zoekt-indexserver-0         /sbin/tini -- zoekt-source ...   Up
zoekt-webserver-0           /sbin/tini -- /bin/sh -c z ...   Up (healthy)
```
4. Stop the deployment, and restart the databases service only to ensure there are no other connections during backup and restore.

```bash
docker-compose down
docker-compose -f db-only-migrate.docker-compose.yaml up -d
```

5. Generate the database dumps

```bash
docker exec pgsql sh -c 'pg_dump -C --username sg sg' > sourcegraph_db.out
docker exec codeintel-db -c 'pg_dump -C --username sg sg' > codeintel_db.out
```

6. Ensure the `sourcegraph_db.out` and `codeintel_db.out` files are moved to a safe and secure location. 

### Restore sourcegraph databases

#### Restoring sourcegraph databases into a new environment

The following instructions apply only if you are restoring your databases into a new deployment of sourcegraph ie: a new virtual machine 

If you are restoring a previously running environment, see the instructions for [restoring a previously running deployment](#restoring-sourcegraph-databases-into-an-existing-environment)

1. Copy the database dump files into the `deploy-sourcegraph-docker/docker-compose` directory. 
2. Start the database services

```bash
docker-compose -f db-only-migrate.docker-compose.yaml up -d
```

3. Copy the database files into the containers

```bash
docker cp sourcegraph_db.out pgsql:/tmp/sourecgraph_db.out
docker cp codeintel_db.out codeintel-db:/tmp/codeintel_db.out
```

4. Restore the databases

```bash
docker exec pgsql sh -c 'psql -v ERROR_ON_STOP=1 --username sg -f /tmp/sourcegraph_db.out sg'
docker exec codeintel-db sh -c 'psql -v ERROR_ON_STOP=1 --username sg -f /tmp/condeintel_db.out sg'
```

5. Start the remaining sourcegraph services

```bash
docker-compose up -d
```

6. Verify the deployment has started 

```bash 
docker-compose ps
          Name                         Command                       State                                                           Ports
-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------
caddy                       caddy run --config /etc/ca ...   Up                      2019/tcp, 0.0.0.0:443->443/tcp, 0.0.0.0:80->80/tcp
cadvisor                    /usr/bin/cadvisor -logtost ...   Up (health: starting)   8080/tcp
codeinsights-db             docker-entrypoint.sh postgres    Up                      5432/tcp
codeintel-db                /postgres.sh                     Up (healthy)            5432/tcp
github-proxy                /sbin/tini -- /usr/local/b ...   Up
gitserver-0                 /sbin/tini -- /usr/local/b ...   Up
grafana                     /entry.sh                        Up                      3000/tcp, 0.0.0.0:3370->3370/tcp
jaeger                      /go/bin/all-in-one-linux - ...   Up                      0.0.0.0:14250->14250/tcp, 14268/tcp, 0.0.0.0:16686->16686/tcp, 5775/udp, 0.0.0.0:5778->5778/tcp,
                                                                                     0.0.0.0:6831->6831/tcp, 6831/udp, 0.0.0.0:6832->6832/tcp, 6832/udp
minio                       /usr/bin/docker-entrypoint ...   Up (healthy)            9000/tcp
pgsql                       /postgres.sh                     Up (healthy)            5432/tcp
precise-code-intel-worker   /sbin/tini -- /usr/local/b ...   Up (health: starting)   3188/tcp
prometheus                  /bin/prom-wrapper                Up                      0.0.0.0:9090->9090/tcp
query-runner                /sbin/tini -- /usr/local/b ...   Up
redis-cache                 /sbin/tini -- redis-server ...   Up                      6379/tcp
redis-store                 /sbin/tini -- redis-server ...   Up                      6379/tcp
repo-updater                /sbin/tini -- /usr/local/b ...   Up
searcher-0                  /sbin/tini -- /usr/local/b ...   Up (healthy)
symbols-0                   /sbin/tini -- /usr/local/b ...   Up (healthy)            3184/tcp
syntect-server              sh -c /http-server-stabili ...   Up (healthy)            9238/tcp
worker                      /sbin/tini -- /usr/local/b ...   Up                      3189/tcp
zoekt-indexserver-0         /sbin/tini -- zoekt-source ...   Up
zoekt-webserver-0           /sbin/tini -- /bin/sh -c z ...   Up (healthy)> docker-compose ps
```

7. Browse to your sourcegraph deployment, login and verify your existing configuration has been restored


#### Restoring sourcegraph databases into an existing environment

1. `cd` to the `deploy-sourcegraph-docker/docker-compose` and stop the previous deployment and remove any existing volumes
```bash
docker-compose down
docker volume rm docker-compose_pgsql
docker volume rm docker-compose_codeintel-db
```

2. Start the databases services only
```bash
docker-compose -f db-only-migrate.docker-compose.yaml up -d
```

3. Copy the database files into the containers

```bash
docker cp sourcegraph_db.out pgsql:/tmp/sourecgraph_db.out
docker cp codeintel_db.out codeintel-db:/tmp/codeintel_db.out
```

4. Restore the databases

```bash
docker exec pgsql sh -c 'psql -v ERROR_ON_STOP=1 --username sg -f /tmp/sourcegraph_db.out sg'
docker exec codeintel-db sh -c 'psql -v ERROR_ON_STOP=1 --username sg -f /tmp/condeintel_db.out sg'
```

5. Start the remaining sourcegraph services

```bash
docker-compose up -d
```

6. Verify the deployment has started 

```bash 
docker-compose ps
          Name                         Command                       State                                                           Ports
-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------
caddy                       caddy run --config /etc/ca ...   Up                      2019/tcp, 0.0.0.0:443->443/tcp, 0.0.0.0:80->80/tcp
cadvisor                    /usr/bin/cadvisor -logtost ...   Up (health: starting)   8080/tcp
codeinsights-db             docker-entrypoint.sh postgres    Up                      5432/tcp
codeintel-db                /postgres.sh                     Up (healthy)            5432/tcp
github-proxy                /sbin/tini -- /usr/local/b ...   Up
gitserver-0                 /sbin/tini -- /usr/local/b ...   Up
grafana                     /entry.sh                        Up                      3000/tcp, 0.0.0.0:3370->3370/tcp
jaeger                      /go/bin/all-in-one-linux - ...   Up                      0.0.0.0:14250->14250/tcp, 14268/tcp, 0.0.0.0:16686->16686/tcp, 5775/udp, 0.0.0.0:5778->5778/tcp,
                                                                                     0.0.0.0:6831->6831/tcp, 6831/udp, 0.0.0.0:6832->6832/tcp, 6832/udp
minio                       /usr/bin/docker-entrypoint ...   Up (healthy)            9000/tcp
pgsql                       /postgres.sh                     Up (healthy)            5432/tcp
precise-code-intel-worker   /sbin/tini -- /usr/local/b ...   Up (health: starting)   3188/tcp
prometheus                  /bin/prom-wrapper                Up                      0.0.0.0:9090->9090/tcp
query-runner                /sbin/tini -- /usr/local/b ...   Up
redis-cache                 /sbin/tini -- redis-server ...   Up                      6379/tcp
redis-store                 /sbin/tini -- redis-server ...   Up                      6379/tcp
repo-updater                /sbin/tini -- /usr/local/b ...   Up
searcher-0                  /sbin/tini -- /usr/local/b ...   Up (healthy)
symbols-0                   /sbin/tini -- /usr/local/b ...   Up (healthy)            3184/tcp
syntect-server              sh -c /http-server-stabili ...   Up (healthy)            9238/tcp
worker                      /sbin/tini -- /usr/local/b ...   Up                      3189/tcp
zoekt-indexserver-0         /sbin/tini -- zoekt-source ...   Up
zoekt-webserver-0           /sbin/tini -- /bin/sh -c z ...   Up (healthy)> docker-compose ps
```

7. Browse to your sourcegraph deployment, login and verify your existing configuration has been restored
