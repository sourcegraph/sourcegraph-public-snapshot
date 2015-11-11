#!/bin/bash

# Check if all code is gofmt'd

DIFF=`find . \( -path ./Godeps -o -path ./vendored \) -prune -o -name '*.go' -exec gofmt -d {} +`;
if [ -z "$DIFF" ]; then
    exit 0;
else
    echo "ERROR: gofmt check failed:";
    echo "$DIFF";
    exit 1;
fi
