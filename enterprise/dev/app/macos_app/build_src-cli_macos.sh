#!/usr/bin/env bash

# shellcheck disable=SC2064

# build src-cli for macOS
# since this is a Golang program, it can be built on any system.
# It uses a non-macOS tool to create the universal binaries, so it can be built on Linux or macOS (or maybe even Windows)

exedir=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)

# shellcheck disable=SC1090 disable=SC1091
. "${exedir}/common_functions.sh"

workdir="${PWD}"
src_version=4.5.0

while [ ${#} -gt 0 ]; do
  case "${1}" in
    --src-cli-version)
      [ -z "${2}" ] && {
        echo "expected a src-cli version" 1>&2
        exit 1
      }
      src_version="${2}"
      shift
      ;;
    --workdir | --work-dir)
      [ -z "${2}" ] && {
        echo "expected a working directory" 1>&2
        exit 1
      }
      workdir="${2}"
      shift
      ;;
    --help)
      echo "$(basename "${BASH_SOURCE[0]}") [--workdir <directory to work in (defaults to PWD)>] [--src-cli-version <version> (defaults to ${src_version})]" 1>&2
      exit 1
      ;;
  esac
  shift
done

cd "${workdir}" || exit 1

# unused, but preserved for my future self
build_from_source() {
  [ -d src-cli ] || git clone https://github.com/sourcegraph/src-cli.git
  cd src-cli || return 1
  git reset --hard
  git clean -dfxq
  git pull
  for arch in arm64 amd64; do
    GOARCH=${arch} go build -o src-${arch} ./cmd/src
  done
  make_fat_binary src-universal src-{arm64,amd64} || return 1
  echo "${PWD}/src-universal"
  return 0
}

build_from_release() {
  curl -fsSLO "https://github.com/sourcegraph/src-cli/releases/download/${src_version}/src-cli_${src_version}_checksums.txt"
  grep -E "src-cli_${src_version}_darwin_(amd|arm)64[.]tar[.]gz" "src-cli_${src_version}_checksums.txt >darwin_checksums.txt"
  echo "downloading the Intel binary"
  curl -fsSLO "https://github.com/sourcegraph/src-cli/releases/download/${src_version}/src-cli_${src_version}_darwin_amd64.tar.gz"
  echo "downloading the Arm binary"
  curl -fsSLO "https://github.com/sourcegraph/src-cli/releases/download/${src_version}/src-cli_${src_version}_darwin_arm64.tar.gz"
  sha256sum -c darwin_checksums.txt >/dev/null || {
    echo "corrupt download!" 1>&2
    return 1
  }
  mkdir amd64 && tar -xzf "src-cli_${src_version}_darwin_amd64.tar.gz" -C amd64
  mkdir arm64 && tar -xzf "src-cli_${src_version}_darwin_arm64.tar.gz" -C arm64
  mkdir "src-universal-${src_version}" 2>/dev/null
  make_fat_binary "src-universal-${src_version}/src amd64/src" arm64/src || return 1
  cd "src-universal-${src_version}" || return 1
  tar cvzf "../src-universal-${src_version}.tar.gz" "src"
  echo "${workdir}/src-universal-${src_version}.tar.gz"
  return 0
}

build_from_release
