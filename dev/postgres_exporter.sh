#!/usr/bin/env bash

# Description: Prometheus collects metrics and aggregates them into graphs.
#

set -euf -o pipefail

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"

PG_EXPORTER_QUERIES=${DIR}/pg_exporter/queries.yaml

if [ ! -e "${PG_EXPORTER_QUERIES}" ]; then
  echo "Could not find postgres exporter config, expected /dev/pg_exporter"
  exit 1
fi


IMAGE=wrouesnel/postgres_exporter:v0.7.0@sha256:785c919627c06f540d515aac88b7966f352403f73e931e70dc2cbf783146a98b
CONTAINER=postgres_exporter

# Use psql to read the effective values for PG* env vars (instead of, e.g., hardcoding the default
# values).
get_pg_env() { psql -c '\set' | grep "$1" | cut -f 2 -d "'"; }
PGHOST=${PGHOST-$(get_pg_env HOST)}
PGUSER=${PGUSER-$(get_pg_env USER)}
PGPORT=${PGPORT-$(get_pg_env PORT)}
# we need to be able to query schema_migrations table
PGDATABASE=${PGDATABASE-$(get_pg_env DATABASE)}

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


exec docker run --rm -p9187:9187 ${NET_ARG} --name="$CONTAINER"  \
     -e DATA_SOURCE_NAME="${DATA_SOURCE_NAME}" \
     --mount type=bind,src="${PG_EXPORTER_QUERIES}",target=/queries.yaml \
     -e PG_EXPORTER_EXTEND_QUERY_PATH=/queries.yaml ${IMAGE}
