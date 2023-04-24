#!/usr/bin/env bash

set -euo pipefail

cd "$(dirname "${BASH_SOURCE[0]}")"

IMAGE=${IMAGE:-sourcegraph/jaeger-all-in-one}

docker build --no-cache -f Dockerfile.wolfi -t "${IMAGE}" . \
  --progress=plain \
  --build-arg COMMIT_SHA \
  --build-arg DATE \
  --build-arg VERSION
