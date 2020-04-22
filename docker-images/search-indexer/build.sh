#!/usr/bin/env bash

set -ex
cd "$(dirname "${BASH_SOURCE[0]}")"

# This merely re-tags the image to match our official versioning scheme. The
# actual image currently lives here: https://github.com/sourcegraph/zoekt/blob/master/Dockerfile.indexserver
docker pull index.docker.io/sourcegraph/zoekt-indexserver:0.0.20200401202737-ef3ec23@sha256:354ed968e62a7d011b647476a63116813aea23bdada0a2fc4322df5381acb6b3
docker tag index.docker.io/sourcegraph/zoekt-indexserver:0.0.20200401202737-ef3ec23@sha256:354ed968e62a7d011b647476a63116813aea23bdada0a2fc4322df5381acb6b3 "$IMAGE"
