#! /usr/bin/env bash

# Make this script independent of where it's called
cd "$(dirname "${BASH_SOURCE[0]}")"/../..

set -eu

echo "--- Running buildchecker"
go run ./dev/buildchecker/ \
  -buildkite.token="$BUILDKITE_TOKEN" \
  -github.token="$GITHUB_TOKEN" \
<<<<<<< HEAD
  -slack.announce-webhook="$SLACK_ANNOUNCE_WEBHOOK" \
  -slack.debug-webhook="$SLACK_DEBUG_WEBHOOK"
=======
  -slack.token="$SLACK_TOKEN" \
  -slack.webhook="$SLACK_WEBHOOK"
>>>>>>> 3ea417d11119985e1d9b4b7dc4e8e91bf6624a6b
