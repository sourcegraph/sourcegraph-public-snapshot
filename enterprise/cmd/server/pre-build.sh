#!/usr/bin/env bash

cd $(dirname "${BASH_SOURCE[0]}")/../../..
set -euxo pipefail

echo "--- (enterprise) pre-build frontend"
./enterprise/cmd/frontend/pre-build.sh
