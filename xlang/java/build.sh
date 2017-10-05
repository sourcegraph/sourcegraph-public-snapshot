#!/bin/bash
set -ex
cd $(dirname "${BASH_SOURCE[0]}")

export IMAGE=us.gcr.io/sourcegraph-dev/xlang-java
export TAG=${TAG-latest}

SKINNY_MODE=false;
if [ -z ${SKINNY} ]; then
    echo "building image with artifacts.";
else
    SKINNY_MODE=true;
    BUILD_FOLDER="docker-skinny";
    echo "building image without artifacts.";
fi

if [ ! -d "java-langserver" ]; then
    git clone git@github.com:sourcegraph/java-langserver.git java-langserver
else
    cd java-langserver && git fetch origin && git checkout origin/master && cd ..
fi

cd java-langserver
mvn clean compile assembly:single

cd ..
mv java-langserver/target/java-language-server.jar ${BUILD_FOLDER-docker}
cp java-langserver/add-android-support-libs.sh ${BUILD_FOLDER-docker}

cd ${BUILD_FOLDER-docker}
if [ "$SKINNY_MODE" = false ]; then
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
fi

docker build -t $IMAGE:$TAG .
gcloud docker -- push $IMAGE:$TAG
