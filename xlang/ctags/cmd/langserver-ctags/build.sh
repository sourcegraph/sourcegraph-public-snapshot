#!/bin/bash
set -e
cd $(dirname "${BASH_SOURCE[0]}")

export IMAGE=us.gcr.io/sourcegraph-dev/xlang-ctags
export TAG=${TAG-latest}

set -x

CGO_ENABLED=0 GOOS=linux GOARCH=386 go build -o langserver-ctags .

docker build -t $IMAGE:$TAG .
gcloud docker -- push $IMAGE:$TAG
