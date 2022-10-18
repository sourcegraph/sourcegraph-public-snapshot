#!/usr/bin/env bash

set -euxo pipefail

TMP=$(mktemp -d)
cleanup() {
  rm -rf "$TMP"
}
trap cleanup EXIT

cd "$GIT_DIR"

# chore: get rid of GNU parallel citation spam
{
  mkdir -p "$HOME"/.parallel && touch "$HOME"/.parallel/will-cite
  echo 'will cite' | parallel --citation &>/dev/null
}

function print_remote_and_default_branch() {
  # set shell options again since they aren't exported
  # when we export the bash function
  set -euo pipefail

  local remote="$1"

  local default_branch
  default_branch="$(git remote show "$remote" | awk '/HEAD branch/ {print $NF}' | grep .)"

  echo "${remote}" "${default_branch}"
}
export -f print_remote_and_default_branch

# discover the current default branch for each remote
{
  defaults_file="${TMP}/remotes_branches.txt"
  git remote |
    parallel \
      --keep-order \
      --line-buffer \
      print_remote_and_default_branch \
      >"$defaults_file"
}

# set the appropriate default branch for each remote
while IFS= read -r line; do
  remote="$(awk '{printf $1}' <<<"$line")"
  default="$(awk '{printf $2}' <<<"$line")"

  git remote set-branches "$remote" "$default"
done <"$defaults_file"
