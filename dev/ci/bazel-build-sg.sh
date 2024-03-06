#!/usr/bin/env bash

set -o errexit -o nounset -o pipefail

aspectRC="/tmp/aspect-generated.bazelrc"
rosetta bazelrc > "$aspectRC"
bazelrc=(--bazelrc="$aspectRC")

echo "--- :bazel: Build sg cli"
bazel "${bazelrc[@]}" build //dev/sg:sg

sg_cli="$(bazel "${bazelrc[@]}" cquery //dev/sg:sg --output files)"
cp "$sg_cli" ./sg
