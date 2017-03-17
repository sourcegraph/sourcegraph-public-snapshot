#!/bin/bash
set -e
cd $(dirname "${BASH_SOURCE[0]}")/../..

export IMAGE=us.gcr.io/sourcegraph-dev/sourcegraph
export TAG=${TAG-latest}
export GOBIN="$PWD/vendor/.bin"
export PATH="$GOBIN:$PATH"

set -x

git clean -fdx ui/assets

cd ui
yarn install
yarn run build
cd ..

go generate ./app/assets ./app/templates

go install sourcegraph.com/sourcegraph/sourcegraph/vendor/github.com/neelance/godockerize
godockerize build -t $IMAGE:$TAG --env VERSION=$TAG ./cmd/src
