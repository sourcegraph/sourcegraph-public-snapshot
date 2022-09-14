#!/usr/bin/env bash

# This script builds the executor image as a GCP boot disk image and as an AWS AMI.

cd "$(dirname "${BASH_SOURCE[0]}")"
set -eu

TMR_WORKDIR=$(mktemp -d -t sgdockerbuild_XXXXXXX)
cleanup() {
  rm -rf "$TMR_WORKDIR"
}
trap cleanup EXIT

echo "--- gcp secret"
gcloud secrets versions access latest --secret=e2e-builder-sa-key --quiet --project=sourcegraph-ci >"$TMR_WORKDIR/builder-sa-key.json"

echo "--- packer build"

# Copy files into workspace.
cp -R ./* "$TMR_WORKDIR"
cp ../../../../.tool-versions "$TMR_WORKDIR"

export NAME
NAME="${IMAGE_FAMILY}-${BUILDKITE_BUILD_NUMBER}"
export AWS_EXECUTOR_AMI_ACCESS_KEY=${AWS_EXECUTOR_AMI_ACCESS_KEY}
export AWS_EXECUTOR_AMI_SECRET_KEY=${AWS_EXECUTOR_AMI_SECRET_KEY}
# This should prevent some occurrences of Failed waiting for AMI failures:
# https://austincloud.guru/2020/05/14/long-running-packer-builds-failing/
export AWS_MAX_ATTEMPTS=480
export AWS_POLL_DELAY_SECONDS=5

pushd "$TMR_WORKDIR" 1>/dev/null
packer build -force docker-mirror.json
popd 1>/dev/null
