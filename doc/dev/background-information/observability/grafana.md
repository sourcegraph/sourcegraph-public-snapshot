# Sourcegraph Grafana

We ship a custom Grafana image as part of a standard Sourcegraph distribution.
Learn more about it in our [monitoring architecture](https://about.sourcegraph.com/handbook/engineering/observability/monitoring_architecture#sourcegraph-grafana).

Adding dashboards, panels, etc. to this image is handled by the [monitoring generator](./monitoring-generator.md).

The image is defined in [`docker-images/grafana`](https://sourcegraph.com/github.com/sourcegraph/sourcegraph/-/tree/docker-images/grafana).
