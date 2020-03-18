#!/usr/bin/env bash

# Description: Prometheus collects metrics and aggregates them into graphs.
#

set -euf -o pipefail

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

CONFIG_DIR="${DIR}/../docker-images/prometheus/config"

PROMETHEUS_DISK="${HOME}/.sourcegraph-dev/data/prometheus"

IMAGE=sourcegraph/prometheus:10.0.9@sha256:35a6d00e8291e365ef62a7a656fa674161582040116b9fa0612ce4baa1866bae
CONTAINER=prometheus

CID_FILE="${PROMETHEUS_DISK}/prometheus.cid"

mkdir -p ${PROMETHEUS_DISK}/logs
rm -f ${CID_FILE}

function finish {
  if test -f ${CID_FILE}; then
      echo 'trapped CTRL-C: stopping docker prometheus container'
      docker stop $(cat ${CID_FILE})
      rm -f  ${CID_FILE}
  fi
  rm -f ${CONFIG_DIR}/prometheus_targets.yml
  docker rm -f $CONTAINER
}
trap finish EXIT

NET_ARG=""
PROM_TARGETS="${DIR}/prometheus/all/prometheus_targets.yml"

if [[ "$OSTYPE" == "linux-gnu" ]]; then
   NET_ARG="--net=host"
   PROM_TARGETS="${DIR}/prometheus/linux/prometheus_targets.yml"
fi

cp ${PROM_TARGETS} ${CONFIG_DIR}/prometheus_targets.yml

docker inspect $CONTAINER > /dev/null 2>&1 && docker rm -f $CONTAINER
docker run --rm ${NET_ARG} --cidfile ${CID_FILE} \
    --name=prometheus \
    --cpus=1 \
    --memory=4g \
    --user=$UID \
    -p 0.0.0.0:9090:9090 \
    -v ${PROMETHEUS_DISK}:/prometheus \
    -v ${CONFIG_DIR}:/sg_prometheus_add_ons \
    -e PROMETHEUS_ADDITIONAL_FLAGS=--web.enable-lifecycle \
    ${IMAGE} >> ${PROMETHEUS_DISK}/logs/prometheus.log 2>&1 &
wait $!
