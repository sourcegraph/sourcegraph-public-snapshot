#!/usr/bin/env bash

set -eu -o pipefail

cd "$(dirname "${BASH_SOURCE[0]}")/../../../../.."

# TODO: Manage these variables properly
GCP_PROJECT="sourcegraph-ci"
GCS_BUCKET="package-repository"
TARGET_ARCH="x86_64"
MAIN_BRANCH="main"
BRANCH="${BUILDKITE_BRANCH:-'default-branch'}"
IS_MAIN=$([ "$BRANCH" = "$MAIN_BRANCH" ] && echo "true" || echo "false")

# shellcheck disable=SC2001
BRANCH_PATH=$(echo "$BRANCH" | sed 's/[^a-zA-Z0-9_-]/-/g')
if [[ "$IS_MAIN" != "true" ]]; then
  BRANCH_PATH="branches/$BRANCH_PATH"
fi

cd wolfi-packages/packages/$TARGET_ARCH

# Check that this exact package does not already exist in the repo - fail if so

echo " * Uploading package to repository"

# List all .apk files under wolfi-packages/packages/$TARGET_ARCH/
error="false"
package_usage_list=""
apks=(*.apk)
for apk in "${apks[@]}"; do
  echo " * Processing $apk"

  # Generate the branch-specific path to upload the package to
  dest_path="gs://$GCS_BUCKET/$BRANCH_PATH/$TARGET_ARCH/"
  echo "   -> File path: ${dest_path}${apk}"

  # Generate the path to the package file on the main branch
  dest_path_main="gs://$GCS_BUCKET/$MAIN_BRANCH/$TARGET_ARCH/"

  # Generate index fragment for this package
  melange index -o "$apk.APKINDEX.tar.gz" "$apk"
  tar zxf "$apk.APKINDEX.tar.gz"
  index_fragment="$apk.APKINDEX.fragment"
  mv APKINDEX "$index_fragment"
  echo "   * Generated index fragment '$index_fragment"

  # Check whether this version of the package already exists in the main package repo
  echo "   * Checking if this package version already exists in the production repo..."
  if gsutil -q -u "$GCP_PROJECT" stat "${dest_path_main}${apk}"; then
    echo "The production package repository already contains a package with this version: $apk" >&2
    echo "   -> Production repository file path: ${dest_path_main}${apk}" >&2
    echo "Resolve this issue by incrementing the \`epoch\` field in the package's YAML file." >&2
    # Soft fail at the end - we still want to allow the package to be uploaded for cases like a Buildkite pipeline being rerun
    error="true"
  else
    echo "   * File does not exist, uploading..."
  fi

  # no-cache to avoid index/packages getting out of sync
  echo "   * Uploading package and index fragment to repo"
  gsutil -u "$GCP_PROJECT" -h "Cache-Control:no-cache" cp "$apk" "$index_fragment" "$dest_path"

  # Concat package names for annotation
  package_name=$(echo "$apk" | sed -E 's/(-[0-9].*)//')
  package_usage_list="$package_usage_list    - ${package_name}@branch\n"
done

# Show package usage message on branches
if [[ "$IS_MAIN" != "true" ]]; then
  if [[ -n "$BUILDKITE" ]]; then
    echo -e "To use this package locally, add the following lines to your base image config:
\`\`\`
contents:
  keyring:
    - https://packages.sgdev.org/sourcegraph-melange-dev.rsa.pub
  repositories:
    - '@branch https://packages.sgdev.org/${BRANCH_PATH}'
  packages:
$package_usage_list
  \`\`\`" | ../../../enterprise/dev/ci/scripts/annotate.sh -s "custom-repo" -m -t "info"
  fi
fi

if [[ "$error" == "true" ]]; then
  exit 222 # Soft fail
fi
