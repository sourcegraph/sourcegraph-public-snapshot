#!/usr/bin/env bash
cd "$(dirname "${BASH_SOURCE[0]}")"
set -ex

# This merely re-tags the image to match our official versioning scheme. The
# actual image currently lives here: https://github.com/sourcegraph/infrastructure/tree/master/docker-images
#
# TODO: Move the image to this directory so it is open-source and built in CI automatically.
docker pull docker.io/sourcegraph/syntect_server:71d2481@sha256:8854665264522a86b711c732d803395478dab30f6df6197b5d0c1c7c21cd6261
docker tag docker.io/sourcegraph/syntect_server:71d2481@sha256:8854665264522a86b711c732d803395478dab30f6df6197b5d0c1c7c21cd6261 "$IMAGE"
