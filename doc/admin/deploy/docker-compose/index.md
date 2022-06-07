# Sourcegraph with Docker Compose

Docker Compose is a tool for defining and running multi-[container](https://www.docker.com/resources/what-container) Docker applications (in this case, Sourcegraph!). With Docker Compose, you use a YAML file to configure your application’s services. Then, with a single command, you create and start all the services from your configuration. To learn more about Docker Compose, head over to [Docker's Docker Compose docs](https://docs.docker.com/compose/).

For Sourcegraph customers who want a simplified single-machine deployment of Sourcegraph with easy configuration and low cost of effort to maintain, Sourcegraph with Docker Compose is an ideal choice.

Not sure if the Docker Compose deployment type is the right for you? Learn more about the various Sourcegraph deployment types in our [Deployment overview section](../index.md).

The rest of this page provides generic instruction for how to install Sourcegraph with Docker Compose on a server, which could be the local machine, a server on a local network, or cloud-hosted server. Follow these links for cloud-specific Sourcegraph installation guides:

- [Deploy Sourcegraph with Docker Compose on Amazon Web Services](../../deploy/docker-compose/aws.md)
- [Deploy Sourcegraph with Docker Compose on Google Cloud](../../deploy/docker-compose/google_cloud.md)
- [Deploy Sourcegraph with Docker Compose on DigitalOcean](../../deploy/docker-compose/digitalocean.md)

## Prerequisites
Complete the following tasks before deploying Sourcegraph with Docker Compose:

- Provision an appropriately configured server to host your Sourcegraph containers. Running Sourcegraph on Windows or ARM / ARM64 images is *not* supported for production deployments.
    - Determine the number and size of the repos in your environment.
    - Determine the number of users and their engagement rate with the repos. 
    - Configure the server resources using the [resource estimator](../resource_estimator.md) to ensure it has sufficient CPUs, memory, and SSD capacity.

>Note: Sourcegraph requires SSD backed storage. 

- Configure ingress firewall rules to enable secure access to the server.
- Configure access for the server to your deployment files, in the examples that follow a Personal Access Token for GitHub is required.   
- Install [Docker Compose](https://docs.docker.com/compose/) on the server. Sourcegraph deployments should *not* be deployed with Docker Swarm

>Note: Minimum Docker [v20.10.0](https://docs.docker.com/engine/release-notes/#20100) and Docker Compose [v1.29.0](https://docs.docker.com/compose/release-notes/#1290)

- Obtain a [Sourcegraph license](https://about.sourcegraph.com/pricing/). You can run through these instructions without one, but you must obtain a license for instances of more than 10 users.

## Installation Process Overview


 - [Fork the Sourcegraph Reference Repository](#fork-the-sourcegraph-reference-repository). 
 - [Clone your fork of the reference repository locally](#clone-your-fork).
 - [Create a release branch on your clone](#configure-a-release-branch). 
 - [Customize the Docker-Compose yaml file to reflect your environment variables](#make-yaml-customizations). 
 - [Publish changes to your release branch](#publish-your-changes-to-the-release-branch)
 - [Clone your release branch onto your server](#clone-your-release-branch-onto-your-server)
 - [Build and start the containers in detached mode](#run-sourcegraph). 
 - [Perform initial management operations](#management-operations)

The following sections describe each step in more detail. 

## Fork the Sourcegraph Reference Repository
The [`sourcegraph/deploy-sourcegraph-docker`](https://github.com/sourcegraph/deploy-sourcegraph-docker/) reference repository contains everything you need to install and configure a Docker Compose Sourcegraph instance, and it will make upgrades far easier. We **strongly** recommend that you create and run Sourcegraph from your own fork of the reference repository to track customizations to the [Sourcegraph Docker Compose YAML](https://github.com/sourcegraph/deploy-sourcegraph-docker/blob/master/docker-compose/docker-compose.yaml). 

>NOTE: The following steps use GitHub as an example. You can fork the reference repo into your chosen code host. 

> WARNING: In GitHub, forks of public repos are also public. Create a private fork if you plan to store secrets (SSL certificates, external Postgres credentials, etc.) within the repository. However, a preferable approach would be to use a Secrets Management Service. 

1\. Use the GithHub GUI to [Create a fork](https://docs.github.com/en/get-started/quickstart/fork-a-repo#forking-a-repository) of the [sourcegraph/deploy-sourcegraph-docker](https://github.com/sourcegraph/deploy-sourcegraph-docker/) reference repository.

*Alternatively*, if you are using GitHub and you want your fork to be private, create a private clone of the reference repository in your code host. This process can be performed on any machine with access to your code host:

1\. Create an [empty private repository](https://docs.github.com/en/repositories/creating-and-managing-repositories/creating-a-new-repository), for example `<you/private-repository>` in GitHub.

2\. Bare clone the reference repository. 

```bash
  git clone --bare https://github.com/sourcegraph/deploy-sourcegraph-docker/
```

3\. Navigate to the bare clone and mirror push it to your private repository.

```bash
  cd deploy-sourcegraph-docker.git
  git push --mirror https://github.com/<you/private-repository>.git
```

4\. Remove your local bare clone. 

```bash
  cd ..
  rm -rf deploy-sourcegraph-docker.git
```

## Clone your fork

Clone your fork of the reference repo to your local machine. In this example, you clone the GitHub repo you created earlier: 

```bash
  git clone https://github.com/<you/private-repository>.git 
```

## Configure a release branch

Add the reference repository as an `upstream` remote so that you can get updates.

```bash
  git remote add upstream https://github.com/sourcegraph/deploy-sourcegraph-docker
```

Create a `release` branch to track all of your customizations to Sourcegraph. This branch will be used to [upgrade Sourcegraph](#upgrade-and-migration) and install your Sourcegraph instance.

```bash
  # Specify the version you want to install
  export SOURCEGRAPH_VERSION="v3.40.1"
  # Check out the selected version for use, in a new branch called "release"
  git checkout $SOURCEGRAPH_VERSION -b release
```

## Make YAML customizations

Make customizations to the Docker Compose YAML as required for your environment and commit the changes to your `release` branch. The following section represents a number of key configuration items for your deployment. For more detailed configuration, see Sourcegraph's [configuration](../../config/index.md) docs.

### Enable tracing
Check that tracing is enabled in the docker-compose.yaml file. The environment variable should be set to `SAMPLING_STRATEGIES_FILE=/etc/jaeger/sampling_strategies.json` in the `jaeger` container section:

```yaml
jaeger:
  container_name: jaeger
  # ...
  environment:
    - 'SAMPLING_STRATEGIES_FILE=/etc/jaeger/sampling_strategies.json'
```
### Git configuration
#### Git SSH configuration
Provide your `gitserver` instance with your SSH / Git configuration (e.g. `.ssh/config`, `.ssh/id_rsa`, `.ssh/id_rsa.pub`, and `.ssh/known_hosts`--but you can also provide other files like `.netrc`, `.gitconfig`, etc. if needed) by mounting a directory that contains this configuration into the `gitserver` container.

For example, in the `gitserver-0` container configuration in your docker-compose.yaml file, add the volume listed in the following example, replacing `~/path/on/host/` with the path on the host machine to the `.ssh` directory:

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

#### Git HTTP(S) authentication

The easiest way to specify HTTP(S) authentication for repositories is to include the username and password in the clone URL itself, such as `https://user:password@example.com/my/repo`. These credentials won't be displayed to non-admin users.

Otherwise, follow the previous steps for mounting SSH configuration to mount a host directory containing the desired `.netrc` file to `/home/sourcegraph/` in the `gitserver` container.

### Expose debug port

To [generate pprof profiling data](../../pprof.md), you must configure your deployment to expose port 6060 on one of your frontend containers, for example:

```diff
  sourcegraph-frontend-0:
    container_name: sourcegraph-frontend-0
    # ...
+   ports:
+     - '0.0.0.0:6060:6060'
```

For specific ports that can be exposed, see the [debug ports section](../../pprof.md#debug-ports) of Sourcegraphs's [generate pprof profiling data](../../pprof.md) docs.

### Use an external database

The Docker Compose configuration has its own internal PostgreSQL and Redis databases. 

To preserve this data when you kill and recreate the containers, review Sourcegraph's External Services for additional information on how you can [use external services](../../external_services/index.md) for persistence.

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

## Publish your changes to the release branch
Publish the customized configuration to the release branch you created earlier:

```bash
  git add .
  git commit -m "customize docker-compose.yaml for environment"
  git push origin release
```

## Clone your release branch onto your server

Now that you have published your changes to your code host you can deploy your customized codebase on the production server. Clone your release branch onto the production server. In this example, you SSH into the server and clone the branch you created earlier: 



  ```bash
  git clone --branch release https://github.com/<you/private-repository>.git 
  ```

## Run Sourcegraph
> NOTE: The `docker-compose.yaml` file currently depends on configuration files which live in the repository, so you must have the entire repository cloned onto your server.

On the production server, move to the configuration directory and run Sourcegraph:

```bash
  cd docker-compose
  docker-compose up -d
```

Once the server is ready (the `sourcegraph-frontend-0` service is healthy when running `docker ps`), navigate to the `sourcegraph-frontend-0` hostname or IP address on port `80`.  

## Management Operations

### Manage storage

The Sourcegraph Docker Compose yaml file uses [Docker volumes](https://docs.docker.com/storage/volumes/) to store its data. These volumes are stored at `/var/lib/docker/volumes` by [default on Linux](https://docs.docker.com/storage/#choose-the-right-type-of-mount).

Guides for managing cloud storage and backups are available in our [cloud-specific installation guides](./index.md#cloud-installation):

- [Storage and backups for Amazon Web Services](./aws.md#storage-and-backups)
- [Storage and backups for Google Cloud](./google_cloud.md#storage-and-backups)
- [Storage and backups for Digital Ocean](./digitalocean.md#storage-and-backups)

### Access the database

The following command allows a user to shell into the Sourcegraph database container and run `psql` to interact with the container's postgres database:

```bash
docker exec -it pgsql psql -U sg #access pgsql container and run psql
docker exec -it codeintel-db psql -U sg #access codeintel-db container and run psql
```
### Database Migrations

> NOTE: The `migrator` service is only available in versions `3.37` and later.

The `frontend` container in the `docker-compose.yaml` file will automatically run on startup and migrate the databases if any changes are required, however administrators may wish to migrate their databases before upgrading the rest of the system when working with large databases. Sourcegraph guarantees database backward compatibility to the most recent minor point release so the database can safely be upgraded before the application code.

To execute the database migrations independently, follow the [docker-compose instructions on how to manually run database migrations](../../how-to/manual_database_migrations.md#docker-compose). Running the `up` (default) command on the `migrator` of the *version you are upgrading to* will apply all migrations required by the next version of Sourcegraph.

### Backup and restore

The following instructions are specific to backing up and restoring the Sourcegraph databases in a Docker Compose deployment. These do not apply to other deployment types.

> WARNING: **Only core data will be backed up**.
>
> These instructions will only back up core data including user accounts, configuration, repository-metadata, etc. Other data will be regenerated automatically:
>
> - Repositories will be re-cloned
> - Search indexes will be rebuilt from scratch
>
> The above may take a while if you have a lot of repositories. In the meantime, searches may be slow or return incomplete results. This process rarely takes longer than 6 hours and is usually **much** faster.

#### Back up Sourcegraph databases

These instructions will back up the primary `sourcegraph` database and the [codeintel](../../../code_intelligence/index.md) database.

1\. `ssh` from your local machine into the machine hosting the `sourcegraph` deployment

2\. `cd` to the `deploy-sourcegraph-docker/docker-compose` directory on the host

3\. Verify the deployment is running:

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
4\. Stop the deployment, and restart the databases service only to ensure there are no other connections during backup and restore.

```bash
docker-compose down
docker-compose -f db-only-migrate.docker-compose.yaml up -d
```

5\. Generate the database dumps

```bash
docker exec pgsql sh -c 'pg_dump -C --username sg sg' > sourcegraph_db.out
docker exec codeintel-db -c 'pg_dump -C --username sg sg' > codeintel_db.out
```

6\. Ensure the `sourcegraph_db.out` and `codeintel_db.out` files are moved to a safe and secure location. 

#### Restore Sourcegraph databases into a new environment

The following instructions apply **only if you are restoring your databases into a new deployment** of Sourcegraph ie: a new virtual machine. If you are restoring a previously running environment, see the instructions for [restoring a previously running deployment](#restoring-sourcegraph-databases-into-an-existing-environment)

1\. Copy the database dump files into the `deploy-sourcegraph-docker/docker-compose` directory. 

2\. Start the database services

```bash
docker-compose -f db-only-migrate.docker-compose.yaml up -d
```

3\. Copy the database files into the containers

```bash
docker cp sourcegraph_db.out pgsql:/tmp/sourecgraph_db.out
docker cp codeintel_db.out codeintel-db:/tmp/codeintel_db.out
```

4\. Restore the databases

```bash
docker exec pgsql sh -c 'psql -v ERROR_ON_STOP=1 --username sg -f /tmp/sourcegraph_db.out sg'
docker exec codeintel-db sh -c 'psql -v ERROR_ON_STOP=1 --username sg -f /tmp/condeintel_db.out sg'
```

5\. Start the remaining Sourcegraph services

```bash
docker-compose up -d
```

6\. Verify the deployment has started 

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

7\. Browse to your Sourcegraph deployment, login and verify your existing configuration has been restored

### Restore Sourcegraph databases into an existing environment

1\. `cd` to the `deploy-sourcegraph-docker/docker-compose` and stop the previous deployment and remove any existing volumes
```bash
docker-compose down
docker volume rm docker-compose_pgsql
docker volume rm docker-compose_codeintel-db
```

2\. Start the databases services only

```bash
docker-compose -f db-only-migrate.docker-compose.yaml up -d
```

3\. Copy the database files into the containers

```bash
docker cp sourcegraph_db.out pgsql:/tmp/sourecgraph_db.out
docker cp codeintel_db.out codeintel-db:/tmp/codeintel_db.out
```

4\. Restore the databases

```bash
docker exec pgsql sh -c 'psql -v ERROR_ON_STOP=1 --username sg -f /tmp/sourcegraph_db.out sg'
docker exec codeintel-db sh -c 'psql -v ERROR_ON_STOP=1 --username sg -f /tmp/condeintel_db.out sg'
```

5\. Start the remaining Sourcegraph services

```bash
docker-compose up -d
```

6\. Verify the deployment has started 

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

7\. Browse to your Sourcegraph deployment, login and verify your existing configuration has been restored

### Monitoring

You can monitor the health of a deployment in several ways:

- Using [Sourcegraph's built-in observability suite](../../observability/index.md), which includes dashboards and alerting for Sourcegraph services.
- Using [`docker ps`](https://docs.docker.com/engine/reference/commandline/ps/) to check on the status of containers within the deployment (any tooling designed to work with Docker containers and/or Docker Compose will work too).
  - This requires direct access to your instance's host machine.

### Upgrade

If you [configured Docker Compose with a release branch](#configure-release-branch), when you upgrade you can merge the corresponding upstream release tag into your `release` branch.

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
