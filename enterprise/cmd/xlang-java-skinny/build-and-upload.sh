#!/bin/bash

set -ex
cd $(dirname "${BASH_SOURCE[0]}")

export IMAGE=${IMAGE-us.gcr.io/sourcegraph-dev/xlang-java-skinny}
export VERSION=$(date '+%Y-%m-%d-%H%M')

./build.sh

gcloud auth configure-docker
docker push "$IMAGE:$VERSION"

if [ -z "$NOTLATEST" ]; then
    # push latest version, too
    docker tag "$IMAGE:$VERSION" "${IMAGE}:latest"
    docker push "${IMAGE}:latest"
fi
