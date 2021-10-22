# Administration

<p class="lead">
Administration guides and documentation for <a href="install">self-hosted Sourcegraph instances</a>.
</p>

Administration is usually handled by site administrators are the admins responsible for deploying, managing, and configuring Sourcegraph for regular users. They have [special privileges](privileges.md) on a Sourcegraph instance.

## [Install Sourcegraph](install/index.md)

- [Best Practices](deployment_best_practices.md)
- [Deploying workers](workers.md)
- [PostgreSQL configuration](config/postgres-conf.md)
- [Upgrading PostgreSQL](postgres.md)
- [Using external services (PostgreSQL, Redis, S3/GCS)](external_services/index.md)
- <span class="badge badge-experimental">Experimental</span> [Validation](validation.md)
- <span class="badge badge-experimental">Experimental</span> [Deploy executors](deploy_executors.md)

## [Upgrade Sourcegraph](updates/index.md)

- [Migrations](migration/index.md)

## [Configuration](config/index.md)

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
  
For deployment configuration, please refer to the relevant [installation guide](./install/index.md).

## [Observability](observability.md)

- [Monitoring guide](how-to/monitoring-guide.md)
- [Metrics and dashboards](./observability/metrics.md)
- [Alerting](./observability/alerting.md)

## Features

- [Code intelligence and language servers](../code_intelligence/index.md)
- [Sourcegraph extensions and extension registry](extensions/index.md)
- [Search](search.md)
- [Federation](federation/index.md)
- [Pings](pings.md)
- [Usage statistics](usage_statistics.md)
- [User feedback surveys](user_surveys.md)
- [Beta and experimental features](beta_and_experimental_features.md)
- [Pricing and subscriptions](subscriptions/index.md)
