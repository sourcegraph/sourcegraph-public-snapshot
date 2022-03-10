# Deployment Overview

Sourcegraph can be installed in a variety of methods to set up a deployment for your private code. Whether you are setting up a proof of concept, or looking for a product-level deployment, this overview will provide you with an introduction to key items to consider.

For a list of all deployment topics, visit our [Deployment Table of Contents](index), and if you're just starting out, you can [**try Sourcegraph Cloud**](https://sourcegraph.com) or [run Sourcegraph locally](docker/index.md). 

## Resource planning

Sourcegraph has provided the [Resource Estimator](a_new_admin/deploy/resource_estimator) as a starting point to determine necessary resources based on the size of your deployment. 

As a recommendation, if you are planning deployment scenario will include very large codebases and a large number of users, our [Kubernetes](a_new_admin/deploy/kubernetes/scale) Deployment option will be your best option.

## Options and scenarios

Using the table below click on the deployment type that best meets your needs.

Of course, if you're just starting out, you can [**try Sourcegraph Cloud**](https://sourcegraph.com) or [run Sourcegraph locally](docker/index.md).

| Deployment Type                                          | Suggested for                                           | Setup time      | Resource isolation | Auto-healing | Multi-machine |
| -------------------------------------------------------- | ------------------------------------------------------- | --------------- | :----------------: | :----------: | :-----------: |
| [**Docker Compose**](../install/docker-compose/index.md) | **Small & medium** production deployments               | 🟢 5 minutes     |         ✅          |      ✅       |       ❌       |
| [**Kubernetes**](../install/kubernetes/index.md)         | **Medium & large** highly-available cluster deployments | 🟠 30-90 minutes |         ✅          |      ✅       |       ✅       |
| [**Single-container**](../install/docker/index.md)       | Local testing                                           | 🟢 1 minute      |         ❌          |      ❌       |       ❌       |


> NOTE: The Single container option is provided for local proof-of-concepts and not intended for testing or deploye at a pre-production/production leve. If you're just starting out, and want to absolute quickest setup time, [**try Sourcegraph Cloud**](https://sourcegraph.com).


## External services

Sourcegraph by default provides versions of services it needs to operate, including:

- A [PostgreSQL](https://www.postgresql.org/) instance for storing long-term information, such as user information when using Sourcegraph's built-in authentication provider instead of an external one.
- A second PostgreSQL instance for storing large-volume precise code intelligence data.
- A [Redis](https://redis.io/) instance for storing short-term information such as user sessions.
- A second Redis instance for storing cache data.
- A [MinIO](https://min.io/) instance that serves as a local S3-compatible object storage to hold user uploads before they can be processed. _This data is for temporary storage and content will be automatically deleted once processed._
- A [Jaeger](https://www.jaegertracing.io/) instance for end-to-end distributed tracing. 

Your Sourcegraph instance can be configured to use an external or managed version of these services. Using a managed version of PostgreSQL can make backups and recovery easier to manage and perform. Using a managed object storage service may decrease your hosting costs as persistent volumes are often more expensive than object storage space.

See the following guides to use an external or managed version of each service type.

- See [Using your own PostgreSQL server](./postgres.md) to replace the bundled PostgreSQL instances.
- See [Using your own Redis server](./redis.md) to replace the bundled Redis instances.
- See [Using a managed object storage service (S3 or GCS)](./object_storage.md) to replace the bundled MinIO instance.
- See [Using an external Jaeger instance](../observability/tracing.md#Use-an-external-Jaeger-instance) to replace the bundled Jaeger instance.

> NOTE: Using Sourcegraph with an external service is a [paid feature](https://about.sourcegraph.com/pricing). [Contact us](https://about.sourcegraph.com/contact/sales) to get a trial license.

### Cloud alternatives

- Amazon Web Services: [AWS RDS for PostgreSQL](https://aws.amazon.com/rds/), [Amazon ElastiCache](https://aws.amazon.com/elasticache/redis/), and [S3](https://aws.amazon.com/s3/) for storing user uploads.
- Google Cloud: [Cloud SQL for PostgreSQL](https://cloud.google.com/sql/docs/postgres/), [Cloud Memorystore](https://cloud.google.com/memorystore/), and [Cloud Storage](https://cloud.google.com/storage) for storing user uploads.
- Digital Ocean: [Digital Ocean Managed Databases](https://www.digitalocean.com/products/managed-databases/) for [Postgres](https://www.digitalocean.com/products/managed-databases-postgresql/), [Redis](https://www.digitalocean.com/products/managed-databases-redis/), and [Spaces](https://www.digitalocean.com/products/spaces/) for storing user uploads.

## Configuration (TBD)

Configuration at the deployment level focuses on ensuring your Sourcegraph runs optimally based on the size of your repositories and number of users. Configuration options will vary based on the type of deployment you choose, so you will want to consult the specific configuration guides for additional information.

If you're looking for configuration at the Administration level, check out the [customization section](a_new_admin\customization.md).


## Updates

A new version of Sourcegraph is released every month (with patch releases in between, released as needed). Check the [Sourcegraph blog](https://about.sourcegraph.com/blog) or the site admin updates page to learn about updates. We actively maintain the two most recent monthly releases of Sourcegraph.

**Regardless of your deployment type**, the following rules apply:

- **Upgrade one minor version at a time**, e.g. v3.26 --> v3.27 --> v3.28.
  - Patches (e.g. vX.X.4 vs. vX.X.5) do not have to be adopted when moving between vX.X versions.
- **Check the [update notes for your deployment type](#update-notes) for any required manual actions** before updating.
- Check your [out of band migration status](../migration/index.md) prior to upgrade to avoid a necessary rollback while the migration finishes.

### Update notes

Please see the instructions for your deployment type:

- [Sourcegraph with Docker Compose](docker_compose.md)
- [Sourcegraph with Kubernetes](kubernetes.md)
- [Single-container Sourcegraph with Docker](server.md)
- [Pure-Docker custom deployments](pure_docker.md)


### Changelog

For product update notes, please refer to the [changelog](../../CHANGELOG.md).

## Reference Repository

For Docker Compose and Kubernetes deployments, Sourcegraph provides reference repositories with branches corresponding to the version of Sourcegraph you wish to deploy. Depending on your deployment type, the reference repository contains everything you need to spin up and configure your instance. This will also assist in your upgrade process going forward. For more information, follow the install and configuration docs for your deployment type (linked below).

- Docker Compose
- Kubernets





