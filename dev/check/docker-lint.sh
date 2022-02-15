#!/usr/bin/env bash

set -e

HADOLINT=./.bin/hadolint
HADOLINT_VERSION=v1.15.0

if [[ ! -f "$HADOLINT" ]]; then
  echo "--- install hadolint"
  mkdir -p .bin
  curl -sL -o $HADOLINT "https://github.com/hadolint/hadolint/releases/download/$HADOLINT_VERSION/hadolint-$(uname -s)-$(uname -m)"
  chmod 700 $HADOLINT
fi

echo "--- hadolint"
git ls-files | grep Dockerfile | xargs $HADOLINT
