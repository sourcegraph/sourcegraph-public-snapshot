#!/bin/bash
set -e
cd $(dirname "${BASH_SOURCE[0]}")

export IMAGE=us.gcr.io/sourcegraph-dev/xlang-javascript-typescript
export TAG=${TAG-latest}

set -x
type yarn > /dev/null 2>&1 || npm install -g yarn

if [ ! -d "javascript-typescript-langserver" ]; then
    git clone https://github.com/sourcegraph/javascript-typescript-langserver javascript-typescript-langserver && cd javascript-typescript-langserver
else
    cd javascript-typescript-langserver && git pull
fi
yarn install
./node_modules/.bin/tsc -p .

cd ..
docker build -t $IMAGE:$TAG .
gcloud docker -- push $IMAGE:$TAG

