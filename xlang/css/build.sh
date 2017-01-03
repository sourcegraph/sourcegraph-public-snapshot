#!/bin/bash
set -e
cd $(dirname "${BASH_SOURCE[0]}")

export IMAGE=us.gcr.io/sourcegraph-dev/xlang-css
export TAG=${TAG-latest}

set -x
if [ ! -d "css-langserver" ]; then
    git clone https://github.com/sourcegraph/css-langserver css-langserver && cd css-langserver/langserver
else
    cd css-langserver && git pull && cd langserver
fi

yarn
./node_modules/.bin/tsc -p .

cd ../..
docker build -t $IMAGE:$TAG .
gcloud docker -- push $IMAGE:$TAG

# Also push latest if on CI
[ -z "$CI" ] || (docker tag $IMAGE:$TAG $IMAGE:latest && gcloud docker -- push $IMAGE:latest)
