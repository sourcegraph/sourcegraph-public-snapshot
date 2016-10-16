#!/bin/bash

# Runs a Sourcegraph server locally for development. This should still
# be run via `make serve-dev` in the parent directory so that the
# credential env vars get set.

set -euf -o pipefail

cd "$(dirname "${BASH_SOURCE[0]}")/.." # cd to repo root dir

GOBIN="$PWD"/vendor/bin go get sourcegraph.com/sourcegraph/sourcegraph/vendor/sourcegraph.com/sqs/rego

export WEBPACK_DEV_SERVER_URL=http://localhost:8080
export WEBPACK_DEV_SERVER_ADDR=127.0.0.1:8080
[ -n "${WEBPACK_DEV_SERVER_URL-}" ] && [ "${WEBPACK_DEV_SERVER_URL-}" != " " ] && (curl -Ss -o /dev/null "$WEBPACK_DEV_SERVER_URL" || (cd ui && npm start &))

export DEBUG=t

type ulimit > /dev/null && ulimit -n 10000
exec "$PWD"/vendor/bin/rego -installenv=GOGC=off,GODEBUG=sbrk=1 -tags="${GOTAGS-}" sourcegraph.com/sourcegraph/sourcegraph/cmd/src ${SRCFLAGS-} serve --reload --app.disable-support-services ${SERVEFLAGS-}
