#!/usr/bin/env bash

set -euxo pipefail

cd "$GIT_DIR"

# chore: get rid of GNU parallel citation spam
{
  mkdir -p "$HOME"/.parallel && touch "$HOME"/.parallel/will-cite
  echo 'will cite' | parallel --citation &>/dev/null
}

function print_remote_and_default_branch() {
  local remote="$1"

  local default_branch
  default_branch="$(git remote show "$remote" | sed -n '/HEAD branch/s/.*: //p')"

  echo "${remote}" "${default_branch}"
}
export -f print_remote_and_default_branch

# discover the current default branch for each remote
mapfile -t remotes_to_default < <(git remote | parallel --keep-order --line-buffer print_remote_and_default_branch)

# set the appropriate default branch for each remote
for line in "${remotes_to_default[@]}"; do
  remote="$(awk '{printf $1}' <<<"$line")"
  default="$(awk '{printf $2}' <<<"$line")"

  git remote set-branches "$remote" "$default"
done
