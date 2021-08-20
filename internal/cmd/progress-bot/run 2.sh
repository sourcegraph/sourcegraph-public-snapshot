#!/usr/bin/env bash

set -euo pipefail

git blame -w --line-porcelain -- CHANGELOG.md |
  progress-bot -since="$SINCE" -dry="$DRY" -channel="$CHANNEL"
