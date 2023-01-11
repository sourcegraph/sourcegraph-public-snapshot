#!/usr/bin/env bash

cd "$(dirname "${BASH_SOURCE[0]}")"
set -ex

# Retag the upstream node-exporter release
VERSION="v1.4.0@sha256:4dc469c325388dee18dd0a9e53ea30194abed43abc6330d4ffd6d451727ba3e6"

docker pull prom/node-exporter:$VERSION
docker tag prom/node-exporter:$VERSION "$IMAGE"
