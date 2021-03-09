#!/usr/bin/env bash

set -eu

cd "$(dirname "${BASH_SOURCE[0]}")/../.."

export GOBIN="$PWD/.bin"
export PATH=$GOBIN:$PATH
export GO111MODULE=on

echo "--- go mod tidy"

go mod tidy

# Check if git diff contains go.mod and/or go.sum. In case the diff does not contain either of these files, `grep -c` will return a non-zero exit code leading to this script exiting as we `set -e` above. To avoid this, when `grep -c` returns no matches and fails we execute `true` to avoid a premature non-zero exit code. Detailed explanation is here:
# https://unix.stackexchange.com/questions/330660/prevent-grep-from-exiting-in-case-of-nomatch
readonly DIFF=$(git diff --name-only | grep -e 'go.mod' -e 'go.sum' -c || true)

if [[ "${DIFF}" -gt 0 ]]; then
  echo 'ERROR: go mod tidy generated a diff in go.mod and/or go.sum. Please run "go mod tidy" and commit the changes'
  exit 1
fi

echo "Success: No changes in go mod tidy"
