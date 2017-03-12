#!/bin/bash
set -e
cd $(dirname "${BASH_SOURCE[0]}")

export IMAGE=us.gcr.io/sourcegraph-dev/xlang-java
export TAG=${TAG-latest}

set -x

if [ ! -d "java-langserver" ]; then
    git clone git@github.com:sourcegraph/java-langserver.git java-langserver
else
    cd java-langserver && git fetch origin && git checkout origin/master && cd ..
fi

cd java-langserver
mvn clean compile assembly:single

cd ..
mv java-langserver/target/java-language-server.jar docker

cd docker
if [ -d artifacts ]; then
    cd ./artifacts && git fetch origin && git checkout origin/master && cd -
else
    git clone --depth 1 https://github.com/sourcegraph/java-artifacts artifacts
fi
docker build -t $IMAGE:$TAG .
gcloud docker -- push $IMAGE:$TAG
