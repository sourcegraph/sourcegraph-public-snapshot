#!/usr/bin/env bash

container="$(docker run -d --name some-postgres -p 5433:5432 postgres:9.6)"
trap "docker rm -f $container" EXIT

while ! env PGPORT=5433 pg_isready; do
    echo "Sleeping 1s to wait for the postgres Docker container to start..."
    sleep 1
done

go generate github.com/sourcegraph/sourcegraph/migrations
env \
  PATH="/usr/local/opt/postgresql@9.6/bin:$PATH" \
  LOG_MIGRATE_TO_STDOUT=true \
  PGPORT=5433 \
  PGDATABASE=postgres \
  PGUSER=postgres \
  GO111MODULE=on \
  go generate github.com/sourcegraph/sourcegraph/cmd/frontend/db
