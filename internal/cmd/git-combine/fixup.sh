#!/usr/bin/env bash

set -eux

cd "$GIT_DIR"

function get_default_branch() {
  local remote="$1"

  git remote show "$remote" | sed -n '/HEAD branch/s/.*: //p'
}

# Remove parallel citation log spam.
echo 'will cite' | parallel --citation &>/dev/null

git remote | while read -r remote; do
  default="$(git remote show "$remote" | sed -n '/HEAD branch/s/.*: //p')"
  git remote set-branches "$remote" "$default"
done
