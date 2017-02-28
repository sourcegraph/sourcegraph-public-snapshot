#!/bin/bash
set -e
cd $(dirname "${BASH_SOURCE[0]}")

export IMAGE=us.gcr.io/sourcegraph-dev/xlang-java
export TAG=${TAG-latest}

set -x

if [ ! -d "java-langserver" ]; then
    git clone git@github.com:sourcegraph/java-langserver.git java-langserver
else
    cd java-langserver && git pull && cd ..
fi

cd java-langserver
mvn clean compile assembly:single

cd ..
mv java-langserver/target/java-language-server.jar docker

cd docker
git submodule update --init artifacts
docker build -t $IMAGE:$TAG .
gcloud docker -- push $IMAGE:$TAG
