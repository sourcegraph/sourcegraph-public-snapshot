#!/usr/bin/env bash

cd "$(dirname "${BASH_SOURCE[0]}")"
set -ex

# Retag the upstream node-exporter release
VERSION="v1.5.0@sha256:fa8e5700b7762fffe0674e944762f44bb787a7e44d97569fe55348260453bf80"

docker pull prom/node-exporter:$VERSION
docker tag prom/node-exporter:$VERSION "$IMAGE"
