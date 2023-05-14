#!/usr/bin/env bash

set -eu

cd "$(dirname "${BASH_SOURCE[0]}")"/../../.. || exit 1

bazelrc() {
  if [[ $(uname -s) == "Darwin" ]]; then
    echo "--bazelrc=.bazelrc --bazelrc=.aspect/bazelrc/ci.macos.bazelrc"
  else
    echo "--bazelrc=.bazelrc --bazelrc=.aspect/bazelrc/ci.bazelrc --bazelrc=.aspect/bazelrc/ci.sourcegraph.bazelrc"
  fi
}

bazel_build() {
  local bazel_cmd
  local platform
  local target_path
  platform=$1
  target_path=$2
  bazel_cmd="bazel"
  if [[ ${CI:-""} == "true" ]]; then
    bazel_cmd="${bazel_cmd} $(bazelrc)"
  fi

  echo "--- :bazel: Building Sourcegraph Backend (${VERSION}) for platform: ${platform}"
  ${bazel_cmd} build //enterprise/cmd/sourcegraph:sourcegraph --stamp --workspace_status_command=./enterprise/dev/app/app_stamp_vars.sh

  out=$(bazel cquery //enterprise/cmd/sourcegraph:sourcegraph --output=files)
  mkdir -p ".bin"
  cp -vf "${out}" "${target_path}"
}

pre_codesign() {
  local binary_path
  binary_path=$1
  # Tauri won't code sign our sidecar sourcegraph-backend Go binary for us, so we need to do it on
  # our own. https://github.com/tauri-apps/tauri/discussions/2269
  # For details on code signing, see doc/dev/background-information/app/codesigning.md
  if [[ ${PLATFORM_IS_MACOS} == 1 ]]; then
    # We expect the same APPLE_ env vars that Tauri does here, see https://tauri.app/v1/guides/distribution/sign-macos
    security create-keychain -p my_temporary_keychain_password my_temporary_keychain.keychain
    security set-keychain-settings my_temporary_keychain.keychain
    security unlock-keychain -p my_temporary_keychain_password my_temporary_keychain.keychain
    security list-keychains -d user -s my_temporary_keychain.keychain "$(security list-keychains -d user | sed 's/["]//g')"

    echo "$APPLE_CERTIFICATE" >cert.p12.base64
    base64 -d -i cert.p12.base64 -o cert.p12

    security import ./cert.p12 -k my_temporary_keychain.keychain -P "$APPLE_CERTIFICATE_PASSWORD" -T /usr/bin/codesign
    security set-key-partition-list -S apple-tool:,apple:, -s -k my_temporary_keychain_password -D "$APPLE_SIGNING_IDENTITY" -t private my_temporary_keychain.keychain

    echo "[Code Signing] binary: ${binary_path}"
    codesign --force -s "$APPLE_SIGNING_IDENTITY" --keychain my_temporary_keychain.keychain --deep "${binary_path}"

    security delete-keychain my_temporary_keychain.keychain
    security list-keychains -d user -s login.keychain
  fi
}

upload_artifacts() {
  local platform
  local target_path
  platform=$1
  target_path=$2
  buildkite-agent artifact upload "${target_path}"
}

platform() {
  # We need to determine the platform string for the sourcegraph-backend binary
  local arch=""
  local platform=""
  local macos=0
  case "$(uname -m)" in
    "amd64")
      arch="x86_64"
      ;;
    "arm64")
      arch="aarch64"
      ;;
    *)
      arch=$(uname -m)
      ;;
  esac

  case "$(uname -s)" in
    "Darwin")
      platform="${arch}-apple-darwin"
      macos=1
      ;;
    "Linux")
      platform="${arch}-unknown-linux-gnu"
      ;;
    *)
      platform="${arch}-unknown-unknown"
      ;;
  esac

  export PLATFORM=${platform}
  export PLATFORM_IS_MACOS=${macos}
}

VERSION=$(./enterprise/dev/app/app_version.sh)
export VERSION

set_platform
target_path=".bin/sourcegraph-backend-${PLATFORM}"

bazel_build "${PLATFORM}" "${target_path}"

# TODO(burmudar) move this to it's own file
if [[ ${CODESIGNING} == 1 ]]; then
  # If on a macOS host, Tauri will invoke the base64 command as part of the code signing process.
  # it expects the macOS base64 command, not the gnutils one provided by homebrew, so we prefer
  # that one here:
  export PATH="/usr/bin/:$PATH"

  pre_codesign "${target_path}"
fi

if [[ ${CI:-""} == "true" ]]; then
  upload_artifacts "${PLATFORM}" ".bin/sourcegraph-backend-*"
fi
