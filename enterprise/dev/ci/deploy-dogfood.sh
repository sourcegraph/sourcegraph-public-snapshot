#!/bin/bash
set -ex

case "$BUILDKITE_BRANCH" in
    master)
        DEPLOYMENT=sourcegraph-frontend
        CONTAINER=frontend
        ;;

    *)
        echo "Expected BUILDKITE_BRANCH to match master, got $BUILDKITE_BRANCH"
        exit 1
esac

IMAGE=$(kubectl get deployment "--namespace=$NAMESPACE" "--context=$CONTEXT" -o 'jsonpath={.spec.template.spec.containers[?(@.name=="'"$CONTAINER"'")].image}' "$DEPLOYMENT" | awk -F ':' '{printf $1}')

kubectl "--namespace=$NAMESPACE" "--context=$CONTEXT" set image "deployment/$DEPLOYMENT" "$CONTAINER=$IMAGE:$VERSION"
kubectl "--namespace=$NAMESPACE" "--context=$CONTEXT" rollout status "deployment/$DEPLOYMENT"

if [ -n "$DEPLOYMENT_BG" ]; then
    kubectl "--namespace=$NAMESPACE" "--context=$CONTEXT" set image "deployment/$DEPLOYMENT_BG" "$CONTAINER_BG=$IMAGE:$VERSION"
    kubectl "--namespace=$NAMESPACE" "--context=$CONTEXT" rollout status "deployment/$DEPLOYMENT_BG"
fi
