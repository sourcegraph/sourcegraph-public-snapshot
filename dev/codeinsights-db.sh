#!/usr/bin/env bash

# Description: Code Insights Postgres DB.

set -euf -o pipefail
pushd "$(dirname "${BASH_SOURCE[0]}")/.." >/dev/null

if [ -n "${DISABLE_CODE_INSIGHTS:-}" ]; then
  echo Not starting Code Insights DB since DISABLE_CODE_INSIGHTS is set.
  exit 0
fi

DISK="${HOME}/.sourcegraph-dev/data/codeinsights-db"
if [ ! -e "${DISK}" ]; then
  mkdir -p "${DISK}"
fi

IMAGE=sourcegraph/codeinsights-db:dev
CONTAINER=codeinsights-db
PORT=5435

docker inspect $CONTAINER >/dev/null 2>&1 && docker rm -f $CONTAINER

# Log file location: since we log outside of the Docker container, we should
# log somewhere that's _not_ ~/.sourcegraph-dev/data/codeinsights-db, since that gets
# volume mounted into the container and therefore has its own ownership
# semantics.
LOGS="${HOME}/.sourcegraph-dev/logs/codeinsights-db"
mkdir -p "${LOGS}"

LOG_FILE="${LOGS}/codeinsights-db.log"

# Quickly build image
echo "codeinsights-db: building ${IMAGE}..."
IMAGE=${IMAGE} CACHE=true ./docker-images/codeinsights-db/build.sh >"${LOG_FILE}" 2>&1 ||
  (BUILD_EXIT_CODE=$? && echo "codeinsights-db build failed; dumping log:" && cat "${LOG_FILE}" && exit $BUILD_EXIT_CODE)

function finish() {
  EXIT_CODE=$?

  # Exit code 2 indicates a normal Ctrl-C termination via goreman, so we'll
  # only dump the log if it's not 0 _or_ 2.
  if [ $EXIT_CODE -ne 0 ] && [ $EXIT_CODE -ne 2 ]; then
    echo "codeinsights-db exited with unexpected code ${EXIT_CODE}; dumping log:"
    cat "${LOG_FILE}"
  fi

  # Ensure that we still return the same code so that goreman can do sensible
  # things once this script exits.
  return $EXIT_CODE
}

echo "codeinsights-db: serving on http://localhost:${PORT}"
echo "codeinsights-db: note that logs are piped to ${LOG_FILE}"
docker run --rm \
  --name=${CONTAINER} \
  --cpus=1 \
  --memory=1g \
  -e POSTGRES_DB=postgres \
  -e POSTGRES_PASSWORD=password \
  -e POSTGRES_USER=postgres \
  -p 0.0.0.0:${PORT}:5432 \
  -v "${DISK}":/var/lib/postgresql/data \
  ${IMAGE} >"${LOG_FILE}" 2>&1 || finish
