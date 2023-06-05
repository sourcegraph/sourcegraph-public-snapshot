#!/usr/bin/env bash

# This script runs the backend integration tests against a running server.

cd "$(dirname "${BASH_SOURCE[0]}")/../../../.."
set -e

URL="${1:-"http://localhost:7080"}"

echo '--- integration test ./dev/gqltest -long'
bazel run //dev/gqltest:gqltest_test -- -long -base-url "$URL"

echo '--- sleep 5s to wait for site configuration to be restored from gqltest'
sleep 5

echo '--- integration test ./dev/authtest -long'
bazel run //dev/authtest:authtest_test -- -long -base-url "$URL" -email "gqltest@sourcegraph.com" -username "gqltest-admin"
