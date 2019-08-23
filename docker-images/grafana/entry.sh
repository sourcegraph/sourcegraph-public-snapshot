#!/bin/bash
set -e

export GF_PATHS_PROVISIONING=/sg_config/provisioning

export GF_PATHS_CONFIG=/sg_config/grafana.ini

if test "$USE_KUBERNETES_DISCOVERY" = 'true'; then
    export GF_PATHS_CONFIG=/sg_config/grafana_k8s.ini
fi

exec "/run.sh"
