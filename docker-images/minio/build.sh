#!/usr/bin/env bash

set -ex
cd "$(dirname "${BASH_SOURCE[0]}")"

# Retag the 2022-05-08 MinIO release
MINIO_RELEASE="RELEASE.2022-05-08T23-50-31Z"
docker pull minio/minio:$MINIO_RELEASE
docker tag minio/minio:$MINIO_RELEASE "$IMAGE"
