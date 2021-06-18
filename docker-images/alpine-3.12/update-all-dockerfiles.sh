#!/usr/bin/env bash

cd "$(dirname "${BASH_SOURCE[0]}")"/../..
set -euxo pipefail

update_image_reference() {
  local old_image_stub="$1"
  local new_tag_and_digest="$2"
  local file="$3"

  # (sourcegraph\/alpine-3.12(\S*))(\s*)((AS)?.*)
  # local original="($old_image_stub:([^:space:]*))([:space:]*)((AS)?.*)"
  local original="$old_image_stub:[[^:space:]]*@(sha256:[[^:space:]]*)"
  local replacement="\1"

  local new_text
  new_text="$(sed -E "s|$original|$replacement|g" "$file")"

  echo "$new_text" >"$file"
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

REPO="sourcegraph/alpine-3.12"

MISSING_MESSAGE="Please provide the image tag either via the 'TAG' environent variable or as a shell script argument"
TAG="${TAG:-${1:?"$MISSING_MESSAGE"}}"

NEW_TAG_AND_DIGEST="$(get_new_tag_and_digest "$REPO" "$TAG")"

DOCKERFILES=()
mapfile -t DOCKERFILES < <(fd --glob Dockerfile .)

for file in "${DOCKERFILES[@]}"; do
  update_image_reference "$REPO" "$NEW_TAG_AND_DIGEST" "$file"
done
