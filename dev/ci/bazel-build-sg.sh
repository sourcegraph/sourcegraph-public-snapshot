#!/usr/bin/env bash

set -o errexit -o nounset -o pipefail

bazelrc=(--bazelrc=.bazelrc --bazelrc=.aspect/bazelrc/ci.bazelrc --bazelrc=.aspect/bazelrc/ci.sourcegraph.bazelrc)

echo "--- :bazel: Build sg cli"
bazel "${bazelrc[@]}" build //dev/sg:sg

sg_cli="$(bazel "${bazelrc[@]}" cquery //dev/sg:sg --output files)"
cp "$sg_cli" ./sg
