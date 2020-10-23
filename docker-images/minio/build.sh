#!/usr/bin/env bash

set -ex
cd "$(dirname "${BASH_SOURCE[0]}")"

# Retag the 2020-10-18 MinIO release
docker pull minio/minio:RELEASE.2020-10-18T21-54-12Z
docker tag minio/minio:RELEASE.2020-10-18T21-54-12Z "$IMAGE"
