#!/usr/bin/env bash

cd $(dirname "${BASH_SOURCE[0]}")/../../..
set -euxo pipefail

parallel_run() {
    log_file=$(mktemp)
    trap "rm -rf $log_file" EXIT

    parallel --jobs 4 --keep-order --line-buffer --tag --joblog $log_file "$@"
    cat $log_file
}

build_frontend_enterprise() {
    echo "--- (enterprise) pre-build frontend"
   ./enterprise/cmd/frontend/pre-build.sh
}
export -f build_frontend_enterprise

echo "--- (enterprise) pre-build frontend in parallel"
parallel_run {} ::: build_frontend_enterprise
