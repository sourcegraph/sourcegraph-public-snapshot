#!/usr/bin/env bash

set -e

RUNNER="$1"
TARBALL="$2"
"$RUNNER" "$($TARBALL)"

exit 1
