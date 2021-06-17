#!/usr/bin/env bash

cd "$(dirname "${BASH_SOURCE[0]}")"/../..
set -euxo pipefail

update_image_reference() {
  local old_image_stub="$1"
  local new_image="$2"
  local file="$3"

  local new_text
  new_text="$(sed -E "s|($old_image_stub.*)(\s*)(.*)|$new_image\2\3|g" "$file")"

  echo "$new_text" >"$file"
}

get_pinned_image() {
  local repo="$1"
  local tag="$2"
  local image="$repo:$tag"

  docker pull "$image" >/dev/null 2>&1

  local digest
  digest="$(docker inspect --format='{{index .RepoDigests 0}}' "$image" | sed "s~$repo@~~g" | tr -d '\n')"
  echo -n "$image@$digest"
}

REPO="sourcegraph/alpine-3.12"

MISSING_MESSAGE="Please provide the image tag either via the 'TAG' environent variable or as a shell script argument"
TAG="${TAG:-${1:?"$MISSING_MESSAGE"}}"

IMAGE="$(get_pinned_image "$REPO" "$TAG")"

DOCKERFILES=()
mapfile -t DOCKERFILES < <(fd --glob Dockerfile .)

for file in "${DOCKERFILES[@]}"; do
  update_image_reference "$REPO" "$IMAGE" "$file"
done
