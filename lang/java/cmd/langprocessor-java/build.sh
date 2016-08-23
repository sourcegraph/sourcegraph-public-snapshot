#!/bin/bash
set -ex

GOOS=linux GOARCH=amd64 go build -o langprocessor-java .

if [ ! -d "java-language-processor" ]; then
    git clone https://github.com/alexsaveliev/java-language-processor && cd java-language-processor
else
    cd java-language-processor && git pull
fi
./gradlew assemble
cd ..
docker build -t us.gcr.io/sourcegraph-dev/langprocessor-java .
gcloud docker push us.gcr.io/sourcegraph-dev/langprocessor-java
