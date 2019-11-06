#!/usr/bin/env bash

cd $(dirname "${BASH_SOURCE[0]}")/../../..
set -euxo pipefail

parallel_run() {
    log_file=$(mktemp)
    trap "rm -rf $log_file" EXIT

    parallel --keep-order --line-buffer --tag --joblog $log_file "$@"
    cat $log_file
}

# We run the the management-console's pre-build script in parallel because it invokes expensive
# yarn/node commands
parallel_run {} ::: ./enterprise/cmd/frontend/pre-build.sh ./cmd/management-console/pre-build.sh
