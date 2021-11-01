#!/usr/bin/env bash

# This script publishes the executor docker registry mirror image built by build.sh

cd "$(dirname "${BASH_SOURCE[0]}")"
set -eu

export AWS_ACCESS_KEY_ID="${AWS_EXECUTOR_AMI_ACCESS_KEY}"
export AWS_SECRET_ACCESS_KEY="${AWS_EXECUTOR_AMI_SECRET_KEY}"

# Point to GCP boot disk image/AMI built by build.sh script
NAME="executor-docker-mirror-$(git log -n1 --pretty=format:%h)-${BUILDKITE_BUILD_NUMBER}"
GOOGLE_IMAGE_NAME="${NAME}"
AWS_AMI_ID=$(aws ec2 describe-images --filter "Name=name,Values=${NAME}" --query 'Images[*].[ImageId]' --output text)

# Mark GCP boot disk as released and make it usable outside of Sourcegraph
gcloud compute images add-labels --project=sourcegraph-ci "${GOOGLE_IMAGE_NAME}" --labels='released=true'
gcloud compute images add-iam-policy-binding --project=sourcegraph-ci "${GOOGLE_IMAGE_NAME}" --member='allAuthenticatedUsers' --role='roles/compute.imageUser'

# Make AMI usable outside of Sourcegraph
aws ec2 modify-image-attribute --image-id "${AWS_AMI_ID}" --launch-permission "Add=[{Group=all}]"
