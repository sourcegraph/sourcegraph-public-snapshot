#!/usr/bin/env bash
set -eu -o pipefail

echo "--- Check static SVGs for optimizations"

# mutex is necessary since CI runs various yarn installs in parallel
yarn --mutex network --frozen-lockfile --ignore-scripts

echo "Checking for potential SVG optimizations"
yarn run -s optimize-svg-assets
git diff --exit-code -- '*.svg' || echo 'Found SVG optimizations. Please run "yarn optimize-svg-assets" and commit the result.'
