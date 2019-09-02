#!/usr/bin/env bash

set -euf -o pipefail

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

mkdir -p ~/sourcegraph-docker/prometheus-disk/logs
rm -f ~/sourcegraph-docker/prometheus-disk/prom.cid

function finish {
  echo 'trapped CTRL-C: stopping docker prometheus container'
  docker stop $(cat ~/sourcegraph-docker/prometheus-disk/prom.cid)
  rm -f  ~/sourcegraph-docker/prometheus-disk/prom.cid
}
trap finish EXIT

NET_ARG=""
CONFIG_SUB_DIR="internal"

if [[ "$OSTYPE" == "linux-gnu" ]]; then
   NET_ARG="--net=host"
   CONFIG_SUB_DIR="local"
fi

# Description: Prometheus collects metrics and aggregates them into graphs.
#
#
docker run --rm ${NET_ARG} --cidfile ~/sourcegraph-docker/prometheus-disk/prom.cid \
    --name=prometheus \
    --cpus=4 \
    --memory=8g \
    -p 0.0.0.0:9090:9090 \
    -v ~/sourcegraph-docker/prometheus-disk:/prometheus \
    -v ${DIR}/prometheus/${CONFIG_SUB_DIR}:/sg_add_ons \
    sourcegraph/prometheus:3.8 >> ~/sourcegraph-docker/prometheus-disk/logs/prometheus.log 2>&1 &
wait $!
