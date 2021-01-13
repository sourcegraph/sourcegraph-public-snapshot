# Sourcegraph cAdvisor

We ship a custom [cAdvisor](https://github.com/google/cadvisor) image as part of the standard Sourcegraph Kubernetes and docker-compose distribution.
cAdvisor exports container monitoring metrics scraped by [Prometheus](./prometheus.md) and visualized in [Grafana](./grafana.md).

The image is defined in [`docker-images/cadvisor`](https://sourcegraph.com/github.com/sourcegraph/sourcegraph/-/tree/docker-images/cadvisor).

## Monitoring

Monitoring on cAdvisor metrics is defined in the [monitoring generator](./monitoring-generator.md).
cAdvisor observables are generally defined as [shared observables](https://sourcegraph.com/github.com/sourcegraph/sourcegraph/-/tree/monitoring/definitions/shared).

## Identifying containers

How relevant containers are identified from exported cAdvisor metrics is documented in [`CadvisorNameMatcher`](https://sourcegraph.com/search?q=repo:%5Egithub%5C.com/sourcegraph/sourcegraph%24+type:symbol+CadvisorNameMatcher&patternType=literal).

## Available metrics

Exported metrics are documented in the [cAdvisor Prometheus metrics list](https://github.com/google/cadvisor/blob/master/docs/storage/prometheus.md#prometheus-container-metrics).
In the list, the column `-disable_metrics parameter` indicates the "group" the metric belongs in.

Container runtime and deployment environment compatability for various metrics seem to be grouped by these groups - before using a metric, ensure that the metric is supported in all relevant environments (for example, both Docker and `containerd` container runtimes).
Support is generally poorly documented, but a search through the [cAdvisor repository issues](https://github.com/google/cadvisor/issues) might provide some hints.
