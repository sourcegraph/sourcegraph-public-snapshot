#!/bin/bash
cd $(dirname "${BASH_SOURCE[0]}")
set -ex

# This merely re-tags the image to match our official versioning scheme. The
# actual image currently lives here: https://github.com/sourcegraph/infrastructure/tree/master/docker-images
#
# TODO: Move the image to this directory so it is open-source and built in CI automatically.
docker pull index.docker.io/sourcegraph/jaeger-all-in-one:1.17.1@sha256:46bfa2ac08dd08181ab443ef966d664048a6c6ac725054bcc1fbfda5bd4040a3
docker tag index.docker.io/sourcegraph/jaeger-all-in-one:1.17.1@sha256:46bfa2ac08dd08181ab443ef966d664048a6c6ac725054bcc1fbfda5bd4040a3 $IMAGE
