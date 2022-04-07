#!/usr/bin/env bash

set -ex
pushd "$(dirname "${BASH_SOURCE[0]}")/../../../.." >/dev/null

docker build \
  -f ./internal/cmd/search-blitz/Dockerfile \
  --platform linux/amd64 \
  --build-arg COMMIT_SHA="$(git rev-parse HEAD)" \
  -t "us.gcr.io/sourcegraph-dev/search-blitz:$1" \
  .
docker push "us.gcr.io/sourcegraph-dev/search-blitz:$1"

popd >/dev/null
