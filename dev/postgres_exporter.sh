#!/usr/bin/env bash

# Description: Prometheus collects metrics and aggregates them into graphs.
#

set -euf -o pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." >/dev/null 2>&1 && pwd)"

IMAGE=postgres_exporter:dev
CONTAINER=postgres_exporter

# Use psql to read the effective values for PG* env vars (instead of, e.g., hardcoding the default
# values).
get_pg_env() { psql -c '\set' | grep "$1" | cut -f 2 -d "'"; }
PGHOST=${PGHOST-$(get_pg_env HOST)}
PGUSER=${PGUSER-$(get_pg_env USER)}
PGPORT=${PGPORT-$(get_pg_env PORT)}
# we need to be able to query migration_logs table
PGDATABASE=${PGDATABASE-$(get_pg_env DBNAME)}

ADJUSTED_HOST=${PGHOST:-127.0.0.1}
if [[ ("$ADJUSTED_HOST" == "localhost" || "$ADJUSTED_HOST" == "127.0.0.1" || -f "$ADJUSTED_HOST") && "$OSTYPE" != "linux-gnu" ]]; then
  ADJUSTED_HOST="host.docker.internal"
fi

NET_ARG=""
DATA_SOURCE_NAME="postgresql://${PGUSER}:${PGPASSWORD}@${ADJUSTED_HOST}:${PGPORT}/${PGDATABASE}?sslmode=${PGSSLMODE:-disable}"

if [[ "$OSTYPE" == "linux-gnu" ]]; then
  NET_ARG="--net=host"
  DATA_SOURCE_NAME="postgresql://${PGUSER}:${PGPASSWORD}@${ADJUSTED_HOST}:${PGPORT}/${PGDATABASE}?sslmode=${PGSSLMODE:-disable}"
fi

echo "Building pg_exporter docker image"
env IMAGE="${IMAGE}" "${REPO_ROOT}/docker-images/postgres_exporter/build.sh"

exec docker run --rm -p9187:9187 ${NET_ARG} --name="$CONTAINER" \
  -e DATA_SOURCE_NAME="${DATA_SOURCE_NAME}" "${IMAGE}"
