#!/bin/bash
set -ex

GOOS=linux GOARCH=amd64 go build -o langserver-ctags .

docker build -t us.gcr.io/sourcegraph-dev/langserver-ctags .
gcloud docker push us.gcr.io/sourcegraph-dev/langserver-ctags
