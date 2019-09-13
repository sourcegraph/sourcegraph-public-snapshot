#!/usr/bin/env bash

# Description: Dashboards and graphs for Grafana metrics.
#

set -euf -o pipefail

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

GRAFANA_DISK="${HOME}/.sourcegraph-dev/data/grafana"

IMAGE=sourcegraph/grafana:6.3.3-1
CONTAINER=grafana

CID_FILE="${GRAFANA_DISK}/grafana.cid"

mkdir -p ${GRAFANA_DISK}/logs
rm -f ${CID_FILE}

function finish {
  if test -f ${CID_FILE}; then
      echo 'trapped CTRL-C: stopping docker grafana container'
      docker stop $(cat ${CID_FILE})
      rm -f  ${CID_FILE}
  fi
  docker rm -f $CONTAINER
}
trap finish EXIT

CONFIG_SUB_DIR="all"

if [[ "$OSTYPE" == "linux-gnu" ]]; then
   CONFIG_SUB_DIR="linux"
fi

docker inspect $CONTAINER > /dev/null 2>&1 && docker rm -f $CONTAINER
docker run --rm  --cidfile ${CID_FILE} \
    --name=grafana \
    --cpus=1 \
    --memory=1g \
    --user=$UID \
    -p 0.0.0.0:3000:3000 \
    -v ${GRAFANA_DISK}:/var/lib/grafana \
    -v ${DIR}/grafana/${CONFIG_SUB_DIR}:/sg_config_grafana/provisioning/datasources \
    -v ${DIR}/../docker-images/grafana/config/provisioning/dashboards/sourcegraph:/sg_grafana_additional_dashboards \
    -e GF_AUTH_ANONYMOUS_ENABLED=true \
    -e GF_AUTH_ANONYMOUS_ORG_NAME='Main Org.' \
    -e GF_AUTH_ANONYMOUS_ORG_ROLE=Editor \
    -e GF_USERS_ALLOW_SIGN_UP='false' \
    -e GF_USERS_AUTO_ASSIGN_ORG='true' \
    -e GF_USERS_AUTO_ASSIGN_ORG_ROLE=Editor \
    ${IMAGE} >> ${GRAFANA_DISK}/logs/grafana.log 2>&1 &
wait $!

# Add the following lines above if you wish to use an auth proxy with Grafana:
#
# -e GF_AUTH_PROXY_ENABLED=true \
# -e GF_AUTH_PROXY_HEADER_NAME='X-Forwarded-User' \
# -e GF_SERVER_ROOT_URL='https://grafana.example.com' \
