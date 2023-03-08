#!/usr/bin/env bash

set -eou pipefail

cd "$(dirname "${BASH_SOURCE[0]}")"

IMAGE=${IMAGE:-sourcegraph/dind}

echo "--- docker build ${IMAGE}"
docker build -t "${IMAGE}" . \
  --progress=plain \
  --build-arg COMMIT_SHA \
  --build-arg DATE \
  --build-arg VERSION
