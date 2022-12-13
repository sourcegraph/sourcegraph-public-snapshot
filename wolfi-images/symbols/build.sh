#!/usr/bin/env bash

# Build base image using apko build container
docker run -v "$PWD":/work \
  cgr.dev/chainguard/apko \
  build apko.yaml sourcegraph/symbols-base:latest sourcegraph-symbols-base.tar ||
  echo "*** Build failed ***"

# Import into Docker
docker load <sourcegraph-symbols-base.tar
