#!/usr/bin/env bash

set -o errexit -o nounset -o pipefail

echo "--- :bazel: Build sg cli"
bazel \
  --bazelrc=.bazelrc \
  --bazelrc=.aspect/bazelrc/ci.bazelrc \
  --bazelrc=.aspect/bazelrc/ci.sourcegraph.bazelrc \
  build \
  //dev/sg:sg

sg_cli="$(bazel cquery //dev/sg:sg --output files)"
cp "$sg_cli" ./sg
