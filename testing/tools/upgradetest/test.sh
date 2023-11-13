#!/usr/bin/env bash

set -e

RUNNER="$1"
TARBALL="$2"
FILES="$(dirname "$3")"
"$RUNNER" "$($TARBALL)" "$FILES"

echo $FILES

exit 1
