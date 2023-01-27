#!/usr/bin/env bash

set -eu -o pipefail

cd "$(dirname "${BASH_SOURCE[0]}")/../../../../.."
# TODO: This must be replaced with a proper K8s managed secret
key_path=$(realpath ./wolfi-packages/temporary-keys/)

# TODO: Manage these variables properly
GCP_PROJECT="sourcegraph-ci"
GCS_BUCKET="package-repository"
TARGET_ARCH="x86_64"
branch="main"

tmpdir=$(mktemp -d -t melange-bin.XXXXXXXX)
function cleanup() {
  echo "Removing $tmpdir"
  rm -rf "$tmpdir"
}
trap cleanup EXIT

# TODO: Install these binaries as part of the buildkite base image
(
  cd "$tmpdir"
  mkdir bin

  # Install melange from Sourcegraph cache
  # Source: https://github.com/chainguard-dev/melange/releases/download/v0.2.0/melange_0.2.0_linux_amd64.tar.gz
  wget https://storage.googleapis.com/package-repository/ci-binaries/melange_0.2.0_linux_amd64.tar.gz
  tar zxf melange_0.2.0_linux_amd64.tar.gz
  mv melange_0.2.0_linux_amd64/melange bin/melange
)

export PATH="$tmpdir/bin:$PATH"

apkindex_build_dir=$(mktemp -d -t apkindex-build.XXXXXXXX)
pushd "$apkindex_build_dir"

# Fetch all APKINDEX fragments from bucket
gsutil -u "$GCP_PROJECT" -m cp "gs://$GCS_BUCKET/packages/$branch/$TARGET_ARCH/*.APKINDEX.fragment" ./

# Concat all fragments into a single APKINDEX and tar.gz it
touch placeholder.APKINDEX.fragment
cat ./*.APKINDEX.fragment >APKINDEX
touch DESCRIPTION
tar zcf APKINDEX.tar.gz APKINDEX DESCRIPTION

# Sign index
melange sign-index --signing-key "$key_path/melange.rsa" APKINDEX.tar.gz

# Upload signed APKINDEX archive
gsutil -u "$GCP_PROJECT" cp APKINDEX.tar.gz "gs://$GCS_BUCKET/packages/$branch/$TARGET_ARCH/"
