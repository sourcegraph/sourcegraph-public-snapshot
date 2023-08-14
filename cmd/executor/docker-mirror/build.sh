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
cp ../../../.tool-versions "$TMR_WORKDIR"

export PKR_VAR_name
PKR_VAR_name="${IMAGE_FAMILY}-${BUILDKITE_BUILD_NUMBER}"
export PKR_VAR_image_family="${IMAGE_FAMILY}"
export PKR_VAR_tagged_release="${EXECUTOR_IS_TAGGED_RELEASE}"
export PKR_VAR_aws_access_key=${AWS_EXECUTOR_AMI_ACCESS_KEY}
export PKR_VAR_aws_secret_key=${AWS_EXECUTOR_AMI_SECRET_KEY}
# This should prevent some occurrences of Failed waiting for AMI failures:
# https://austincloud.guru/2020/05/14/long-running-packer-builds-failing/
export PKR_VAR_aws_max_attempts=480
export PKR_VAR_aws_poll_delay_seconds=5

pushd "$TMR_WORKDIR" 1>/dev/null

export PKR_VAR_aws_regions
if [ "${EXECUTOR_IS_TAGGED_RELEASE}" = "true" ]; then
  PKR_VAR_aws_regions="$(jq -r '.' <aws_regions.json)"
else
  PKR_VAR_aws_regions='["us-west-2"]'
fi

packer init docker-mirror.pkr.hcl
packer build -force docker-mirror.pkr.hcl

popd 1>/dev/null
