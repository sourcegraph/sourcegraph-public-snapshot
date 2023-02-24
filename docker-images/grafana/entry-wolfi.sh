#!/usr/bin/env bash
set -e

export GF_PATHS_PROVISIONING=/sg_config_grafana/provisioning
export GF_PATHS_CONFIG=/sg_config_grafana/grafana.ini
export GF_PATHS_DATA=/var/lib/grafana
export GF_PATHS_HOME=/usr/share/grafana
export GF_PATHS_LOGS=/var/log/grafana
export GF_PATHS_PLUGINS=/sg_config_grafana/provisioning/plugins
export PATH=/usr/share/grafana/bin:/usr/sbin:/sbin:/usr/bin:/bin

exec "/run.sh"
