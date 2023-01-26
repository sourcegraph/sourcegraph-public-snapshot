#!/usr/bin/env bash

set -eu -o pipefail

cd "$(dirname "${BASH_SOURCE[0]}")/../../../../.."
# TODO: This must be replaced with a proper K8s managed secret
key_path=$(realpath ./wolfi-packages/temporary-keys/)

# Fetch all the index fragments
# Combine to create the index file
# Sign index file using private key
# Upload apkindex bundle

# TODO: Manage these variables properly
GCP_PROJECT="sourcegraph-ci"
GCS_BUCKET="package-repository"
ARCH="x86_64"
branch="main"

echo "[Placeholder]"

tmpdir=$(mktemp -d -t melange-bin.XXXXXXXX)
function cleanup() {
  echo "Removing $tmpdir"
  # rm -rf "$tmpdir"
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

# Fetch all fragments from bucket
gsutil -u "$GCP_PROJECT" cp "gs://$GCS_BUCKET/packages/$branch/$ARCH/*.APKINDEX.fragment" ./

# Concat all fragments into a single index and tar.gz it
# TODO: Handle case where there are no fragments more cleanly
touch placeholder.APKINDEX.fragment
cat ./*.APKINDEX.fragment >APKINDEX
touch DESCRIPTION
tar zcf APKINDEX.tar.gz APKINDEX DESCRIPTION

# Sign index using a key from somewhere
melange sign-index --signing-key "$key_path" APKINDEX.tar.gz

# Upload signed APKINDEX
gsutil -u "$GCP_PROJECT" cp APKINDEX.tar.gz "gs://$GCS_BUCKET/packages/$branch/$ARCH/"
