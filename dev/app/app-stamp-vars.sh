#!/usr/bin/env bash

# This CANNOT be 0.0.0+dev, or else the binary will not start:
# https://github.com/sourcegraph/sourcegraph/issues/50958
# Note this also must be > any OOB migration version so that they run.
stamp_version="${VERSION:-"2023.0.0+dev"}"

echo STABLE_VERSION "$stamp_version"
echo VERSION_TIMESTAMP "$(date +%s)"
