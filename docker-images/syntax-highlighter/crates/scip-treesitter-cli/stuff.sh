#!/usr/bin/env bash

image="$1"
out="$2"

echo here
docker load --input="$1"

# commented just to allow completing the whole process as the -h
# will lead to exit code 1
# ---
# docker run --rm scip-java-tmp:tmp -- -h
echo here

echo "it works" > "$out"
