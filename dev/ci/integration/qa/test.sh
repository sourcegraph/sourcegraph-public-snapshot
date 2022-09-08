#!/usr/bin/env bash

export SOURCEGRAPH_BASE_URL="${1:-"http://localhost:7080"}"

# shellcheck disable=SC1091
if [[ $(id -u) -eq 1 ]]; then
    source /root/.profile
else
    # shellcheck disable=SC1090
    source "${HOME}/.profile"
fi

cd "$(dirname "${BASH_SOURCE[0]}")/../../../.."

set -e

echo "--- init sourcegraph"
pushd internal/cmd/init-sg
go build
./init-sg initSG
popd
# Load variables set up by init-server, disabling `-x` to avoid printing variables
set +x
# shellcheck disable=SC1091
if [[ $(id -u) -eq 1 ]]; then
    source /root/.sg_envrc
else
    # shellcheck disable=SC1090
    source "${HOME}/.sg_envrc"
fi

echo "--- TEST: Checking Sourcegraph instance is accessible"
curl -f "${SOURCEGRAPH_BASE_URL}"
curl -f "${SOURCEGRAPH_BASE_URL}/healthz"
echo "--- TEST: Running tests"
# Run all tests, and error if one fails
test_status=0
pushd client/web
yarn run test:regression || test_status=1
popd
exit $test_status
