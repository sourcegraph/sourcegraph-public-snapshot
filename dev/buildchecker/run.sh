#! /usr/bin/env bash

# Make this script independent of where it's called
cd "$(dirname "${BASH_SOURCE[0]}")"/../..

set -eu

echo "--- Running buildchecker"
go run ./dev/buildchecker/ \
  -buildkite.token="$BUILDKITE_TOKEN" \
  -github.token="$GITHUB_TOKEN" \
  -slack.token="$SLACK_TOKEN" \
  -slack.webhook="$SLACK_WEBHOOK"
