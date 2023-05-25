#!/usr/bin/env bash
set -eu

cd "$(dirname "${BASH_SOURCE[0]}")"/../../.. || exit 1

download_artifacts() {
  src=$1
  dest=$2
  mkdir -p "${dest}"
  buildkite-agent artifact download "${src}" "${dest}"
}

# check that the directory exists and that is has files in it
if [[ ! -d "./dist" ||  -z $(ls dist/) ]]; then
  download_artifacts "dist/*" dist/
else
  echo "missing dist artefacts - not creating release"
  exit 1
fi

VERSION=$(./enterprise/dev/app/app_version.sh)
echo "--- :github: Creating GitHub release for Sourcegraph App (${VERSION})"
echo "Release will have to following assets:"
ls -al ./dist

# On CI it is assumed this command runs in a stateless agent, where the GITHUB_TOKEN is injected
gh release create "app-v${VERSION}" \
  --prerelease \
  --draft \
  --title "Sourcegraph App v${VERSION}" \
  --notes "A new Sourcegraph App version is available" \
  ./dist/*
