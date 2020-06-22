#!/bin/bash
set -e

export GF_PATHS_PROVISIONING=/sg_config_grafana/provisioning
export GF_PATHS_CONFIG=/sg_config_grafana/grafana.ini

# create relevant directories if they are not available - the grafana startup
# script expects these to be present.
function create_if_set() {
  DIR=$1
  if
    [ -n "$DIR" ] &
    [ ! -d "$DIR" ]
  then
    echo "Creating $DIR"
    mkdir -p "$DIR"
  fi
}

create_if_set "$GF_PATHS_DATA"
create_if_set "$GF_PATHS_HOME"

exec "/usr/share/grafana-wrapper"
