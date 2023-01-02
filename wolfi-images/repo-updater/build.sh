#!/usr/bin/env bash

# Build base image using apko build container
docker run \
  -v "$PWD":/work \
  -v "$PWD/../../dependencies/packages":/work/packages \
  -v "$PWD/../../dependencies/keys":/work/keys \
  cgr.dev/chainguard/apko \
  build --debug -k /work/keys/melange.rsa.pub apko.yaml \
  sourcegraph/repo-updater-base:latest sourcegraph-repo-updater-base.tar ||
  echo "*** Build failed ***"

# Import into Docker
docker load <sourcegraph-repo-updater-base.tar
