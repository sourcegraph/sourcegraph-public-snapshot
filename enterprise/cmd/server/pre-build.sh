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

build_management_console() {
    echo "--- pre- build management-console"
    ./cmd/management-console/pre-build.sh
}
export -f build_management_console

# We run the the management-console's pre-build script in parallel because it invokes expensive
# yarn/node commands
echo "--- (enterprise) pre-build frontend and management console in parallel"
parallel_run {} ::: build_frontend_enterprise build_management_console
