#!/usr/bin/env bash
set -e

# See https://stackoverflow.com/a/5148851
require_clean_work_tree() {
  git rev-parse --verify HEAD >/dev/null || exit 1
  git update-index -q --ignore-submodules --refresh
  err=0

  if ! git diff-files --quiet --ignore-submodules; then
    echo >&2 "Cannot $1: You have unstaged changes."
    err=1
  fi

  if ! git diff-index --cached --quiet --ignore-submodules HEAD --; then
    if [ $err = 0 ]; then
      echo >&2 "Cannot $1: Your index contains uncommitted changes."
    else
      echo >&2 "Additionally, your index contains uncommitted changes."
    fi
    err=1
  fi

  if [ $err = 1 ]; then
    test -n "$2" && echo >&2 "$2"
    exit 1
  fi
}

# Must be on master branch.
BRANCH=$(git rev-parse --abbrev-ref HEAD)
if [[ "$BRANCH" != "master" ]]; then
  echo 'Must be on master branch.'
  exit 1
fi

require_clean_work_tree "publish"

while true; do
  read -r -p "Did you already run ./build.sh? [y/n] " yn
  case $yn in
    [Yy]*) break ;;
    [Nn]*) echo "Please run ./build.sh first." && exit ;;
    *) echo "Please answer yes or no." ;;
  esac
done

VERSION=$(git rev-parse --short HEAD)

echo docker push sourcegraph/syntect_server
docker push sourcegraph/syntect_server

docker tag sourcegraph/syntect_server sourcegraph/syntect_server:"$VERSION"
echo docker push sourcegraph/syntect_server:"$VERSION"
docker push sourcegraph/syntect_server:"$VERSION"

docker tag sourcegraph/syntect_server us.gcr.io/sourcegraph-dev/syntect_server:"$VERSION"
echo docker push us.gcr.io/sourcegraph-dev/syntect_server:"$VERSION"
docker push us.gcr.io/sourcegraph-dev/syntect_server:"$VERSION"
