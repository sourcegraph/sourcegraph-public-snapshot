#!/bin/bash
cd $(dirname "${BASH_SOURCE[0]}")
set -ex

rm -rf observability
cp -R ../../observability .

docker build -t ${IMAGE:-sourcegraph/prometheus} . \
    --progress=plain \
    --build-arg COMMIT_SHA \
    --build-arg DATE \
    --build-arg VERSION
