#! /usr/bin/env bash

# Make this script independent of where it's called
cd "$(dirname "${BASH_SOURCE[0]}")"/../..

set -eu

echo "--- Running 'pr-auditor'"
go run ./dev/pr-auditor/ \
  -github.payload-path="$GITHUB_EVENT_PATH" \
  -github.token="$GITHUB_TOKEN" \
  -github.run-url="$GITHUB_RUN_URL" \
  -skip-check-test-plan="${SKIP_CHECK_TEST_PLAN:-False}" \
  -skip-check-reviews="${SKIP_CHECK_REVIEWS:-False}"
