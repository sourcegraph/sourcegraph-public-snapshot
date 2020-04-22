#!/usr/bin/env bash

set -ex

cd "$(dirname "${BASH_SOURCE[0]}")"

# We copy just the monitoring directory and the root go.mod/go.sum so that we
# do not need to send the entire repository as build context to Docker. Additionally,
# we do not use a separate go.mod/go.sum in the monitoring/ directory because
# editor tooling would occassionally include and not include it in the root
# go.mod/go.sum.
rm -rf monitoring
cp -R ../../monitoring .
cp ../../go.* ./monitoring

docker build --no-cache -t "${IMAGE:-sourcegraph/grafana}" . \
  --progress=plain \
  --build-arg COMMIT_SHA \
  --build-arg DATE \
  --build-arg VERSION
