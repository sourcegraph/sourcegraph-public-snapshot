#!/bin/bash
set -ex

export GOOS=linux GOARCH=amd64 CGO_ENABLED=0

go build -o langproxy-js .

docker build -t us.gcr.io/sourcegraph-dev/langproxy-js .
gcloud docker push us.gcr.io/sourcegraph-dev/langproxy-js
