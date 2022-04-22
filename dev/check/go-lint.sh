#!/usr/bin/env bash

echo "--- golangci-lint"
trap 'rm -f "$TMPFILE"' EXIT
set -e
TMPFILE=$(mktemp)

echo "0" >"$TMPFILE"

cd "$(dirname "${BASH_SOURCE[0]}")/../.."
base="$(pwd)"

export GOBIN="$PWD/.bin"
export PATH=$GOBIN:$PATH
export GO111MODULE=on

config_file="$(pwd)/.golangci.yml"
lint_script="$(pwd)/dev/golangci-lint.sh"

run() {
  LINTER_ARG=${1}

  set +e
  OUT=$("$lint_script" --config "$config_file" run "$LINTER_ARG")
  EXIT_CODE=$?
  set -e

  echo -e "$OUT"

  if [ $EXIT_CODE -ne 0 ]; then
    # We want to return after running all tests, we don't want to fail fast, so
    # we store the EXIT_CODE (in a tmp file as this is running in a sub-shell).
    echo "$EXIT_CODE" >"$TMPFILE"
    mkdir -p "$base/annotations"
    echo -e "$OUT" >"$base/annotations/go-lint"
    echo "^^^ +++"
  fi
}

# Used to ignore directories (for example, when using submodules)
#   (It appears to be unused, but it's actually used doing -v below)
#
# shellcheck disable=SC2034
declare -A IGNORED_DIRS=(
  ["./docker-images/syntax-highlighter"]=1
  ["./internal/codeintel/dependencies/internal/lockfiles/testdata/parse"]=1
)

# If no args are given, traverse through each project with a `go.mod`
if [ $# -eq 0 ]; then
  find . -name go.mod -type f -exec dirname '{}' \; | while read -r d; do

    # Skip any ignored directories.
    if [ -v "IGNORED_DIRS[$d]" ]; then
      continue
    fi

    pushd "$d" >/dev/null

    echo "--- golangci-lint $d"

    run "./..."

    popd >/dev/null
  done
else
  run "$@"
fi

read -r EXIT_CODE <"$TMPFILE"
exit "$EXIT_CODE"
