#!/usr/bin/env bash

set -euo pipefail

cd /var/lib/postgresql

# This pg_ctl function allows us to start Postgres listening
# only on a UNIX socket. This is needed for intermediary upgrade operations
# to run without interference from external clients via TCP.
function pg_ctl() {
    pg_ctl -w -l "/var/lib/postgresql/pg_ctl.log" \
        -D "$PG_DATA" \
        -U "postgres" \
        -o "-p 5432 -c listen_addresses='' -c unix_socket_permissions=0700 -c unix_socket_directories='/var/run/postgresql'" \
        "$1"
}

pg_ctl start

# Apply post pg_upgrade fixes and optimizations.
if [ -e reindex_hash.sql ]; then
    echo "[INFO] Re-indexing hash based indexes"
    psql -U "postgres" -d postgres -f reindex_hash.sql
fi

# Rebuild optimizer statistics
if [ -e ./analyze_new_cluster.sh ]; then
    echo "[INFO] Re-building optimizer statistics"
    ./analyze_new_cluster.sh
fi

pg_ctl stop
