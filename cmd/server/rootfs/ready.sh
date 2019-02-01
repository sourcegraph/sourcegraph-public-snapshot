#!/usr/bin/env bash

set -euo pipefail

# This script checks to see if postgres is alive. It is expected to be used by
# a Kubernetes ready probe.

# We check if the TCP port is available since that is how clients will
# connect. While upgrading only the unix port will be available, so we
# specifically want to avoid reporting ready in that case.

if [ -n "$POSTGRES_PASSWORD" ]; then
    export PGPASSWORD="$POSTGRES_PASSWORD"
fi

export PGHOST="$(hostname)"
export PGUSER="$PGUSERNEW"
export PGDATABASE="$PGDATABASENEW"

# Check if we can run a simple query. If it fails the reason will be printed
# to stderr and we will have a non-zero exit code.
$PGBINNEW/psql --no-password --tuples-only --no-align -c 'select 1;' > /dev/null
