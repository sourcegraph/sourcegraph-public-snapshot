#!/bin/bash
set -ex

VERSION=$(date +"%y%m%d")-$CIRCLE_BUILD_NUM-${CIRCLE_SHA1:0:7}
make dist PACKAGEFLAGS=$VERSION

cp release/$VERSION/linux-amd64 deploy/sourcegraph/src
echo $GCLOUD_SERVICE_ACCOUNT | base64 --decode > gcloud-service-account.json
gcloud auth activate-service-account --key-file gcloud-service-account.json
gcloud config set project sourcegraph-dev
docker build -t us.gcr.io/sourcegraph-dev/sourcegraph:$CIRCLE_SHA1 deploy/sourcegraph
docker tag us.gcr.io/sourcegraph-dev/sourcegraph:$CIRCLE_SHA1 us.gcr.io/sourcegraph-dev/sourcegraph:latest
gcloud docker push us.gcr.io/sourcegraph-dev/sourcegraph:$CIRCLE_SHA1
gcloud docker push us.gcr.io/sourcegraph-dev/sourcegraph:latest
