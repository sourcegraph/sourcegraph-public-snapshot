#!/usr/bin/env bash

set -euxo pipefail
cd "$(dirname "${BASH_SOURCE[0]}")"/../../..

echo "--- (enterprise) pre-build frontend"
./enterprise/cmd/frontend/pre-build.sh
