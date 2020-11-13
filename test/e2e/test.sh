#!/usr/bin/env bash

# shellcheck disable=SC1091
source /root/.profile
cd "$(dirname "${BASH_SOURCE[0]}")/../.." || exit

set -x

test/setup-deps.sh
test/setup-display.sh

# ==========================

pushd enterprise || exit
./cmd/server/pre-build.sh
./cmd/server/build.sh
popd || exit
./dev/ci/e2e.sh
docker image rm -f "${IMAGE}"

# ==========================

test/cleanup-display.sh
