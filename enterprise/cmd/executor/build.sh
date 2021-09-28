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

echo "--- packer build"

cat <<EOF >"$OUTPUT/cloudbuild.yaml"
timeout: 1200s
steps:
  - name: gcr.io/cloud-builders/gcloud
    entrypoint: bash
    args: ['-c', 'gcloud secrets versions access latest --secret=e2e-builder-sa-key --quiet --project=sourcegraph-ci > /workspace/builder-sa-key.json']
  - name: index.docker.io/hashicorp/packer:1.6.6
    timeout: 1200s
    env:
      - 'NAME=executor-$(git log -n1 --pretty=format:%h)-${BUILDKITE_BUILD_NUMBER}'
      - 'SRC_CLI_VERSION=${SRC_CLI_VERSION}'
      - 'AWS_EXECUTOR_AMI_ACCESS_KEY=${AWS_EXECUTOR_AMI_ACCESS_KEY}'
      - 'AWS_EXECUTOR_AMI_SECRET_KEY=${AWS_EXECUTOR_AMI_SECRET_KEY}'
      # This should prevent some occurrences of Failed waiting for AMI failures:
      # https://austincloud.guru/2020/05/14/long-running-packer-builds-failing/
      - 'AWS_MAX_ATTEMPTS=240'
      - 'AWS_POLL_DELAY_SECONDS=5'
    args: ['build', 'executor.json']
EOF

# Copy cloudbuild files into workspace.
cp -R ./cloudbuild/* "$OUTPUT"

# Run gcloud image build.
gcloud builds submit --config="$OUTPUT/cloudbuild.yaml" "$OUTPUT" --project="sourcegraph-ci" --timeout=20m
