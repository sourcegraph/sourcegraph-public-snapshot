# Site administration documentation

> NOTE: Upgrading to `3.0.1+`? Read our [migration guide](migration/3_0.md) for `2.x` and `3.0.0`.

Site administrators are the admins responsible for deploying, managing, and configuring Sourcegraph for regular users.

- [Install Sourcegraph](install.md)
  - [Install Sourcegraph with Docker](install/docker.md)
  - [Install Sourcegraph on a cluster](install/cluster.md)
- Management, deployment, and configuration:
  - [Configuration](config/index.md)
  - [Adding Git repositories](repo/add.md) (from a code host or clone URL)
  - [Repository webhooks](repo/webhooks.md)
  - [Management console](management_console.md)
  - [User authentication](auth.md)
  - [Upgrading Sourcegraph](updates.md)
  - [nginx HTTP server configuration](nginx.md)
  - [Setting the URL for your instance](url.md)
  - [Monitoring and tracing](monitoring_and_tracing.md)
  - [Repository permissions](repo/permissions.md)
  - [Upgrading PostgreSQL](postgres.md)
  - [Using external databases (PostgreSQL and Redis)](external_database.md)
- Features:
  - [Code intelligence and language servers](../user/code_intelligence/index.md)
  - [Sourcegraph extensions and extension registry](extensions.md)
  - [Search](search.md)
  - [Federation](federation.md)
  - [Pings](pings.md)
  - [Usage statistics](../user/usage_statistics.md)
  - [User feedback surveys](../user/user_surveys.md)
- Integrations:
  - [GitHub and GitHub Enterprise](../integration/github.md)
  - [GitLab](../integration/gitlab.md)
  - [Bitbucket Server](../integration/bitbucket_server.md)
  - [AWS CodeCommit](../integration/aws_codecommit.md)
  - [Phabricator](../integration/phabricator.md)
  - [All integrations](../integration.md)
- Migration guides:
  - [From OpenGrok to Sourcegraph](migration/opengrok.md)
  - [Migrating to Sourcegraph 3.0.1+](migration/3_0.md)
- [Pricing and subscriptions](subscriptions/index.md)
- [FAQ](faq.md)
