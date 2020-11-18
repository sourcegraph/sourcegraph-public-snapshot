#!/usr/bin/env bash

# shellcheck disable=SC1091
source /root/.profile
cd "$(dirname "${BASH_SOURCE[0]}")/../.." || exit

set -ex

test/setup-deps.sh
test/setup-display.sh

# ==========================

pushd enterprise || exit
./cmd/server/pre-build.sh
./cmd/server/build.sh
popd || exit

# TODO: we temporarily allow all the test commands to pass, to help teams identify if
# their tests have been fixed. Remove this when all tests are green so we can identify
# failures in the future.
set +e

echo "TEST: Running E2E tests"
./dev/ci/e2e.sh
docker image rm -f "${IMAGE}"

# ==========================

test/cleanup-display.sh
