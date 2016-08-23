#!/bin/bash
set -ex

GOOS=linux GOARCH=amd64 go build -o langprocessor-js .

if [ ! -d "js-language-processor" ]; then
    git clone https://github.com/antonina-cherednichenko/poc-jslang-server js-language-processor && cd js-language-processor
else
    cd js-language-processor && git pull
fi
npm install
tsc -p .
cd ..
docker build -t us.gcr.io/sourcegraph-dev/langprocessor-js .
#gcloud docker push us.gcr.io/sourcegraph-dev/langprocessor-js
