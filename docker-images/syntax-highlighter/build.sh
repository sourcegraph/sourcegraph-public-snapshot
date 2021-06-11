#!/usr/bin/env bash
cd "$(dirname "${BASH_SOURCE[0]}")"
set -ex

# This merely re-tags the image to match our official versioning scheme. The
# actual image currently lives here: https://github.com/sourcegraph/infrastructure/tree/master/docker-images
#
# TODO: Move the image to this directory so it is open-source and built in CI automatically.
docker pull docker.io/sourcegraph/syntect_server:137d7de@sha256:36e5f85519052bc660345753cdb4774ea38f81b3b3ba65eec184ac53a3ad1c7b
docker tag docker.io/sourcegraph/syntect_server:137d7de@sha256:36e5f85519052bc660345753cdb4774ea38f81b3b3ba65eec184ac53a3ad1c7b "$IMAGE"
