#!/usr/bin/env bash
set -e

# Description: Dashboards and graphs for Prometheus metrics.
#
#
docker run --detach \
    --name=grafana \
    --cpus=1 \
    --memory=1g \
    -p 0.0.0.0:3000:3000 \
    -v ~/sourcegraph-docker/grafana-disk:/var/lib/grafana \
    -v $(pwd)/grafana:/etc/grafana \
    -e GF_AUTH_ANONYMOUS_ENABLED=true \
    -e GF_AUTH_ANONYMOUS_ORG_NAME=Sourcegraph \
    -e GF_AUTH_ANONYMOUS_ORG_ROLE=Editor \
    -e GF_USERS_ALLOW_SIGN_UP='false' \
    -e GF_USERS_AUTO_ASSIGN_ORG='true' \
    -e GF_USERS_AUTO_ASSIGN_ORG_ROLE=Editor \
    grafana/grafana:6.1.1@sha256:e7a513bf7f33ef9681b2d35a799136e1ce9330f9055f75dfa2101d812946184b

# Add the following lines above if you wish to use an auth proxy with Grafana:
#
# -e GF_AUTH_PROXY_ENABLED=true \
# -e GF_AUTH_PROXY_HEADER_NAME='X-Forwarded-User' \
# -e GF_SERVER_ROOT_URL='https://grafana.example.com' \
