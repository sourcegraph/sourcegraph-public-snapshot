#!/usr/bin/env bash
set -e

export GF_PATHS_PROVISIONING=/sg_config_grafana/provisioning
export GF_PATHS_CONFIG=/sg_config_grafana/grafana.ini

exec grafana-server \
  --homepath="$GF_PATHS_HOME" \
  --config="$GF_PATHS_CONFIG" \
  --packaging=docker \
  "$@" \
  cfg:default.log.mode="console" \
  cfg:default.paths.data="$GF_PATHS_DATA" \
  cfg:default.paths.logs="$GF_PATHS_LOGS" \
  cfg:default.paths.plugins="$GF_PATHS_PLUGINS" \
  cfg:default.paths.provisioning="$GF_PATHS_PROVISIONING"
