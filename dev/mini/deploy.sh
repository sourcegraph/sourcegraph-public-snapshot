#!/bin/bash

# Script to deploy initializer image to Google Cloud Registry

cd ./init
gcloud docker -- push us.gcr.io/sourcegraph-dev/initializer:latest
docker tag $(docker build . | tail -n1 | awk -F ' ' {'print $3'}) us.gcr.io/sourcegraph-dev/initializer:latest
cd ..
