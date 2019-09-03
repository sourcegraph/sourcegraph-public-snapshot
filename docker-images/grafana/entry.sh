#!/bin/bash
set -e

export GF_PATHS_PROVISIONING=/sg_config_grafana/provisioning

export GF_PATHS_CONFIG=/sg_config_grafana/grafana.ini

if test "$USE_KUBERNETES_DISCOVERY" = 'true'; then
    export GF_PATHS_CONFIG=/sg_config_grafana/grafana_k8s.ini
fi

exec "/run.sh"
