#!/bin/bash
set -ex
cd $(dirname "${BASH_SOURCE[0]}")

env GOBIN=$PWD/../../../../vendor/.bin go install sourcegraph.com/sourcegraph/sourcegraph/vendor/github.com/neelance/godockerize
../../../../vendor/.bin/godockerize build -t us.gcr.io/sourcegraph-dev/langprocessor-go . 
gcloud docker -- push us.gcr.io/sourcegraph-dev/langprocessor-go
