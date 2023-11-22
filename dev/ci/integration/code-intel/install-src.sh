#!/usr/bin/env bash

# This script is called by test.sh to install an up-to-date
# version of src-cli as required by the codeintel-qa pipeline. The target binary
# is installed to {REPO_ROOT}/.bin/src.

set -eux
cd "$(dirname "${BASH_SOURCE[0]}")/../../../.."
root_dir="$(pwd)"

# By default, version of src-cli that builds with 1.19.8
VERSION=${1:-'58b3f701691cbdbd10b54161d9bfca88b781480d'}

TEMP=$(mktemp -d -t sgdockerbuild_XXXXXXX)
cleanup() {
  rm -rf "${TEMP}"
}
trap cleanup EXIT

# TODO: migrate upstream to bazel
# bazel build @com_github_sourcegraph_src-cli//cmd/src:src
# out=$(bazel cquery @com_github_sourcegraph_src-cli//cmd/src:src --output=files)
# cp "$out" "$root_dir/.bin/src"

git clone git@github.com:sourcegraph/src-cli.git "${TEMP}" --depth 1
pushd "${TEMP}"
git fetch origin "${VERSION}" --depth 1
git checkout "${VERSION}"
mkdir -p "${root_dir}/.bin"
go build -o "${root_dir}/.bin" ./cmd/src
popd
