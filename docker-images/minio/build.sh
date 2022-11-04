#!/usr/bin/env bash

set -ex
cd "$(dirname "${BASH_SOURCE[0]}")"

# Retag the 2022-05-08 MinIO release
MINIO_RELEASE="RELEASE.2022-09-17T00-09-45Z"
docker pull minio/minio:$MINIO_RELEASE
docker rmi "$IMAGE" || true
docker tag minio/minio:$MINIO_RELEASE "$IMAGE"
