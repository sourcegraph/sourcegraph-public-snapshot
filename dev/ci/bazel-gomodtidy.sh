#!/usr/bin/env bash

set -eu
EXIT_CODE=0

runGoModTidy() {
  local dir
  dir=$1
  cd "$dir"
  echo "--- :bazel: Running go mod tidy in $dir"
  bazel run @go_sdk//:bin/go -- mod tidy
  cd -
}

# search for go.mod and run `go mod tidy` in the directory containing the go.mod
find . -name go.mod -type f -exec dirname '{}' \; | while read -r dir; do runGoModTidy "${dir}"; done
# check if go.mod got updated
git ls-files --exclude-standard --others | grep go.mod | xargs git add --intent-to-add

diffFile=$(mktemp)
trap 'rm -f "${diffFile}"' EXIT

git diff --color=always --output="${diffFile}" --exit-code || EXIT_CODE=$?

# if we have a diff, go.mod got updated so we notify people
if [[ $EXIT_CODE -ne 0 ]]; then
  echo "--- :x: One or more go.mod files are out of date. Please see the diff for the directory and run 'go mod tidy'"
  cat "${diffFile}"
  exit 1
fi
