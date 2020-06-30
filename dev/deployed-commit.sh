#!/bin/bash

# This script determines the commit of sourcegraph/sourcegraph that is currently
# running on sourcegraph.com.

set -e

# Determine the image that is currently running in the k8s prod context. This will
# return the last sourcegraph/frontend image sorted by age (i.e., the newest pod).
IMAGE=$(
  kubectl -n prod get pods -l app=sourcegraph-frontend --sort-by=.metadata.creationTimestamp -o jsonpath='{..image}' |
    tr '[:space:]' '\n' |
    grep index.docker.io/sourcegraph/frontend |
    sort | uniq | tail -n1
)

# Get image locally so we can inspect
docker pull -q "${IMAGE}" >/dev/null

# Extract rev from pulled image
docker inspect "${IMAGE}" | jq -r '.[0].Config.Labels["org.opencontainers.image.revision"]'
