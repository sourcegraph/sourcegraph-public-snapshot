#!/bin/bash

set -e
unset CDPATH
cd "$(dirname "${BASH_SOURCE[0]}")/.." # cd to repo root dir

parallel_run() {
    log_file=$(mktemp)
    trap "rm -rf $log_file" EXIT

    parallel --jobs 4 --keep-order --line-buffer --joblog $log_file "$@"
    echo "--- done - displaying job log:"
    cat $log_file
}

echo "--- 🚨 Buildkite's timing information is misleading! Only consider the job timing that's printed after 'done'"

export ARGS="$@"

DIRS=(
   web \
   shared \
   browser \
   packages \
   sourcegraph-extension-api \
   packages/@sourcegraph/extension-api-types \
   lsif dev/release
)

run_command() {
    dir=$1
    echo "--- $dir: ${dir}"
    (set -x; cd "$dir" && $ARGS)
}
export -f run_command

parallel_run run_command {} ::: "${DIRS[@]}"
