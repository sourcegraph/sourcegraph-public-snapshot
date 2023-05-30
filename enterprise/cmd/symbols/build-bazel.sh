#!/usr/bin/env bash

# This script builds the symbols docker image.

cd "$(dirname "${BASH_SOURCE[0]}")/../../.."
set -eu

OUTPUT=$(mktemp -d -t sgdockerbuild_XXXXXXX)
cleanup() {
  rm -rf "$OUTPUT"
}
trap cleanup EXIT

bazelrc=(
  --bazelrc=.bazelrc
)
if [[ ${CI:-""} == "true" ]]; then
  bazelrc+=(
    --bazelrc=.aspect/bazelrc/ci.bazelrc
    --bazelrc=.aspect/bazelrc/ci.sourcegraph.bazelrc
  )
fi

echo "--- bazel build"
bazel "${bazelrc[@]}" \
  build \
  //enterprise/cmd/symbols \
  --stamp \
  --workspace_status_command=./dev/bazel_stamp_vars.sh \
  --config incompat-zig-linux-amd64

out=$(
  bazel \
    "${bazelrc[@]}" \
    cquery \
    //enterprise/cmd/symbols \
    --stamp \
    --workspace_status_command=./dev/bazel_stamp_vars.sh \
    --config incompat-zig-linux-amd64 \
    --output=files
)
cp "$out" "$OUTPUT"
cp cmd/symbols/ctags-install-alpine.sh "$OUTPUT"

docker build -f cmd/symbols/Dockerfile.bazel -t "$IMAGE" "$OUTPUT" \
  --progress=plain \
  --build-arg COMMIT_SHA \
  --build-arg DATE \
  --build-arg VERSION
