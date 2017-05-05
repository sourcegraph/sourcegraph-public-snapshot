#!/bin/bash
set -e
cd $(dirname "${BASH_SOURCE[0]}")

export IMAGE=us.gcr.io/sourcegraph-dev/xlang-swift
export TAG=${TAG-latest}

set -x

if [ ! -d "swift-langserver" ]; then
    git clone git@github.com:sourcegraph/swift-langserver.git swift-langserver
else
    cd swift-langserver && git pull && cd ..
fi

docker build -t $IMAGE:$TAG .
gcloud docker -- push $IMAGE:$TAG
