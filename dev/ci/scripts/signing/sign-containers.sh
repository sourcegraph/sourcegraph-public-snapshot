#!/usr/bin/env bash

set -euxo pipefail

echo "Placeholder for signing containers..."

which buildkite-agent

mkdir -p tmp/
buildkite-agent artifact download pushed-images.txt tmp/ --step simulate-push-images

ls -al tmp/
cat tmp/pushed-images.txt
