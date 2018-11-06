#!/bin/bash

set -ex
cd $(dirname "${BASH_SOURCE[0]}")

export IMAGE=${IMAGE-us.gcr.io/sourcegraph-dev/xlang-java-skinny}
export VERSION=${VERSION-dev}

BUILD_FOLDER="docker";
echo "building image without artifacts.";

if [ ! -d "java-langserver" ]; then
    git clone git@github.com:sourcegraph/java-langserver.git ./java-langserver
else
    pushd ./java-langserver && git fetch origin && git checkout origin/master && popd
fi

pushd ./java-langserver
mvn clean compile assembly:single

popd
mv ./java-langserver/target/java-language-server.jar "$BUILD_FOLDER"
cp ./java-langserver/add-android-support-libs.sh "$BUILD_FOLDER"

pushd "./$BUILD_FOLDER"
docker build -t $IMAGE:$VERSION .
