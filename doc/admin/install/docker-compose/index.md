# Install Sourcegraph with Docker Compose

If you need scalability and high-availability beyond what a single-node [Docker Compose](https://docs.docker.com/compose/) can offer, use the [Kubernetes cluster deployment option](https://github.com/sourcegraph/deploy-sourcegraph) instead.

If you want to migrate from the single-container server (`sourcegraph/server`) to the Docker Compose deployment, refer to this [migration guide](./migrate.md).

## Requirements

- [Sourcegraph Enterprise license](configure.md#add-license-key). _You can run through these instructions without one, but you must obtain a license for instances of more than 10 users._
- [Docker Compose](https://docs.docker.com/compose/).
- A dedicated host with for your deployment.
  - Use the resource estimator to ensure you provision [enough capacity](../resource_estimator.md)
  - Sourcegraph requires SSD backed storage.

> WARNING: You need to create a [fork of our deployment reference.](configure.md#fork-this-repository)
### Storage

The [Sourcegraph Docker Compose definition](https://github.com/sourcegraph/deploy-sourcegraph-docker/blob/master/docker-compose/docker-compose.yaml) uses [Docker volumes](https://docs.docker.com/storage/volumes/) to store its data. These volumes are stored at `/var/lib/docker/volumes` by [default on Linux](https://docs.docker.com/storage/#choose-the-right-type-of-mount).

### Note About Windows Installation

> WARNING: Running Sourcegraph on Windows is not supported for production deployments.

The Docker Compose installation requires a minimum of 8 CPU cores (logical) on the host machine in order to complete successfully. If using the Docker for Windows app, the default CPU count is limited to 2 which will result in errors during installation. You can go into the Docker app Settings->Resources window to increase the CPU count to > 8 to resolve this issue.

## Steps

It takes less than 5 minutes to run and install Sourcegraph using Docker Compose:

```bash
# 🚨 The master branch tracks development. Use the branch of this repository corresponding to the version of Sourcegraph you wish to deploy, e.g. git checkout v3.24.1

git clone https://github.com/sourcegraph/deploy-sourcegraph-docker
cd deploy-sourcegraph-docker/docker-compose
SOURCEGRAPH_VERSION="v3.25.1"
git checkout $SOURCEGRAPH_VERSION
docker-compose up -d
```

Once the server is ready (the `sourcegraph-frontend-0` service is healthy when running `docker ps`), navigate to the hostname or IP address on port `80`.  Create the admin account, then you'll be guided through setting up Sourcegraph for code searching and navigation.

For next steps and further configuration options, visit the [site administration documentation](../../index.md).

> NOTE: If you get stuck or need help, [file an issue](https://github.com/sourcegraph/sourcegraph/issues/new?&title=Improve+Sourcegraph+quickstart+guide), [tweet (@srcgraph)](https://twitter.com/srcgraph) or [email](mailto:support@sourcegraph.com?subject=Sourcegraph%20quickstart%20guide).

## Cloud installation guides

Cloud specific Sourcegraph installation guides for AWS, Google Cloud and Digital Ocean.

- [Install Sourcegraph with Docker Compose on AWS](../../install/docker-compose/aws.md)
- [Install Sourcegraph with Docker Compose on Google Cloud](../../install/docker-compose/google_cloud.md)
- [Install Sourcegraph with Docker Compose on DigitalOcean](../../install/docker-compose/digitalocean.md)
