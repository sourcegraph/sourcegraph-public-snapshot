#!/usr/bin/env bash

# set -e

# Setup the correct pnpm version before actually building stuff
# https://community.render.com/t/how-to-specify-pnpm-version/8743/6
echo "--- pnpm version setup"
mkdir -p bin
corepack enable --install-directory bin
export PATH="$PWD/bin:$PATH"

pnpm_version=$(pnpm --version)
echo "pnpm version=$pnpm_version"

# echo "--- running pnpm-build.sh"
# ./dev/ci/pnpm-build.sh "$1"
