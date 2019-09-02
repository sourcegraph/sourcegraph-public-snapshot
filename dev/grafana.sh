#!/usr/bin/env bash

set -euf -o pipefail

GRAFANA_DISK="${HOME}/sourcegraph-docker/grafana-disk"

CID_FILE="${GRAFANA_DISK}/grafana.cid"

mkdir -p ${GRAFANA_DISK}/logs
rm -f ${CID_FILE}

function finish {
  if test -f ${CID_FILE}; then
      echo 'trapped CTRL-C: stopping docker grafana container'
      docker stop $(cat ${CID_FILE})
      rm -f  ${CID_FILE}
  fi
}
trap finish EXIT

# Description: Dashboards and graphs for grafana metrics.
#
docker run --rm  --cidfile ${CID_FILE} \
    --name=grafana \
    --cpus=1 \
    --memory=1g \
    -p 0.0.0.0:3000:3000 \
    -v ${GRAFANA_DISK}:/var/lib/grafana \
    -e GF_AUTH_ANONYMOUS_ENABLED=true \
    -e GF_AUTH_ANONYMOUS_ORG_NAME=Sourcegraph \
    -e GF_AUTH_ANONYMOUS_ORG_ROLE=Editor \
    -e GF_USERS_ALLOW_SIGN_UP='false' \
    -e GF_USERS_AUTO_ASSIGN_ORG='true' \
    -e GF_USERS_AUTO_ASSIGN_ORG_ROLE=Editor \
    sourcegraph/grafana:3.8 >> ${GRAFANA_DISK}/logs/grafana.log 2>&1 &

# Add the following lines above if you wish to use an auth proxy with Grafana:
#
# -e GF_AUTH_PROXY_ENABLED=true \
# -e GF_AUTH_PROXY_HEADER_NAME='X-Forwarded-User' \
# -e GF_SERVER_ROOT_URL='https://grafana.example.com' \
