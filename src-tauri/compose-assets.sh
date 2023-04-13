#!/usr/bin/env bash
cd "$(dirname "${BASH_SOURCE[0]}")"/
set -ex

rm -rf .assets/
mkdir -p assets/.assets
cp -R ../ui/assets/* assets/.assets
cd assets/.assets/
mv index.html ..
rm -rf *.go *.gz *.br *.map chunks/*.map
