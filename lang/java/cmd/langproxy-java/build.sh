#!/bin/bash
set -ex

export GOOS=linux GOARCH=amd64 CGO_ENABLED=0

go build -o langproxy-java .

docker build -t us.gcr.io/sourcegraph-dev/langproxy-java .
gcloud docker push us.gcr.io/sourcegraph-dev/langproxy-java
