#!/usr/bin/env bash

set -eu

ACCESS_TOKEN=$(gcloud auth print-access-token)
echo "Access token is $ACCESS_TOKEN"

registry="us.gcr.io/sourcegraph-dev/"
image="wolfi-postgresql-12-base"
tag="latest-main"

DIGEST=$(curl -s -L -H "Authorization: Bearer $ACCESS_TOKEN" -H "Accept: application/vnd.docker.distribution.manifest.v2+json" "https://${registry}${image}/manifests/${tag}")

echo "Digest for ${registry}/${image}:${tag} is $DIGEST"
