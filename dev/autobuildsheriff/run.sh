#! /usr/bin/env bash

# Make this script independent of where it's called
cd "$(dirname "${BASH_SOURCE[0]}")"/../..

set -eu

pushd dev/buildsheriff

echo "--- Running buildsheriff"
go run main.go \
  -buildkite.token="$BUILDKITE_TOKEN" \
  -github.token="$GITHUB_TOKEN" \
  -slack.webhook="$SLACK_WEBHOOK"

popd
