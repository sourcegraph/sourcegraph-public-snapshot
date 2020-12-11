# Generating Grafana dashboards with Jsonnet + Grafonnet

[Jsonnet](https://jsonnet.org/) is a data-templating language that produces YAML or JSON output. [Grafonnet](https://github.com/grafana/grafonnet-lib) is a small library on top of Jsonnet used to produce Grafana dashboards.

## Installation

1. Install Jsonnet: `brew install jsonnet`. This will install the binaries `jsonnet` and `jsonnetfmt` into your path.
2. Install Grafonnet by cloning it somewhere on your local machine: `git clone https://github.com/grafana/grafonnet-lib.git`.

## Building dashboards

Run `jsonnet` to execute a `.jsonnet` file and produce a JSON payload on standard out. This should be redirected to the correct directory for provisioned dashboards. For example:

```
jsonnet -J /path/to/grafonnet-lib ./dashboard.jsonnet > ../../../docker-images/grafana/config/provisioning/dashboards/sourcegraph/dashboard.json
```

## Libsonnet

The `common.libsonnet` file should be imported and used as a base for creating panels and graphs. It handles some common functionality around visualizing metrics conforming to [the](https://grafana.com/blog/2018/08/02/the-red-method-how-to-instrument-your-services/) [RED](https://www.weave.works/blog/the-red-method-key-metrics-for-microservices-architecture/) [method](https://thenewstack.io/monitoring-microservices-red-method/).

See the inline documentation for usage, or existing dashboard that imports it for an example.
Hello World
