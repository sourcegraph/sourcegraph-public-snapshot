#!/usr/bin/env bash

set -e

echo "--- Pack ./ui ./client ./dev/ci/*.sh and root files"
mkdir -p client-pack
find . -maxdepth 1 -type f -print0 | xargs echo ui client dev/ci/*.sh | xargs tar -czf 'client-pack/client.tar.gz'

echo "--- Upload pre-built client artifact"
cd client-pack
buildkite-agent artifact upload 'client.tar.gz'
