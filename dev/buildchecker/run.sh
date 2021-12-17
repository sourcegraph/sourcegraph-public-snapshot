#! /usr/bin/env bash

# Make this script independent of where it's called
cd "$(dirname "${BASH_SOURCE[0]}")"/../..

set -eu

pushd dev/buildchecker

echo "--- Running buildchecker"
go run main.go \
  -buildkite.token="$BUILDKITE_TOKEN" \
  -github.token="$GITHUB_TOKEN" \
  -slack.webhook="$SLACK_WEBHOOK"

popd
