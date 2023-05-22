#!/usr/bin/env bash

set -ex
cd "$(dirname "${BASH_SOURCE[0]}")"

# This image is identical to our "sourcegraph/postgres-12-alpine" image,
# but runs with a different uid to avoid migration issues
IMAGE="${IMAGE:-sourcegraph/codeinsights-db}" POSTGRES_UID=70 PING_UID=700 ../postgres-12-alpine/build-wolfi.sh
