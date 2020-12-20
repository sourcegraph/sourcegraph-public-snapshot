# Sourcegraph monitoring generator

The Sourcegraph monitoring generator uses [`Container` definitions](./monitoring/README.md#type-container) to generate integrations with [Sourcegraph's monitoring architecture](https://about.sourcegraph.com/handbook/engineering/observability/monitoring_architecture).
It also aims to help codify guidelines defined in the [Sourcegraph monitoring pillars](https://about.sourcegraph.com/handbook/engineering/observability/monitoring_pillars).

This page primarily documents the [generator's current capabilities](#features) - in other words, and what you get for free by declaring Sourcegraph service monitoring in this package - as well as [how to make changes to the generator itself](#development).

To learn about how to find, add, and use monitoring, see the [Sourcegraph monitoring developer guide](https://about.sourcegraph.com/handbook/engineering/observability/monitoring).

## Usage

From this directory:

```sh
go generate ./...
```

Logging output supports the [Sourcegraph log level flags](https://docs.sourcegraph.com/admin/observability#logs).
Other configuration options can be customized via flags declared in [`main.go`](./main.go).

## Features

### Documentation generation

The generator automatically creates documentation from monitoring definitions, such as [alert solutions references](https://docs.sourcegraph.com/admin/observability/alert_solutions), that customers and engineers can reference.

Links to generated documentation can be provided in our other generated integrations - for example, [Slack alerts](https://docs.sourcegraph.com/admin/observability/alerting#setting-up-alerting) will provide a link to the appropriate alert solutions entry.

### Grafana integration

The generator automatically generates and ships dashboards from monitoring definitions within the [Sourcegraph Grafana distribution](https://about.sourcegraph.com/handbook/engineering/observability/monitoring_architecture#sourcegraph-grafana).

It also takes care of the following:

- Graphs within rows are sized appropriately
- Alerts visualization through the [`ObservableAlertDefinition` API](./monitoring/README.md#type-observablealertdefinition):
  - Overview graphs for alerts (both Sourcegraph-wide and per-service)
  - Threshold lines for alerts of all levels are rendered in graphs
- Formatting of units, labels, and more (using either the defaults, or the [`ObservablePanelOptions` API](./monitoring/README.md#type-observablepaneloptions))
- Maintaining a uniform look and feel across all dashboards

Links to generated documentation can be provided in our other generated integrations - for example, [Slack alerts](https://docs.sourcegraph.com/admin/observability/alerting#setting-up-alerting) will provide a link to the appropriate service's dashboard.

### Prometheus integration

The generator automatically generates and ships Prometheus recording rules and alerts within the [Sourcegraph Prometheus distribution](https://about.sourcegraph.com/handbook/engineering/observability/monitoring_architecture#sourcegraph-prometheus). This includes the [`alert_count` recording rules](https://about.sourcegraph.com/handbook/engineering/observability/monitoring_architecture#alert-count-metrics) and native Prometheus alerts, all with appropriate and consistent labels.

Generated Prometheus recording rules are leveraged by the [Grafana integration](#grafana-integration).

### Alertmanager integration

The generator's [Prometheus integration](#prometheus-integration) is a critical part of the [Sourcegraph's alerting capabilities](https://about.sourcegraph.com/handbook/engineering/observability/monitoring_architecture#alert-notifications), which handles alert routing by level and formatting of alert messages to include links to [documentation](#documentation-generation) and [dashboards](#grafana-integration). Learn more about using Sourcegraph alerting in the [alerting documentation](https://docs.sourcegraph.com/admin/observability/alerting).

At Sourcegraph, routing based on team ownership (as defined by [`ObservableOwner`](./monitoring/README.md#type-observableowner)) is used to route customer support requests and [on-call events through OpsGenie](https://about.sourcegraph.com/handbook/engineering/incidents/on_call).

## Development

The Sourcegraph monitoring generator consists of three components:

- The [main program](./main.go) - this is the primary entrypoint to the generator.
- _Definitions_, defined in the top-level [`monitoring/definitions` package](./definitions/).
  This is where the all service monitoring definitions lives.
  If you are editing monitoring, this is probably where you want to look - see the [Sourcegraph monitoring developer guide](https://about.sourcegraph.com/handbook/engineering/observability/monitoring).
- _Generator_, defined in the nested [`monitoring/monitoring` package](./monitoring/README.md) package.
  This is where the API for service monitoring definitions is defined, as well as the generator code.

All features and capabilities for developed for the generator should align with the [Sourcegraph monitoring pillars](https://about.sourcegraph.com/handbook/engineering/observability/monitoring_pillars).
