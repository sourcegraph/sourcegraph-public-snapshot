#!/usr/bin/env bash

# Description: Prometheus collects metrics and aggregates them into graphs.
#

set -euf -o pipefail

IMAGE=wrouesnel/postgres_exporter:v0.7.0@sha256:785c919627c06f540d515aac88b7966f352403f73e931e70dc2cbf783146a98b
CONTAINER=postgres_exporter

if [[ ("$PGHOST" == "localhost" || "$PGHOST" == "127.0.0.1") &&  "$OSTYPE" != "linux-gnu" ]]; then
    PGHOST="host.docker.internal"
fi

NET_ARG=""
DATA_SOURCE_NAME="postgresql://${PGUSER:-sourcegraph}:${PGPASSWORD:-sourcegraph}@${PGHOST:-host.docker.internal}:${PGPORT:-5432}/postgres?sslmode=${PGSSLMODE:-disable}"

if [[ "$OSTYPE" == "linux-gnu" ]]; then
   NET_ARG="--net=host"
   DATA_SOURCE_NAME="postgresql://${PGUSER:-sourcegraph}:${PGPASSWORD:-sourcegraph}@${PGHOST:-127.0.0.1}:${PGPORT:-5432}/postgres?sslmode=${PGSSLMODE:-disable}"
fi

exec docker run --rm -p9187:9187 ${NET_ARG} --name=postgres_exporter -e DATA_SOURCE_NAME=${DATA_SOURCE_NAME} ${IMAGE}
