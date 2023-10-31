#!/usr/bin/env bash

set -eu

runGoModTidy() {
  local exit_code
  local dir
  exit_code=0
  dir=$1

  cd $dir
  echo "--- :bazel: Running go mod tidy in $dir"
  bazel run @go_sdk//:bin/go -- mod tidy
  cd -
}

# search for go.mod and run `go mod tidy` in the directory containing the go.mod
find . -name go.mod -type f -exec dirname '{}' \; | while read -r d; do runGoModTidy $d; done
# check if go.mod got updated
git ls-files --exclude-standard --others | grep go.mod | xargs git add --intent-to-add
git diff --exit-code || exit_code=$? # do not fail on non-zero exit

# if we have a diff, go.mod got updated so we notify people
if [[ $exit_code -ne 0 ]]; then
  echo "--- :x: One or more go.mod files are out of date. Please see the diff for the directory and run 'go mod tidy'"
  exit 1
fi
