#!/usr/bin/env bash

set -ex

BUILDDIR=$(mktemp -d -t sgdockerbuild_XXXXXXX)
TMP=$(mktemp -d -t sggraf_tmp_XXXXXXX)
cleanup() {
  rm -rf "$BUILDDIR"
  rm -rf "$TMP"

}
trap cleanup EXIT

./dev/ci/bazel.sh build //monitoring:generate_config
monitoring_cfg=$(./dev/ci/bazel.sh cquery //monitoring:generate_config --output=files)

cp "$monitoring_cfg" "$TMP"
pushd "$TMP"
unzip "monitoring.zip"
popd

cp -r docker-images/grafana/entry-alpine.sh "$BUILDDIR/"
cp -r docker-images/grafana/config "$BUILDDIR/"
cp -r "$TMP/monitoring/grafana" "$BUILDDIR/"

# # shellcheck disable=SC2086
docker build -f docker-images/grafana/Dockerfile.bazel -t "${IMAGE:-sourcegraph/grafana}" "$BUILDDIR" \
  --progress=plain \
  --build-arg COMMIT_SHA \
  --build-arg DATE \
  --build-arg VERSION
