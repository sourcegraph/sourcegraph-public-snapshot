#!/usr/bin/env bash

declare -r mydir=$(dirname "$0")

if ! "${mydir}/../../../windows/check_requirements.cmd"; then
  echo "STOP! Requirements missing. Please fix before proceeding."
  exit 1
fi

echo "IN BUILD SCRIPT"
set -eux
#version="$(./enterprise/dev/app/app-version.sh)"
version="23.7.1"
echo "Building version: ${version}"

echo "--- :chrome: Building web"
pnpm install
NODE_ENV=production ENTERPRISE=1 SOURCEGRAPH_APP=1 pnpm run build-web

export PATH=$PATH:/c/msys64/ucrt64/bin
platform=${PLATFORM:-"not-defined"}

export GO111MODULE=on

ldflags="-s -w"
ldflags="$ldflags -X github.com/sourcegraph/sourcegraph/internal/version.version=${version}"
ldflags="$ldflags -X github.com/sourcegraph/sourcegraph/internal/version.timestamp=$(date +%s)"
ldflags="$ldflags -X github.com/sourcegraph/sourcegraph/internal/conf/deploy.forceType=app"

echo "--- :go: Building Sourcegraph Backend (${version}) for platform: ${platform}"
GOOS=windows GOARCH=amd64 CGO_ENABLED=1 \
  go build \
  -o ".bin/sourcegraph-backend-${platform}.exe" \
  -trimpath \
  -tags dist \
  -ldflags "$ldflags" \
  ./enterprise/cmd/sourcegraph

pnpm tauri build
