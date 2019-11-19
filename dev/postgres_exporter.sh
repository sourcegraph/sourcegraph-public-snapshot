#!/usr/bin/env bash

# Description: Prometheus collects metrics and aggregates them into graphs.
#

set -euf -o pipefail

IMAGE=wrouesnel/postgres_exporter:v0.7.0@sha256:785c919627c06f540d515aac88b7966f352403f73e931e70dc2cbf783146a98b
CONTAINER=postgres_exporter

NET_ARG=""
DATA_SOURCE_NAME="postgresql://sourcegraph:sourcegraphd@host.docker.internal:5432/postgres?sslmode=disable"

if [[ "$OSTYPE" == "linux-gnu" ]]; then
   NET_ARG="--net=host"
   DATA_SOURCE_NAME="postgresql://sourcegraph:sourcegraphd@127.0.0.1:5432/postgres?sslmode=disable"
fi

exec docker run --rm -p9187:9187 ${NET_ARG} --name=postgres_exporter -e DATA_SOURCE_NAME=${DATA_SOURCE_NAME} ${IMAGE}
