# Sourcegraph Grafana

We ship a custom Grafana image as part of a standard Sourcegraph distribution.
Learn more about it in our [monitoring architecture](https://about.sourcegraph.com/handbook/engineering/observability/monitoring_architecture#sourcegraph-grafana).

Adding dashboards, panels, etc. to this image is handled by the [monitoring generator](./monitoring-generator.md).

The image is defined in [`docker-images/grafana`](https://sourcegraph.com/github.com/sourcegraph/sourcegraph/-/tree/docker-images/grafana).

## Upgrading Grafana

To upgrade Grafana, make the appropriate version change to the [`sourcegraph/grafana` Dockerfile](https://sourcegraph.com/search?q=repo:%5Egithub%5C.com/sourcegraph/sourcegraph%24+FROM+grafana/grafana::%5Bversion.%5D&patternType=structural) and:

* Ensure the image still builds: `./docker-images/grafana/build.sh`
* [Run the monitoring stack locally](../../how-to/monitoring_local_dev.md) and verify that all generated Grafana dashboards still render correctly
