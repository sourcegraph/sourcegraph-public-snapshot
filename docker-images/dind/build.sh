#!/usr/bin/env bash

cd "$(dirname "${BASH_SOURCE[0]}")"
set -ex

# Retag the upstream dind (docker-in-docker) release
VERSION="20.10.22-dind@sha256:03f2d563100b9776283de1e18f10a1f0b66d2fdc7918831bf8db1cda767d6b37"

docker pull docker:$VERSION
docker tag docker:$VERSION "$IMAGE"
