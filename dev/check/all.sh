#!/usr/bin/env bash

set -ex
cd $(dirname "${BASH_SOURCE[0]}")

parallel_run() {
    log_file=$(mktemp)
    trap "rm -rf $log_file" EXIT

    parallel --jobs 4 --keep-order --line-buffer --tag --joblog $log_file "$@"
    cat $log_file
}

go version
go env

CHECKS=(
    ./yarn-deduplicate.sh \
    ./docsite.sh \
    ./gofmt.sh \
    ./template-inlines.sh \
    ./go-enterprise-import.sh \
    ./go-dbconn-import.sh \
    ./mgmt-console-conf-import.sh \
    ./go-generate.sh \
    ./go-lint.sh \
    ./todo-security.sh \
    ./no-localhost-guard.sh \
    ./bash-syntax.sh \
    ./check-owners.sh
)

parallel_run {} ::: "${CHECKS[@]}"

# TODO(sqs): Reenable this check when about.sourcegraph.com is reliable. Most failures come from its
# downtime, not from broken URLs.
#
# ./broken-urls.bash

echo "--- done"
