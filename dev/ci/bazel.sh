#!/bin/bash

if [[ "${CI:-false}" == "true" ]]; then
  bazel \
    --bazelrc=.bazelrc \
    --bazelrc=.aspect/bazelrc/ci.bazelrc \
    --bazelrc=.aspect/bazelrc/ci.sourcegraph.bazelrc \
    "$@" \
    --remote_cache="${CI_BAZEL_REMOTE_CACHE}" \
    --google_credentials=/mnt/gcloud-service-account/gcloud-service-account.json
else
  bazel "$@"
fi
