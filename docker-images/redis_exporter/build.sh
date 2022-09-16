#!/usr/bin/env bash

set -ex
cd "$(dirname "${BASH_SOURCE[0]}")"

# Retag the v1.35.0 redis-exporter release
VERSION="v1.35.0@sha256:edb0c9b19cacd90acc78f13f0908a7e6efd1df704e401805c24bffd241285f70"
docker pull oliver006/redis_exporter:$VERSION
docker tag oliver006/redis_exporter:$VERSION "$IMAGE"
