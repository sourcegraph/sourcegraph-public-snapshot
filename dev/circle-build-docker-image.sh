#!/bin/bash
set -ex

VERSION=$1

make dist PACKAGEFLAGS="--os=linux $VERSION"

cp release/$VERSION/linux-amd64 deploy/sourcegraph/src
docker build -t us.gcr.io/sourcegraph-dev/sourcegraph:$VERSION deploy/sourcegraph

echo $GCLOUD_SERVICE_ACCOUNT | base64 --decode > gcloud-service-account.json
gcloud auth activate-service-account --key-file gcloud-service-account.json
gcloud config set project sourcegraph-dev
gcloud docker -- push us.gcr.io/sourcegraph-dev/sourcegraph:$VERSION
