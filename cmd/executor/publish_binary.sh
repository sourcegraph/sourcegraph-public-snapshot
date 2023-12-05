#!/usr/bin/env bash

# --- begin runfiles.bash initialization v3 ---
# Copy-pasted from the Bazel Bash runfiles library v3.
set -uo pipefail; set +e; f=bazel_tools/tools/bash/runfiles/runfiles.bash
source "${RUNFILES_DIR:-/dev/null}/$f" 2>/dev/null || \
  source "$(grep -sm1 "^$f " "${RUNFILES_MANIFEST_FILE:-/dev/null}" | cut -f2- -d' ')" 2>/dev/null || \
  source "$0.runfiles/$f" 2>/dev/null || \
  source "$(grep -sm1 "^$f " "$0.runfiles_manifest" | cut -f2- -d' ')" 2>/dev/null || \
  source "$(grep -sm1 "^$f " "$0.exe.runfiles_manifest" | cut -f2- -d' ')" 2>/dev/null || \
  { echo>&2 "ERROR: cannot find $f"; exit 1; }; f=; set -e
# --- end runfiles.bash initialization v3 ---

## Setting up tools
gsutil=$(rlocation sourcegraph_workspace/dev/tools/gsutil)

## Setting up inputs
executor_binary="$1"

## Script begins here
set -eu

GIT_COMMIT="$(git rev-parse HEAD)"

## Prepare artifacts
mkdir -p workdir/linux-amd64
workdir_abs="$(pwd)/workdir"
trap "rm -Rf $workdir_abs" EXIT

cp "$executor_binary" workdir/linux-amd64
echo >>workdir/info.txt
git log -n1 >>"workdir/info.txt"
sha256sum workdir/linux-amd64/executor >>workdir/linux-amd64/executor_SHA256SUM

# Set GCP Project to sourcegraph-dev, that's where the GCS bucket lives.
export CLOUDSDK_CORE_PROJECT="sourcegraph-dev"

echo "Uploading binaries for this commit"
"$gsutil" cp -r workdir/* "gs://sourcegraph-artifacts/executor/${GIT_COMMIT}"

if [ "${BUILDKITE_BRANCH}" = "main" ]; then
  echo "Uploading binaries as latest"
  # Drop the latest folder, as we'll overwrite it.
  "$gsutil" rm -rf gs://sourcegraph-artifacts/executor/latest || true
  "$gsutil" cp -r workdir/* gs://sourcegraph-artifacts/executor/latest/
fi

# If this is a tagged release, we want to create a directory for it.
if [ "${EXECUTOR_IS_TAGGED_RELEASE}" = "true" ]; then
  echo "Uploading binaries for the ${BUILDKITE_TAG} tag"
  # Drop the tag if existing, allowing for rebuilds.
  "$gsutil" rm -rf "gs://sourcegraph-artifacts/executor/${BUILDKITE_TAG}" || true
  "$gsutil" cp -r workdir/* "gs://sourcegraph-artifacts/executor/${BUILDKITE_TAG}/"
fi

# Just in case.
"$gsutil" iam ch allUsers:objectViewer gs://sourcegraph-artifacts
