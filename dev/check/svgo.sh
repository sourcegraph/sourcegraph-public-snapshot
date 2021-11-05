#!/usr/bin/env bash
set -eu -o pipefail

cd "$(dirname "${BASH_SOURCE[0]}")"/../..

echo "--- Check static SVGs for optimizations"

# mutex is necessary since CI runs various yarn installs in parallel
yarn --mutex network --frozen-lockfile --ignore-scripts

echo "Checking for potential SVG optimizations"

allSvgsOptimized() {
  for file in ./ui/assets/img/*.svg; do
    diff -w -q "$file" <(yarn run -s optimize-svg-assets -i "$file" -o -) || return 1
  done
}

if ! allSvgsOptimized; then
  echo 'Found SVG optimizations. Please run "yarn optimize-svg-assets ui/assets/img" and commit the result.'
  exit 1
fi
