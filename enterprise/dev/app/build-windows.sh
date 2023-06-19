#!/usr/bin/env bash

echo "--- :chrome: Building web"
pnpm install
NODE_ENV=production ENTERPRISE=1 SOURCEGRAPH_APP=1 pnpm run build-web

platform="x86_64-pc-windows-msvc" # This is the name Tauri expects for the Windows executable
version="$(./enterprise/dev/app/app_version.sh)"

export GO111MODULE=on

ldflags="-s -w"
ldflags="$ldflags -X github.com/sourcegraph/sourcegraph/internal/version.version=${version}"
ldflags="$ldflags -X github.com/sourcegraph/sourcegraph/internal/version.timestamp=$(date +%s)"
ldflags="$ldflags -X github.com/sourcegraph/sourcegraph/internal/conf/deploy.forceType=app"

echo "--- :go: Building Sourcegraph Backend (${version}) for platform: ${platform}"
GOOS=windows GOARCH=amd64 go build \
  -o ".bin/sourcegraph-backend-${platform}.exe" \
  -trimpath \
  -tags dist \
  -ldflags "$ldflags" \
  ./enterprise/cmd/sourcegraph

pnpm tauri build
