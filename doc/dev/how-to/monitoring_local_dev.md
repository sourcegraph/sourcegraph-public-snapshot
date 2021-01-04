# Set up local Sourcegraph monitoring development

This guide documents how to spin up and develop Sourcegraph's monitoring stack locally.
Sourcegraph employees should also refer to the [handbook's monitoring section](https://about.sourcegraph.com/handbook/engineering/observability/monitoring) for Sourcegraph-specific documentation.
The [developing observability page](../background-information/observability/index.md) contains relevant documentation as well, including background about the components listed here.

> NOTE: For how to *use* Sourcegraph's observability and an overview of our observability features, refer to the [observability for site administrators documentation](../../admin/observability/index.md).

## Running monitoring components

### With all services

The monitoring stack is included in the `./dev/start.sh` and `./enterprise/dev/start.sh` scripts.
Learn more about these in the [general development getting started guide](../getting-started/index.md).

### Without all services

For convenience, there are a number of ways to spin up Sourcegraph's monitoring services *without* having to start up every other service as well.

#### Grafana

Running just Grafana is a convenient way to validate dashboards.
When doing so, you may wish to connect Grafana to a remote Prometheus instance that you have administrator access to (such as [Sourcegraph's instances](https://about.sourcegraph.com/handbook/engineering/deployments/instances)), to show more real data than is available on your dev server.
For Kubernetes deployments, you can do this by getting `kubectl` connected to a Sourcegraph cluster and then port-forwarding Prometheus via:

```sh
kubectl port-forward svc/prometheus 9090:30090
```

Then, you can start up a standalone Grafana using:

```sh
./dev/grafana.sh
```

Dashboards will be available at `localhost:3030`.

Note that instead of `kubectl`, you can use whichever port-forwarding mechanism you wish to connect to a remote Prometheus instance as well, as long as Prometheus is available on port `9090` locally.
The dev targets for Grafana are defined in the following files:

* Non-Linux: [`dev/grafana/all/datasources.yaml`](https://sourcegraph.com/github.com/sourcegraph/sourcegraph/-/blob/dev/grafana/all/datasources.yaml)
* Linux: [`dev/grafana/linux/datasources.yaml`](https://sourcegraph.com/github.com/sourcegraph/sourcegraph/-/blob/dev/grafana/linux/datasources.yaml)

#### Prometheus

Running just Prometheus is a convenient way to validate the generated recording and alert rules.
You can start up a standalone Prometheus using:

```sh
./dev/prometheus.sh
```

The loaded generated recording and alert rules are available at `http://localhost:9090/rules`.
The bundled Alertmanager is available at `http://localhost:9090/alertmanager`.

Some configuration options are available:

* `DISABLE_SOURCEGRAPH_CONFIG`: when `true`, disables the prom-wrapper's [integration with the Sourcegraph frontend](#frontend-integration).
* `DISABLE_ALERTMANAGER`: when `true`, disables the bundled Alertmanager entirely.
  This includes the behaviour of `DISABLE_SOURCEGRAPH_CONFIG=true`.

Note that without services to scrape, running a standalone Prometheus will not provide any metrics - if you need to test metrics as well, you should [start all services](#with-all-services) instead.
The dev targets for Prometheus are defined in the following files:

* Non-Linux: [`dev/prometheus/all/prometheus_targets.yml`](https://sourcegraph.com/github.com/sourcegraph/sourcegraph/-/blob/dev/prometheus/all/prometheus_targets.yml)
* Linux: [`dev/prometheus/linux/prometheus_targets.yml`](https://sourcegraph.com/github.com/sourcegraph/sourcegraph/-/blob/dev/prometheus/linux/prometheus_targets.yml)

##### Frontend integration

The Sourcegraph Prometheus service features an integration with the Sourcegraph frontend that requires a frontend instance to be running to develop or test these features.
Note that the Prometheus service will still run without additional configuration even if no frontend is accessible.

One way to do this is to [start up Prometheus alongside all Sourcegraph services](#with-all-services).
You can alternatively spin up just the frontend separately:

```sh
./dev/start.sh --only frontend
```

This should be sufficient to access the frontend API and the admin console (`/site-admin`), which is where most of the integration is.

#### Docsite

The docsite is used to serve generated monitoring documentation, such as the [alert solutions reference](../../admin/observability/alert_solutions.md).
You can spin it up by running:

```sh
yarn docsite:serve
```

Learn more about docsite development in the [product documentation implementation guide](./documentation_implementation.md).

## Using the monitoring generator

> NOTE: Looking to add monitoring first? Refer to the [how to add monitoring](./add_monitoring.md) guide!

The dev startup scripts used in this guide all mount relevant configuration directories into each monitoring service.
This means that you can:

* Update your monitoring definitions
* Run the generator to regenerate and reload monitoring services
* Validate the result of your changes immediately (for example, by checking Prometheus rules in `/rules` or Grafana dashboards in `/-/debug/grafana`)

To run the generator and trigger a reload:

```sh
RELOAD=true go generate ./monitoring
```

Make sure to provide the following parameters as well, where relevant:

* `GRAFANA_DIR=''`, if you are *not* running Grafana
* `PROMETHEUS_DIR=''`, if you are *not* running Prometheus
* `SRC_LOG_LEVEL=dbug` to enable potentially helpful output for debugging issues
