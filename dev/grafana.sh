#!/usr/bin/env bash

# Description: Dashboards and graphs for Grafana metrics.
#

set -euf -o pipefail

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

GRAFANA_DISK="${HOME}/.sourcegraph-dev/data/grafana"

IMAGE=sourcegraph/grafana:10.0.11
CONTAINER=grafana

mkdir -p ${GRAFANA_DISK}/logs

CONFIG_SUB_DIR="all"

if [[ "$OSTYPE" == "linux-gnu" ]]; then
   CONFIG_SUB_DIR="linux"
fi

docker inspect $CONTAINER > /dev/null 2>&1 && docker rm -f $CONTAINER
docker run --rm \
    --name=grafana \
    --cpus=1 \
    --memory=1g \
    --user=$UID \
    -p 0.0.0.0:3370:3370 \
    -v ${GRAFANA_DISK}:/var/lib/grafana \
    -v ${DIR}/grafana/${CONFIG_SUB_DIR}:/sg_config_grafana/provisioning/datasources \
    -v ${DIR}/../docker-images/grafana/jsonnet:/sg_grafana_additional_dashboards \
    ${IMAGE} >> ${GRAFANA_DISK}/logs/grafana.log 2>&1 &
wait $!

# Add the following lines above if you wish to use an auth proxy with Grafana:
#
# -e GF_AUTH_PROXY_ENABLED=true \
# -e GF_AUTH_PROXY_HEADER_NAME='X-Forwarded-User' \
# -e GF_SERVER_ROOT_URL='https://grafana.example.com' \
