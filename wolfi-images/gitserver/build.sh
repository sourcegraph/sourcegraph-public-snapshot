#!/usr/bin/env bash

# Build base image using apko build container
docker run \
  -v "$PWD":/work \
  -v "$PWD/../../dependencies/packages":/work/packages \
  -v "$PWD/../../dependencies/keys":/work/keys \
  cgr.dev/chainguard/apko \
  build --debug -k /work/keys/melange.rsa.pub apko.yaml sourcegraph/gitserver-base:latest sourcegraph-gitserver-base.tar ||
  echo "*** Build failed ***"

# Import into Docker
docker load <sourcegraph-gitserver-base.tar
