# Executor Monitoring Architecture Overview

![image](imgs/executor_monitoring.svg)

[QubitProducts/exporter_exporter](https://github.com/QubitProducts/exporter_exporter) to proxy all targets to a single port. Simplifies the Prometheus configuration. Untested, but may require additional relabeling etc to reshape the metrics.

[prometheus/node_exporter](https://github.com/prometheus/node_exporter) to surface compute machine metrics.
