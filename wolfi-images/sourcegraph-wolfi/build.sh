#!/usr/bin/env bash

# Build base image using apko build container
docker run -v "$PWD":/work \
  cgr.dev/chainguard/apko \
  build apko.yaml sourcegraph/wolfi-base:latest sourcegraph-wolfi-base.tar ||
  echo "*** Build failed ***"

# Import into Docker
docker load <sourcegraph-wolfi-base.tar
