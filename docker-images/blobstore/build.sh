#!/usr/bin/env bash
cd "$(dirname "${BASH_SOURCE[0]}")"
set -ex

docker build -t "${IMAGE:-"sourcegraph/blobstore"}" . \
  --platform linux/amd64 \
  --progress=plain \
  --build-arg COMMIT_SHA \
  --build-arg DATE \
  --build-arg VERSION
