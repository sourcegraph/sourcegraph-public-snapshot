#!/usr/bin/env bash

# Description: Dashboards and graphs for Grafana metrics.

set -euf -o pipefail
pushd "$(dirname "${BASH_SOURCE[0]}")/.." >/dev/null

GRAFANA_DISK="${HOME}/.sourcegraph-dev/data/grafana"

IMAGE=sourcegraph/grafana:61380_2020-04-17_046b2c8@sha256:7797b81d6e6dc061e745bfafbb89608d52b07107bc946139bc58f8295103b379
CONTAINER=grafana

mkdir -p ${GRAFANA_DISK}/logs

CONFIG_SUB_DIR="all"

if [[ "$OSTYPE" == "linux-gnu" ]]; then
  CONFIG_SUB_DIR="linux"
fi

docker inspect $CONTAINER >/dev/null 2>&1 && docker rm -f $CONTAINER

# Generate Grafana dashboards
pushd observability
DEV=true RELOAD=false go generate
popd

docker run --rm \
  --name=grafana \
  --cpus=1 \
  --memory=1g \
  --user=$UID \
  -p 0.0.0.0:3370:3370 \
  -v ${GRAFANA_DISK}:/var/lib/grafana \
  -v $(pwd)/dev/grafana/${CONFIG_SUB_DIR}:/sg_config_grafana/provisioning/datasources \
  -v $(pwd)/docker-images/grafana/jsonnet:/sg_grafana_additional_dashboards \
  ${IMAGE} >>${GRAFANA_DISK}/logs/grafana.log 2>&1 &
wait $!

# Add the following lines above if you wish to use an auth proxy with Grafana:
#
# -e GF_AUTH_PROXY_ENABLED=true \
# -e GF_AUTH_PROXY_HEADER_NAME='X-Forwarded-User' \
# -e GF_SERVER_ROOT_URL='https://grafana.example.com' \
