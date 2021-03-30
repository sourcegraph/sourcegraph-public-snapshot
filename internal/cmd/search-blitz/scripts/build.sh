#!/usr/bin/env bash
yes | gcloud auth configure-docker
docker build --build-arg COMMIT_SHA="$(git rev-parse HEAD)" -t us.gcr.io/sourcegraph-dev/search-blitz:$1 .
docker push us.gcr.io/sourcegraph-dev/search-blitz:$1
