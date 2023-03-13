#!/bin/bash

# This script is called by test.sh to install an up-to-date
# version of src-cli as required by the codeintel-qa pipeline. The target binary
# is installed to {REPO_ROOT}/.bin/src.

set -eux
cd "$(dirname "${BASH_SOURCE[0]}")/../../../.."
root_dir="$(pwd)"

# By default, the commit that added handleSCIP support
VERSION=${1:-'1c70d536b4ab3187b5aed41af8f259f1b8ceba6b'}

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
