#!/usr/bin/env bash

set -ex

cd "$(dirname "${BASH_SOURCE[0]}")"

rm -rf monitoring
cp -R ../../monitoring .

docker build --no-cache -t "${IMAGE:-sourcegraph/grafana}" . \
  --progress=plain \
  --build-arg COMMIT_SHA \
  --build-arg DATE \
  --build-arg VERSION
