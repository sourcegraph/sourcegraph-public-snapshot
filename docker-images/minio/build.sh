#!/usr/bin/env bash

set -ex
cd "$(dirname "${BASH_SOURCE[0]}")"

# Retag the 2021-11-09 MinIO release
MINIO_RELEASE="RELEASE.2021-11-09T03-21-45Z"
docker pull minio/minio:$MINIO_RELEASE
docker tag minio/minio:$MINIO_RELEASE "$IMAGE"
