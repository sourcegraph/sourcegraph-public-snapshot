#!/usr/bin/env bash

# We want to build multiple go binaries, so we use a custom build step on CI.
cd "$(dirname "${BASH_SOURCE[0]}")"/../../..
set -ex

OUTPUT=$(mktemp -d -t sgdockerbuild_XXXXXXX)
cleanup() {
  rm -rf "$OUTPUT"
}
trap cleanup EXIT

SRC_CLI_VERSION="$(go run ./internal/cmd/src-cli-version/main.go)"

# Environment for building linux binaries
export GO111MODULE=on
export GOARCH=amd64
export GOOS=linux
export CGO_ENABLED=0

if [[ "${DOCKER_BAZEL:-false}" == "true" ]]; then

  TARGETS=(
    //enterprise/cmd/batcheshelper
    //enterprise/cmd/executor
  )
  ./dev/ci/bazel.sh build "${TARGETS[@]}"
  for TARGET in "${TARGETS[@]}"; do
    out=$(./dev/ci/bazel.sh cquery "$TARGET" --output=files)
    cp "$out" "$OUTPUT"
    echo "copying $TARGET"
  done

  docker build -f enterprise/cmd/batcheshelper/Dockerfile -t "$IMAGE" "$OUTPUT" \
    --progress=plain \
    --build-arg COMMIT_SHA \
    --build-arg DATE \
    --build-arg VERSION
  exit $?
fi

pushd ./enterprise/cmd/executor 1>/dev/null
pkg="github.com/sourcegraph/sourcegraph/enterprise/cmd/executor"
go build -trimpath -ldflags "-X github.com/sourcegraph/sourcegraph/internal/version.version=$VERSION -X github.com/sourcegraph/sourcegraph/internal/version.timestamp=$(date +%s)" -buildmode exe -tags dist,shell -o "$OUTPUT/$(basename $pkg)" "$pkg"
popd 1>/dev/null

pushd ./enterprise/cmd/batcheshelper 1>/dev/null
pkg="github.com/sourcegraph/sourcegraph/enterprise/cmd/batcheshelper"
go build -trimpath -ldflags "-X github.com/sourcegraph/sourcegraph/internal/version.version=$VERSION -X github.com/sourcegraph/sourcegraph/internal/version.timestamp=$(date +%s)" -buildmode exe -tags dist -o "$OUTPUT/$(basename $pkg)" "$pkg"
popd 1>/dev/null

docker build -f enterprise/cmd/bundled-executor/Dockerfile -t "$IMAGE" "$OUTPUT" \
  --progress=plain \
  --build-arg SRC_CLI_VERSION="${SRC_CLI_VERSION}" \
  --platform=linux/amd64 \
  --build-arg COMMIT_SHA \
  --build-arg DATE \
  --build-arg VERSION
