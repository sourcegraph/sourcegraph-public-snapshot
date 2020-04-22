#!/usr/bin/env bash

# Manual publishing of Docker image until automated through CI

set -e

export COMMIT_SHA
COMMIT_SHA=$(git rev-parse --short HEAD)

export VERSION=$COMMIT_SHA

export DATE
DATE=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

export IMAGE=sourcegraph/src-expose

. ./build.sh

docker image push sourcegraph/src-expose:latest
docker image tag sourcegraph/src-expose:latest sourcegraph/src-expose:"$VERSION"
docker image push sourcegraph/src-expose:"$VERSION"
