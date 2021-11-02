#!/usr/bin/env bash

# This script runs the backend integration tests against a candidate server image.

cd "$(dirname "${BASH_SOURCE[0]}")/../../"
SG_ROOT=$(pwd)
set -ex

# Setup single-server instance and run tests
# backend integration tests requires a Github Enterprise Token
GITHUB_TOKEN=$GHE_GITHUB_TOKEN
GITHUB_TOKEN=$GITHUB_TOKEN ./dev/ci/run-integration.sh "${SG_ROOT}/dev/ci/backend-integration-against-server.sh"
