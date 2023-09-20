#!/usr/bin/env bash

# This is a very basic safeguard to catch before committing if a file contains a token. It's only a heuristic, and
# is complemented by a GitHub action that runs a real scanner and not this poor's man version of it.
# Still the value of this script lies in the fact that it can catch mistakes before committing the token locally,
# so it doesn't need to be rotated, as it was never committed.

set -eu

files=$(git diff --name-only --staged)

function check() {
  local file="$1"

  if grep -qE "s(?:g[psd]|lk)_[0-9a-fA-F]{40,}" "$file"; then
    echo "Found a Sourcegraph token in git staged file: $file. Please remove it."
    exit 1
  fi
  if grep -qE "gh[pousr]_[0-9a-zA-Z]{36}" "$file"; then
    echo "Found a GitHub token in git staged file: $file. Please remove it."
    exit 1
  fi
  if grep -qE "(?:(ABIA)|(ACCA)|(AGPA)|(AIDA)|(AIPA)|(AKIA)|(ANPA)|(ANVA)|(APKA)|(AROA)|(ASCA)|(ASIA))[0-9A-Z]{16}" "$file"; then
    echo "Found an AWS token in git staged file: $file. Please remove it."
    exit 1
  fi
  if grep -qE "sk-[0-9a-zA-Z]{48}" "$file"; then
    echo "Found an OpenAI token in git staged file: $file. Please remove it."
    exit 1
  fi
  if grep -qE "sk-[0-9a-zA-Z-]{86}" "$file"; then
    echo "Found an Anthropic token in git staged file: $file. Please remove it."
    exit 1
  fi
  if grep -qE "https://hooks.slack.com/services/[0-9A-Za-z/]{44}" "$file"; then
    echo "Found a Slack webhook in git staged file: $file. Please remove it."
    exit 1
  fi
  if grep -qE "xoxb-[0-9a-zA-Z-]{49}" "$file"; then
    echo "Found a Slack bot token in git staged file: $file. Please remove it."
    exit 1
  fi
  if grep -qE "xapp-[0-9a-zA-Z-]{92}" "$file"; then
    echo "Found a Slack app token in git staged file: $file. Please remove it."
    exit 1
  fi
  if grep -qE "service_account" "$file" && grep -qE "[0-9a-zA-Z-]*@[0-9a-zA-Z-]*\.iam\.gserviceaccount\.com" "$file"; then
    echo "Found a Google service account key file in git staged file: $file. Please remove it."
    exit 1
  fi
}

export -f check
parallel check ::: "$files"
