#!/bin/bash
set -e
cd $(dirname "${BASH_SOURCE[0]}")

export IMAGE=us.gcr.io/sourcegraph-dev/xlang-typescript
export TAG=${TAG-latest}

set -x
type yarn > /dev/null 2>&1 || npm install -g yarn

if [ ! -d "langserver-typescript" ]; then
    git clone https://github.com/sourcegraph/langserver-typescript langserver-typescript && cd langserver-typescript
else
    cd langserver-typescript && git pull
fi
yarn install
./node_modules/.bin/tsc -p .

cd ..
docker build -t $IMAGE:$TAG .
gcloud docker -- push $IMAGE:$TAG

