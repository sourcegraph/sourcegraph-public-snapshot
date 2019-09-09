# Grafana image

Vanilla Grafana image with one addition: embedded Sourcegraph provisioning.

# Image API

```shell script
docker run  \
    -v ${GRAFANA_DISK}:/var/lib/grafana \
    -v %{GRAFANA_DATA_SOURCES}:/sg_config_grafana/provisioning/datasources \
    sourcegraph/grafana
```

Image expects two volumes mounted:

- at `/var/lib/grafana` a data directory where logs, the Grafana db and other Grafana data files will live
- at `/sg_config_grafana/provisioning/datasources` a directory with data source yaml files.

Additional behavior can be controlled with
[environmental variables](https://grafana.com/docs/installation/configuration/).
