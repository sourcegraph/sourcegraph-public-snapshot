#!/usr/bin/env bash

set -ex
cd "$(dirname "${BASH_SOURCE[0]}")"

# Retag the 2021-04-22 MinIO release
MINIO_RELEASE="RELEASE.2021-04-22T15-44-28Z"
docker pull minio/minio:$MINIO_RELEASE
docker tag minio/minio:$MINIO_RELEASE "$IMAGE"
