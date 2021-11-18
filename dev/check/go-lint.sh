#!/usr/bin/env bash

echo "--- golangci-lint"

trap "echo ^^^ +++" ERR

set -ex
cd "$(dirname "${BASH_SOURCE[0]}")/../.."

export GOBIN="$PWD/.bin"
export PATH=$GOBIN:$PATH
export GO111MODULE=on

config_file="$(pwd)/.golangci.yml"
lint_script="$(pwd)/dev/golangci-lint.sh"

run() {
  LINTER_ARG=${1}

  OUT=$("$lint_script" --config "$config_file" run "$LINTER_ARG")
  EXIT_CODE=$?
  echo -e "$OUT"

  if [ $EXIT_CODE -ne 0 ]; then
    echo -e "$OUT" | ./dev/ci/annotate.sh -s "golangci-lint"
    echo "^^^ +++"

    exit $EXIT_CODE
  else
    echo $EXIT_CODE
  fi
}

# If no args are given, traverse through each project with a `go.mod`
if [ $# -eq 0 ]; then
  find . -name go.mod -exec dirname '{}' \; | while read -r d; do
    pushd "$d" >/dev/null

    echo "--- golangci-lint $d"

    run "./..."

    popd >/dev/null
  done
else
  run "$@"
fi
