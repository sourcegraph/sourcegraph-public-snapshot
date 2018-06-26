#!/bin/bash
set -ex

case "$BUILDKITE_BRANCH" in
    master)
        DEPLOYMENT=sourcegraph-frontend
        CONTAINER=frontend
    docker-images/*)
        DEPLOYMENT=$($BUILDKITE_BRANCH | awk -F '/' '{printf $2}')
        CONTAINER=$DEPLOYMENT
    *)
        echo "Expected BUILDKITE_BRANCH to match master or docker-images/*, got $BUILDKITE_BRANCH"
        exit 1
esac

case "$DEPLOYMENT" in
    gitserver)
        DEPLOYMENT="gitserver-1"
    xlang-javascript-typescript)
        DEPLOYMENT="xlang-typescript"
esac

IMAGE=$(kubectl get deployment "--namespace=$NAMESPACE" "--context=$CONTEXT" -o 'jsonpath={.spec.template.spec.containers[?(@.name="'"$CONTAINER"'")].image}' "$DEPLOYMENT" | awk -F ':' '{printf $1}')

kubectl "--namespace=$NAMESPACE" "--context=$CONTEXT" set image "deployment/$DEPLOYMENT" "$DEPLOYMENT=$IMAGE:$VERSION"
kubectl "--namespace=$NAMESPACE" "--context=$CONTEXT" rollout status "deployment/$DEPLOYMENT"
