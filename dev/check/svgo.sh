#!/usr/bin/env bash
set -eu -o pipefail

echo "--- Check static SVGS for optimizations"

# Prevent duplicates in yarn.lock/node_modules that lead to errors and bloated bundle sizes

# mutex is necessary since CI runs various yarn installs in parallel
yarn --mutex network --frozen-lockfile --ignore-scripts

echo "Checking for potential SVG optimizations"
yarn run -s optimize-svg-assets
git diff --exit-code -- '*.svg'
