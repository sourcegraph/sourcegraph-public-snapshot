#!/usr/bin/env bash
# Update all Docker images with a new alpine base image.

# Before running this script, update the alpine image and merge into main. CI will then push the new image
# to Docker Hub. Observe the tag of the new image and call this script to update all Docker images.
#
# Usage:
# $ ./update-all-dockerfiles.sh <old image> <new image> <tag>
#
# Example:
# $ ./update-all-dockerfiles.sh sourcegraph/alpine-1.12 sourcegraph/alpine-1.14 117627_2021-11-24_7ab43a9

OLD_IMAGE=$1
NEW_IMAGE=$2
NEW_TAG=$3

cd "$(dirname "${BASH_SOURCE[0]}")"/../..
set -euo pipefail

check_sd_installed() {
  if ! command -v sd &>/dev/null; then
    echo "'sd' command not installed. Please install 'sd' by following the instructions on https://github.com/chmln/sd#installation"
    exit 1
  fi
}

check_fd_installed() {
  if ! command -v fd &>/dev/null; then
    echo "'fd' command not installed. Please install 'fd' by following the instructions on https://github.com/sharkdp/fd"
    exit 1
  fi
}

update_image_reference() {
  local old_image_stub="$1"
  local new_image_stub="$2"
  local new_tag_and_digest="$3"
  local file="$4"

  local original="$old_image_stub:(\S*@sha256:\S*)"
  local replacement="$new_image_stub:$new_tag_and_digest"

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
check_fd_installed

MISSING_MESSAGE="Please provide the image tag either via the 'TAG' environent variable or as a shell script argument"
TAG="${TAG:-${NEW_TAG:?"$MISSING_MESSAGE"}}"

NEW_TAG_AND_DIGEST="$(get_new_tag_and_digest "$NEW_IMAGE" "$TAG")"

DOCKERFILES=$(fd --glob Dockerfile\* .)
for file in $DOCKERFILES; do
  update_image_reference "$OLD_IMAGE" "$NEW_IMAGE" "$NEW_TAG_AND_DIGEST" "$file"
  echo "$file"
done
