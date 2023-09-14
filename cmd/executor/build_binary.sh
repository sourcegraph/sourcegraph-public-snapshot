#!/usr/bin/env bash

# This script builds the executor binary.

cd "$(dirname "${BASH_SOURCE[0]}")"/../..
set -eu

OUTPUT=$(mktemp -d -t sgbuild_XXXXXXX)
cleanup() {
  rm -rf "$OUTPUT"
}
trap cleanup EXIT

GIT_COMMIT="$(git rev-parse HEAD)"

mkdir -p "${OUTPUT}/${GIT_COMMIT}/linux-amd64"

# Environment for building linux binaries
export GO111MODULE=on
export GOARCH=amd64
export GOOS=linux
export CGO_ENABLED=0
export VERSION

echo "--- go build"
pushd ./cmd/executor 1>/dev/null
pkg="github.com/sourcegraph/sourcegraph/cmd/executor"
bin_name="${OUTPUT}/${GIT_COMMIT}/linux-amd64/$(basename $pkg)"
go build -trimpath -ldflags "-X github.com/sourcegraph/sourcegraph/internal/version.version=$VERSION -X github.com/sourcegraph/sourcegraph/internal/version.timestamp=$(date +%s)" -buildmode exe -tags dist -o "$bin_name" "$pkg"
popd 1>/dev/null

echo "--- create binary artifacts"
INFO_PATH="${OUTPUT}/${GIT_COMMIT}/info.txt"

echo "executor built from https://github.com/sourcegraph/sourcegraph" >"${INFO_PATH}"
echo >>"${INFO_PATH}"
git log -n1 >>"${INFO_PATH}"
sha256sum "${OUTPUT}/${GIT_COMMIT}/linux-amd64/executor" >>"${OUTPUT}/${GIT_COMMIT}/linux-amd64/executor_SHA256SUM"

# Upload the new release folder
echo "--- upload binary artifacts"
gsutil cp -r "${OUTPUT}/${GIT_COMMIT}" gs://sourcegraph-artifacts/executor
gsutil iam ch allUsers:objectViewer gs://sourcegraph-artifacts
