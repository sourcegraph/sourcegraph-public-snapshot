#!/usr/bin/env bash

GOARCH=amd64 GOOS=linux go build .

docker build --platform linux/amd64 -t us-central1-docker.pkg.dev/sourcegraph-ci/build-tracker/build-tracker .
docker push  us-central1-docker.pkg.dev/sourcegraph-ci/build-tracker/build-tracker
