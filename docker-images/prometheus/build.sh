#!/usr/bin/env bash

set -ex
cd "$(dirname "${BASH_SOURCE[0]}")"

rm -rf observability
cp -R ../../observability .

docker build --no-cache -t "${IMAGE:-sourcegraph/prometheus}" . \
  --progress=plain \
  --build-arg COMMIT_SHA \
  --build-arg DATE \
  --build-arg VERSION
