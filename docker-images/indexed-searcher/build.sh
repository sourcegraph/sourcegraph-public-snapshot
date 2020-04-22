#!/usr/bin/env bash

set -ex
cd "$(dirname "${BASH_SOURCE[0]}")"

# This merely re-tags the image to match our official versioning scheme. The
# actual image currently lives here: https://github.com/sourcegraph/zoekt/blob/master/Dockerfile.webserver
docker pull index.docker.io/sourcegraph/zoekt-webserver:0.0.20200401202737-ef3ec23@sha256:d48de388d28899fd0c3ad0d6f84d466b3a1f533f6b967a713918d438ab8bc63c
docker tag index.docker.io/sourcegraph/zoekt-webserver:0.0.20200401202737-ef3ec23@sha256:d48de388d28899fd0c3ad0d6f84d466b3a1f533f6b967a713918d438ab8bc63c "$IMAGE"
