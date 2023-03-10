#!/bin/bash

# This script is called by test.sh and preprod-run.sh to install an up-to-date
# version of src-cli as required by the codeintel-qa pipeline. The target binary
# is installed to {REPO_ROOT}/.bin/src.

set -eux
cd "$(dirname "${BASH_SOURCE[0]}")/../../../.."
root_dir="$(pwd)"

# By default, version of src-cli that builds with 1.19.6
VERSION=${1:-'85115b1a8a2e1bc174075eefacbae6ad9d19af1f'}

TEMP=$(mktemp -d -t sgdockerbuild_XXXXXXX)
cleanup() {
  rm -rf "${TEMP}"
}
trap cleanup EXIT

git clone git@github.com:sourcegraph/src-cli.git "${TEMP}" --depth 1
pushd "${TEMP}"
git fetch origin "${VERSION}" --depth 1
git checkout "${VERSION}"
mkdir -p "${root_dir}/.bin"
go build -o "${root_dir}/.bin" ./cmd/src
popd
