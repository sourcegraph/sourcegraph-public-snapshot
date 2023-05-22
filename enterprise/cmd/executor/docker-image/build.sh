#!/usr/bin/env bash

# We want to build multiple go binaries, so we use a custom build step on CI.
cd "$(dirname "${BASH_SOURCE[0]}")"/../../../..
set -ex

OUTPUT=$(mktemp -d -t sgdockerbuild_XXXXXXX)
cleanup() {
  rm -rf "$OUTPUT"
}
trap cleanup EXIT

if [[ "${DOCKER_BAZEL:-false}" == "true" ]]; then
  ./dev/ci/bazel.sh build //enterprise/cmd/executor //internal/cmd/src-cli-version
  out=$(./dev/ci/bazel.sh cquery //enterprise/cmd/executor --output=files)
  cp "$out" "$OUTPUT"

  src_cli=$(./dev/ci/bazel.sh cquery //internal/cmd/src-cli-version --output=files)
  SRC_CLI_VERSION=$(eval "$src_cli")

  docker build -f enterprise/cmd/executor/docker-image/Dockerfile -t "$IMAGE" "$OUTPUT" \
    --progress=plain \
    --build-arg SRC_CLI_VERSION="${SRC_CLI_VERSION}" \
    --build-arg COMMIT_SHA \
    --build-arg DATE \
    --build-arg VERSION

  exit $?
fi

SRC_CLI_VERSION="$(go run ./internal/cmd/src-cli-version/main.go)"

# Environment for building linux binaries
export GO111MODULE=on
export GOARCH=amd64
export GOOS=linux
export CGO_ENABLED=0

pushd ./enterprise/cmd/executor 1>/dev/null
pkg="github.com/sourcegraph/sourcegraph/enterprise/cmd/executor"
go build -trimpath -ldflags "-X github.com/sourcegraph/sourcegraph/internal/version.version=$VERSION  -X github.com/sourcegraph/sourcegraph/internal/version.timestamp=$(date +%s)" -buildmode exe -tags dist -o "$OUTPUT/$(basename $pkg)" "$pkg"
popd 1>/dev/null

docker build -f enterprise/cmd/executor/docker-image/Dockerfile -t "$IMAGE" "$OUTPUT" \
  --progress=plain \
  --build-arg SRC_CLI_VERSION="${SRC_CLI_VERSION}" \
  --build-arg COMMIT_SHA \
  --build-arg DATE \
  --build-arg VERSION
