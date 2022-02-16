#!/usr/bin/env bash

set -e

echo "--- install hadolint"
curl -sL -o hadolint "https://github.com/hadolint/hadolint/releases/download/v1.15.0/hadolint-$(uname -s)-$(uname -m)"
chmod 700 hadolint

echo "--- hadolint"
git ls-files | grep Dockerfile | xargs ./hadolint
