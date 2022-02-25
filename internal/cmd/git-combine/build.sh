#!/usr/bin/env bash

set -e
cd "$(dirname "${BASH_SOURCE[0]}")"

version="$1"
commit="$(git rev-parse HEAD)"
image="us.gcr.io/sourcegraph-dev/git-combine:$version"

if [[ ! ($version =~ ^[0-9]+\.[0-9]+\.[0-9]+$) ]]; then
  echo -e "USAGE: build.sh VERSION\n\nVERSION is a string like 0.0.1"
  exit 1
fi

set -x

docker build --platform=linux/amd64 \
  --build-arg VERSION="$version" \
  --build-arg COMMIT_SHA="$commit" \
  --tag "$image" \
  .

docker push "$image"
