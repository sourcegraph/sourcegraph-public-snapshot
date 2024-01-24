#!/usr/bin/env bash

set -e

RUNNER="$1"
MIGRATOR_TARBALL="$2"
FRONTEND_TARBALL="$3"

# TODO why do we need this ? Is it hardcoded?
FILES="$(dirname "$4")"

# Loads the tarball for the migrator in docker, will be migrator:candidate
"$MIGRATOR_TARBALL"
# Loads the tarball for the migrator in docker, will be frontend:candidate
"$FRONTEND_TARBALL"

"$RUNNER" standard # "$FILES"

exit 1
