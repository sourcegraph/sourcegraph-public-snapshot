#!/usr/bin/env bash

set -ex
cd "$(dirname "${BASH_SOURCE[0]}")"

echo "!!!!!!!!!!!!!!!!!!"
echo "!!! DEPRECATED !!!"
echo "!!!!!!!!!!!!!!!!!!"
echo "This script is deprecated!"
echo "Add your checks to 'dev/sg/linters' instead."

parallel_run() {
  ../ci/parallel_run.sh "$@"
}

go version
go env

CHECKS=(
  ./gofmt.sh
  ./template-inlines.sh
  ./go-dbconn-import.sh
  ./go-lint.sh
  ./no-localhost-guard.sh
  ./bash-syntax.sh
  ./shfmt.sh
  ./shellcheck.sh
  ./submodule.sh
)

echo "--- 🚨 Buildkite's timing information is misleading! Only consider the job timing that's printed after 'done'"

MAYBE_TIME_PREFIX=""
if [[ "${CI_DEBUG_PROFILE:-"false"}" == "true" ]]; then
  MAYBE_TIME_PREFIX="env time -v"
fi

parallel_run "${MAYBE_TIME_PREFIX}" {} ::: "${CHECKS[@]}"

# TODO(sqs): Reenable this check when about.sourcegraph.com is reliable. Most failures come from its
# downtime, not from broken URLs.
#
# ./broken-urls.bash
