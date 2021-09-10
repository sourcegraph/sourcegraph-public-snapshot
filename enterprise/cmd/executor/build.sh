#!/usr/bin/env bash

# This script builds the executor google cloud image.

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
steps:
  - name: gcr.io/cloud-builders/gcloud
    entrypoint: bash
    args: ['-c', 'gcloud secrets versions access latest --secret=e2e-builder-sa-key --quiet --project=sourcegraph-ci > /workspace/builder-sa-key.json']
  - name: index.docker.io/hashicorp/packer:1.6.6
    env:
      - 'VERSION=$(git log -n1 --pretty=format:%h)'
      - 'BUILD_TIMESTAMP=$BUILD_TIMESTAMP'
      - 'SRC_CLI_VERSION=$SRC_CLI_VERSION'
      - 'AWS_EXECUTOR_AMI_ACCESS_KEY=$AWS_EXECUTOR_AMI_ACCESS_KEY'
      - 'AWS_EXECUTOR_AMI_SECRET_KEY=$AWS_EXECUTOR_AMI_SECRET_KEY'
    args: ['build', 'executor.json']
EOF

# Copy cloudbuild files into workspace.
cp -R ./cloudbuild/* "$OUTPUT"

# Run gcloud image build.
gcloud builds submit --config="$OUTPUT/cloudbuild.yaml" "$OUTPUT" --project="sourcegraph-ci"
