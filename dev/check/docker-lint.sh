#!/usr/bin/env bash

set -e

echo "--- install hadolint"
mkdir -p .bin
curl -sL -o .bin/hadolint "https://github.com/hadolint/hadolint/releases/download/v1.15.0/hadolint-$(uname -s)-$(uname -m)"
chmod 700 .bin/hadolint

echo "--- hadolint"
git ls-files | grep Dockerfile | xargs ./.bin/hadolint
