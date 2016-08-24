#!/bin/bash
set -ex

GOOS=linux GOARCH=amd64 go build -o langproxy-js .

docker build -t us.gcr.io/sourcegraph-dev/langproxy-js .
gcloud docker push us.gcr.io/sourcegraph-dev/langproxy-js
