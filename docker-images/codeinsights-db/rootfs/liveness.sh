#!/usr/bin/env bash

set -euxo pipefail
# shellcheck source=./env.sh
source /env.sh

# This script checks to see if postgres is alive. It uses the ready check, but
# additionally ignores upgrades to give the container enough time to
# re-compute indexes. It is expected to be used by a Kubernetes liveness probe.

# Ensure we are in the same dir ready.sh
cd "$(dirname "${BASH_SOURCE[0]}")"

if [ -s "${PGDATA}/PG_VERSION" ] && [ ! -s "${REINDEX_COMPLETED_FILE}" ]; then
  echo "[INFO] Postgres is re-indexing..."
  exit 0
fi

./ready.sh
