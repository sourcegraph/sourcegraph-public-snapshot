#!/usr/bin/env bash

set -ex

# We build out of tree to prevent triggering dev watch scripts when we copy go
# files.
BUILDDIR=$(mktemp -d -t sgdockerbuild_XXXXXXX)
TMP=$(mktemp -d -t sgprom_tmp_XXXXXXX)

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
cp "$monitoring_cfg" "$TMP"
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
