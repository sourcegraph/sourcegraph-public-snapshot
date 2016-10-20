#!/bin/bash
set -ex

# This command builds a release build of langserver-ctags and uploads it to the
# gcloud docker registry

CGO_ENABLED=0 GOOS=linux GOARCH=386 go build -o langserver-ctags .

docker build -t us.gcr.io/sourcegraph-dev/langserver-ctags .
gcloud docker -- push us.gcr.io/sourcegraph-dev/langserver-ctags
