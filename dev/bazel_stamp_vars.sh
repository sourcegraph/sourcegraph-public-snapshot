#!/usr/bin/env bash

build_number="${BUILDKITE_BUILD_NUMBER:-000000}"
date_fragment="$(date +%y-%m-%d)"
latest_tag="5.0"
stamp_version="${VERSION:-${build_number}_${date_fragment}_${latest_tag}-$(git rev-parse HEAD)}"

echo STABLE_VERSION "$stamp_version"
echo VERSION_TIMESTAMP "$(date +%s)"
