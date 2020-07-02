#!/usr/bin/env bash

# Description: Prometheus collects metrics and aggregates them into graphs.
# Most of this script is very similar to /dev/grafana.sh - refer to inline documentation there.

set -euf -o pipefail
pushd "$(dirname "${BASH_SOURCE[0]}")/.." >/dev/null

PROMETHEUS_DISK="${HOME}/.sourcegraph-dev/data/prometheus"
if [ ! -e "${PROMETHEUS_DISK}" ]; then
  mkdir -p ${PROMETHEUS_DISK}
fi

IMAGE=sourcegraph/prometheus:61407_2020-04-18_9aa5791@sha256:9b31cc8832defb66cd29e5a298422983179495193745324902660073b3fdc835
CONTAINER=prometheus

CONFIG_DIR="$(pwd)/docker-images/prometheus/config"
DOCKER_NET=""
DOCKER_USER=""
PROM_TARGETS="dev/prometheus/all/prometheus_targets.yml"

if [[ "$OSTYPE" == "linux-gnu" ]]; then
  DOCKER_USER="--user=$UID"
  DOCKER_NET="--net=host"
  PROM_TARGETS="dev/prometheus/linux/prometheus_targets.yml"
fi

docker inspect $CONTAINER >/dev/null 2>&1 && docker rm -f $CONTAINER

cp ${PROM_TARGETS} "${CONFIG_DIR}"/prometheus_targets.yml

pushd monitoring
DEV=true RELOAD=false go generate
popd

PROMETHEUS_LOGS="${HOME}/.sourcegraph-dev/logs/prometheus"
mkdir -p "${PROMETHEUS_LOGS}"
PROMETHEUS_LOG_FILE="${PROMETHEUS_LOGS}/prometheus.log"

function finish() {
  PROMETHEUS_EXIT_CODE=$?

  if [ $PROMETHEUS_EXIT_CODE -ne 0 ] && [ $PROMETHEUS_EXIT_CODE -ne 2 ]; then
    echo "Prometheus exited with unexpected code ${PROMETHEUS_EXIT_CODE}; dumping log:"
    cat "${PROMETHEUS_LOG_FILE}"
  fi

  rm -f "${CONFIG_DIR}"/prometheus_targets.yml
  return $PROMETHEUS_EXIT_CODE
}

docker run --rm ${DOCKER_NET} ${DOCKER_USER} \
  --name=${CONTAINER} \
  --cpus=1 \
  --memory=4g \
  -p 0.0.0.0:9090:9090 \
  -v "${PROMETHEUS_DISK}":/prometheus \
  -v "${CONFIG_DIR}":/sg_prometheus_add_ons \
  -e PROMETHEUS_ADDITIONAL_FLAGS=--web.enable-lifecycle \
  ${IMAGE} >"${PROMETHEUS_LOG_FILE}" 2>&1 || finish
