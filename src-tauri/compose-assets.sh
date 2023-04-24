#!/usr/bin/env bash
cd "$(dirname "${BASH_SOURCE[0]}")"/ || exit 1
set -x

rm -rf assets/
mkdir -p assets
cp -r ../client/app-shell/dist/* assets/
