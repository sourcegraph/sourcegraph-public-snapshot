#!/usr/bin/env bash

set -ex
cd "$(dirname "${BASH_SOURCE[0]}")"

docker tag sourcegraph/alpine-3.12:"$VERSION" sourcegraph/alpine-3.12:latest
docker push sourcegraph/alpine-3.12:"$VERSION"
docker push sourcegraph/alpine-3.12:latest
