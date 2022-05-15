# Deployment Overview

Sourcegraph supports two main deployment types: [Docker Compose](docker-compose/index.md) and [Kubernetes](kubernetes/index.md). Each deployment type will require a different level of investment and technical understanding. What works best depends on the needs and desired outcomes for your business. 

If you aren't currently working with our Customer Engineering team, this overview will provide a high-level view of what's available and needed depending on the deployment type you choose. 

Specifically, the table provided in the [Deployment types](#deployment-types) section will provide some high-level guidance, followed by more detailed descriptions for each type.

Sourcegraph also provides a [resource estimator](#resource-planning) to help predict and plan the required resource for your deployment. This tool ensures you provision appropriate resources to scale your instance.

If you are short on time and looking for a quick way to test Sourcegraph locally, consider running Sourcegraph via our [Docker Single Container](docker-single-container/index.md). 

Or, if you don't want to bother with setup and configuration [try Sourcegraph Cloud](https://sourcegraph.com) instead.

## Resource planning

Sourcegraph has provided the [Resource Estimator](resource_estimator.md) as a starting point to determine necessary resources based on the size of your deployment. 

We recommend the Kubernetes deployment type if your deployment scenario includes a large codebase and many users. The [Kubernetes docs](kubernetes/index.md) provide additional information for [scaling Kubernetes deployments](kubernetes/scale.md).

## Deployment types

| Deployment Type | Suggested for | Setup time | Resource isolation | Auto-healing | Multi-machine | Complexity |
| ---------------------------------------------------------------------- | ---------------------------------------------------------------- | ----------------- | :----------------: | :----------: | :-----------: | :--------: |
| [Kubernetes with Helm](kubernetes/helm.md) | Production deployments of any size | 5 - 90 minutes | YES | YES | YES | Easy - Hard |
| [Docker Compose](docker-compose/index.md) | Production deployments where Kubernetes with Helm is not viable | 5 - 30 minutes | YES | YES | NO | Easy - Medium |
| [Kubernetes without Helm](kubernetes/index.md) | Production deployments of any size | 30 - 90 minutes | YES | YES | YES| Medium - Hard |
| [Docker Single Container](docker-single-container/index.md) | Local testing _(Not recommended for production)_ | 1 minute | NO | NO | NO | Easy |

Each of the deployment types listed in the table above provides a different level of capability. As mentioned previously, base your deployment type on the needs of your business. However, you should also consider the technical expertise available for your deployment. The sections below provide more detailed recommendations for each deployment type.

### [Kubernetes with Helm](kubernetes/helm.md)

We recommend Kubernetes with Helm for most production deployments. 

Kubernetes provides resource isolation (from other services or applications), automated-healing, and far greater ability to scale.

Helm provides a simple mechanism for deployment customizations, as well as a much simpler upgrade experience.

### [Docker Compose](docker-compose/index.md)

Docker Compose is recommended for small and medium-size deployments where Kubernetes with Helm is not a viable option. 

It does not provide multi-machine capability such as high availability, but will require less setup time overall.

### [Kubernetes without Helm](kubernetes/index.md)

Before making a decision to deploy via Kubernetes without Helm, checkout our [Kubernetes with Helm docs](kubernetes/helm.md) for additional information on why we [recommend using Helm](kubernetes/helm.md#why-use-helm).

If you are unable to use Helm to deploy, but still want to use Kubernetes, follow our [Kubernetes deployment documentation](kubernetes/index.md). 

This path will require advanced knowledge of Kubernetes. For team's without the ability to support this, please speak to your Sourcegraph contact about using Docker Compose instead. 

### [Docker Single Container](docker-single-container/index.md) 

The Docker Single container option is provided for **local proof-of-concept only** and is **not intended for testing or deployment at a pre-production/production level**. 

Some features, such as [Code Insights](../../code_insights/index.md), are not available when using this deployment type. 

If you're just starting out and want the absolute quickest setup time, [try Sourcegraph Cloud](https://sourcegraph.com).

## Reference repositories

For [Docker Compose](https://github.com/sourcegraph/deploy-sourcegraph-docker/),[Kubernetes with Helm](https://docs.sourcegraph.com/admin/install/kubernetes/helm), and [Kubernetes without Helm](https://github.com/sourcegraph/deploy-sourcegraph/) deployments, Sourcegraph provides reference repositories with branches corresponding to the version of Sourcegraph you wish to deploy. The reference repository contains everything you need to spin up and configure your instance depending on your deployment type, which also assists in your upgrade process going forward.

Before starting, you will need to decide your deployment type, including if you would like to use Kubernetes with Helm (vs. a more manual customization path). In the case of Kubernetes once you choose Helm (or not), it **can't be changed afterwards**. 

For more information, follow the install and configuration docs for your specific deployment type: [Docker Compose](docker-compose/index.md), [Kubernetes with Helm](kubernetes/helm.md), or [Kubernetes without Helm](kubernetes/index.md).


## External services

By default, Sourcegraph provides versions of services it needs to operate, including:

- A [PostgreSQL](https://www.postgresql.org/) instance for storing long-term information, such as user data, when using Sourcegraph's built-in authentication provider instead of an external one.
- A second PostgreSQL instance for storing large-volume precise code intelligence data.
- A [Redis](https://redis.io/) instance for storing short-term information such as user sessions.
- A second Redis instance for storing cache data.
- A [MinIO](https://min.io/) instance that serves as a local S3-compatible object storage to hold user uploads before processing. _This data is for temporary storage, and content will be automatically deleted once processed._
- A [Jaeger](https://www.jaegertracing.io/) instance for end-to-end distributed tracing. 

> NOTE: As a best practice, configure your Sourcegraph instance to use an external or managed version of these services. Using a managed version of PostgreSQL can make backups and recovery easier to manage and perform. Using a managed object storage service may decrease hosting costs as persistent volumes are often more expensive than object storage space.

### External services guides
See the following guides to use an external or managed version of each service type.

- [PostgreSQL Guide](../postgres.md)
- See [Using your PostgreSQL server](../external_services/postgres.md) to replace the bundled PostgreSQL instances.
- See [Using your Redis server](../external_services/redis.md) to replace the bundled Redis instances.
- See [Using a managed object storage service (S3 or GCS)](../external_services/object_storage.md) to replace the bundled MinIO instance.
- See [Using an external Jaeger instance](../observability/tracing.md#use-an-external-jaeger-instance) in our [tracing documentation](../observability/tracing.md) to replace the bundled Jaeger instance.Use-an-external-Jaeger-instance

> NOTE: Using Sourcegraph with an external service is a [paid feature](https://about.sourcegraph.com/pricing). [Contact us](https://about.sourcegraph.com/contact/sales) to get a trial license.

### Cloud alternatives

- Amazon Web Services: [AWS RDS for PostgreSQL](https://aws.amazon.com/rds/), [Amazon ElastiCache](https://aws.amazon.com/elasticache/redis/), and [S3](https://aws.amazon.com/s3/) for storing user uploads.
- Google Cloud: [Cloud SQL for PostgreSQL](https://cloud.google.com/sql/docs/postgres/), [Cloud Memorystore](https://cloud.google.com/memorystore/), and [Cloud Storage](https://cloud.google.com/storage) for storing user uploads.
- Digital Ocean: [Digital Ocean Managed Databases](https://www.digitalocean.com/products/managed-databases/) for [Postgres](https://www.digitalocean.com/products/managed-databases-postgresql/), [Redis](https://www.digitalocean.com/products/managed-databases-redis/), and [Spaces](https://www.digitalocean.com/products/spaces/) for storing user uploads.

## Configuration

Configuration at the deployment level focuses on ensuring your Sourcegraph runs optimally, based on the size of your repositories and the number of users. Configuration options will vary based on the type of deployment you choose. Consult the specific configuration deployment sections for additional information.

In addition you can review our [Configuration docs](../config/index.md) for overall Sourcegraph configuration.

## Operation

In general, operation activities for your Sourcegraph deployment will consist of storage management, database access, database migrations, and backup and restore. Details are provided with the instructions for each deployment type.

## Monitoring

Sourcegraph provides a number of options to monitor the health and usage of your deployment. While high-level guidance is provided as part of your deployment type, you can also review our [Observability docs](../observability/index.md) for more detailed instruction.

## Upgrades

A new version of Sourcegraph is released every month (with patch releases in between as needed). We actively maintain the two most recent monthly releases of Sourcegraph. The [changelog](../../CHANGELOG.md) provides all information related to any changes that are/were in a release.

Depending on your current version and the version you are looking to upgrade, there may be specific upgrade instruction and requirements. Checkout the [Upgrade docs](../updates/index.md) for additional information and instructions.

