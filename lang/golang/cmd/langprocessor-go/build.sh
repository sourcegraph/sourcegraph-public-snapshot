#!/bin/bash
set -ex

GOOS=linux GOARCH=amd64 go build -o langprocessor-go .

docker build -t us.gcr.io/sourcegraph-dev/langprocessor-go .
gcloud docker push us.gcr.io/sourcegraph-dev/langprocessor-go
