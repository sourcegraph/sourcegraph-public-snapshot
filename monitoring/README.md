# Sourcegraph monitoring generator

The Sourcegraph monitoring generator uses [`Container`](#TODO) definitions in this package to generate integrations with [Sourcegraph's monitoring architecture](https://about.sourcegraph.com/handbook/engineering/observability/monitoring_architecture). It also aims to help codify guidelines defined in the [Sourcegraph monitoring pillars](https://about.sourcegraph.com/handbook/engineering/observability/monitoring_pillars).

This page primarily documents the [generator's current capabilities](#features) - in other words, and what you get for free by declaring Sourcegraph service monitoring in this package - as well as [how to make changes to the generator itself](#development).

To learn about how to find, add, and use monitoring, see the [Sourcegraph monitoring developer guide](https://about.sourcegraph.com/handbook/engineering/observability/monitoring).

## Features

TODO

## Development

The generator program is defined entirely in [`generator.go`](./generator.go).
