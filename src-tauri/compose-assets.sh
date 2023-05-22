#!/usr/bin/env bash
cd "$(dirname "${BASH_SOURCE[0]}")"/ || exit 1
set -x

rm -rf assets/
mkdir -p assets
cd .. && pnpm build-app-shell && cd - ||exit 1
cp -r ../client/app-shell/dist/* assets/
