#!/usr/bin/env bash

# This script builds the native binary single program.

cd "$(dirname "${BASH_SOURCE[0]}")/../../.."
set -eu

# Environment for building linux binaries
#export GOARCH=amd64
#export GOOS=linux

# macOS
#export GOARCH=arm64
#export GOOS=darwin
#export CGO_ENABLED=1
#export CC=clang
#export CGO_CFLAGS='-target arm64-apple-darwin21.6.0'

test() {
  grep -q "^${1}$" <<<"${supported_platforms}"
}

build() {
  local os arch ext binary
  os=${1%%/*}
  arch=${1##*/}
  ext=""
  [[ ${os} == windows ]] && ext=".exe"
  binary=".bin/$(basename ${pkg})-${os}-${arch}-dist${ext}"
  echo "--- go build for ${os}/${arch}"
  (ENTERPRISE=1 \
    DEV_WEB_BUILDER="esbuild yarn run build-web" \
    GOOS="${os}" \
    GOARCH="${arch}" \
    go build -trimpath \
    -ldflags "${ldflags[*]}" \
    -buildmode exe \
    -tags dist \
    -o "${binary}" "${pkg}")
  success=$?
  if [ ${success} -eq 0 ]; then
    printf "Go build succeeded for %s/%s" "${os}" "${arch}"
    if [ -s "${binary}" ]; then
      printf " the binary file is %s" "${binary}"
    else
      printf " but the binary file %s is missing" "${binary}"
    fi
    printf "\n"
  else
    printf "Go build failed for %s/%s" "${os}" "${arch}"
    [ -s "${binary}" ] && printf " the binary file %s is still in place - it's probably an old build" "${binary}"
    printf "\n"
  fi
}

pkg="github.com/sourcegraph/sourcegraph/enterprise/cmd/sourcegraph"
ldflags=("-X github.com/sourcegraph/sourcegraph/internal/version.version=${VERSION-0.0.0+dev}")
ldflags+=("-X github.com/sourcegraph/sourcegraph/internal/version.timestamp=$(date +%s)")
ldflags+=("-X github.com/sourcegraph/sourcegraph/internal/conf/deploy.forceType=single-program")

native_platform="$(go env GOOS)/$(go env GOARCH)"
supported_platforms="$(go tool dist list)"

target_platforms=()

while [ $# -gt 0 ]; do
  case "${1%%=*}" in
    -h | --help)
      cat 1>&2 <<-EOF
Builds native binaries for local installs.
Defaults to building a binary for the local platform,
but can build binaries for many platforms.
Some common examples:
  - darwin/arm64
  - darwin/amd64
  - windows/amd64
  - windows/arm64
  - linux/amd64
  - linux/arm64
The target platforms are limited only by what Go is capable of.
All of the aforementioned should be supported out of the box.
To specify desired target platforms, use the --target options
on the command line. Multiple targets in a single run are supported.
Example:
  ${0} --target=darwin/arm64 --target linux/amd64
Building for Windows is not yet supported.
EOF
      exit 1
      ;;
    --target)
      if [[ ${1} == --target=* ]]; then
        platform=${1##*=}
      elif [[ $# -gt 1 ]]; then
        shift
        platform=${1}
      else
        echo "--target requires a value" 1>&2
        exit 1
      fi
      [[ ${platform} == */* ]] || {
        echo "${platform} is not a valid platform specifier" 1>&2
        exit 1
      }
      test "${platform}" || {
        echo "not able to build for platform ${platform}" 1>&2
        exit 1
      }
      target_platforms+=("${platform}")
      ;;
    *)
      echo "unsupported option ${1}" 1>&2
      exit 1
      ;;
  esac
  shift
done

[ ${#target_platforms[@]} -eq 0 ] && target_platforms=("${native_platform}")

for platform in "${target_platforms[@]}"; do
  build "${platform}" &
done

wait
