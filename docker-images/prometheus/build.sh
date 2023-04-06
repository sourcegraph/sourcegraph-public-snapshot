#!/usr/bin/env bash

set -ex

# We build out of tree to prevent triggering dev watch scripts when we copy go
# files.
BUILDDIR=$(mktemp -d -t sgdockerbuild_XXXXXXX)
cleanup() {
  rm -rf "$BUILDDIR"
  rm -rf "$TMP"
}
trap cleanup EXIT

# Enable image build caching via CACHE=true
BUILD_CACHE="--no-cache"
if [[ "$CACHE" == "true" ]]; then
  BUILD_CACHE=""
fi

if [[ "$DOCKER_BAZEL" == "true" ]]; then

  bazel build //docker-images/prometheus/cmd/prom-wrapper //monitoring:generate_config \
    --stamp \
    --workspace_status_command=./dev/bazel_stamp_vars.sh

  out=$(bazel cquery //docker-images/prometheus/cmd/prom-wrapper --output=files)
  pwd
  cp "$out" "$BUILDDIR"

  out=$(bazel cquery //monitoring:generate_config --output=files)

  TMP=$(mktemp -d -t sgprom_tmp_XXXXXXX)
  cp "$out" "$TMP"

  monitoring_cfg=$(bazel cquery //monitoring:generate_config --output=files)
  sudo cp "$monitoring_cfg" "$TMP"
  pushd "$TMP"
  unzip "monitoring.zip"
  popd

  cp -r docker-images/prometheus/config "$BUILDDIR/sg_config_prometheus"
  cp docker-images/prometheus/*.sh "$BUILDDIR/"
  cp -r "$TMP/monitoring/prometheus"/* "$BUILDDIR/sg_config_prometheus"
  mkdir "$BUILDDIR/sg_prometheus_add_ons"
  cp dev/prometheus/linux/prometheus_targets.yml "$BUILDDIR/sg_prometheus_add_ons"

  docker build ${BUILD_CACHE} -f docker-images/prometheus/Dockerfile.bazel -t "${IMAGE:-sourcegraph/prometheus}" "$BUILDDIR" \
    --platform linux/amd64 \
    --progress=plain \
    --build-arg BASE_IMAGE \
    --build-arg COMMIT_SHA \
    --build-arg DATE \
    --build-arg VERSION

  exit $?
fi

cd "$(dirname "${BASH_SOURCE[0]}")"

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
pushd "../../monitoring"
go build \
  -trimpath \
  -o "$BUILDDIR"/.bin/monitoring-generator .

# Final pre-build stage.
pushd "$BUILDDIR"

# Note: This chmod is so that both the `sourcegraph` user and host system user (what `whoami` reports on
# Linux) both have access to the files in the container AND files mounted by `-v` into the container without it
# running as root. For more details, see:
# https://github.com/sourcegraph/sourcegraph/pull/11832#discussion_r451109637
chmod -R 777 config

# shellcheck disable=SC2086
docker build ${BUILD_CACHE} -t "${IMAGE:-sourcegraph/prometheus}" . \
  --platform linux/amd64 \
  --progress=plain \
  --build-arg BASE_IMAGE \
  --build-arg COMMIT_SHA \
  --build-arg DATE \
  --build-arg VERSION

# cd out of $BUILDDIR for cleanup
popd
