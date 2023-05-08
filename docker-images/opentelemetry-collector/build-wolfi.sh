#!/usr/bin/env bash

set -ex
cd "$(dirname "${BASH_SOURCE[0]}")"

docker build -f Dockerfile.wolfi -t "${IMAGE:-sourcegraph/opentelemetry-collector}" . \
  --platform linux/amd64 \
  --build-arg COMMIT_SHA \
  --build-arg DATE \
  --build-arg VERSION
