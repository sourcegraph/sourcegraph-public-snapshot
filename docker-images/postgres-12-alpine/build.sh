#!/usr/bin/env bash

set -ex
cd "$(dirname "${BASH_SOURCE[0]}")"

POSTGRES_UID=${POSTGRES_UID:-999}
PING_UID=${PING_UID:-99}

docker build -t "${IMAGE:-index.docker.io/sourcegraph/postgres-12-alpine}" --build-arg POSTGRES_UID="$POSTGRES_UID" --build-arg PING_UID="$PING_UID" .
