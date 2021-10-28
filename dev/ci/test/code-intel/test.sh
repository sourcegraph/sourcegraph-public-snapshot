#!/usr/bin/env bash

# This script runs the codeintel-qa test utility against a candidate server image.

cd "$(dirname "${BASH_SOURCE[0]}")/../../../.."
SG_ROOT=$(pwd)
set -ex

# Use candidate image built by main pipeline
export IMAGE="us.gcr.io/sourcegraph-dev/server:${CANDIDATE_VERSION}"

# TODO(JH) Remove that before merging the PR
if [[ -n "$DUP_GITHUB_TOKEN" ]]; then
  GITHUB_TOKEN=$DUP_GITHUB_TOKEN
fi

# Setup single-server instance and run tests
GITHUB_TOKEN=$GITHUB_TOKEN ./dev/ci/run-integration.sh "${SG_ROOT}/dev/ci/test/code-intel/test-against-server.sh"
