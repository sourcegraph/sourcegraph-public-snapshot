#!/bin/bash

# Runs a Sourcegraph server locally for development. This should still
# be run via `make serve-dev` in the parent directory so that the
# credential env vars get set.

set -euf -o pipefail

go get sourcegraph.com/sqs/rego

# This ulimit check is for the large number of open files from rego; we need
# this here even though the `src` sysreq package also checks for ulimit (for
# the app itself).
[ `ulimit -n` -ge 10000 ] || (echo "Error: Please increase the open file limit by running\n\n  ulimit -n 10000\n" 1>&2; exit 1)

export WEBPACK_DEV_SERVER_URL=http://localhost:8080
export WEBPACK_DEV_SERVER_ADDR=127.0.0.1:8080
[ -n "${WEBPACK_DEV_SERVER_URL-}" ] && [ "${WEBPACK_DEV_SERVER_URL-}" != " " ] && (curl -Ss -o /dev/null "$WEBPACK_DEV_SERVER_URL" || (cd ui && npm start &))

export DEBUG=t

exec rego -installenv=GOGC=off,GODEBUG=sbrk=1 -tags="${GOTAGS-}" sourcegraph.com/sourcegraph/sourcegraph/cmd/src ${SRCFLAGS-} serve --reload --app.disable-support-services ${SERVEFLAGS-}
