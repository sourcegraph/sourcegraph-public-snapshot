---
title: Deployment Overview
---

# Deployment Overview

Sourcegraph offers multiple deployment options to suit different needs. The appropriate option for your organization depends on your goals and requirements, as well as the technical expertise and resources available. The following sections overview the available options and their associated investments and technical demands.

## Deployment types

Carefully consider your organization's needs and technical expertise when selecting a Sourcegraph deployment method. The method you choose cannot be changed for a running instance, so make an informed decision. The available methods have different capabilities, and the following sections provide recommendations to help you choose.

### [Sourcegraph Cloud](https://signup.sourcegraph.com/)

**For Enterprises looking for a cloud solution.**

A cloud instance hosted and maintained by Sourcegraph

<div>
  <a class="cloud-cta" href="https://signup.sourcegraph.com" target="_blank" rel="noopener noreferrer">
    <div class="cloud-cta-copy">
      <h2>Get Sourcegraph on your code.</h2>
      <h3>A single-tenant instance managed by Sourcegraph.</h3>
      <p>Sign up for a 30 day trial for your team.</p>
    </div>
    <div class="cloud-cta-btn-container">
      <div class="visual-btn">Get free trial now</div>
    </div>
  </a>
</div>
  
### [Machine Images](machine-images/index.md) 

**For Enterprises looking for a self-hosted solution.** 

<span class="badge badge-note">RECOMMENDED</span> An option to run Sourcegraph on your own infrastructure using pre-configured machine images.

Customized machine images allow you to spin up a preconfigured and customized Sourcegraph instance with just a few clicks, all in less than 10 minutes! Currently available in the following hosts:

<div class="getting-started">
  <a class="btn btn-secondary text-center" href="machine-images/aws-ami"><span>AWS AMIs</span></a>
  <a class="btn btn-secondary text-center" href="machine-images/azure"><span>Azure Images</span></a>
  <a class="btn btn-secondary text-center" href="machine-images/gce"><span>Google Compute Images</span></a>
</div>

### [Install-script](single-node/script.md)

The [install-script](single-node/script.md) is recommended for on-premises deployments and deployments to unsupported cloud environments. You can also use it to set up Linux VMs for your own machine images.

>NOTE: Deploying with machine images require the necessary technical expertise and resources to maintain and manage their own infrastructure.

### [Kubernetes](kubernetes/index.md)

**For large Enterprises that require a multi-node, self-hosted solution.**

- **Kustomize** utilizes the built-in features of kubectl to provide maximum flexibility in configuring your deployment
- **Helm** offers a simpler deployment process but with less customization flexibility

We highly recommend deploying Sourcegraph on Kubernetes with Kustomize due to the flexibility it provides.

<div class="getting-started">
  <a class="btn btn-secondary text-center" href="kubernetes/index"><span>Kustomize</span></a>
  <a class="btn btn-secondary text-center" href="kubernetes/helm"><span>Helm</span></a>
</div>

>NOTE: Given the technical knowledge required to deploy and maintain on Kubernetes, teams without these resources should contact their Sourcegraph representative at [sales@sourcegraph.com](mailto:sales@sourcegraph.com) to discuss alternative deployment options.

## Non-production environments on local machines

  - [Docker Compose](docker-compose/index.md) - A deployment option using Docker Compose
  - [Docker Single Container](docker-single-container/index.md) - A deployment option using a single Docker container
  - [Minikube](single-node/minikube.md) - A deployment option using Minikube with Docker container

---

## ARM / ARM64 support

Running Sourcegraph on ARM / ARM64 images is not supported for production deployments.

## Reference repositories

Sourcegraph provides reference repositories with branches corresponding to the version of Sourcegraph you wish to deploy for each supported deployment type. The reference repository contains everything you need to spin up and configure your instance depending on your deployment type, which also assists in your upgrade process going forward.

For more information, please read [our docs on setting up your own copy of the reference repository](repositories.md) for deploying purpose, and then follow the installation and configuration docs for your specific deployment type.

## Configuration

