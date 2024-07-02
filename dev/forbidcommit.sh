#!/usr/bin/env bash

set -eu

# Select the files to inspect. We don't want to list files which are deleted, as it makes no sense
# to look for a token being committed in those.
#
# We use --diff-filter, that tells git to only include in the diff files that were:
# - "A" added
# - "C" copied
# - "M" modified
# - "R" renamed
files=$(git diff --name-only --staged --diff-filter ACMR)

function check() {
  local file="$1"
  if [[ "$file" == "dev/forbidcommit.sh" || "$file" == ".pre-commit-config.yaml" ]]; then
    exit 0
  fi

  # -i means case-insensitive
  # -F means searsch for a fixed string, i.e. no pattern matching.
  if grep -iF 'FORBIDCOMMIT' "$file"; then
    echo "🔴 Found a FORBIDCOMMIT string in git staged file: $file."
    echo "You most likely added this yourself to prevent yourself from accidentally committing that file."
    echo "Please look at your changes carefully before removing the FORBIDCOMMIT pragma."
    exit 1
  fi
}

export -f check
parallel check ::: "$files"
