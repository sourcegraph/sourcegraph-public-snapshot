#!/usr/bin/env bash

set -euo pipefail

cd /var/lib/postgresql

# Allow the container to be started with root in Kubernetes and change permissions
# of the parent volume directory to be owned entirely by the postgres user.
if [ "$(id -u)" = '0' ]; then
  mkdir -p "$PGDATA"
  chown -R postgres "$(dirname "$PGDATA")"
  chmod 700 "$(dirname "$PGDATA")" "$PGDATA"
  exec gosu postgres "${BASH_SOURCE[0]}" "$@"
fi

if [ ! -s "$PGDATA/PG_VERSION" ]; then
  echo "[INFO] Initializing Postgres database '$POSTGRES_DB' from scratch in $PGDATA"
  /initdb.sh
fi

/conf.sh

exec postgres
