#!/usr/bin/env bash
cd "$(dirname "${BASH_SOURCE[0]}")"
set -ex

docker build --no-cache -f Dockerfile.wolfi -t "${IMAGE:-"sourcegraph/cadvisor"}" . \
  --progress=plain \
  --build-arg COMMIT_SHA \
  --build-arg DATE \
  --build-arg VERSION
