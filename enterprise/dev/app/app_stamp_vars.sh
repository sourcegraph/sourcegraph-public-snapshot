#!/usr/bin/env bash

# This CANNOT be 0.0.0+dev, or else the binary will not start:
# https://github.com/sourcegraph/sourcegraph/issues/50958
stamp_version="${VERSION:-"1.0.0+dev"}"

echo STABLE_VERSION "$stamp_version"
echo VERSION_TIMESTAMP "$(date +%s)"
