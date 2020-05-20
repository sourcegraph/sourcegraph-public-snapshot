#!/usr/bin/env bash

set -ex
cd "$(dirname "${BASH_SOURCE[0]}")"

docker tag sourcegraph/alpine:"$VERSION" sourcegraph/alpine:latest
docker push sourcegraph/alpine:"$VERSION"
docker push sourcegraph/alpine:latest
