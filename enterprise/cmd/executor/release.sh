#!/usr/bin/env bash

# Add released label to the image built by the build.sh command
gcloud compute images add-labels --project=sourcegraph-ci "executor-$(git log -n1 --pretty=format:%h)-${BUILD_TIMESTAMP}" --labels='released=true'
