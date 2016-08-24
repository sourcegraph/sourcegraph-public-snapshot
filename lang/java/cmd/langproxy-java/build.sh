#!/bin/bash
set -ex

GOOS=linux GOARCH=amd64 go build -o langproxy-java .

docker build -t us.gcr.io/sourcegraph-dev/langproxy-java .
gcloud docker push us.gcr.io/sourcegraph-dev/langproxy-java
