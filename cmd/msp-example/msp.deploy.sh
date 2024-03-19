#!/usr/bin/env bash

set -eux -o pipefail

# Service Specific Parameters
GCP_PROJECT="msp-testbed-robert-7be9"
GCP_REGION="us-central1"
GCP_DELIVERY_PIPELINE="msp-testbed-us-central1-rollout"
GCP_CLOUDRUN_SKAFFOLD_SOURCE="gs://msp-testbed-robert-7be9-cloudrun-skaffold/source.tar.gz"
REPOSITORY="us.gcr.io/sourcegraph-dev/msp-example"

# Env Variable Defaults
: "${BUILDKITE_BUILD_NUMBER:?"BUILDKITE_BUILD_NUMBER is required"}"
: "${BUILDKITE_COMMIT:?"BUILDKITE_COMMIT is required"}"

# TODO: figure out a good way to capture author details
# can contain only lowercase letters, numeric characters, underscores, and dashes.
# All characters must use UTF-8 encoding, and international characters are allowed.
# Keys must start with a lowercase letter or international character
# : ${BUILDKITE_BUILD_AUTHOR_EMAIL:?"BUILDKITE_BUILD_AUTHOR_EMAIL is required"}

# Computed Variables
SHORT_SHA="${BUILDKITE_COMMIT:0:12}"
TAG="${SHORT_SHA}_${BUILDKITE_BUILD_NUMBER}"
# resource ids must be lower-case letters, numbers, and hyphens,
# with the first character a letter, the last a letter or a number,
# and a 63 character maximum
RELEASE_NAME="deploy-${SHORT_SHA}-${BUILDKITE_BUILD_NUMBER}"

# Commands are passed as args to the script
push=$1
gcloud=$2

# Push a known tag so it is guaranteed to exist for the deploy
1>&2 "${push}" --tag "${TAG}" --repository "${REPOSITORY}"

# Create the Cloud Deploy release
1>&2 "${gcloud}" deploy releases create "${RELEASE_NAME}" \
    --project="${GCP_PROJECT}" \
    --region="${GCP_REGION}" \
    --delivery-pipeline=${GCP_DELIVERY_PIPELINE} \
    --source="${GCP_CLOUDRUN_SKAFFOLD_SOURCE}" \
    --labels="commit=${BUILDKITE_COMMIT}" \
    --deploy-parameters="customTarget/tag=${TAG}"
