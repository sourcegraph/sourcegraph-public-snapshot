#!/usr/bin/env bash

# Description: An S3-compatible filesystem.
# Most of this script is very similar to /dev/grafana.sh - refer to inline documentation there.

set -euf -o pipefail
pushd "$(dirname "${BASH_SOURCE[0]}")/.." >/dev/null

MINIO_DISK="${HOME}/.sourcegraph-dev/data/minio"
if [ ! -e "${MINIO_DISK}" ]; then
  mkdir -p "${MINIO_DISK}"
fi
IMAGE=sourcegraph/minio
CONTAINER=minio

docker inspect $CONTAINER >/dev/null 2>&1 && docker rm -f $CONTAINER

MINIO_LOGS="${HOME}/.sourcegraph-dev/logs/minio"
mkdir -p "${MINIO_LOGS}"
MINIO_LOG_FILE="${MINIO_LOGS}/minio.log"

# Quickly build image
IMAGE=${IMAGE} CACHE=true ./docker-images/minio/build.sh >"${MINIO_LOG_FILE}" 2>&1 ||
  (BUILD_EXIT_CODE=$? && echo "build failed; dumping log:" && cat "${MINIO_LOG_FILE}" && exit $BUILD_EXIT_CODE)

function finish() {
  MINIO_EXIT_CODE=$?

  # Exit code 2 indicates a normal Ctrl-C termination via goreman, so we'll
  # only dump the log if it's not 0 _or_ 2.
  if [ $MINIO_EXIT_CODE -ne 0 ] && [ $MINIO_EXIT_CODE -ne 2 ]; then
    echo "MinIO exited with unexpected code ${MINIO_EXIT_CODE}; dumping log:"
    cat "${MINIO_LOG_FILE}"
  fi

  # Ensure that we still return the same code so that goreman can do sensible
  # things once this script exits.
  return $MINIO_EXIT_CODE
}

docker run --rm \
  --name=${CONTAINER} \
  --cpus=1 \
  --memory=1g \
  -p 0.0.0.0:9000:9000 \
  -e 'MINIO_ACCESS_KEY=AKIAIOSFODNN7EXAMPLE' \
  -e 'MINIO_SECRET_KEY=wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY' \
  -v "${MINIO_DISK}":/data \
  ${IMAGE} server /data >"${MINIO_LOG_FILE}" 2>&1 || finish
