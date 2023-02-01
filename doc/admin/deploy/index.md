---
title: Deployment Overview
---

# Deployment Overview

Sourcegraph provides a variety of deployment options to meet the diverse needs of our users. Each method requires varying levels of technical expertise and investment. The appropriate choice for your organization will depend on your specific goals and requirements. The following section provides an overview of the available deployment options and the associated level of investment and technical understanding required.

## Deployment types

When planning to deploy Sourcegraph, it is important to first determine the appropriate deployment method for your organization. For example, if you choose to deploy using Kubernetes, you will need to make a decision on whether or not to use Helm. Once a deployment type is selected and implemented, it **cannot** be changed.

The deployment types available for Sourcegraph have varying levels of capabilities. When selecting a deployment type, it is important to consider both the needs of your business as well as the technical expertise available within your organization. As previously mentioned, it is not possible to change the deployment type of a running instance, so it's important to make an informed decision. The following sections provide more detailed recommendations for each deployment type to assist you in making your decision.

### [Install-script for single-machine](single-node/script.md) 

Quickly install Sourcegraph onto a single Linux machine using our install script.

### [Machine Images](machine-images/index.md) 

<span class="badge badge-note">RECOMMENDED</span> Customized machine images allow you to spin up a preconfigured and customized Sourcegraph instance with just a few clicks, all in less than 10 minutes!

Currently available in the following hosts:

<div class="getting-started">
  <a class="btn btn-secondary text-center" href="machine-images/aws-ami"><span>AWS AMIs</span></a>
  <a class="btn btn-secondary text-center" href="machine-images/azure"><span>Azure Images</span></a>
  <a class="btn btn-secondary text-center" href="machine-images/gce"><span>Google Compute Images</span></a>
</div>

Machine images are our recommended deployment method for organizations that require a simple on-premise solution, but would still like to maintain and manage their own infrastructure. 

### [Kubernetes](kubernetes/index.md)

Kubernetes is recommended for non-standard deployments where deploying Sourcegraph with our Machine Images or install-script is not a viable option.

Below are the supported methods to deploy Sourcegraph on Kubernetes:

- **Kustomize** our recommended approach. Kustomize provides a simple, yet highly flexible way to configure your deployment using the native features of `kubectl`.
- **Helm** provides a widely used mechanism for deployment customizations and managing upgrades.

Both paths require in-depth Kubernetes knowledge to set up and maintain. For teams without the ability to support this, please speak to your Sourcegraph representative to learn about other deployment methods that we offer.

<div class="getting-started">
  <a class="btn btn-secondary text-center" href="kubernetes/index"><span>Kustomize</span></a>
  <a class="btn btn-secondary text-center" href="kubernetes/helm"><span>Helm</span></a>
</div>

 We strongly recommended you to contact us at [sales@sourcegraph.com](mailto:sales@sourcegraph.com) to learn more about the technical complexities involved to make sure Kubernetes is the best option for you and your teams. We do not recommend this method if you do not have an established infrastructure that is built on Kubernetes.

### ARM / ARM64 support

Running Sourcegraph on ARM / ARM64 images is not supported for production deployments at this time.

---

## Recommendations

Some of the options below are recommended for individuals/teams who have the necessary technical expertise and resources to maintain and manage their own infrastructure, while Sourcegraph Cloud and Sourcegraph App are more suitable for users who don't want to manage the infrastructure themselves.

### Cloud solution

For Enterprises looking for a cloud solution:

  - [Sourcegraph Cloud](https://signup.sourcegraph.com/) - A cloud instance hosted and maintained by Sourcegraph

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
  
### Self-hosted solution - single-node

For Enterprises looking for a self-hosted solution:
  
  - [Machine images](machine-images/index.md) - An option to run Sourcegraph on your own infrastructure using pre-configured machine images
  - Linux Virtual Machines with our [install-script](single-node/script.md) - An option to set up a single-node Sourcegraph deployment using a script provided by Sourcegraph
  
### Self-hosted solution - multi-node

For large Enterprises that require a multi-node, self-hosted solution

  - [Kubernetes with Kustomize](kubernetes/index.md) - A multi-node deployment option using Kubernetes and Kustomize
  - [Kubernetes with Helm](kubernetes/helm.md) - A multi-node deployment option using Kubernetes and Helm

We strongly recommended you to contact us at [sales@sourcegraph.com](mailto:sales@sourcegraph.com) to learn more about the technical complexities involved to make sure Kubernetes is the best option for you and your teams.

### Non-production environments on local machines

  - Sourcegraph App (Coming soon) - A standalone application for local development and experimentation
  - [Docker Compose](docker-compose/index.md) - A deployment option using Docker Compose
  - [Docker Single Container](docker-single-container/index.md) - A deployment option using a single Docker container
  
---

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
