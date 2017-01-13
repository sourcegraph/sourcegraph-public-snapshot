#!/bin/bash
set -e
cd $(dirname "${BASH_SOURCE[0]}")

export IMAGE=us.gcr.io/sourcegraph-dev/xlang-java
export TAG=${TAG-latest}

set -x

if [ ! -d "java-langserver" ]; then
    git clone https://github.com/sourcegraph/java-langserver java-langserver
else
    cd java-langserver && git pull && cd ..
fi

cd java-langserver

mvn clean compile assembly:single

docker build -t $IMAGE:$TAG .
gcloud docker -- push $IMAGE:$TAG
