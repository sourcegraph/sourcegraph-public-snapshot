#!/usr/bin/env bash

set -euxo pipefail
cd /

# shellcheck source=./env.sh
source /env.sh

# Allow the container to be started with root in Kubernetes and change permissions
# of the parent volume directory to be owned entirely by the postgres user.
if [ "$(id -u)" = '0' ]; then
  mkdir -p "$PGDATA"
  chown -R postgres:postgres "$(dirname "$PGDATA")"
  chmod 750 "$(dirname "$PGDATA")" "$PGDATA"
  su-exec postgres "${BASH_SOURCE[0]}" "$@"
fi

if [ ! -s "$PGDATA/PG_VERSION" ]; then
  echo "[INFO] Initializing Postgres database '$POSTGRES_DB' from scratch in $PGDATA"
  /initdb.sh
fi

/conf.sh
/patch-conf.sh

if [ ! -s "${REINDEX_COMPLETED_FILE}" ]; then
  echo "[INFO] Re-creating all indexes for database '$POSTGRES_DB'"
  /reindex.sh
fi

exec postgres
