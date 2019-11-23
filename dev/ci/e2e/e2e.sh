#!/usr/bin/env bash

cd $(dirname "${BASH_SOURCE[0]}")
set -ex

if [ -z "$IMAGE_VERSION" ]; then
    echo "Must specify \$IMAGE_VERSION."
    exit 1
fi

if [ -z "$ID" ]; then
    echo "Must specify \$ID."
    exit 1
fi

gcloud container clusters get-credentials ci-e2e --zone us-central1-a --project sourcegraph-dev

YAML=$(cat sourcegraph-server.Pod.yaml | sed "s/{{ID}}/$ID/g" | sed "s/{{IMAGE_VERSION}}/$IMAGE_VERSION/g")

echo "$YAML"
time echo "$YAML" | kubectl apply -f -

read -n 1 -p "Teardown..."
time echo "$YAML" | kubectl delete -f -


