#!/usr/bin/env bash

set -euo pipefail

# Required to stop .git ownership error
# https://github.com/actions/runner/issues/2033
git config --global --add safe.directory /github/workspace

git blame -w --line-porcelain -- CHANGELOG.md |
  progress-bot -since="$SINCE" -dry="$DRY" -channel="$CHANNEL"
