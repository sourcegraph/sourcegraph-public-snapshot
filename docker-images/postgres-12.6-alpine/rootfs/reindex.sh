#!/usr/bin/env bash

cd "$(dirname "${BASH_SOURCE[0]}")"
set -Eexo pipefail

# shellcheck source=./env.sh
source /env.sh

file_env() {
  local var="$1"
  local fileVar="${var}_FILE"
  local def="${2:-}"
  if [ "${!var:-}" ] && [ "${!fileVar:-}" ]; then
    echo >&2 "error: both $var and $fileVar are set (but are exclusive)"
    exit 1
  fi
  local val="$def"
  if [ "${!var:-}" ]; then
    val="${!var}"
  elif [ "${!fileVar:-}" ]; then
    val="$(<"${!fileVar}")"
  fi
  export "$var"="$val"
  unset "$fileVar"
}

prepare_env() {
  file_env 'POSTGRES_USER' 'postgres'
  file_env 'POSTGRES_PASSWORD'
  export PGPASSWORD="${PGPASSWORD:-$POSTGRES_PASSWORD}"
  export PGUSER="${PGUSER:-$POSTGRES_USER}"
}

unset_env() {
  unset PGPASSWORD
}

postgres_start() {
  # internal start of server in order to allow set-up using psql-client
  # - does not listen on external TCP/IP and waits until start finishes
  # - "-P" prevents Postgres from using indexes for system catalog lookups -
  #        see https://www.postgresql.org/docs/12/sql-reindex.html

  pg_ctl -D "$PGDATA" \
    -o "-c listen_addresses=''" \
    -o "-P" \
    -w start
}

postgres_stop() {
  pg_ctl -D "$PGDATA" -m fast -w stop
}

cleanup() {
  postgres_stop
  unset_env
}

reindex() {
  reindexdb --no-password --verbose --echo "$@"
}

# allow the container to be started with `--user`
if [ "$(id -u)" = '0' ]; then
  # TODO@davejrt is this fix what you meant?
  su-exec postgres "${BASH_SOURCE[0]}" "$@"
fi

# look specifically for REINDEX_COMPLETED_FILE, as it is expected in the DB dir
if [ ! -s "${REINDEX_COMPLETED_FILE}" ]; then
  prepare_env
  postgres_start
  trap cleanup EXIT

  echo
  echo 'PostgresSQL must rebuild its indexes. This process can take up to a few hours on systems with a large dataset.'
  echo

  reindex --system
  reindex --all

  # mark reindexing process as done
  echo "Re-indexing for 3.31 release completed successfully at $(date)" >"${REINDEX_COMPLETED_FILE}"

  echo
  echo 'PostgreSQL reindexing process complete - ready for start up.'
  echo
fi
