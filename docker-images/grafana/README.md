# Grafana image

Vanilla Grafana image with two additions: provisioned Sourcegraph dashboards and config, and a wrapper program that polls the frontend so alerts can be configured in Sourcegraph site configuration nicely. For more details, refer to [the handbook](https://about.sourcegraph.com/handbook/engineering/distribution/observability/monitoring#grafana).

## Image API

```shell script
docker run  \
    -v ${GRAFANA_DISK}:/var/lib/grafana \
    -v %{GRAFANA_DATA_SOURCES}:/sg_config_grafana/provisioning/datasources \
    sourcegraph/grafana
```

Image expects two volumes mounted:

- at `/var/lib/grafana` a data directory where logs, the Grafana db and other Grafana data files will live
- at `/sg_config_grafana/provisioning/datasources` a directory with data source yaml files.

A directory containing dashboard json specifications can be mounted at
`/sg_grafana_additional_dashboards` and they will be picked up automatically. Changes to files in that directory
will be detected automatically while Grafana is running.

More behavior can be controlled with
[environmental variables](https://grafana.com/docs/installation/configuration/).

Note that to run Grafana without access to the frontend, set `DISABLE_SOURCEGRAPH_CONFIG=true` to disable Sourcegraph site config subscription. For more details, see [`cmd/grafana-wrapper`](./cmd/grafana-wrapper/main.go).
