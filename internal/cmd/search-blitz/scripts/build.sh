#!/usr/bin/env bash

set -e
pushd "$(dirname "${BASH_SOURCE[0]}")/.." >/dev/null

if [ -z "$1" ]; then
  echo "USAGE $0 VERSION"
  echo "To see current versions visit https://console.cloud.google.com/gcr/images/sourcegraph-dev/us/search-blitz?project=sourcegraph-dev"
  gcloud container images list-tags us.gcr.io/sourcegraph-dev/search-blitz
  exit 1
fi

set -x

CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o searchblitz .

docker build \
  --platform linux/amd64 \
  --build-arg COMMIT_SHA="$(git rev-parse HEAD)" \
  -t "us.gcr.io/sourcegraph-dev/search-blitz:$1" \
  .
docker push "us.gcr.io/sourcegraph-dev/search-blitz:$1"

popd >/dev/null
