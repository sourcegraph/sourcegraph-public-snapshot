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

## Features

### Documentation pages

TODO

### Grafana integration

TODO

### Prometheus integration

TODO

### Alertmanager integration

TODO

## Development

The Sourcegraph monitoring generator consists of three components:

* The [main program](./main.go).
* *Definitions*, defined in the top-level [`monitoring/definitions` package](./definitions/).
  This is where the all service monitoring definitions lives.
  If you are editing monitoring, this is probably where you want to look - see the [Sourcegraph monitoring developer guide](https://about.sourcegraph.com/handbook/engineering/observability/monitoring).
* *Generator*, defined in the nested [`monitoring/monitoring` package](./monitoring/README.md) package.
  This is where the API for service monitoring definitions is defined, as well as the generator code.
