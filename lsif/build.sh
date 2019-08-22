#!/usr/bin/env bash

cd "$(dirname "${BASH_SOURCE[0]}")/../.."
set -ex

yarn --cwd lsif
yarn --cwd lsif run build

docker build -f lsif/Dockerfile -t "$IMAGE" lsif/out \
    --build-arg COMMIT_SHA \
    --build-arg DATE \
    --build-arg VERSION
