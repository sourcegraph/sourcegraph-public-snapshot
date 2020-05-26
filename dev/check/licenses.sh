#!/usr/bin/env bash

echo "--- licenses"
set -e

cd "$(dirname "${BASH_SOURCE[0]}")"/../..

yarn --mutex network --frozen-lockfile

./dev/licenses-npm.sh
