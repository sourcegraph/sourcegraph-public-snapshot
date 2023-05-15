#!/usr/bin/env bash

set -euo pipefail

cd "$(dirname "${BASH_SOURCE[0]}")"

IMAGE=${IMAGE:-sourcegraph/jaeger-agent}

docker build --no-cache -f Dockerfile.wolfi -t "${IMAGE}" . \
  --progress=plain \
  --build-arg COMMIT_SHA \
  --build-arg DATE \
  --build-arg VERSION
