#!/usr/bin/env bash

set -eu -o pipefail

cd "$(dirname "${BASH_SOURCE[0]}")/../../../../.."

# TODO: Manage these variables properly
GCP_PROJECT="sourcegraph-ci"
GCS_BUCKET="package-repository"
TARGET_ARCH="x86_64"
branch="main"

cd wolfi-packages/packages/$TARGET_ARCH

# Use GCP tooling to upload new package to repo, ensuring it's on the right branch.
# Check that this exact package does not already exist in the repo - fail if so

# TODO: Support branches for uploading
# TODO: Check for existing files only if we're on main - overwriting is permitted on branches

echo " * Uploading package to repository"

# List all .apk files under wolfi-packages/packages/$TARGET_ARCH/
apks=(*.apk)
for apk in "${apks[@]}"; do
  echo " * Processing $apk"
  dest_path="gs://$GCS_BUCKET/packages/$branch/$TARGET_ARCH/"
  echo "   -> File path: $dest_path / $apk"

  # Generate index fragment for this package
  melange index -o "$apk.APKINDEX.tar.gz" "$apk"
  tar zxf "$apk.APKINDEX.tar.gz"
  index_fragment="$apk.APKINDEX.fragment"
  mv APKINDEX "$index_fragment"
  echo "   * Generated index fragment '$index_fragment"

  # Check if this version of the package already exists in bucket
  echo "   * Checking if this package version already exists in repo..."
  if gsutil -q -u "$GCP_PROJECT" stat "$dest_path/$apk"; then
    echo "$apk: A package with this version already exists, and cannot be overwritten."
    echo "Resolve this issue by incrementing the \`epoch\` field in the package's YAML file."
    # exit 1
  else
    echo "   * File does not exist, uploading..."
  fi

  # TODO: Pass -n when on main to avoid accidental overwriting
  echo "   * Uploading package and index fragment to repo"
  gsutil -u "$GCP_PROJECT" cp "$apk" "$index_fragment" "$dest_path"
done
