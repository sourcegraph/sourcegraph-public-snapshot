#!/usr/bin/env bash

set -euf -o pipefail

read -p 'Have you read DEVELOPMENT.md? [y/N] ' -n 1 -r
echo
case "$REPLY" in
  Y | y) ;;
  *)
    echo 'Please read the Releasing section of DEVELOPMENT.md before running this script.'
    exit 1
    ;;
esac

if ! echo "$VERSION" | grep -Eq '^[0-9]+\.[0-9]+\.[0-9]+(-rc\.[0-9]+)?$'; then
  echo "\$VERSION is not in MAJOR.MINOR.PATCH format"
  exit 1
fi

# Create a new tag and push it, this will trigger the goreleaser workflow in .github/workflows/goreleaser.yml
git tag "${VERSION}" -a -m "release v${VERSION}"
# We use `--atomic` so that we push the tag and the commit if the commit was or wasn't pushed before
git push --atomic origin main "${VERSION}"
