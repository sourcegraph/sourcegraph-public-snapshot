#!/usr/bin/env bash

set -ex
cd "$(dirname "${BASH_SOURCE[0]}")"

ZOEKT_VERSION=$(go mod edit -print | awk '/sourcegraph\/zoekt/ {print substr($2, 2)}')

docker build --no-cache -f Dockerfile.wolfi -t "${IMAGE:-"sourcegraph/indexed-searcher"}" . \
  --progress=plain \
  --build-arg ZOEKT_VERSION="$ZOEKT_VERSION" \
  --build-arg COMMIT_SHA \
  --build-arg DATE \
  --build-arg VERSION
