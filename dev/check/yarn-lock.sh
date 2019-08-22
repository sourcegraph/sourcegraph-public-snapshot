#!/bin/bash

cd "$(dirname "${BASH_SOURCE[0]}")"

set -eu -o pipefail

echo "--- make sure yarn.lock doesn't change when running yarn"

yarn --cwd cmd/management-console
git diff --exit-code -- cmd/management-console/yarn.lock ':!go.sum'

yarn --cwd lsif/server
git diff --exit-code -- lsif/server/yarn.lock ':!go.sum'
