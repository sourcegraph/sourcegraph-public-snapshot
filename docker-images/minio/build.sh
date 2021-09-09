#!/usr/bin/env bash

set -ex
cd "$(dirname "${BASH_SOURCE[0]}")"

# Retag the 2021-08-31 MinIO release
MINIO_RELEASE="RELEASE.2021-08-31T05-46-54Z"
docker pull minio/minio:$MINIO_RELEASE
docker tag minio/minio:$MINIO_RELEASE "$IMAGE"
