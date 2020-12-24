# Sourcegraph Prometheus

We ship a custom Prometheus image as part of a standard Sourcegraph distribution.
It currently bundles Alertmanager as well as integrations to the Sourcegraph web application.
Learn more about it in our [monitoring architecture](https://about.sourcegraph.com/handbook/engineering/observability/monitoring_architecture#sourcegraph-prometheus).

Adding recording rules, alerts, etc. to this image is handled by the [monitoring generator](./monitoring-generator.md).

The image is defined in [`docker-images/prometheus`](https://sourcegraph.com/github.com/sourcegraph/sourcegraph/-/tree/docker-images/prometheus).

## Prom-wrapper

The entrypoint of the image is a sidecar program called the prom-wrapper.
Learn more about it [here](https://about.sourcegraph.com/handbook/engineering/observability/monitoring_architecture#prom-wrapper).

The source code for this program is currently kept in [`docker-images/prometheus/cmd/prom-wrapper`](https://sourcegraph.com/github.com/sourcegraph/sourcegraph/-/tree/docker-images/prometheus/cmd/prom-wrapper).

## Upgrading Prometheus or Alertmanager

To upgrade Prometheus or Alertmanager, make the appropriate version and sum changes to the [`sourcegraph/prometheus` Dockerfile](https://sourcegraph.com/search?q=repo:%5Egithub%5C.com/sourcegraph/sourcegraph%24+file:go.mod+prometheus/alertmanager+OR+prometheus/client_golang&patternType=literal) and make sure to:

* Upgrade the [Alertmanager and Prometheus Go client dependencies](https://sourcegraph.com/search?q=repo:%5Egithub%5C.com/sourcegraph/sourcegraph%24+file:go.mod+prometheus/alertmanager+OR+prometheus/client_golang&patternType=literal) where appropriate
* Ensure the image still builds: `./docker-images/prometheus/build.sh`
* [Run the monitoring stack locally](../../how-to/monitoring_local_dev.md) and verify that:
  * all Prometheus rules are evaluated successfully (`localhost:9090/rules`)
  * Alertmanager starts up correctly (`localhost:9090/alertmanager/#/status`)
  * [`observability.alerts` can be configured](../../../admin/observability/alerting.md) via the Sourcegraph web application
