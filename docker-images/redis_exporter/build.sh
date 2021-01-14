#!/usr/bin/env bash

set -ex
cd "$(dirname "${BASH_SOURCE[0]}")"

# Retag the v1.15.1 redis-exporter release
VERSION="v1.15.1@sha256:f3f51453e4261734f08579fe9c812c66ee443626690091401674be4fb724da70"
docker pull oliver006/redis_exporter:$VERSION
docker tag oliver006/redis_exporter:$VERSION "$IMAGE"
