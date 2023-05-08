#!/usr/bin/env bash

set -eu

GCLOUD_APP_CREDENTIALS_FILE=${GCLOUD_APP_CREDENTIALS_FILE-$HOME/.config/gcloud/application_default_credentials.json}
cd "$(dirname "${BASH_SOURCE[0]}")"/../../.. || exit 1

bazel_build() {
  platform=$1
  bazel build //enterprise/cmd/sourcegraph:sourcegraph \
  --stamp \
  --workspace_status_command=./dev/app_stamp_vars.sh \

  out=$(bazel cquery //enterprise/cmd/sourcegraph:sourcegraph --output=files)
  cp -vf "${out}" .bin/sourcegraph-backend-${platform}
}

# We need to determine the platform string for the sourcegraph-backend binary
case "$(uname -m)" in
  "amd64")
    ARCH="x86_64"
    ;;
  "arm64")
    ARCH="aarch64"
    ;;
  *)
    ARCH=$(uname -m)
esac

case "$(uname -s)" in
  "Darwin")
    PLATFORM="${ARCH}-apple-darwin"
    ;;
  "Linux")
    PLATFORM="${ARCH}-unknown-linux-gnu"
    ;;
  *)
    PLATFORM="${ARCH}-unknown-unknown"
esac

echo "[Bazel] Building Sourcegraph for platform: ${PLATFORM}"
bazel_build ${PLATFORM}
echo "[Tauri] Building Application"]
NODE_ENV=production pnpm run build-app-shell
pnpm tauri build
