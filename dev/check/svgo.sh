#!/usr/bin/env bash
set -eu -o pipefail

cd "$(dirname "${BASH_SOURCE[0]}")"/../..

echo "--- Check static SVGs for optimizations"

# mutex is necessary since CI runs various yarn installs in parallel
yarn install

echo "Checking for potential SVG optimizations"

allSvgsOptimized() {
  for file in ./ui/assets/img/*.svg; do
    # Instead of letting svgo update the actual files, we output to STDOUT and manually check each diff ourselves.
    # By ensuring no files are actually modified, we can ensure that this lint check does not affect other checks.
    diff -w -q "$file" <(yarn --silent run optimize-svg-assets -i "$file" -o -) || return 1
  done
}

if ! allSvgsOptimized; then
  echo 'Found SVG optimizations. Please run "yarn optimize-svg-assets ui/assets/img" and commit the result.'
  exit 1
fi
