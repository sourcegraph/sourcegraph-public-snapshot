#!/usr/bin/env bash
cd "$(dirname "${BASH_SOURCE[0]}")"
set -ex

# This merely re-tags the image to match our official versioning scheme.
docker pull gcr.io/google-containers/cadvisor:v0.35.0@sha256:4074c8bc608b78af3ca3d6e60b3794369a190ab2efd992e31b3079b075401efa
docker tag gcr.io/google-containers/cadvisor:v0.35.0@sha256:4074c8bc608b78af3ca3d6e60b3794369a190ab2efd992e31b3079b075401efa "$IMAGE"
