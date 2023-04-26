#!/usr/bin/env bash

set -ex

cd "$(dirname "${BASH_SOURCE[0]}")/../.."

# We build out of tree to prevent triggering dev watch scripts when we copy go
# files.
BUILDDIR=$(mktemp -d -t sgdockerbuild_XXXXXXX)
TMP=$(mktemp -d -t sgprom_tmp_XXXXXXX)

cleanup() {
  rm -rf "$BUILDDIR"
  rm -rf "$TMP"
}
trap cleanup EXIT

./dev/ci/bazel.sh build //docker-images/prometheus/cmd/prom-wrapper //monitoring:generate_config
out=$(./dev/ci/bazel.sh cquery //docker-images/prometheus/cmd/prom-wrapper --output=files)
cp "$out" "$BUILDDIR"
monitoring_cfg=$(./dev/ci/bazel.sh cquery //monitoring:generate_config --output=files)
cp "$monitoring_cfg" "$TMP/"
pushd "$TMP"
unzip "monitoring.zip"
popd

cp -r docker-images/prometheus/config "$BUILDDIR/sg_config_prometheus"
cp docker-images/prometheus/*.sh "$BUILDDIR/"
cp -r "$TMP/monitoring/prometheus"/* "$BUILDDIR/sg_config_prometheus"
mkdir "$BUILDDIR/sg_prometheus_add_ons"
cp dev/prometheus/linux/prometheus_targets.yml "$BUILDDIR/sg_prometheus_add_ons"

docker build -f docker-images/prometheus/Dockerfile.bazel -t "${IMAGE:-sourcegraph/prometheus}" "$BUILDDIR" \
  --platform linux/amd64 \
  --progress=plain \
  --build-arg BASE_IMAGE \
  --build-arg COMMIT_SHA \
  --build-arg DATE \
  --build-arg VERSION
