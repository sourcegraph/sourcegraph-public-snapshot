#!/usr/bin/env bash

cd "$(dirname "${BASH_SOURCE[0]}")/.."

migrate -database postgres://$PGHOST:$PGPORT/$PGDATABASE -path ./migrations "$@"
