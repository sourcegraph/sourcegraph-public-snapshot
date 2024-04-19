#!/usr/bin/env bash

# This script builds the sourcegraph/executor-vm docker image.

set -eux
cd "$(dirname "${BASH_SOURCE[0]}")"

# If we're running in CI, generate aspect bazelrc
if [[ ${CI:-} == "true" ]]; then
  aspectRC="/tmp/aspect-generated.bazelrc"
  rosetta bazelrc >"${aspectRC}"
  bazelrcs=(--bazelrc="${aspectRC}")
fi

SRC_CLI_VERSION="$(bazel "${bazelrcs[@]}" run //internal/cmd/src-cli-version:src-cli-version)"

echo "--- docker build"
docker build -t "$IMAGE" . \
  --progress=plain \
  --build-arg SRC_CLI_VERSION="${SRC_CLI_VERSION}" \
  --build-arg COMMIT_SHA \
  --build-arg DATE \
  --build-arg VERSION
