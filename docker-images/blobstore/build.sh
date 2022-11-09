#!/usr/bin/env bash

set -ex
cd "$(dirname "${BASH_SOURCE[0]}")"

# Retag the release
S3PROXY_RELEASE="sha-ba0fd6d"
docker pull andrewgaul/s3proxy:$S3PROXY_RELEASE
docker rmi "$IMAGE" || true
docker tag andrewgaul/s3proxy:$S3PROXY_RELEASE "$IMAGE"
