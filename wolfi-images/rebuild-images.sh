#! /usr/bin/env bash

set -eu -o pipefail

# Run this script to rebuild all Wolfi base images.
#
# Normally Buildkite only rebuilds a base image when its YAML file changes - this scripts bumps a
#   value in a comment to force a rebuild.
# This is helpful to fix vulnerabilities as it will fetch the latest versions of packages.

cd "$(dirname "${BASH_SOURCE[0]}")" || exit

COMMENT_TOKEN="MANUAL REBUILD"
DATE=$(date)

# Search Wolfi YAML files - if no matching comment exists, append it
grep -L "$COMMENT_TOKEN" ./*/apko.yaml | xargs -I {} sh -c "echo \"\n# $COMMENT_TOKEN: \" >> {}"

# Update comment to include the current date & time
sed -i '' "s/# $COMMENT_TOKEN: .*/# $COMMENT_TOKEN: $DATE/" ./*/apko.yaml

echo "Buildkite will rebuild the following base images on next push:"
grep -l "# $COMMENT_TOKEN: $DATE" ./*/apko.yaml | sed 's/\.\/\(.*\)\/apko\.yaml/ 🐳 \1/'
