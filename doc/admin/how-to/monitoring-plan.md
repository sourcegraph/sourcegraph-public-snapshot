# Monitoring Guide
Please visit our [Observability Docs](./observability) for more in-depth information about observability.

## What should I be monitoring?
Sourcegraph comes with built-in monitoring in the form of Grafana. monitoring dashboard via Grafana.

## When should I be monitoring?
Whenever you see an alert from the Grafana dashboard.

## What are the key values / alerts to look for when looking at the Grafana Dashboard?
All key values are defined as either warnings or critical. Please visit our [Observability Docs](./observability/alerting#understanding-alerts) to 
learn how they are defined. Warnings are typically not as important as critical but should be investigated. 
All alerts turn red when they fire. Critical alerts should be investigated and reported if they are occurring repeatedly.

## How do I know when more resources is needed for a specified service?
All resources contain a dashboard called `Provisioning indicators` that provide information about the current resource usage of containers. These can be used to determine if a scale-up is needed.

## What’s the threshold for each resource?
All resources have provisioning indicators that will fire up alerts when the container uses 80% of the container memory limit and 90% of the specified CPU limit.

## How much more resources should I add after receiving the alerts?


## What are some of the important alerts that I should beaware of?
All alerts are important, but more important when they are in red as that indicates there is an issue occuring continuously.

## How to set up alerts?
See our Observability Docs on [setting up alerting](./observability/alerting#setting-up-alerting).

## How to create a custom alert?
Creating a customer alert is not recommended and currently not supported by Sourcegraph.

## Troubleshooting
