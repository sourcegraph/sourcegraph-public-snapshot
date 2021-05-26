# Monitoring Guide
Please visit our [Observability Docs](./observability) for more in-depth information about observability.

## What should I be monitoring?
Sourcegraph comes with built-in monitoring in the form of [Grafana](./observability/metrics#grafana) using [Prometheus](./observability/metrics#prometheus) for metrics and alerting. Generally, you can access Grafana by visiting `https://<your-sourcegraph-url>/-/debug/grafana` for monitoring purpose.

## When should I be monitoring?
Whenever you see an alert on the Grafana dashboard.

## What are the key values / alerts to look for when looking at the Grafana Dashboard?
All key values are defined as either warnings or critical. Please visit our [Observability Docs](./observability/alerting#understanding-alerts) to 
learn how they are defined. Warnings are typically not as important as critical but should be investigated. 
All alerts turn red when they fire. Critical alerts should be investigated and reported if they are occurring repeatedly.

## How do I know when more resources is needed for a specified service?
All resources contain a dashboard called `Provisioning indicators` that provide information about the current resource usage of containers. These can be used to determine if a scale-up is needed.

## What does this <ALERT-MESSAGE> mean?
See [Alert solutions](https://docs.sourcegraph.com/admin/observability/alert_solutions) to learn about each alert and their possible solutions. 

## What’s the threshold for each resource?
All resources have provisioning indicators that will fire up alerts when the container uses 80% of the container memory limit and 90% of the specified CPU limit.

## How much resources should I add after receiving alerts about running out of resources?
You should make the decision based on the metrics from your Grafana Dashboard. 

## What are some of the important alerts that I should beaware of?
All alerts are important, but more important when they are in red as that indicate there are issues occuring continuously that require attention.

## How to set up alerts?
See our Observability Docs on [setting up alerting](./observability/alerting#setting-up-alerting).

## How to create a custom alert?
Creating a customer alert is not recommended and currently not supported by Sourcegraph.
