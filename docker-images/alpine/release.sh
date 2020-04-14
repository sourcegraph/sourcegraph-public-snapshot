#!/bin/bash
cd $(dirname "${BASH_SOURCE[0]}")
set -ex

docker tag sourcegraph/alpine:$VERSION sourcegraph/alpine:latest
docker push sourcegraph/alpine:$VERSION
docker push sourcegraph/alpine:latest
