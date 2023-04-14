#!/usr/bin/env bash

set -ex
cd "$(dirname "${BASH_SOURCE[0]}")/../.."

BUILDDIR=$(mktemp -d -t sgdockerbuild_XXXXXXX)
cleanup() {
  rm -rf "$BUILDDIR"
}
trap cleanup EXIT

./dev/ci/bazel.sh build //docker-images/syntax-highlighter:syntect_server \
  --stamp \
  --workspace_status_command=./dev/bazel_stamp_vars.sh

out=$(bazel cquery //docker-images/syntax-highlighter:syntect_server --output=files)

cp "$out" "$BUILDDIR"

# # shellcheck disable=SC2086
docker build -f docker-images/syntax-highlighter/Dockerfile.bazel -t "${IMAGE:-sourcegraph/syntax-highlighter}" "$BUILDDIR" \
  --progress=plain \
  --build-arg COMMIT_SHA \
  --build-arg DATE \
  --build-arg VERSION
