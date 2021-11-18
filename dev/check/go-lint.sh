#!/usr/bin/env bash

echo "--- golangci-lint"

set -e

cd "$(dirname "${BASH_SOURCE[0]}")/../.."

export GOBIN="$PWD/.bin"
export PATH=$GOBIN:$PATH
export GO111MODULE=on

config_file="$(pwd)/.golangci.yml"
lint_script="$(pwd)/dev/golangci-lint.sh"
global_exit_code=0
annotate_script="$(pwd)/dev/ci/annotate.sh"

run() {
  LINTER_ARG=${1}

  set +e
  OUT=$("$lint_script" --config "$config_file" run "$LINTER_ARG")
  EXIT_CODE=$?
  set -e

  echo -e "$OUT"

  if [ $EXIT_CODE -ne 0 ]; then
    global_exit_code="$EXIT_CODE"

    echo -e "$OUT" | "$annotate_script" -s "golangci-lint"
    echo "^^^ +++"
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

exit $global_exit_code
