#!/usr/bin/env bash

set -exuo pipefail
cd $(dirname "${BASH_SOURCE[0]}")/../../..

parallel_run() {
    log_file=$(mktemp)
    trap "rm -rf $log_file" EXIT

    parallel --jobs 4 --keep-order --line-buffer --tag --joblog $log_file "$@"
    cat $log_file
}

echo "--- yarn root"
# mutex is necessary since frontend and the management-console can
# run concurrent "yarn" installs
# TODO: This is no longer needed since the management console was removed.
yarn --mutex network --frozen-lockfile --network-timeout 60000

build_browser() {
    echo "--- yarn browser"
    (cd browser && TARGETS=phabricator yarn build)
}

build_web() {
    echo "--- yarn web"
    (cd web && NODE_ENV=production yarn -s run build --color)
}

export -f build_browser
export -f build_web

echo "--- (enterprise) build browser and web concurrently"
parallel_run ::: build_browser build_web

echo "--- (enterprise) generate"
./enterprise/dev/generate.sh
