#!/usr/bin/env bash

# This script builds the symbols docker image.

cd "$(dirname "${BASH_SOURCE[0]}")/../../.."
set -eu

env \
  PKG=github.com/sourcegraph/sourcegraph/enterprise/cmd/symbols \
  cmd/symbols/build.sh
