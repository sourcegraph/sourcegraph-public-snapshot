#!/usr/bin/env bash

cd "$(dirname "${BASH_SOURCE[0]}")"
set -ex

# Retag the upstream dind (docker-in-docker) release
VERSION="23.0.1-dind@sha256:ed6220b0de0f309f0844cf8cf1a6b861e981fb7f5c28bec6acc97abc910bd0a8"

docker pull docker:$VERSION
docker tag docker:$VERSION "$IMAGE"
