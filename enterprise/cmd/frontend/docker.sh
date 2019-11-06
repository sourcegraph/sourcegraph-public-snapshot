#!/usr/bin/env bash

cd "$(dirname "${BASH_SOURCE[0]}")/../../.."
set -euxo pipefail

export FRONTEND_PKG="github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend"
export PRE_BUILD_SCRIPT="enterprise/cmd/frontend/pre-build.sh"

export ENTERPRISE="true"

./cmd/frontend/docker.sh
