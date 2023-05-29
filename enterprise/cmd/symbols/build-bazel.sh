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
bazel \
  --bazelrc=.bazelrc \
    --bazelrc=.aspect/bazelrc/ci.bazelrc \
    --bazelrc=.aspect/bazelrc/ci.sourcegraph.bazelrc \
  build //enterprise/cmd/symbols \
    --stamp \
  --workspace_status_command=./dev/bazel_stamp_vars.sh \
  --platforms @zig_sdk//platform:linux_amd64 \
  --extra_toolchains @zig_sdk//toolchain:linux_amd64_musl

out=$(
  bazel --bazelrc=.bazelrc \
    --bazelrc=.aspect/bazelrc/ci.bazelrc \
    --bazelrc=.aspect/bazelrc/ci.sourcegraph.bazelrc \
    cquery //enterprise/cmd/symbols \
    --stamp \
    --workspace_status_command=./dev/bazel_stamp_vars.sh \
    --platforms @zig_sdk//platform:linux_amd64 \
    --extra_toolchains @zig_sdk//toolchain:linux_amd64_musl \
    --output=files
)
cp -v "$out" "$OUTPUT"

# we can't build scip-ctags with symbols since the platform args conflict
echo "--- :bazel: bazel build for target //docker-images/syntax-highlighter:scip-ctags"
bazel \
  --bazelrc=.bazelrc \
    --bazelrc=.aspect/bazelrc/ci.bazelrc \
    --bazelrc=.aspect/bazelrc/ci.sourcegraph.bazelrc \
  build //docker-images/syntax-highlighter:scip-ctags \
    --stamp \
  --workspace_status_command=./dev/bazel_stamp_vars.sh \

out=$(
  bazel --bazelrc=.bazelrc \
    --bazelrc=.aspect/bazelrc/ci.bazelrc \
    --bazelrc=.aspect/bazelrc/ci.sourcegraph.bazelrc \
    cquery //docker-images/syntax-highlighter:scip-ctags \
    --stamp \
    --workspace_status_command=./dev/bazel_stamp_vars.sh \
    --output=files
)
cp -v "$out" "$OUTPUT"

cp cmd/symbols/ctags-install-alpine.sh "$OUTPUT"

echo "--- DEBUG"
shasum cmd/symbols/Dockerfile.bazel
ls -lah $OUTPUT
git checkout cmd/symbols/Dockerfile.bazel
echo "--- :docker: docker build for symbols"
docker build -f cmd/symbols/Dockerfile.bazel -t "$IMAGE" "$OUTPUT" \
  --progress=plain \
  --build-arg COMMIT_SHA \
  --build-arg DATE \
  --build-arg VERSION
