#!/usr/bin/env bash

set -euf -o pipefail

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

mkdir -p ~/sourcegraph-docker/prometheus-disk/logs

function finish {
  echo 'trapped CTRL-C: stopping docker prometheus container'
  docker stop $(cat ~/sourcegraph-docker/prometheus-disk/prom.cid)
  rm -f  ~/sourcegraph-docker/prometheus-disk/prom.cid
}
trap finish EXIT

# Description: Prometheus collects metrics and aggregates them into graphs.
#
#
docker run --rm --cidfile ~/sourcegraph-docker/prometheus-disk/prom.cid \
    --name=prometheus \
    -p 0.0.0.0:9090:9090 \
    -v ~/sourcegraph-docker/prometheus-disk:/prometheus \
    -v ${DIR}/prometheus/internal:/sg_add_ons \
    sourcegraph/prometheus:v2.12.0-1 >> ~/sourcegraph-docker/prometheus-disk/logs/prometheus.log 2>&1 &
wait $!
