# Monitoring Guide

Please visit our [Observability Docs](https://docs.sourcegraph.com/admin/observability) for more in-depth information about observability in Sourcegraph.

## Prerequisites

This document assumes that you are a [site administrator](https://docs.sourcegraph.com/admin).

## FAQs

### What should I look at when my instance is having performance issues?

Sourcegraph comes with built-in monitoring in the form of [Grafana](https://docs.sourcegraph.com/admin/observability/metrics#grafana), connected to [Prometheus](https://docs.sourcegraph.com/admin/observability/metrics#prometheus) for metrics and alerting.

Generally, Grafana should be the first stop you make when experiencing a system performance issue. From there you can look for system alerts or metrics that would provide you with more insights on what’s causing the performance issue. You can learn more about [accessing Grafana here](https://docs.sourcegraph.com/admin/observability/metrics#grafana).

### What are the key values/alerts to look for when looking at the Grafana Dashboard?

Please refer to the [Dashboards](https://docs.sourcegraph.com/admin/observability/metrics#dashboards) guide for more on how to use our Grafana dashboards.

Please refer to [Understanding alerts](https://docs.sourcegraph.com/admin/observability/alerting#understanding-alerts) for examples and suggested actions for alerts.

### How do I know when more resources are needed for a specified service?

All resource dashboards contain a section called `Provisioning indicators` that provide information about the current resource usage of containers. These can be used to determine if a scale-up is needed ([example panel](https://docs.sourcegraph.com/admin/observability/dashboards#frontend-provisioning-container-cpu-usage-long-term)).

More information on each available panel in the dashboards is available in the [Dashboards reference](https://docs.sourcegraph.com/admin/observability/dashboards).

### What does this `<ALERT-MESSAGE>` mean?

See [Alert solutions](https://docs.sourcegraph.com/admin/observability/alert_solutions) to learn about each alert and their possible solutions.

### What’s the threshold for each resource?

All resources dashboards contain a section called `Container monitoring` that indicate thresholds at which alerts will fire for each resource ([example alert](https://docs.sourcegraph.com/admin/observability/alert_solutions#frontend-container-cpu-usage)).

More information on each available panel in the dashboards is available in the [Dashboards reference](https://docs.sourcegraph.com/admin/observability/dashboards).

### How much resources should I add after receiving alerts about running out of resources?

You should make the decision based on the metrics from the relevant Grafana dashboard linked in each alert.
  
### What are some of the important alerts that I should be aware of?

We recommend paying closer attention to [critical alerts](https://docs.sourcegraph.com/admin/observability/alerting#understanding-alerts).

### How do I set up alerts?

Please refer to our guide on [setting up alerting](https://docs.sourcegraph.com/admin/observability/alerting#setting-up-alerting).

### How do I create a custom alert?

Creating a custom alert is not recommended and currently not supported by Sourcegraph. However, please provide feedback on the monitoring dashboards and alerts if you find anything could be improved via our issue tracker.

More advanced users can also refer to [our FAQ item about custom consumption of Sourcegraph metrics](https://docs.sourcegraph.com/admin/faq#can-i-consume-sourcegraph-s-metrics-in-my-own-monitoring-system-datadog-new-relic-etc).
