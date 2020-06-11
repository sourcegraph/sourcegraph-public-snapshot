#!/usr/bin/env bash

# Description: Dashboards and graphs for Grafana metrics.

set -euf -o pipefail
pushd "$(dirname "${BASH_SOURCE[0]}")/.." >/dev/null

GRAFANA_DISK="${HOME}/.sourcegraph-dev/data/grafana"
IMAGE=sourcegraph/grafana:dev
CONTAINER=grafana

mkdir -p "${GRAFANA_DISK}"/logs

# quickly build image - should do this because the image has a Sourcegraph wrapper program
# see /docker-images/grafana/cmd/grafana-wrapper for more details
IMAGE=${IMAGE} CACHE=true ./docker-images/grafana/build.sh

# docker containers must access things via docker host on non-linux platforms
CONFIG_SUB_DIR="all"
SRC_FRONTEND_INTERNAL="host.docker.internal:3090"

if [[ "$OSTYPE" == "linux-gnu" ]]; then
  SRC_FRONTEND_INTERNAL="localhost:3090"
  CONFIG_SUB_DIR="linux"
fi

docker inspect $CONTAINER >/dev/null 2>&1 && docker rm -f $CONTAINER

# Generate Grafana dashboards
pushd monitoring
DEV=true RELOAD=false go generate
popd

docker run --rm \
  --name=grafana \
  --cpus=1 \
  --memory=1g \
  -p 0.0.0.0:3370:3370 \
  -v "${GRAFANA_DISK}":/var/lib/grafana \
  -v "$(pwd)"/dev/grafana/${CONFIG_SUB_DIR}:/sg_config_grafana/provisioning/datasources \
  -v "$(pwd)"/docker-images/grafana/jsonnet:/sg_grafana_additional_dashboards \
  -e SRC_FRONTEND_INTERNAL="${SRC_FRONTEND_INTERNAL}" \
  ${IMAGE} >>"${GRAFANA_DISK}"/logs/grafana.log 2>&1 &
wait $!

# Add the following lines above if you wish to use an auth proxy with Grafana:
#
# -e GF_AUTH_PROXY_ENABLED=true \
# -e GF_AUTH_PROXY_HEADER_NAME='X-Forwarded-User' \
# -e GF_SERVER_ROOT_URL='https://grafana.example.com' \
