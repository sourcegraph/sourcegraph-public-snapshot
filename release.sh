#!/usr/bin/env bash

set -euf -o pipefail

if ! echo "$VERSION" | grep -Eq '^[0-9]+\.[0-9]+\.[0-9]+$'; then
  echo "\$VERSION is not in MAJOR.MINOR.PATCH format"
  exit 1
fi

# Push a new tag, this will trigger the goreleaser workflow in .github/workflows/goreleaser.yml
git tag "${VERSION}" -a -m "release v${VERSION}"
git tag latest -f -a -m "release v${VERSION}"
git push -f --tags
