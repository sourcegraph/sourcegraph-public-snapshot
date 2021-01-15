# Sourcegraph Prometheus

> WARNING: Looking for Prometheus documentation for Sourcegraph administrators?
> See the [metrics and dashboards documentation](../../../admin/observability/metrics.md#prometheus).

We ship a custom Prometheus image as part of a standard Sourcegraph distribution.
It currently bundles Alertmanager as well as integrations to the Sourcegraph web application.
Learn more about it in our [monitoring architecture](https://about.sourcegraph.com/handbook/engineering/observability/monitoring_architecture#sourcegraph-prometheus).

Adding recording rules, alerts, etc. to this image is handled by the [monitoring generator](./monitoring-generator.md).

The image is defined in [`docker-images/prometheus`](https://sourcegraph.com/github.com/sourcegraph/sourcegraph/-/tree/docker-images/prometheus).

## Metrics

See the [metrics and dashboards documentation](../../../admin/observability/metrics.md#grafana).

To learn more about developing metrics, see the [observability developer guides](./index.md#guides).

## Prom-wrapper

The entrypoint of the image is a sidecar program called the prom-wrapper.
Learn more about it [here](https://about.sourcegraph.com/handbook/engineering/observability/monitoring_architecture#prom-wrapper).

The source code for this program is currently kept in [`docker-images/prometheus/cmd/prom-wrapper`](https://sourcegraph.com/github.com/sourcegraph/sourcegraph/-/tree/docker-images/prometheus/cmd/prom-wrapper).

To learn more about developing our observability stack, see the [local Sourcegraph monitoring development guide](../../how-to/monitoring_local_dev.md).

## Upgrading Prometheus or Alertmanager

To upgrade Prometheus or Alertmanager:

1. Make the appropriate version and sum changes to the [`sourcegraph/prometheus` Dockerfile](https://sourcegraph.com/search?q=repo:%5Egithub%5C.com/sourcegraph/sourcegraph%24+FROM+prom/:%5Bimg%7Eprometheus%7Calertmanager%5D::%5Bversion.%5D+OR+FROM+prom/alertmanager::%5Bversion.%5D+OR+LABEL+com.sourcegraph.:%5Bimg%7Eprometheus%7Calertmanager%5D.version%3D:%5Bversion.%5D&patternType=structural)
1. Ensure no image update steps are required by checking upstream Dockerfiles where required as noted in the [`sourcegraph/prometheus` Dockerfile](https://sourcegraph.com/github.com/sourcegraph/sourcegraph/-/blob/docker-images/prometheus/Dockerfile) where appropriate
1. Upgrade the [Alertmanager and Prometheus Go client dependencies](https://sourcegraph.com/search?q=repo:%5Egithub%5C.com/sourcegraph/sourcegraph%24+file:go.mod+prometheus/alertmanager+OR+prometheus/client_golang&patternType=literal) where appropriate
   1. For the Alertmanager dependency, the fork needs to be upgraded to the appropriate version first: [`sourcegraph/alertmanager`](https://github.com/sourcegraph/alertmanager)
1. Ensure the image still builds: `./docker-images/prometheus/build.sh`
1. [Run the monitoring stack locally](../../how-to/monitoring_local_dev.md) and verify that:
   1. If upgrading Prometheus: all Prometheus rules are evaluated successfully (`localhost:9090/rules`)
   1. If upgrading Alertmanager: Alertmanager starts up correctly (`localhost:9090/alertmanager/#/status`), and [`observability.alerts` can be configured](../../../admin/observability/alerting.md) via the Sourcegraph web application
