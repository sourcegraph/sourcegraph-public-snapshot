#!/usr/bin/env bash

cd "$(dirname "${BASH_SOURCE[0]}")"/../..
set -ex

function run_tests() {
  echo '--- integration test ./dev/gqltest -long'
  go test ./dev/gqltest -long

  echo '--- sleep 5s to wait for site configuration to be restored from gqltest'
  sleep 5

  echo '--- integration test ./dev/authtest -long'
  go test ./dev/authtest -long -email "gqltest@sourcegraph.com" -username "gqltest-admin"
}

# Setup single-server instance and run tests
./dev/ci/backend-integration-setup.sh run_tests
