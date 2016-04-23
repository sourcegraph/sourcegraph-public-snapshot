#!/bin/bash

set -ex

VERSION=$1
SERVER=$2

make dist PACKAGEFLAGS="--os=linux $VERSION"

cp release/$VERSION/linux-amd64 deploy/sourcegraph/src
docker build -t us.gcr.io/sourcegraph-dev/sourcegraph:$VERSION deploy/sourcegraph
gcloud config set project sourcegraph-dev
gcloud docker push us.gcr.io/sourcegraph-dev/sourcegraph:$VERSION

curl http://deploy-bot.sourcegraph.com/set-branch-version -F "token=$DEPLOY_BOT_TOKEN" -F "branch=$SERVER" -F "version=$VERSION"
