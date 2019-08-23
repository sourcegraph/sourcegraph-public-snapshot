#!/usr/bin/env bash

cd "$(dirname "${BASH_SOURCE[0]}")/../.."
set -ex

yarn --cwd lsif/server --frozen-lockfile
yarn --cwd lsif/server run build

docker build -f lsif/server/Dockerfile -t "$IMAGE" lsif/server/out \
    --build-arg COMMIT_SHA \
    --build-arg DATE \
    --build-arg VERSION
