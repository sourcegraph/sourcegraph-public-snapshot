#!/usr/bin/env bash

set -e
pushd "$(dirname "${BASH_SOURCE[0]}")/../../../.." >/dev/null

if [ -z "$1" ]; then
  echo "USAGE $0 VERSION"
  echo "To see current versions visit https://console.cloud.google.com/gcr/images/sourcegraph-dev/us/search-blitz?project=sourcegraph-dev"
  exit 1
fi

set -x

docker build \
  -f ./internal/cmd/search-blitz/Dockerfile \
  --platform linux/amd64 \
  --build-arg COMMIT_SHA="$(git rev-parse HEAD)" \
  -t "us.gcr.io/sourcegraph-dev/search-blitz:$1" \
  .
docker push "us.gcr.io/sourcegraph-dev/search-blitz:$1"

popd >/dev/null
