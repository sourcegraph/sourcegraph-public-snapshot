#!/usr/bin/env bash

cd "$(dirname "${BASH_SOURCE[0]}")"/../..
set -euo pipefail

check_sd_installed() {
  if ! command -v sd &>/dev/null; then
    echo "'sd' command not installed. Please install 'sd' by following the instructions on https://github.com/chmln/sd#installation"
    exit 1
  fi
}

update_image_reference() {
  local old_image_stub="$1"
  local new_tag_and_digest="$2"
  local file="$3"

  local original="(?P<repo>$old_image_stub:)(\S*@sha256:\S*)"
  local replacement="\${repo}$new_tag_and_digest"

  sd "$original" "$replacement" "$file"
}

get_new_tag_and_digest() {
  local repo="$1"
  local tag="$2"
  local image="$repo:$tag"

  docker pull "$image" >/dev/null 2>&1

  local digest
  digest="$(docker inspect --format='{{index .RepoDigests 0}}' "$image" | sed "s~$repo@~~g" | tr -d '\n')"
  echo -n "$tag@$digest"
}

check_sd_installed

REPO="sourcegraph/alpine-3.12"

MISSING_MESSAGE="Please provide the image tag either via the 'TAG' environent variable or as a shell script argument"
TAG="${TAG:-${1:?"$MISSING_MESSAGE"}}"

NEW_TAG_AND_DIGEST="$(get_new_tag_and_digest "$REPO" "$TAG")"

DOCKERFILES=()
mapfile -t DOCKERFILES < <(fd --glob Dockerfile .)

for file in "${DOCKERFILES[@]}"; do
  update_image_reference "$REPO" "$NEW_TAG_AND_DIGEST" "$file"
done
