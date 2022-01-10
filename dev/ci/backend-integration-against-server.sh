#!/usr/bin/env bash

# This script runs the backend integration tests against a running server.
# This script is invoked by ./dev/ci/run-integration.sh after running an instance.

cd "$(dirname "${BASH_SOURCE[0]}")/../.."
set -ex

echo '--- integration test ./dev/gqltest -long'
go test ./dev/gqltest -long

echo '--- sleep 5s to wait for site configuration to be restored from gqltest'
sleep 5

echo '--- integration test ./dev/authtest -long'
go test ./dev/authtest -long -email "gqltest@sourcegraph.com" -username "gqltest-admin"
