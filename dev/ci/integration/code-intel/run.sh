#!/usr/bin/env bash

# This script runs the codeintel-qa test utility against a candidate server image.

cd "$(dirname "${BASH_SOURCE[0]}")/../../../.."
root_dir=$(pwd)
set -ex

echo "--- :terminal: Installing src-cli latest release"
curl -L https://sourcegraph.com/.api/src-cli/src_linux_amd64 -o /usr/local/bin/src
chmod +x /usr/local/bin/src
src version

echo "--- :spiral_note_pad: test.sh"
export IMAGE="us.gcr.io/sourcegraph-dev/server:${CANDIDATE_VERSION}"
./dev/ci/integration/run-integration.sh "${root_dir}/dev/ci/integration/code-intel/test.sh"
