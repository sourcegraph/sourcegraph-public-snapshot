# Prometheus image

Vanilla Prometheus image with one addition: embedded Sourcegraph configs.

# Image API

```shell script
docker run \
    -v ${PROMETHEUS_DISK}:/prometheus \
    -v ${CONFIG_DIR}:/sg_prometheus_add_ons \
    sourcegraph/prometheus
```

Image expects two volumes mounted:

- at `/prometheus` a data directory where logs, the tsdb and other prometheus data files will live
- at `/sg_prometheus_add_ons` a directory that contains additional config files of two types:
  - rule files which must have the suffix `_rules.yml` in their filename (ie `gitserver_rules.yml`)
  - target files which must have the suffix `_targets.yml` in their filename (ie `local_targets.yml`)
  - if this directory contains a file named `prometheus.yml` it will be used as the main prometheus config file
