#!/bin/bash
set -ex

# This command builds a release build of xlang-ctags and uploads it to the
# gcloud docker registry

CGO_ENABLED=0 GOOS=linux GOARCH=386 go build -o langserver-ctags .

docker build -t us.gcr.io/sourcegraph-dev/xlang-ctags .
gcloud docker -- push us.gcr.io/sourcegraph-dev/xlang-ctags
