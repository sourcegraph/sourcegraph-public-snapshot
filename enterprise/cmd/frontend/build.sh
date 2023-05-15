#!/usr/bin/env bash

# This script builds the frontend docker image.

cd "$(dirname "${BASH_SOURCE[0]}")/../../.."
set -eu

OUTPUT=$(mktemp -d -t sgdockerbuild_XXXXXXX)
cleanup() {
  rm -rf "$OUTPUT"
}
trap cleanup EXIT

if [[ "${DOCKER_BAZEL:-false}" == "true" ]]; then
  ./dev/ci/bazel.sh build //enterprise/cmd/frontend
  out=$(./dev/ci/bazel.sh cquery //enterprise/cmd/frontend --output=files)
  cp "$out" "$OUTPUT"

  docker build -f enterprise/cmd/frontend/Dockerfile -t "$IMAGE" "$OUTPUT" \
    --progress=plain \
    --build-arg COMMIT_SHA \
    --build-arg DATE \
    --build-arg VERSION
  exit $?
fi

# Environment for building linux binaries
export GO111MODULE=on
export GOARCH=amd64
export GOOS=linux
export CGO_ENABLED=0

echo "--- go build"
pkg="github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend"
go build -trimpath -ldflags "-X github.com/sourcegraph/sourcegraph/internal/version.version=$VERSION -X github.com/sourcegraph/sourcegraph/internal/version.timestamp=$(date +%s)" -buildmode exe -tags dist -o "$OUTPUT/$(basename $pkg)" "$pkg"

echo "--- docker build"
docker build -f enterprise/cmd/frontend/Dockerfile -t "$IMAGE" "$OUTPUT" \
  --progress=plain \
  --build-arg COMMIT_SHA \
  --build-arg DATE \
  --build-arg VERSION
