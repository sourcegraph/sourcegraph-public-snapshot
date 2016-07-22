#!/bin/bash
set -ex

GOOS=linux GOARCH=amd64 go build -o langserver-go .

docker build -t us.gcr.io/sourcegraph-dev/langserver-go .
gcloud docker push us.gcr.io/sourcegraph-dev/langserver-go