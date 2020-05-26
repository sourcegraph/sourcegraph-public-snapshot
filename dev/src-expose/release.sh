#!/bin/bash

set -e

mkdir -p artifacts && cd artifacts
mkdir -p "src-expose/$(git rev-parse HEAD)"
cd "src-expose/$(git rev-parse HEAD)"
echo "src-expose built from https://github.com/sourcegraph/sourcegraph" >info.txt
echo >>info.txt
git log -n1 >>info.txt
mkdir linux-amd64 darwin-amd64
CGO_ENABLED=0 GOOS=linux go build -trimpath -o linux-amd64/src-expose github.com/sourcegraph/sourcegraph/dev/src-expose
CGO_ENABLED=0 GOOS=darwin go build -trimpath -o darwin-amd64/src-expose github.com/sourcegraph/sourcegraph/dev/src-expose
cd -
rm -rf src-expose/latest
cp -r "src-expose/$(git rev-parse HEAD)" src-expose/latest
gsutil cp -r src-expose gs://sourcegraph-artifacts
gsutil iam ch allUsers:objectViewer gs://sourcegraph-artifacts
