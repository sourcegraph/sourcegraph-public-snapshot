# Observability developer documentation

This documentation is for generalized, usecase-agnostic development of Sourcegraph's observability.
Sourcegraph employees should also refer to the [handbook's observability section](https://about.sourcegraph.com/handbook/engineering/observability) for Sourcegraph-specific documentation.

Observability includes:

- Monitoring - how you know _when something is wrong_, which includes:
  - Dashboards & metrics
  - Alerting
  - Health checks
- Debugging - how you debug _what is wrong_, which includes:
  - Distributed tracing
  - Logging

> NOTE: For how to *use* Sourcegraph's observability and an overview of our observability features, refer to the [observability for administrators documentation](../../admin/observability/index.md).

## Concepts

- [Sourcegraph monitoring pillars](https://about.sourcegraph.com/handbook/engineering/observability/monitoring_pillars)
- [Sourcegraph monitoring architecture](https://about.sourcegraph.com/handbook/engineering/observability/monitoring_architecture)

## Background

- [Monitoring generator](./monitoring-generator.md)

## Guides

- [How to find monitoring](../how-to/find_monitoring.md)
- [How to add monitoring](../how-to/add_monitoring.md)
