#!/usr/bin/env bash

set -x
# Copies postgresql.conf over from /conf/postgresql.conf if it exists

if [ ! -d "/conf" ] || [ ! -f "/conf/postgresql.conf" ]; then
  exit 0
fi

cp /conf/postgresql.conf "$PGDATA/postgresql.conf"

# allow the container to be started with `--user`
if [ "$(id -u)" = '0' ]; then
  chown postgres:postgres "$PGDATA/postgresql.conf"
fi
