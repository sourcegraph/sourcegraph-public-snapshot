#!/usr/bin/env bash

# We want to build multiple go binaries, so we use a custom build step on CI.
cd "$(dirname "${BASH_SOURCE[0]}")"/../..
set -ex

OUTPUT=$(mktemp -d -t sgdockerbuild_XXXXXXX)
cleanup() {
  rm -rf "$OUTPUT"
}
trap cleanup EXIT

if [[ "${DOCKER_BAZEL:-false}" == "true" ]]; then

  bazel build //cmd/loadtest \
    --stamp \
    --workspace_status_command=./dev/bazel_stamp_vars.sh \
    --platforms=@io_bazel_rules_go//go/toolchain:linux_amd64

  out=$(bazel cquery //cmd/loadtest --output=files)
  cp "$out" "$OUTPUT"

  docker build -f cmd/loadtest/Dockerfile.wolfi -t "$IMAGE" "$OUTPUT" \
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

pkg="github.com/sourcegraph/sourcegraph/cmd/loadtest"
go build -trimpath -ldflags "-X github.com/sourcegraph/sourcegraph/internal/version.version=$VERSION  -X github.com/sourcegraph/sourcegraph/internal/version.timestamp=$(date +%s)" -buildmode exe -tags dist -o "$OUTPUT/$(basename $pkg)" "$pkg"

docker build -f cmd/loadtest/Dockerfile.wolfi -t "$IMAGE" "$OUTPUT" \
  --progress=plain \
  --build-arg COMMIT_SHA \
  --build-arg DATE \
  --build-arg VERSION
