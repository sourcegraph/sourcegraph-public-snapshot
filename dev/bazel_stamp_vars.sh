#!/usr/bin/env bash

build_number="${BUILDKITE_BUILD_NUMBER:-000000}"
date_fragment="$(date +%Y-%m-%d)"
latest_tag="5.0"
stamp_version="${build_number}_${date_fragment}_${latest_tag}-$(git rev-parse --short HEAD)"

echo STABLE_VERSION "$stamp_version"
echo VERSION_TIMESTAMP "$(date +%s)"

# Unstable Buildkite env vars
echo "BUILDKITE $BUILDKITE"
echo "BUILDKITE_COMMIT $BUILDKITE_COMMIT"
echo "BUILDKITE_BRANCH $BUILDKITE_BRANCH"
echo "BUILDKITE_PULL_REQUEST_REPO $BUILDKITE_PULL_REQUEST_REPO"
echo "BUILDKITE_PULL_REQUEST $BUILDKITE_PULL_REQUEST"
