# Metrics guide

## High-level alerting metrics

Sourcegraph's metrics include a single high-level metric `alert_count` which indicates the number of `level=critical` and `level=warning` alerts each service has fired over time for each Sourcegraph service. This is the same metric presented on the **Overview** Grafana dashboard:

![Overview Grafana dashboard screenshot](https://user-images.githubusercontent.com/3173176/71050700-21912400-2103-11ea-86fb-cf6d2dbd3d0a.png)

### `alert_count`

**Description:** The number of alerts each service has fired and their severity level. The severity levels are defined as follows:

- `critical`: something is _definitively_ wrong with Sourcegraph.
  - **Examples:** Database inaccessible, running out of disk space, running out of memory.
  - **Suggested action:** Page a site administrator to investigate.
- `warning`: something _could_ be wrong with Sourcegraph.
  - **Examples:** High latency, high search timeouts.
  - **Suggested action:** Email a site administrator to investigate and monitor when convenient.

**Values:**

- Although the values of `alert_count` are floating-point numbers, only their whole numbers have meaning. For example: `0.5` and `0.7` indicate no alerts are firing, while `1.2` indicates exactly one alert is firing and `3.0` indicates exactly three alerts firing.

**Labels:**

- `level`: either `critical` or `warning`, as defined above.
- `service_name`: the name of the service that fired the alert, one of the following constants:
  - `"frontend"`
  - `"github-proxy"`
  - `"gitserver"`
  - `"lsif-server"`
  - `"query-runner"`
  - `"replacer"`
  - `"repo-updater"`
  - `"searcher"`
  - `"symbols"`
  - `"zoekt-indexserver"`
  - `"zoekt-webserver"`
  - `"syntect-server"`
- `name`: the name of the alert that the service fired (chosen by the service)
- `description`: a human-readable description of the alert
- `instance`: identifies the Kubernetes pod, Docker container, or host machine from which the alert came.

## Complete reference

A complete reference of Sourcegraph's vast set of Prometheus metrics is not yet available. If you are interested in this, please reach out by filing an issue or contacting us at support@sourcegraph.com.
