#!/usr/bin/env bash

set -e
unset CDPATH
cd "$(dirname "${BASH_SOURCE[0]}")/.." # cd to repo root dir

parallel_run() {
  ./dev/ci/parallel_run.sh "$@"
}

export ARGS=$*

DIRS=(
  client/web
  client/shared
  client/branded
  client/browser
  client/build-config
  client/common
  client/search
  client/search-ui
  client/http-client
  client/codeintellify
  client/wildcard
  client/template-parser
  client/extension-api
  client/eslint-plugin-sourcegraph
  client/extension-api-types
  client/storybook
  client/client-api
  dev/release
)

run_command() {
  local MAYBE_TIME_PREFIX=""
  if [[ "${CI_DEBUG_PROFILE:-"false"}" == "true" ]]; then
    MAYBE_TIME_PREFIX="env time -v"
  fi

  dir=$1
  echo "--- $dir: $ARGS"
  (
    set -x
    cd "$dir" && eval "${MAYBE_TIME_PREFIX} ${ARGS}"
  )
  ecode="$?"

  # shellcheck disable=SC2181
  # We are checking the sub-shell, following SC2181 would make this unreadable
  if [[ $ecode -ne 0 ]]; then
    echo "^^^ +++"
    exit $ecode
  fi
}
export -f run_command

if [[ "${CI:-"false"}" == "true" ]]; then
  echo "--- 🚨 Buildkite's timing information is misleading! Only consider the job timing that's printed after 'done'"

  parallel_run run_command {} ::: "${DIRS[@]}"
else
  for dir in "${DIRS[@]}"; do
    run_command "$dir"
  done
fi
