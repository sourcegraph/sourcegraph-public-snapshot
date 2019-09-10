#!/usr/bin/env bash

set -euf -o pipefail

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

PROMETHEUS_DISK="${HOME}/.sourcegraph-dev/data/prometheus"
CID_FILE="${PROMETHEUS_DISK}/prometheus.cid"

mkdir -p ${PROMETHEUS_DISK}/logs
rm -f ${CID_FILE}

function finish {
  if test -f ${CID_FILE}; then
      echo 'trapped CTRL-C: stopping docker prometheus container'
      docker stop $(cat ${CID_FILE})
      rm -f  ${CID_FILE}
  fi
}
trap finish EXIT

NET_ARG=""
CONFIG_SUB_DIR="all"

if [[ "$OSTYPE" == "linux-gnu" ]]; then
   NET_ARG="--net=host"
   CONFIG_SUB_DIR="linux"
fi

# Description: Prometheus collects metrics and aggregates them into graphs.
#
#
docker run --rm ${NET_ARG} --cidfile ${CID_FILE} \
    --name=prometheus \
    --cpus=4 \
    --memory=4g \
    --user=$UID \
    -p 0.0.0.0:9090:9090 \
    -v ${PROMETHEUS_DISK}:/prometheus \
    -v ${DIR}/prometheus/${CONFIG_SUB_DIR}:/sg_prometheus_add_ons \
    sourcegraph/prometheus:v2.12.0 >> ${PROMETHEUS_DISK}/logs/prometheus.log 2>&1 &
wait $!