Configuration at the deployment level focuses on ensuring your Sourcegraph deployment runs optimally, based on the size of your repositories and the number of users. You can find your instance size using the size chart in our [Instance Size docs](instance-size.md). Configuration options will vary based on the type of deployment you choose. Consult the specific configuration deployment sections for additional information.

Sourcegraph also provides a [resource estimator](resource_estimator.md) to help predict and plan the required resource for your deployment. This tool ensures you provision appropriate resources to scale your instance.

In addition you can review our [configuration docs](../config/index.md) for overall Sourcegraph configuration.

## Operation

In general, operation activities for your Sourcegraph deployment will consist of storage management, database access, database migrations, and backup and restore. Details are provided with the instructions for each deployment type.

## Monitoring

Sourcegraph provides a number of options to monitor the health and usage of your deployment. While high-level guidance is provided as part of your deployment type, you can also review our [Observability docs](../observability/index.md) for more detailed instruction.

## Upgrades

A new version of Sourcegraph is released every month (with patch releases in between as needed). We actively maintain the two most recent monthly releases of Sourcegraph. The [changelog](../../CHANGELOG.md) provides all information related to any changes that are/were in a release.

Depending on your current version and the version you are looking to upgrade, there may be specific upgrade instruction and requirements. Checkout the [Upgrade docs](../updates/index.md) for additional information and instructions.

## External services

By default, Sourcegraph provides versions of services it needs to operate, including:

- A [PostgreSQL](https://www.postgresql.org/) instance for storing long-term information, such as user data, when using Sourcegraph's built-in authentication provider instead of an external one.
- A second PostgreSQL instance for storing large-volume code graph data.
- A [Redis](https://redis.io/) instance for storing short-term information such as user sessions.
- A second Redis instance for storing cache data.
- A `sourcegraph/blobstore` instance that serves as a local S3-compatible object storage to hold user uploads before processing. _This data is for temporary storage, and content will be automatically deleted once processed._
- A [Jaeger](https://www.jaegertracing.io/) instance for end-to-end distributed tracing.

> NOTE: As a best practice, configure your Sourcegraph instance to use an external or managed version of these services. Using a managed version of PostgreSQL can make backups and recovery easier to manage and perform. Using a managed object storage service may decrease hosting costs as persistent volumes are often more expensive than object storage space.

### External services guides

See the following guides to use an external or managed version of each service type.

- [PostgreSQL Guide](../postgres.md)
- See [Using your PostgreSQL server](../external_services/postgres.md) to replace the bundled PostgreSQL instances.
- See [Using your Redis server](../external_services/redis.md) to replace the bundled Redis instances.
- See [Using a managed object storage service (S3 or GCS)](../external_services/object_storage.md) to replace the bundled blobstore instance.
- See [Using an external Jaeger instance](../observability/tracing.md#use-an-external-jaeger-instance) in our [tracing documentation](../observability/tracing.md) to replace the bundled Jaeger instance.Use-an-external-Jaeger-instance

> NOTE: Using Sourcegraph with an external service is a [paid feature](https://about.sourcegraph.com/pricing). [Contact us](https://about.sourcegraph.com/contact/sales) to get a trial license.

### Cloud alternatives

- Amazon Web Services: [AWS RDS for PostgreSQL](https://aws.amazon.com/rds/), [Amazon ElastiCache](https://aws.amazon.com/elasticache/redis/), and [S3](https://aws.amazon.com/s3/) for storing user uploads.
- Google Cloud: [Cloud SQL for PostgreSQL](https://cloud.google.com/sql/docs/postgres/), [Cloud Memorystore](https://cloud.google.com/memorystore/), and [Cloud Storage](https://cloud.google.com/storage) for storing user uploads.
- Digital Ocean: [Digital Ocean Managed Databases](https://www.digitalocean.com/products/managed-databases/) for [Postgres](https://www.digitalocean.com/products/managed-databases-postgresql/), [Redis](https://www.digitalocean.com/products/managed-databases-redis/), and [Spaces](https://www.digitalocean.com/products/spaces/) for storing user uploads.
