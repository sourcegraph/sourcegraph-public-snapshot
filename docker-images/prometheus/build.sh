#!/usr/bin/env bash

set -ex

cd "$(dirname "${BASH_SOURCE[0]}")"

# We build out of tree to prevent triggering dev watch scripts when we copy go
# files.
BUILDDIR=$(mktemp -d -t sgdockerbuild_XXXXXXX)
cleanup() {
  rm -rf "$BUILDDIR"
}
trap cleanup EXIT

# Copy assets
cp -R . "$BUILDDIR"

# Build args for Go cross-compilation.
export GO111MODULE=on
export GOARCH=amd64
export GOOS=linux
export CGO_ENABLED=0

# Cross-compile prom-wrapper before building the image.
go build \
  -trimpath \
  -installsuffix netgo \
  -tags "dist netgo" \
  -ldflags "-X github.com/sourcegraph/sourcegraph/internal/version.version=$VERSION -X github.com/sourcegraph/sourcegraph/internal/version.timestamp=$(date +%s)" \
  -o "$BUILDDIR"/.bin/prom-wrapper ./cmd/prom-wrapper

# Cross-compile monitoring generator before building the image.
go build \
  -trimpath \
  -o "$BUILDDIR"/.bin/monitoring-generator ../../monitoring

pushd "$BUILDDIR"

# Note: This chmod is so that both the `sourcegraph` user and host system user (what `whoami` reports on
# Linux) both have access to the files in the container AND files mounted by `-v` into the container without it
# running as root. For more details, see:
# https://github.com/sourcegraph/sourcegraph/pull/11832#discussion_r451109637
chmod -R 777 config

# Enable image build caching via CACHE=true
BUILD_CACHE="--no-cache"
if [[ "$CACHE" == "true" ]]; then
  BUILD_CACHE=""
fi

# shellcheck disable=SC2086
docker build ${BUILD_CACHE} -t "${IMAGE:-sourcegraph/prometheus}" . \
  --progress=plain \
  --build-arg BASE_IMAGE \
  --build-arg COMMIT_SHA \
  --build-arg DATE \
  --build-arg VERSION

# cd out of $BUILDDIR for cleanup
popd
