#!/usr/bin/env bash

set -euo pipefail

cd "$(dirname "${BASH_SOURCE[0]}")/../../../.."

export PATH="$PWD/../deploy-sourcegraph-cloud:$PATH"

image="us.gcr.io/sourcegraph-dev/search-blitz:$1"

set -x

cd ../deploy-sourcegraph-cloud
update-images.py "$image"

cd ../deploy-sourcegraph-dogfood-k8s
update-images.py "$image"
