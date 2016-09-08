#!/bin/bash
set -ex

GOOS=linux GOARCH=amd64 go build -o langprocessor-ctags .

docker build -t us.gcr.io/sourcegraph-dev/langprocessor-ctags .
gcloud docker push us.gcr.io/sourcegraph-dev/langprocessor-ctags
