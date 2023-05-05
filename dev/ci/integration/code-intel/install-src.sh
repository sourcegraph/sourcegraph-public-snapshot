#!/bin/bash

# This script is called by test.sh to install an up-to-date
# version of src-cli as required by the codeintel-qa pipeline. The target binary
# is installed to {REPO_ROOT}/.bin/src.

set -eux
cd "$(dirname "${BASH_SOURCE[0]}")/../../../.."
root_dir="$(pwd)"

TEMP=$(mktemp -d -t sgdockerbuild_XXXXXXX)
cleanup() {
  rm -rf "${TEMP}"
}
trap cleanup EXIT

bazel build @com_github_sourcegraph_src-cli//cmd/src:src
out=$(bazel cquery @com_github_sourcegraph_src-cli//cmd/src:src --output=files)
cp "$out" "$root_dir/.bin/src"
