#!/usr/bin/env bash

cd "$(dirname "${BASH_SOURCE[0]}")/../../.."
set -eu

env \
  PKG=github.com/sourcegraph/sourcegraph/enterprise/cmd/symbols \
  cmd/symbols/go-build.sh "$@"
