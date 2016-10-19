#!/bin/bash
set -ex
cd $(dirname "${BASH_SOURCE[0]}")

if [ ! -d "langserver-typescript" ]; then
    git clone https://github.com/sourcegraph/langserver-typescript langserver-typescript && cd langserver-typescript
else
    cd langserver-typescript && git pull
fi
yarn install
./node_modules/.bin/tsc -p .

cd ..
TAG=${TAG-latest}
docker build -t us.gcr.io/sourcegraph-dev/xlang-typescript:$TAG .
gcloud docker -- push us.gcr.io/sourcegraph-dev/xlang-typescript:$TAG

