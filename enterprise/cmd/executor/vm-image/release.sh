#!/usr/bin/env bash

# This script publishes the executor images built by build.sh.

cd "$(dirname "${BASH_SOURCE[0]}")"
set -eu

export AWS_ACCESS_KEY_ID="${AWS_EXECUTOR_AMI_ACCESS_KEY}"
export AWS_SECRET_ACCESS_KEY="${AWS_EXECUTOR_AMI_SECRET_KEY}"

# Point to GCP boot disk image/AMI built by build.sh script
NAME="${IMAGE_FAMILY}-${BUILDKITE_BUILD_NUMBER}"
GOOGLE_IMAGE_NAME="${NAME}"

# Mark GCP boot disk as released and make it usable outside of Sourcegraph.
echo "Publishing GCP compute image"
gcloud compute images add-iam-policy-binding --project=sourcegraph-ci "${GOOGLE_IMAGE_NAME}" --member='allAuthenticatedUsers' --role='roles/compute.imageUser'
echo "Made GCP compute image public"
gcloud compute images update --project=sourcegraph-ci "${GOOGLE_IMAGE_NAME}" --family="${IMAGE_FAMILY}"
echo "Added GCP compute image to image family ${IMAGE_FAMILY}"

# Set the AMIs to public.
if [ "${EXECUTOR_IS_TAGGED_RELEASE}" = "true" ]; then
  for region in $(jq -r '.[]' <aws_regions.json); do
    echo "Getting AMI ID in region ${region}"
    AWS_AMI_ID=$(aws ec2 --region="${region}" describe-images --filter "Name=name,Values=${NAME}" --query 'Images[*].[ImageId]' --output text)
    echo "Found AMI! ID: ${AWS_AMI_ID}"
    # Make executor AMI usable outside of Sourcegraph.
    aws ec2 --region="${region}" modify-image-attribute --image-id "${AWS_AMI_ID}" --launch-permission "Add=[{Group=all}]"
    echo "Published AMI ${AWS_AMI_ID}"
  done
else
  echo "Getting AMI ID in region us-west-2"
  AWS_AMI_ID=$(aws ec2 --region="us-west-2" describe-images --filter "Name=name,Values=${NAME}" --query 'Images[*].[ImageId]' --output text)
  echo "Found AMI! ID: ${AWS_AMI_ID}"
  # Make executor AMI usable outside of Sourcegraph.
  aws ec2 --region="us-west-2" modify-image-attribute --image-id "${AWS_AMI_ID}" --launch-permission "Add=[{Group=all}]"
  echo "Published AMI ${AWS_AMI_ID}"
fi
