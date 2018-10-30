#!/usr/bin/env bash

set -e
cd $(dirname "${BASH_SOURCE[0]}")/..

if [ ! -d ../sourcegraph ]; then
    echo "OSS repo not found at ../sourcegraph"
    exit 1
fi

echo "Linking OSS backend"
go mod edit -replace=github.com/sourcegraph/sourcegraph=../sourcegraph

echo "OSS repo linked. To unlink, run ./dev/unlink-oss.sh"
echo "Do not commit the change to go.mod"
echo "Make sure that the OSS and enterprise repos are at compatible commits"
