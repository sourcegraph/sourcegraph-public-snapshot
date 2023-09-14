#!/usr/bin/env bash

# We start up a postgres tuned for performance vs persistence. This is more
# likely to lose data, but this seems like a fine trade-off for tests/devenv.

# Use a different data dir depending on PG version. Avoids the need to migrate
# when we upgrade at the cost of a fresh dev environment.
PGVER=$(pg_ctl -V | awk '{print $NF}')

export PGHOST="${SG_DATA_DIR:-$HOME/.sourcegraph}/postgres"
export PGDATA="${PGHOST}/${PGVER}"
export PGDATABASE=postgres
export PGDATASOURCE="postgresql:///postgres?host=${PGHOST}"
export PGUSER="${USER}"

if [ ! -d "$PGHOST" ]; then
  mkdir -p "$PGHOST"
fi
if [ ! -d "$PGDATA" ]; then
  echo 'Initializing postgresql database...'
  initdb "$PGDATA" --nosync --encoding=UTF8 --no-locale --auth=trust >/dev/null
  cat <<-EOF >>"$PGDATA"/postgresql.conf
	    unix_socket_directories = '$PGHOST'
	    listen_addresses = 'localhost'
	    max_connections = 250
	    shared_buffers = 12MB
	    fsync = off
	    synchronous_commit = off
	    full_page_writes = off
EOF
fi
if ! pg_isready --quiet; then
  echo 'Starting postgresql database...'
  pg_ctl start -l "$PGHOST/log" 3>&-
fi
