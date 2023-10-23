#!/usr/bin/env bash

set -eu

cd "$(dirname "${BASH_SOURCE[0]}")"/../.. || exit 1

# We need the current go build since when cross compiling using bazel
# the zig compiler or bazel is unable to find system libraries
go_build() {
  local platform
  local version
  platform=$1
  version=$2

  if [ -z "${SKIP_BUILD_WEB-}" ]; then
    echo "--- :chrome: Building web"
    # esbuild is faster
    pnpm install
    NODE_ENV=production CODY_APP=1 pnpm run build-web
  fi

  export GO111MODULE=on
  export CGO_ENABLED=1

  local ldflags
  ldflags="-s -w"
  ldflags="$ldflags -X github.com/sourcegraph/sourcegraph/internal/version.version=${version}"
  ldflags="$ldflags -X github.com/sourcegraph/sourcegraph/internal/version.timestamp=$(date +%s)"
  ldflags="$ldflags -X github.com/sourcegraph/sourcegraph/internal/conf/deploy.forceType=app"

  echo "--- :go: Building Sourcegraph Backend (${version}) for platform: ${platform}"
  GOOS=darwin GOARCH=amd64 go build \
    -o ".bin/sourcegraph-backend-${platform}" \
    -trimpath \
    -tags dist \
    -ldflags "$ldflags" \
    ./cmd/sourcegraph
}

bazelrc() {
  if [[ $(uname -s) == "Darwin" ]]; then
    echo "--bazelrc=.bazelrc --bazelrc=.aspect/bazelrc/ci.bazelrc --bazelrc=.aspect/bazelrc/ci.macos.bazelrc"
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
  bazel_target="//cmd/sourcegraph:sourcegraph"
  bazel_opts="--stamp --workspace_status_command=./dev/app/app-stamp-vars.sh"
  platform=$1
  bin_dir=$2

  if [[ ${CI:-""} == "true" ]]; then
    bazel_cmd="${bazel_cmd} $(bazelrc)"
  fi

  # we need special flags and targets when cross compiling
  # for more info see the BUILD.bazel file in cmd/sourcegraph
  if [[ ${CROSS_COMPILE_X86_64_MACOS:-0} == 1 ]]; then
    bazel_target="//cmd/sourcegraph:sourcegraph_x86_64_darwin"
    # we don't use the incompat-zig-linux-amd64 bazel config here, since we need bazel to pick up the host cc
    bazel_opts="${bazel_opts} --platforms @zig_sdk//platform:darwin_amd64 --extra_toolchains @zig_sdk//toolchain:darwin_amd64"
  fi

  echo "--- :bazel: Building Sourcegraph Backend (${VERSION}) for platform: ${platform}"
  # shellcheck disable=SC2086
  ${bazel_cmd} build ${bazel_target} ${bazel_opts}

  out=$(bazel cquery //cmd/sourcegraph:sourcegraph --output=files)
  mkdir -p "${bin_dir}"
  chmod +x "${out}"
  cp -vf "${out}" "${bin_dir}/sourcegraph-backend-${platform}"
}

upload_artifacts() {
  local platform
  local target_path
  platform=$1
  target_path=$2
  buildkite-agent artifact upload "${target_path}"
}

# determine platform if it is not set
PLATFORM=${PLATFORM:-"$(./dev/app/detect-platform.sh)"}
export PLATFORM

VERSION="$(./dev/app/app-version.sh)"
export VERSION

if [[ ${CROSS_COMPILE_X86_64_MACOS:-0} == 1 ]]; then
  # TODO(burmudar) fix the bazel build - the --incompatible toolchain flag in the root .bazelrc is breaking it
  go_build "${PLATFORM}" "${VERSION}"
else
  bazel_build "${PLATFORM}" ".bin"
fi

# TODO(burmudar) move this to it's own file

if [[ ${CI:-""} == "true" ]]; then
  upload_artifacts "${PLATFORM}" ".bin/sourcegraph-backend-*"
fi
