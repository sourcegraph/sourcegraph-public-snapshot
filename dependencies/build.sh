#!/usr/bin/env bash

if [ $# -eq 0 ]; then
  echo "No arguments supplied - provide the melange YAML file to build"
  exit 0
fi

docker run --privileged \
  -v "$PWD":/work \
  # -v "$HOME/scratch/tmp":/tmp \ # Useful for debugging
  cgr.dev/chainguard/melange build $1 --arch x86_64 --signing-key keys/melange.rsa
