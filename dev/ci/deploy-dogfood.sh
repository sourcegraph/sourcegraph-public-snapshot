#!/bin/bash
set -ex

case "$BUILDKITE_BRANCH" in
    master)
        DEPLOYMENT=sourcegraph-frontend
        CONTAINER=frontend
        ;;

    docker-images/gitserver)
        DEPLOYMENT="gitserver-1"
        CONTAINER="gitserver-1"
        ;;

    docker-images/xlang-*)
        DEPLOYMENT=$(echo $BUILDKITE_BRANCH | awk -F '/' '{printf $2}')
        CONTAINER=$DEPLOYMENT
        # All of the language servers managed in this repo have a background deployment.
        DEPLOYMENT_BG=$DEPLOYMENT-bg
        CONTAINER_BG=$DEPLOYMENT_BG
        ;;

    docker-images/*)
        DEPLOYMENT=$(echo $BUILDKITE_BRANCH | awk -F '/' '{printf $2}')
        CONTAINER=$DEPLOYMENT
        ;;

    *)
        echo "Expected BUILDKITE_BRANCH to match master or docker-images/*, got $BUILDKITE_BRANCH"
        exit 1
esac

IMAGE=$(kubectl get deployment "--namespace=$NAMESPACE" "--context=$CONTEXT" -o 'jsonpath={.spec.template.spec.containers[?(@.name=="'"$CONTAINER"'")].image}' "$DEPLOYMENT" | awk -F ':' '{printf $1}')

kubectl "--namespace=$NAMESPACE" "--context=$CONTEXT" set image "deployment/$DEPLOYMENT" "$CONTAINER=$IMAGE:$VERSION"
kubectl "--namespace=$NAMESPACE" "--context=$CONTEXT" rollout status "deployment/$DEPLOYMENT"

if [ -n "$DEPLOYMENT_BG" ]; then
    kubectl "--namespace=$NAMESPACE" "--context=$CONTEXT" set image "deployment/$DEPLOYMENT_BG" "$CONTAINER_BG=$IMAGE:$VERSION"
    kubectl "--namespace=$NAMESPACE" "--context=$CONTEXT" rollout status "deployment/$DEPLOYMENT_BG"
fi
