#!/usr/bin/env bash

stamp_version="${VERSION:-$(git rev-parse HEAD)}"

echo STABLE_VERSION "$stamp_version"
echo VERSION_TIMESTAMP "$(date +%s)"
