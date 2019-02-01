#!/usr/bin/env bash

set -euo pipefail

cd /var/lib/postgresql

# This pg_ctl function allows us to start Postgres listening
# only on a UNIX socket. This is needed for intermediary upgrade operations
# to run without interference from external clients via TCP.
function pg_ctl() {
    local bindir
    local datadir
    local user

    if [ "$1" == "old" ]; then
        bindir="$PGBINOLD"
        datadir="$PGDATAOLD"
        user="$PGUSEROLD"
    else
        bindir="$PGBINNEW"
        datadir="$PGDATANEW"
        user="$PGUSERNEW"
    fi

    "$bindir/pg_ctl" -w -l "/var/lib/postgresql/pg_ctl_$1.log" \
        -D "$datadir" \
        -U "$user" \
        -o "-p 5432 -c listen_addresses='' -c unix_socket_permissions=0700 -c unix_socket_directories='/var/run/postgresql'" \
        "$2"
}

# The oid of the install user is always 10.
function pg_install_user() {
    echo "select rolname from pg_catalog.pg_roles where oid = 10;" \
        | psql -U "$1" -d "$2" -tA
}

function pg_role_exists() {
    printf "select 1 from pg_catalog.pg_roles where rolname = '%s';" "$1" \
        | psql -U "$2" -d "$3" -tA
}

function pg_database_exists() {
    printf "select 1 from pg_catalog.pg_database where datname = '%s';" "$1" \
        | psql -U "$2" -d "$3" -tA
}

# Allow the container to be started with root in Kubernetes and change permissions
# of the parent volume directory to be owned entirely by the postgres user.
if [ "$(id -u)" = '0' ]; then
  mkdir -p "$PGDATAOLD"
  chown -R postgres "$(dirname "$PGDATAOLD")"
  chmod 700 "$(dirname "$PGDATAOLD")" "$PGDATAOLD"
  exec gosu postgres "${BASH_SOURCE[0]}" "$@"
fi

if [ -s "$PGDATAOLD/PG_VERSION" ] && [ ! -s "$PGDATAOLD.upgraded" ] \
    && [ ! -s "$PGDATANEW/PG_VERSION" ]; then
    # We start and stop the old database to ensure that the
    # its shutdown was clean. sourcegraph.com snapshot backups
    # are taken while the database is running, so this is necessary.
    #
    # Additionally, we need to create the new database with the same
    # install user as the old database, so we must perform a query.
    # The install user of the old database and PGUSEROLD may not be the same.
    pg_ctl old start
    pg_user="$(pg_install_user "$PGUSEROLD" "$PGDATABASEOLD")"
    pg_ctl old stop

    echo "[INFO] $PGDATAOLD found with installation user $pg_user"

    echo "[INFO] Initialising new Postgres database to migrate data to $PGDATANEW"
    POSTGRES_USER="$pg_user" POSTGRES_DB=postgres PGDATA="$PGDATANEW" /initdb.sh

    echo "[INFO] Upgrading old data files in $PGDATAOLD to $PGDATANEW"
    "$PGBINNEW/pg_upgrade" \
        --old-bindir "$PGBINOLD" \
        --new-bindir "$PGBINNEW" \
        --old-datadir "$PGDATAOLD" \
        --new-datadir "$PGDATANEW" \
        --username "$pg_user" \
        --jobs "$(nproc --all)" \
        --verbose \
        --retain

    pg_ctl new start

    # Apply post pg_upgrade fixes and optimizations.
    if [ -e reindex_hash.sql ]; then
        echo "[INFO] Re-indexing hash based indexes"
        psql -U "$pg_user" -d postgres -f reindex_hash.sql
    fi

    # Rebuild optimizer statistics
    if [ -e ./analyze_new_cluster.sh ]; then
        echo "[INFO] Re-building optimizer statistics"
        ./analyze_new_cluster.sh
    fi

    # Ensure $PGUSERNEW exists, even if it didn't in the old db.
    if [ "$(pg_role_exists "$PGUSERNEW" "$pg_user" "postgres")" != "1" ]; then
        echo "[INFO] Creating user $PGUSERNEW in new database"
        printf '%s\n%s\n' "$POSTGRES_PASSWORD" "$POSTGRES_PASSWORD" \
            | "$PGBINNEW/createuser" -e -U "$pg_user" -P "$PGUSERNEW"
    fi

    # Ensure $PGDATABASENEW exists, even if it didn't in the old db.
    if [ "$(pg_database_exists "$PGDATABASENEW" "$pg_user" "postgres")" != "1" ]; then
        echo "[INFO] Creating $PGDATABASENEW in new database"
        "$PGBINNEW/createdb" -e -U "$pg_user" -O "$PGUSERNEW" "$PGDATABASENEW"
    fi

    pg_ctl new stop

    echo "[INFO] Creating marker file $PGDATAOLD.upgraded"
    date > "$PGDATAOLD.upgraded"

elif [ -s "$PGDATAOLD/PG_VERSION" ] && [ ! -s "$PGDATAOLD.upgraded" ] \
    && [ -s "$PGDATANEW/PG_VERSION" ]; then
    # An interrupted upgrade. We need an operator to intervene to decide the
    # best path forward. We could automatically delete $PGDATANEW, but that
    # will likely result in repeated migration failures.
    echo "[FATAL] Detected an interrupted upgrade. $PGDATAOLD/PG_VERSION and $PGDATANEW/PG_VERSION exist, but $PGDATAOLD.upgraded does not."
    echo "Either:"
    echo " - Remove $PGDATANEW to restart migration."
    echo " - Roll-back to using $PGBINOLD."
    exit 1

elif [ ! -s "$PGDATANEW/PG_VERSION" ]; then
    # Completely new Postgres installation, no upgrades to do.
    echo "[INFO] Initializing Postgres database '$PGDATABASENEW' from scratch in $PGDATANEW"
    POSTGRES_USER="$PGUSERNEW" POSTGRES_DB="$PGDATABASENEW" PGDATA="$PGDATANEW" /initdb.sh
fi

PGDATA="$PGDATANEW" exec postgres
