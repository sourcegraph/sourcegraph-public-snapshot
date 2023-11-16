#!/usr/bin/env bash

set -e

RUNNER="$1"
MIGRATOR_TARBALL="$2"
FRONTEND_TARBALL="$3"
FILES="$(dirname "$4")"
"$RUNNER" "$($MIGRATOR_TARBALL)" "$($FRONTEND_TARBALL)" "$FILES"

exit 1
