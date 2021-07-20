# Sourcegraph with Docker Compose

<p class="lead">
Sourcegraph with Docker Compose is an ideal choice for small to medium sized deployments.
</p>

Not sure if Docker Compose is the right choice for you? Learn more about the various [Sourcegraph installation options](../index.md).

<div class="cta-group">
<a class="btn btn-primary" href="#installation">★ Installation</a>
<a class="btn" href="operations">Operations guides</a>
<a class="btn" href="../../../#get-help">Get help</a>
</div>

## Installation

Before you get started, we recommend [learning about how Sourcegraph with Docker Compose works](#about).

### Cloud installation

Deploy Sourcegraph with Docker Compose to a cloud of your choice.

You will need:

- A dedicated host with for your deployment.
  - Use the resource estimator to ensure you provision [enough capacity](../resource_estimator.md)
  - Sourcegraph requires SSD backed storage.
  - The configured host must have [Docker Compose](https://docs.docker.com/compose/).
- [Sourcegraph license](configure.md#add-license-key). You can run through these instructions without one, but you must obtain a license for instances of more than 10 users.

We offer cloud-specific Sourcegraph installation guides:

- [Install Sourcegraph with Docker Compose on Amazon Web Services](../../install/docker-compose/aws.md)
- [Install Sourcegraph with Docker Compose on Google Cloud](../../install/docker-compose/google_cloud.md)
- [Install Sourcegraph with Docker Compose on DigitalOcean](../../install/docker-compose/digitalocean.md)

For next steps and further configuration options, visit the [site administration documentation](../../index.md).

### Direct installation

Deploy Sourcegraph with Docker Compose to your machine.

You will need:

- [Docker Compose](https://docs.docker.com/compose/) installed.
- [Sourcegraph license](configure.md#add-license-key). You can run through these instructions without one, but you must obtain a license for instances of more than 10 users.

To get started:

```bash
git clone https://github.com/sourcegraph/deploy-sourcegraph-docker
cd deploy-sourcegraph-docker/docker-compose
export SOURCEGRAPH_VERSION="v3.29.1"
git checkout $SOURCEGRAPH_VERSION
docker-compose up -d
```

> WARNING: Make sure to use the branch or tag of the deployment reference repository corresponding to the version of Sourcegraph you wish to deploy, e.g. `git checkout $SOURCEGRAPH_VERSION`, as indicated above.

<span class="virtual-br"></span>

> WARNING: To configure your Sourcegraph instance, you must create and use a fork of the reference repository - refer to [Configuring Sourcegraph with Docker Compose](./operations.md#configure) for more details.

Once the server is ready (the `sourcegraph-frontend-0` service is healthy when running `docker ps`), navigate to the hostname or IP address on port `80`.  Create the admin account, then you'll be guided through setting up Sourcegraph for code searching and navigation.

For next steps and further configuration options, visit the [site administration documentation](../../index.md).

> NOTE: Need help? [Reach out to us](../../../index.md#get-help)!

## About

### Docker Compose

Docker Compose is a tool for defining and running multi-[container](https://www.docker.com/resources/what-container) Docker applications (in this case, Sourcegraph!). With Docker Compose, you use a YAML file to configure your application’s services. Then, with a single command, you create and start all the services from your configuration.

Learn more about it [here](https://docs.docker.com/compose/).

### Reference repository

Sourcegraph for Docker Compose is configured using our [`sourcegraph/deploy-sourcegraph-docker` repository](https://github.com/sourcegraph/deploy-sourcegraph-docker/). This repository contains everything you need to spin up a Docker Compose Sourcegraph instance.

### Windows support

> WARNING: Running Sourcegraph on Windows is not supported for production deployments.

The Docker Compose installation requires a minimum of 8 CPU cores (logical) on the host machine in order to complete successfully. If using the Docker for Windows app, the default CPU count is limited to 2 which will result in errors during installation. You can go into the Docker app Settings->Resources window to increase the CPU count to > 8 to resolve this issue.
