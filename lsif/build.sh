#!/usr/bin/env bash

cd "$(dirname "${BASH_SOURCE[0]}")/.."
set -ex

docker build -f lsif/Dockerfile -t "$IMAGE" lsif \
    --progress=plain \
    --build-arg COMMIT_SHA \
    --build-arg DATE \
    --build-arg VERSION
