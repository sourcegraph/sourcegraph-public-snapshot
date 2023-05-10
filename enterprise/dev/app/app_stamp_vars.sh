#!/usr/bin/env bash

stamp_version="${VERSION:-"0.0.0+dev"}"

echo STABLE_VERSION "$stamp_version"
echo VERSION_TIMESTAMP "$(date +%s)"
