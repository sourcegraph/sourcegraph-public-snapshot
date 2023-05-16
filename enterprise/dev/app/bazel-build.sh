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
  local bazel_target
  local bazel_opts
  local platform
  local bin_dir
  bazel_cmd="bazel"
  bazel_target="//enterprise/cmd/sourcegraph:sourcegraph"
  bazel_opts="--stamp --workspace_status_command=./enterprise/dev/app/app_stamp_vars.sh"
  platform=$1
  bin_dir=$2

  if [[ ${CI:-""} == "true" ]]; then
    bazel_cmd="${bazel_cmd} $(bazelrc)"
  fi

  # we need special flags and targets when cross compiling
  # for more info see the BUILD.bazel file in enterprise/cmd/sourcegraph
  if [[ ${CROSS_COMPILE_X86_64_MACOS:-0} == 1 ]]; then
    bazel_target="//enterprise/cmd/sourcegraph:sourcegraph_x86_64_darwin"
    bazel_opts="${bazel_opts} --platform @zig_sdk//platform:darwin_amd64 --extra_toolchains @zig_sdk//toolchain:darwin_amd64"
  fi

  echo "--- :bazel: Building Sourcegraph Backend (${VERSION}) for platform: ${platform}"
  ${bazel_cmd} build ${bazel_target} ${bazel_opts}

  out=$(bazel cquery //enterprise/cmd/sourcegraph:sourcegraph --output=files)
  mkdir -p "${bin_dir}"
  cp -vf "${out}" "${bin_dir}/sourcegraph-backend-${platform}"
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
  case "$(uname -m)" in
    "amd64")
      arch="x86_64"
      ;;
    "arm64")
      arch="aarch64"
      ;;
    *)
      arch=$(uname -m)
  esac

  case "$(uname -s)" in
    "Darwin")
      platform="${arch}-apple-darwin"
      ;;
    "Linux")
      platform="${arch}-unknown-linux-gnu"
      ;;
    *)
      platform="${arch}-unknown-unknown"
  esac

  echo ${platform}
}

# determine platform if it is not set
PLATFORM=${PLATFORM:-$(platform)}
export PLATFORM
export PLATFORM_IS_MACOS=if [[ $(uname -s) == "Darwin" ]]; then 1; else 0 fi;


VERSION=$(./enterprise/dev/app/app_version.sh)
export VERSION

bazel_build "${PLATFORM}" ".bin"

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
