#!/usr/bin/env bash

set -exuo pipefail
cd $(dirname "${BASH_SOURCE[0]}")/../../..

parallel_run() {
    log_file=$(mktemp)
    trap "rm -rf $log_file" EXIT

    parallel --keep-order --line-buffer --tag --joblog $log_file "$@"
    cat $log_file
}

echo "--- yarn root"
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

parallel_run ::: build_browser build_web

# Start postgres (for the dev/generate.sh scripts)
gosu postgres /usr/lib/postgresql/9.6/bin/pg_ctl initdb
## Allow pgsql to listen to all IPs
## See https://stackoverflow.com/a/52381997 for more information
# Try changing lock location to /tmp since Kaniko doesn't allow writes to '/var/run'
# See https://github.com/GoogleContainerTools/kaniko/issues/506 and https://forums.postgresql.fr/viewtopic.php?id=3984
gosu postgres /usr/lib/postgresql/9.6/bin/pg_ctl -o "-c listen_addresses='*' -c unix_socket_directories='/tmp'"  -w start
export PGHOST='/tmp'

echo "--- generate"
./enterprise/dev/generate.sh
