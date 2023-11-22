#!/usr/bin/env bash

version="$(./dev/app/app-version.sh)"
echo "Building version: ${version}"

echo "--- :chrome: Building web"
pnpm install
NODE_ENV=production CODY_APP=1 pnpm run build-web

export PATH=$PATH:/c/msys64/ucrt64/bin
platform="x86_64-pc-windows-msvc" # This is the name Tauri expects for the Windows executable

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
  ./cmd/sourcegraph

pnpm tauri build
