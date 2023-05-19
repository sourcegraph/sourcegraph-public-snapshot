#! /usr/bin/env bash

set -eu -o pipefail

# Run this script to rebuild all Wolfi base images.
#
# Normally Buildkite only rebuilds a base image when its YAML file changes - this scripts bumps a
#   value in a comment to force a rebuild.
# This is helpful to fix vulnerabilities as it will fetch the latest versions of packages.

IMAGE=${1:-'*'}
IMAGE="${IMAGE%.yaml}"

IMAGE_REGEX=$IMAGE
if [ "$IMAGE" == "*" ]; then
  IMAGE_REGEX=".*"
fi

cd "$(dirname "${BASH_SOURCE[0]}")" || exit

COMMENT_TOKEN="MANUAL REBUILD"
DATE=$(date)

# Search Wolfi YAML files - if no matching comment exists, append it
# shellcheck disable=SC2086
grep -L "$COMMENT_TOKEN" ./${IMAGE}.yaml | xargs -I {} sh -c "echo \"\n# $COMMENT_TOKEN: \" >> {}"

# Update comment to include the current date & time
# shellcheck disable=SC2086
sed -i '' "s/# $COMMENT_TOKEN: .*/# $COMMENT_TOKEN: $DATE/" ./${IMAGE}.yaml

echo "Buildkite will rebuild the following base images on next push:"
# shellcheck disable=SC2086
grep -l "# $COMMENT_TOKEN: $DATE" ./${IMAGE}.yaml | sed "s/\.\/\(${IMAGE_REGEX}\)\.yaml/ 🐳 \1/"
