#!/bin/sh
set -e

CONFIG_FILE=/sg_config_prometheus/prometheus.yml

if test "$USE_KUBERNETES_DISCOVERY" = 'true'; then
    CONFIG_FILE=/sg_config_prometheus/prometheus_k8s.yml
fi

sh -c "/bin/prometheus --config.file=$CONFIG_FILE --storage.tsdb.path=/prometheus --web.console.libraries=/usr/share/prometheus/console_libraries --web.console.templates=/usr/share/prometheus/consoles"
