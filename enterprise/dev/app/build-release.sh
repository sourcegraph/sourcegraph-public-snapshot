#!/usr/bin/env bash

set -eu

GCLOUD_APP_CREDENTIALS_FILE=${GCLOUD_APP_CREDENTIALS_FILE-$HOME/.config/gcloud/application_default_credentials.json}
cd "$(dirname "${BASH_SOURCE[0]}")"/../../.. || exit 1

go_build() {
  platform=$1
  if [ -z "${SKIP_BUILD_WEB-}" ]; then
    # esbuild is faster
    pnpm install
    NODE_ENV=production ENTERPRISE=1 SOURCEGRAPH_APP=1 DEV_WEB_BUILDER=esbuild pnpm run build-web
  fi

  export GO111MODULE=on
  export CGO_ENABLED=1

  export GO111MODULE=on
  export CGO_ENABLED=1

  if [[ -z ${VERSION:-} ]]; then
    # get the last non rc tag
    # fetching the tag with the following command does not work in github actions so we just cat a file now
    # git tag -l --sort=creatordate | grep -E "^v[0-9]+.[0-9]+.[0-9]+$" | tail -n 1
    VERSION=$(cat ./enterprise/dev/app/VERSION)
  fi

  echo "[Sourcegraph Backend] version: ${VERSION}"

  ldflags="-X github.com/sourcegraph/sourcegraph/internal/version.version=${VERSION}"
  ldflags="$ldflags -X github.com/sourcegraph/sourcegraph/internal/version.timestamp=$(date +%s)"
  ldflags="$ldflags -X github.com/sourcegraph/sourcegraph/internal/conf/deploy.forceType=app"
  go build \
    -o .bin/sourcegraph-backend-${platform} \
    -trimpath \
    -tags dist \
    -ldflags "$ldflags" \
    ./enterprise/cmd/sourcegraph


}


bazel_build() {
  platform=$1
  bazel build //enterprise/cmd/sourcegraph:sourcegraph \
  --stamp \
  --workspace_status_command=./dev/bazel_stamp_vars.sh \

    #--//:assets_bundle_type=enterprise
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

NODE_ENV=production pnpm run build-app-shell
if [[ $# -eq 0 ]]; then
  # Default to "bazel" if no argument was provided
  echo "[Bazel] Building Sourcegraph for platform: ${PLATFORM}"
  bazel_build ${PLATFORM}
else
  echo "[Go] Building Sourcegraph for platform: ${PLATFORM}"
  go_build ${PLATFORM}
fi
pnpm tauri build
