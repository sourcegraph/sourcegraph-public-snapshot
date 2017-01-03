#!/bin/bash
set -e
cd $(dirname "${BASH_SOURCE[0]}")

export IMAGE=us.gcr.io/sourcegraph-dev/xlang-javascript-typescript
export TAG=${TAG-latest}

set -x
type yarn > /dev/null 2>&1 || npm install -g yarn

cd ./buildserver
yarn
./node_modules/.bin/tsc -p .

cd ..
docker build -t $IMAGE:$TAG .
gcloud docker -- push $IMAGE:$TAG

# Also push latest if on CI
[ -z "$CI" ] || (docker tag $IMAGE:$TAG $IMAGE:latest && gcloud docker -- push $IMAGE:latest)
