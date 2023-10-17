#!/usr/bin/env bash

# This script publishes the executor binary.

cd "$(dirname "${BASH_SOURCE[0]}")"/../..
set -eu

# Copy uploaded binary to 'latest'.
# TODO: Revise this. latest should probably not exist at all.
if [ "${BUILDKITE_BRANCH}" = "main" ]; then
  gsutil rm -rf gs://sourcegraph-artifacts/executor/latest || true
  gsutil cp -r "gs://sourcegraph-artifacts/executor/$(git rev-parse HEAD)" gs://sourcegraph-artifacts/executor/latest
  gsutil iam ch allUsers:objectViewer gs://sourcegraph-artifacts
fi

# If this is a tagged release, we want to create a directory for it.
if [ "${EXECUTOR_IS_TAGGED_RELEASE}" = "true" ]; then
  gsutil rm -rf "gs://sourcegraph-artifacts/executor/${BUILDKITE_TAG}" || true
  gsutil cp -r "gs://sourcegraph-artifacts/executor/$(git rev-parse HEAD)" "gs://sourcegraph-artifacts/executor/${BUILDKITE_TAG}"
  gsutil iam ch allUsers:objectViewer gs://sourcegraph-artifacts
fi
