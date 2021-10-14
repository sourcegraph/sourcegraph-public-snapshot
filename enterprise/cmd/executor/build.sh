#!/usr/bin/env bash

# This script builds the executor image as a GCP boot disk image and as an AWS AMI.

cd "$(dirname "${BASH_SOURCE[0]}")"
set -eu

OUTPUT=$(mktemp -d -t sgdockerbuild_XXXXXXX)
cleanup() {
  rm -rf "$OUTPUT"
}
trap cleanup EXIT

# Capture src cli version before we reconfigure go environment.
SRC_CLI_VERSION="$(go run ../../../internal/cmd/src-cli-version/main.go)"

# Environment for building linux binaries
export GO111MODULE=on
export GOARCH=amd64
export GOOS=linux
export CGO_ENABLED=0

echo "--- go build"
pkg="github.com/sourcegraph/sourcegraph/enterprise/cmd/executor"
go build -trimpath -ldflags "-X github.com/sourcegraph/sourcegraph/internal/version.version=$VERSION -X github.com/sourcegraph/sourcegraph/internal/version.timestamp=$(date +%s)" -buildmode exe -tags dist -o "$OUTPUT/$(basename $pkg)" "$pkg"

echo "--- gcp secret"
gcloud secrets versions access latest --secret=e2e-builder-sa-key --quiet --project=sourcegraph-ci >"$OUTPUT/builder-sa-key.json"

echo "--- packer build"

# Copy files into workspace.
cp -R ./image/* "$OUTPUT"
cp ../../../.tool-versions "$OUTPUT"

export NAME
NAME=executor-$(git log -n1 --pretty=format:%h)-${BUILDKITE_BUILD_NUMBER}
export SRC_CLI_VERSION=${SRC_CLI_VERSION}
export AWS_EXECUTOR_AMI_ACCESS_KEY=${AWS_EXECUTOR_AMI_ACCESS_KEY}
export AWS_EXECUTOR_AMI_SECRET_KEY=${AWS_EXECUTOR_AMI_SECRET_KEY}
# This should prevent some occurrences of Failed waiting for AMI failures:
# https://austincloud.guru/2020/05/14/long-running-packer-builds-failing/
export AWS_MAX_ATTEMPTS=240
export AWS_POLL_DELAY_SECONDS=5

pushd "$OUTPUT" 1>/dev/null
packer build -force executor.json
popd 1>/dev/null
