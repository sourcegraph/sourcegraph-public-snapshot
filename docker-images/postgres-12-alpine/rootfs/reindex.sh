#!/usr/bin/env bash

cd "$(dirname "${BASH_SOURCE[0]}")"
set -Eexo pipefail

# If REINDEX_COMPLETED_FILE is not set, load it from env.sh
if [[ -z $REINDEX_COMPLETED_FILE ]]; then
  # shellcheck source=./env.sh
  source /env.sh
fi
# Check REINDEX_COMPLETED_FILE has been set
if [[ -z $REINDEX_COMPLETED_FILE ]]; then
  echo "Envar REINDEX_COMPLETED_FILE is undefined"
  exit 1
fi

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
  file_env 'POSTGRES_DB'
  export PGUSER="${PGUSER:-$POSTGRES_USER}"
  export PGPASSWORD="${PGPASSWORD:-$POSTGRES_PASSWORD}"
  export PGDATABASE="${PGDATABASE:-$POSTGRES_DB}"
}

unset_env() {
  unset PGPASSWORD
}

postgres_is_running() {
  local pid="$1"

  local proc_entry="/proc/$pid/comm"
  grep -q '^postgres$' "$proc_entry"
}

postgres_stop_cleanly() {
  pg_ctl -D "$PGDATA" -m fast -w stop
}

postgres_stop() {
  # This logic handles the case where we've restored a snapshot
  # that was taken from a still running or improperly shutdown
  # postgres instance. We'll need to check to see if postgres is
  # actually still running under the pid specified in the postmaster.pid
  # file. If it is, we shut it down properly. If it isn't, we
  # delete the pid file so that we can start up properly.
  local postmaster_file="$PGDATA/postmaster.pid"

  if ! [[ -s "$postmaster_file" ]]; then
    # postgres isn't running - nothing to do
    return 0
  fi

  local pid
  pid="$(head -1 "$postmaster_file")"

  if postgres_is_running "$pid"; then
    # postgres is currently running in the container - shut it down cleanly
    postgres_stop_cleanly
    return 0
  fi

  # we have a postmaster file, but a postgres process isn't running anymore.
  # remove the postmaster file - we can't do any better here
  rm "$postmaster_file" || true
}

postgres_start() {
  # internal start of server in order to allow set-up using psql-client
  # - does not listen on external TCP/IP and waits until start finishes
  # - "-P" prevents Postgres from using indexes for system catalog lookups -
  #        see https://www.postgresql.org/docs/12/sql-reindex.html

  pg_ctl -D "$PGDATA" \
    -o "-c listen_addresses=''" \
    -o "-P" \
    -w restart
}

cleanup() {
  postgres_stop_cleanly
  unset_env
}

reindex() {
  reindexdb --no-password --verbose --echo "$@"
}

# allow the container to be started with `--user`
if [ "$(id -u)" = '0' ]; then
  su-exec postgres "${BASH_SOURCE[0]}" "$@"
  # Exit original process running as root
  exit
fi

# look specifically for REINDEX_COMPLETED_FILE, as it is expected in the DB dir
if [ ! -s "${REINDEX_COMPLETED_FILE}" ]; then
  prepare_env
  postgres_stop
  postgres_start
  trap cleanup EXIT

  echo
  echo 'PostgresSQL must rebuild its indexes. This process can take up to a few hours on systems with a large dataset.'
  echo

  reindex --system
  reindex --all

  # mark reindexing process as done
  echo "Writing to '$REINDEX_COMPLETED_FILE"
  echo "Re-indexing completed successfully at $(date)" >"${REINDEX_COMPLETED_FILE}"

  echo
  echo 'PostgreSQL reindexing process complete - ready for start up.'
  echo
fi
