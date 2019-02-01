#!/usr/bin/env bash

set -eo pipefail

cd "$(dirname "${BASH_SOURCE[0]}")/.."

hash migrate 2>/dev/null || {
    if [[ $(uname) == "Darwin" ]]; then
        brew install golang-migrate
    else
        echo "You need to install the 'migrate' tool: https://github.com/golang-migrate/migrate/"
        exit 1
    fi
}

migrate -database "postgres://${PGHOST}:${PGPORT}/${PGDATABASE}" -path ./migrations "$@"
