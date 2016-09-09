#!/bin/bash
set -ex

export GOOS=linux GOARCH=amd64 CGO_ENABLED=0

go build -o langprocessor-go .

docker build -t us.gcr.io/sourcegraph-dev/langprocessor-go .
gcloud docker push us.gcr.io/sourcegraph-dev/langprocessor-go
