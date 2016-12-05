#!/bin/bash
set -e
cd $(dirname "${BASH_SOURCE[0]}")

export IMAGE=us.gcr.io/sourcegraph-dev/lsp-proxy
export TAG=${TAG-latest}
export GOBIN="$PWD/../../../vendor/.bin"
export PATH="$GOBIN:$PATH"

set -x
go install sourcegraph.com/sourcegraph/sourcegraph/vendor/github.com/neelance/godockerize
godockerize build -t $IMAGE:$TAG .
gcloud docker -- push $IMAGE:$TAG

[ -z "$CI" ] || (docker tag $IMAGE:$TAG $IMAGE:latest && gcloud docker -- push $IMAGE:latest)
