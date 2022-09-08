# Administration

Administration is usually handled by site administrators are the admins responsible for deploying, managing, and configuring Sourcegraph for regular users. They have [special privileges](privileges.md) on a Sourcegraph instance. Check out this [quickstart guide](how-to/site-admin-quickstart.md) for more info on Site Administration.

## [Deploy and Configure Sourcegraph](deploy/index.md)

- [Deployment overview](deploy/index.md)
  - [Kubernetes with Helm](deploy/kubernetes/helm.md)
  - [Docker Compose](deploy/docker-compose/index.md)
  - [See all deployment options](deploy/index.md#deployment-types)
- [Best practices](deployment_best_practices.md)
- [Deploying workers](workers.md)
- [PostgreSQL configuration](config/postgres-conf.md)
- [Using external services (PostgreSQL, Redis, S3/GCS)](external_services/index.md)
- <span class="badge badge-experimental">Experimental</span> [Validation](validation.md)
- <span class="badge badge-experimental">Experimental</span> [Executors](executors.md)
- <span class="badge badge-experimental">Experimental</span> [Deploy executors](deploy_executors.md)

## [Upgrade Sourcegraph](updates/index.md)

> NOTE: The Sourcegraph 4.0 [`migrator`](./how-to/manual_database_migrations.md) can now perform upgrades across 
multiple versions on any instance 3.20 or later.

- [Upgrade Sourcegraph](updates/index.md)
  - [Single-minor-version "standard" upgrades](updates/index.md#standard-upgrades)
  - [Multi-version upgrades](updates/index.md#multi-version-upgrades)
- [Upgrade notes](updates/index.md#update-notes) (by distribution method):
  - [Sourcegraph with Docker Compose](updates/docker_compose.md)
  - [Sourcegraph with Kubernetes](updates/kubernetes.md)
  - [Single-container Sourcegraph with Docker](updates/server.md)
  - [Pure-docker custom deployments](updates/pure_docker.md)
- [Migration guides](migration/index.md)
- [Upgrading PostgreSQL](postgres.md#upgrading-postgresql)

## [Configuration](config/index.md)

- [Site Administrator Quickstart](how-to/site-admin-quickstart.md)
- [Integrations](../integration/index.md)
- [Adding Git repositories](repo/add.md) (from a code host or clone URL)
  - [Monorepo](monorepo.md)
  - [Repository webhooks](repo/webhooks.md)
- [HTTP and HTTPS/SSL configuration](http_https_configuration.md)
  - [Adding SSL (HTTPS) to Sourcegraph with a self-signed certificate](ssl_https_self_signed_cert_nginx.md)
- [User authentication](auth/index.md)
  - [User data deletion](user_data_deletion.md)
- [Setting the URL for your instance](url.md)
- [Repository permissions](repo/permissions.md)
  - [Row-level security](repo/row_level_security.md)
- [Batch Changes](../batch_changes/how-tos/site_admin_configuration.md)

For deployment configuration, please refer to the relevant [installation guide](deploy/index.md).

## [Observability](observability.md)

- [Monitoring guide](how-to/monitoring-guide.md)
- [Metrics and dashboards](./observability/metrics.md)
- [Alerting](./observability/alerting.md)
- [Tracing](./observability/tracing.md)
- [Logs](./observability/logs.md)

## Features

- <span class="badge badge-experimental">Experimental</span> [Analytics](./analytics.md)
- [Batch Changes](../batch_changes/index.md)
- [Beta and experimental features](beta_and_experimental_features.md)
- [Code navigation](../code_navigation/index.md)
- [Federation](federation/index.md)
- [Pings](pings.md)
- [Pricing and subscriptions](subscriptions/index.md)
- [Search](search.md)
- [Sourcegraph extensions and extension registry](extensions/index.md)
- [Usage statistics](usage_statistics.md)
- [User feedback surveys](user_surveys.md)

