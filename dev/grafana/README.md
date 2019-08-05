# Grafana configuration

This directory contains configuration for the [Grafana](https://grafana.com/) dashboards deployment.

You need to add a source manually. Go to "Settings" -> "Data Sources" and add a prometheus source with URL "http://host.docker.internal:9090".

This directory is mounted into the `grafana` container. After making your changes to this directory,
simply `docker restart grafana` for your changes to take effect.
