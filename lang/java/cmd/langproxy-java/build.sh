#!/bin/bash
set -ex

env GOBIN=$PWD/../../../../vendor/.bin go install sourcegraph.com/sourcegraph/sourcegraph/vendor/github.com/neelance/godockerize
../../../../vendor/.bin/godockerize build -t us.gcr.io/sourcegraph-dev/langproxy-java .
gcloud docker -- push us.gcr.io/sourcegraph-dev/langproxy-java
