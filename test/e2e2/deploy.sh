#!/bin/bash

docker build -t us.gcr.io/sourcegraph-dev/e2e2 . ;
gcloud docker -- push us.gcr.io/sourcegraph-dev/e2e2;
kubectl delete pod $(kubectl get pods | grep -E 'e2e-chrome|e2e-firefox' | awk '{ print $1 }');
