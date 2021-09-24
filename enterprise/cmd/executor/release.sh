#!/usr/bin/env bash

# Point to GCP boot disk image built by build.sh
IMAGE_NAME="executor-$(git log -n1 --pretty=format:%h)-${BUILD_TIMESTAMP}"

# Add released label to the image
gcloud compute images add-labels --project=sourcegraph-ci "${IMAGE_NAME}" --labels='released=true'

# Make image publicly accessible
gcloud compute images add-iam-policy-binding "${IMAGE_NAME}" --member='allAuthenticatedUsers' --role='roles/compute.imageUser'
