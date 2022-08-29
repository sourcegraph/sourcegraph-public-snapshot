# Sourcegraph with Docker Compose

Docker Compose is a tool for defining and running [multi-container](https://www.docker.com/resources/what-container) Docker applications (in this case, Sourcegraph!). With Docker Compose, you use a YAML file to configure your application’s services. Then, with a single command, you create and start all the services from your configuration. To learn more about Docker Compose, head over to [Docker's Docker Compose docs](https://docs.docker.com/compose/).

For Sourcegraph customers who want a simplified single-machine deployment of Sourcegraph with easy configuration and low cost of effort to maintain, Sourcegraph with Docker Compose is an ideal choice.

Not sure if the Docker Compose deployment type is the right for you? Learn more about the various Sourcegraph deployment types in our [Deployment overview section](../index.md).

## Installing Sourcegraph on Docker Compose

This section provides instruction for how to install Sourcegraph with Docker Compose on a server, which could be the local machine, a server on a local network, or cloud-hosted server. 

Alternatively, follow these links for cloud-specific guides on preparing the environment and installing Sourcegraph:

- [Deploy Sourcegraph with Docker Compose on Amazon Web Services](../../deploy/docker-compose/aws.md)
- [Deploy Sourcegraph with Docker Compose on Google Cloud](../../deploy/docker-compose/google_cloud.md)
- [Deploy Sourcegraph with Docker Compose on DigitalOcean](../../deploy/docker-compose/digitalocean.md)

## Prerequisites

  - Use the [resource estimator](../resource_estimator.md) to configure a server to host your Sourcegraph containers
  - Configure ingress firewall rules to enable secure access to the server
  - Configure access for the server to your deployment files
    - We will be using a Personal Access Token from GitHub in this example
  - Install [Docker Compose](https://docs.docker.com/compose/) on the server 
    - Minimum Docker [v20.10.0](https://docs.docker.com/engine/release-notes/#20100) and Docker Compose [v1.29.0](https://docs.docker.com/compose/release-notes/#1290)
  - Obtain a [Sourcegraph license](https://about.sourcegraph.com/pricing/)
    - You can run through these instructions without one, but you must obtain a license for instances of more than 10 users.

>WARNING: Running Sourcegraph on Windows or ARM / ARM64 images is not supported for production deployments.

## Installation

A step by step guide to install Sourcegraph with Docker Compose.

>WARNING: Docker Swarm mode is not supported.

### Overview

 1. [Fork the Sourcegraph Docker Compose deployment repository](#step-1-fork-the-sourcegraph-docker-compose-deployment-repository)
 2. [Clone your fork of the reference repository locally](#step-2-clone-the-forked-repository-locally)
 3. [Create a release branch on your clone](#step-3-configure-the-release-branch)
 4. [Customize the Docker-Compose YAML file](#step-4-configure-the-yaml-file)
 5. [Publish changes to your release branch](#step-5-update-your-release-branch)
 6. [Clone your release branch onto your server](#step-6-clone-the-release-branch-remotely)
 7. [Build and start the containers in detached mode](#step-7-start-sourcegraph)

### Step 1: Fork the Sourcegraph Docker Compose deployment repository

>NOTE: The following steps use GitHub as an example. You can fork the [reference repository](https://github.com/sourcegraph/deploy-sourcegraph-docker/) into your chosen code host. 

The [`sourcegraph/deploy-sourcegraph-docker`](https://github.com/sourcegraph/deploy-sourcegraph-docker/) repository contains everything you need to install and configure a Sourcegraph Docker Compose instance, and it will make upgrades far easier. We **strongly** recommend that you create and run Sourcegraph from your own fork of the reference repository to track customizations made to the [Sourcegraph Docker Compose YAML file](https://github.com/sourcegraph/deploy-sourcegraph-docker/blob/master/docker-compose/docker-compose.yaml). 

> WARNING: In GitHub, forks of public repositories are also public. Create a private fork if you plan to store secrets (SSL certificates, external Postgres credentials, etc.) within the repository. However, a preferable approach would be to use a Secrets Management Service. 

1\. Use the GithHub GUI to [Create a fork](https://docs.github.com/en/get-started/quickstart/fork-a-repo#forking-a-repository) of the [sourcegraph/deploy-sourcegraph-docker](https://github.com/sourcegraph/deploy-sourcegraph-docker/) reference repository.

*Alternatively*, if you are using GitHub and you want your fork to be private, create a private clone of the reference repository in your code host. This process can be performed on any machine with access to your code host:

2\. Create an [empty private repository](https://docs.github.com/en/repositories/creating-and-managing-repositories/creating-a-new-repository), for example `<you/private-repository>` in GitHub.

3\. Bare clone the reference repository. 

```bash
  git clone --bare https://github.com/sourcegraph/deploy-sourcegraph-docker/
```

4\. Navigate to the bare clone and mirror push it to your private repository.

```bash
  cd deploy-sourcegraph-docker.git
  git push --mirror https://github.com/<you/private-repository>.git
```

5\. Remove your local bare clone. 

```bash
  cd ..
  rm -rf deploy-sourcegraph-docker.git
```

### Step 2: Clone the forked repository locally

Clone the forked repository to your local machine. 

```bash
  git clone https://github.com/<you/private-repository>.git 
```

### Step 3: Configure the release branch

1. Add the reference repository as the remote `upstream` in order to keep your fork synced with the upstream repository.

```bash
  git remote add upstream https://github.com/sourcegraph/deploy-sourcegraph-docker
```

2. Create a `release` branch to track all of your customizations to Sourcegraph. This branch will be used to [upgrade Sourcegraph](#upgrade-and-migration) and install your Sourcegraph instance.

```bash
  # Specify the version you want to install
  export SOURCEGRAPH_VERSION="v3.43.1"
  # Check out the selected version for use, in a new branch called "release"
  git checkout $SOURCEGRAPH_VERSION -b release
```

### Step 4: Configure the YAML file

The reference repository includes a docker-compose.yaml file with a basic configuration. Adjust the service resources for your environment using the [resource estimator](../resource_estimator.md) then commit the changes to your `release` branch. The following section represents a number of key configuration items for your deployment. 

>NOTE: For configuration of Sourcegraph, see Sourcegraph's [configuration](../../config/index.md) docs.

#### Enable tracing
Check that tracing is enabled in the docker-compose.yaml file. The environment variable should be set to `SAMPLING_STRATEGIES_FILE=/etc/jaeger/sampling_strategies.json` in the `jaeger` container section:

```yaml
jaeger:
  container_name: jaeger
  # ...
  environment:
    - 'SAMPLING_STRATEGIES_FILE=/etc/jaeger/sampling_strategies.json'
```

#### Git configuration

##### Git SSH configuration

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

##### Git HTTP(S) authentication

The easiest way to specify HTTP(S) authentication for repositories is to include the username and password in the clone URL itself, such as `https://user:password@example.com/my/repo`. These credentials won't be displayed to non-admin users.

Otherwise, follow the previous steps for mounting SSH configuration to mount a host directory containing the desired `.netrc` file to `/home/sourcegraph/` in the `gitserver` container.

#### Expose debug port

To [generate pprof profiling data](../../pprof.md), you must configure your deployment to expose port 6060 on one of your frontend containers, for example:

```diff
  sourcegraph-frontend-0:
    container_name: sourcegraph-frontend-0
    # ...
+   ports:
+     - '0.0.0.0:6060:6060'
```

For specific ports that can be exposed, see the [debug ports section](../../pprof.md#debug-ports) of Sourcegraphs's [generate pprof profiling data](../../pprof.md) docs.

#### Use an external database

The Docker Compose configuration has its own internal PostgreSQL and Redis databases. 

To preserve this data when you kill and recreate the containers, review Sourcegraph's External Services for additional information on how you can [use external services](../../external_services/index.md) for persistence.

#### Set environment variables

Add/modify the environment variables to all of the sourcegraph-frontend-* services and the sourcegraph-frontend-internal service in the [Docker Compose YAML file](https://github.com/sourcegraph/deploy-sourcegraph-docker/blob/master/docker-compose/docker-compose.yaml):

```yaml
sourcegraph-frontend-0:
  # ...
  environment:
    # ...
    - (YOUR CODE)
    # ...
```

See ["Environment variables in Compose"](https://docs.docker.com/compose/environment-variables/) for other ways to pass these environment variables to the relevant services (including from the command line, a .env file, etc.).

### Step 5: Update your release branch

Publish the customized configuration to the release branch you created earlier:

```bash
  git add .
  git commit -m "customize docker-compose.yaml for environment"
  git push origin release
```

### Step 6: Clone the release branch remotely

Now that you have published your changes to your code host you deploy your customized codebase on the production server. Clone the newly configured release branch onto the production server. 

> NOTE: The `docker-compose.yaml` file currently depends on configuration files which live in the repository, so you must have the entire repository cloned onto your server.

Clone the `release` branch you've created earlier onto your server: 

```bash
  git clone --branch release https://github.com/<you/private-repository>.git 
```

### Step 7: Start Sourcegraph

On the production server, run the following command inside the [docker-compose configuration directory](https://github.com/sourcegraph/deploy-sourcegraph-docker/tree/master/docker-compose) to start Sourcegraph in a detached mode:

```bash
  # Go to the docker-compose configuration directory
  cd docker-compose
  # Start Sourcegraph with Docker Compose in a detached mode
  docker-compose up -d
```

To check if the server is ready, the `sourcegraph-frontend-0` service must be displayed as healthy:

```bash
  # Check the health status for sourcegraph-frontend-0
  docker ps --filter="name=sourcegraph-frontend-0"
```

Once the server is ready, navigate to the `sourcegraph-frontend-0` hostname or IP address on port `80`.  

[<p style="text-align:right;">Next: Management Operations ></p>](operations.md)
