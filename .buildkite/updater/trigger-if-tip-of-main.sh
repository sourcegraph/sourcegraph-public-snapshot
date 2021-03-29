#!/usr/bin/env bash

cd "$(dirname "${BASH_SOURCE[0]}")/../.."
set -euo pipefail

if ! .buildkite/updater/is-tip-of-main.sh; then
  echo "🚨 This commit is not the tip of main (either it's behind or an unrelated commit). Skipping deployment triggers..."
  exit 0 # This is not a failure condition.
fi

buildkite-agent pipeline upload '.buildkite/updater/pipeline.update-trigger.yaml'
