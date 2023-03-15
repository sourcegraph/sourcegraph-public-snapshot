#!/usr/bin/env bash

# This script builds the build-tracker docker image.

cd "$(dirname "${BASH_SOURCE[0]}")/../.."
set -eu

IMAGE="us-central1-docker.pkg.dev/sourcegraph-ci/build-tracker/build-tracker"

echo "--- docker build build-tracker $(pwd)"
docker build -f dev/build-tracker/Dockerfile -t "$IMAGE" "$(pwd)" \

#docker push $IMAGE
