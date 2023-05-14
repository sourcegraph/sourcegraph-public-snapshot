#!/usr/bin/env bash
set -eu

cd "$(dirname "${BASH_SOURCE[0]}")"/../../.. || exit 1

download_artifacts() {
  local src
  local target
  src=$1
  dest=$2
  mkdir -p "${dest}"
  buildkite-agent artifact download "${src}/sourcegraph-backend-*" "${dest}"
}

# check that the directory exists and that is has files in it
if [[ ! -d "./dist" ||  -z $(ls dist/) ]]; then
  download_artifacts "dist" dist/
fi

VERSION=$(./enterprise/dev/app/app_version.sh)
echo "--- Creating GitHub release for Sourcegraph App (${VERSION})"
echo "Release will have to following assets:"
ls -al ./dist
gh release create "app-v${VERSION}" \
  --prerelease \
  --draft \
  --title "Sourcegraph App v${VERSION}" \
  --notes "A new Sourcegraph App version is available" \
  ./dist/*
