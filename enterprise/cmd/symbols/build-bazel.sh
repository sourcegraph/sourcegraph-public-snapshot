#!/usr/bin/env bash

# This script builds the symbols docker image.

cd "$(dirname "${BASH_SOURCE[0]}")/../../.."
set -eu

OUTPUT=$(mktemp -d -t sgdockerbuild_XXXXXXX)
cleanup() {
  rm -rf "$OUTPUT"
}
trap cleanup EXIT

echo "--- :bazel: bazel build for targets //enterprise/cmd/symbols"

bazelrc=(
  --bazelrc=.bazelrc
)
if [[ ${CI:-""} == "true" ]]; then
  bazelrc+=(
    --bazelrc=.aspect/bazelrc/ci.bazelrc
    --bazelrc=.aspect/bazelrc/ci.sourcegraph.bazelrc
  )
fi

bazel "${bazelrc[@]}" \
  build \
  //enterprise/cmd/symbols \
  --stamp \
  --workspace_status_command=./dev/bazel_stamp_vars.sh \
  --config incompat-zig-linux-amd64

out=$(
  bazel "${bazelrc[@]}" \
    cquery //enterprise/cmd/symbols \
    --stamp \
    --workspace_status_command=./dev/bazel_stamp_vars.sh \
    --config incompat-zig-linux-amd64 \
    --output=files
)
cp -v "$out" "$OUTPUT"

# we can't build scip-ctags with symbols since the platform args conflict
# NOTE: cmd/symbols/cargo-config.sh sets some specific config when running on arm64
# since this bazel run typically runs on CI that config change isn't made
echo "--- :bazel: bazel build for target //docker-images/syntax-highlighter:scip-ctags"
bazel "${bazelrc[@]}" \
  build //docker-images/syntax-highlighter:scip-ctags \
  --stamp \
  --workspace_status_command=./dev/bazel_stamp_vars.sh

out=$(
  bazel "${bazelrc[@]}" \
    cquery //docker-images/syntax-highlighter:scip-ctags \
    --stamp \
    --workspace_status_command=./dev/bazel_stamp_vars.sh \
    --output=files
)
cp -v "$out" "$OUTPUT"

cp cmd/symbols/ctags-install-alpine.sh "$OUTPUT"

echo ":docker: context directory contains the following:"
ls -lah "$OUTPUT"
echo "--- :docker: docker build for symbols"
docker build -f cmd/symbols/Dockerfile.bazel -t "$IMAGE" "$OUTPUT" \
  --progress=plain \
  --build-arg COMMIT_SHA \
  --build-arg DATE \
  --build-arg VERSION
