#!/usr/bin/env bash

cd "$(dirname "${BASH_SOURCE[0]}")"/../../..
set -ex

OUTPUT=$(mktemp -d -t sgdockerbuild_XXXXXXX)
cleanup() {
  rm -rf "$OUTPUT"
}
trap cleanup EXIT

# Environment for building linux binaries
export GO111MODULE=on
export GOARCH=amd64
export GOOS=linux
export CGO_ENABLED=0

if [[ "${DOCKER_BAZEL:-false}" == "true" ]]; then
  ./dev/ci/bazel.sh build //enterprise/cmd/batcheshelper \
    --stamp \
    --workspace_status_command=./dev/bazel_stamp_vars.sh \
    --platforms=@io_bazel_rules_go//go/toolchain:linux_amd64

  out=$(./dev/ci/bazel.sh cquery //enterprise/cmd/batcheshelper --output=files)
  cp "$out" "$OUTPUT"

  docker build -f enterprise/cmd/batcheshelper/Dockerfile -t "$IMAGE" "$OUTPUT" \
    --progress=plain \
    --build-arg COMMIT_SHA \
    --build-arg DATE \
    --build-arg VERSION
  exit $?
fi

pkg="github.com/sourcegraph/sourcegraph/enterprise/cmd/batcheshelper"
go build -trimpath -ldflags "-X github.com/sourcegraph/sourcegraph/internal/version.version=$VERSION  -X github.com/sourcegraph/sourcegraph/internal/version.timestamp=$(date +%s)" -buildmode exe -tags dist -o "$OUTPUT/$(basename $pkg)" "$pkg"

docker build -f enterprise/cmd/batcheshelper/Dockerfile -t "$IMAGE" "$OUTPUT" \
  --platform="${PLATFORM:-linux/amd64}" \
  --progress=plain \
  --build-arg COMMIT_SHA \
  --build-arg DATE \
  --build-arg VERSION
