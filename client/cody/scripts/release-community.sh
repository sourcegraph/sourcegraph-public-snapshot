#!/bin/bash

set -ex

cd "$(dirname "${BASH_SOURCE[0]}")/.."

jq '.name = "cody-ai-community" | .displayName = "Cody Community"' package.json > package.community.json
mv package.community.json package.json
trap 'git checkout ./package.json' EXIT

pnpm run vsce:package
pnpm run vsce:publish
