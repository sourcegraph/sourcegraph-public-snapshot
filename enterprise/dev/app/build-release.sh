#!/usr/bin/env bash

set -eu

GCLOUD_APP_CREDENTIALS_FILE=${GCLOUD_APP_CREDENTIALS_FILE-$HOME/.config/gcloud/application_default_credentials.json}
cd "$(dirname "${BASH_SOURCE[0]}")"/../../.. || exit 1

if [ -z "${SKIP_BUILD_WEB-}" ]; then
  # esbuild is faster
  pnpm install
  NODE_ENV=production ENTERPRISE=1 SOURCEGRAPH_APP=1 DEV_WEB_BUILDER=esbuild pnpm run build-web
fi

if [ -z "${VERSION-}" ]; then
  echo "Note: VERSION not set; defaulting to dev version"
  VERSION="$(date '+%Y.%m.%d+dev')"
fi

export GO111MODULE=on
export CGO_ENABLED=1

ldflags="-X github.com/sourcegraph/sourcegraph/internal/version.version=$VERSION"
ldflags="$ldflags -X github.com/sourcegraph/sourcegraph/internal/version.timestamp=$(date +%s)"
ldflags="$ldflags -X github.com/sourcegraph/sourcegraph/internal/conf/deploy.forceType=app"

NODE_ENV=production pnpm run build-app-shell
go build \
  -o .bin/sourcegraph-backend-aarch64-apple-darwin \
  -trimpath \
  -tags dist \
  -ldflags "$ldflags" \
  ./enterprise/cmd/sourcegraph

pnpm tauri build
