#!/usr/bin/env bash

set -eu

HADOLINT=./.bin/hadolint
HADOLINT_VERSION=v1.15.0

arch=$(uname -m)
os=$(uname -s)

# Previous versions of this script may have downloaded a 404 as the binary
if [[ -f "$HADOLINT" ]]; then
  if head -c 10 "$HADOLINT" | grep "Not Found" >/dev/null; then
    rm $HADOLINT
  fi
fi

# Hadolint does not release arm64 binaries, rely on Rosetta instead.
if [[ "$os" == "Darwin" ]] && [[ "$arch" = "arm64" ]]; then
  arch="x86_64"
fi

if [[ ! -f "$HADOLINT" ]]; then
  echo "--- install hadolint"
  mkdir -p .bin
  curl -sLf -o $HADOLINT -w "Downloading hadolint from %{url}\n%{http_code}\n" "https://github.com/hadolint/hadolint/releases/download/$HADOLINT_VERSION/hadolint-$os-$arch"
  chmod 700 $HADOLINT
fi

echo "--- hadolint"
git ls-files | grep Dockerfile | xargs $HADOLINT
