#!/usr/bin/env bash

function usage() {
  echo "Usage:   ./git-diff-no-enterprise.sh <base> <head>"
  echo "Example: ./git-diff-no-enterprise.sh oss/master HEAD"
  echo 'Debug:   Set VERBOSE=1'
}

if [ -z "$1" ] || [ -z "$2" ]; then
  usage
  exit 1
fi

if [ -n "$VERBOSE" ]; then
  set -x
fi
set -euo pipefail

git diff --stat "$1" "$2" ":(exclude)$(git rev-parse --show-toplevel)/enterprise" ":(exclude)$(git rev-parse --show-toplevel)/.github"
