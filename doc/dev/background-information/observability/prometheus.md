# Sourcegraph Prometheus

We ship a custom Prometheus image as part of a standard Sourcegraph distribution.
It currently bundles Alertmanager as well as integrations to the Sourcegraph web application.
Learn more about it in our [monitoring architecture](https://about.sourcegraph.com/handbook/engineering/observability/monitoring_architecture#sourcegraph-prometheus).

Adding recording rules, alerts, etc. to this image is handled by the [monitoring generator](./monitoring-generator.md).

The image is defined in [`docker-images/prometheus`](https://sourcegraph.com/github.com/sourcegraph/sourcegraph/-/tree/docker-images/prometheus).
