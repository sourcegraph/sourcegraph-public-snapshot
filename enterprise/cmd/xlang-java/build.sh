#!/bin/bash

set -ex
cd $(dirname "${BASH_SOURCE[0]}")

export IMAGE=${IMAGE-us.gcr.io/sourcegraph-dev/xlang-java}
export VERSION=${VERSION-dev}

BUILD_FOLDER="docker";
echo "building image with artifacts.";

if [ ! -d "java-langserver" ]; then
    git clone git@github.com:sourcegraph/java-langserver.git java-langserver
else
    cd java-langserver && git fetch origin && git checkout origin/master && cd ..
fi

cd java-langserver
mvn clean compile assembly:single

cd ..
mv java-langserver/target/java-language-server.jar "$BUILD_FOLDER"
cp java-langserver/add-android-support-libs.sh "$BUILD_FOLDER"
cd "$BUILD_FOLDER"

# Add artifacts
if [ -d artifacts ]; then
    cd ./artifacts && git fetch origin && git checkout origin/master && cd -
else
    git clone --depth 1 https://github.com/sourcegraph/java-artifacts artifacts
fi
if [ -d android-sdk-jars ]; then
    cd ./android-sdk-jars && git fetch origin && git checkout origin/master && cd -
else
    git clone --depth 1 https://github.com/sourcegraph/android-sdk-jars
fi

docker build -t $IMAGE:$VERSION .
