#!/usr/bin/env bash

# shellcheck disable=SC2064

# build src-cli for macOS
# since this is a Golang program, it can be built on any system,
# but this approach requires building on macOS
# because it uses lipo to make a universal/fat binary

DIR="${1}"
[[ $# -lt 1 ]] || [[ -z "${1}" ]] && {
  DIR="${PWD}"
}

cd "${DIR}" || exit 1

make_fat_binary() {
  local workdir temp_src dest_path dest_file arch_file arch_files

  workdir="${PWD}"

  dest_path="${1}"
  shift

  [[ ${dest_path} = /* ]] || dest_path="${PWD}/${dest_path}"

  pathchk -- "${dest_path}" || return 1

  dest_file=$(basename "${dest_path}")

  temp_src=$(mktemp -d 2>/dev/null || mktemp -d -t src-cli-XXXXX 2>/dev/null)

  pushd "${temp_src}" 1>/dev/null || return 1

  if [ -n "${ZSH_VERSION}" ]; then
    trap "popd 1>/dev/null && rm -rf \"${temp_src}\"" EXIT
  else
    trap "popd 1>/dev/null && rm -rf \"${temp_src}\"" RETURN
  fi

  # download and build the tool used to make multi-arch (fat) binaries
  curl -fsSL https://github.com/randall77/makefat/archive/refs/heads/master.zip -o makefat.zip
  unzip -o makefat.zip
  docker run --rm \
    -v "${PWD}/makefat-master:/makefat" \
    -w "/makefat" \
    golang:alpine \
    go build || return 1

  # use alpine to make the fat binary so that the OS and arch of makefat match
  # could do pattern matching on the output of `file $(which docker)`
  # to determine local OS and arch but that seems fragile
  rm -rf fatbinary
  mkdir fatbinary
  cp makefat-master/makefat fatbinary/makefat

  arch_files=()
  for arch_file in "${@}"; do
    [[ "${arch_file}" = /* ]] || arch_file="${workdir}/${arch_file}"
    cp "${arch_file}" fatbinary || return 1
    arch_files+=("$(basename "${arch_file}")")
  done

  docker run --rm \
    -v "${PWD}/fatbinary:/fatbinary" \
    -w "/fatbinary" \
    alpine:latest \
    ./makefat "${dest_file}" "${arch_files[@]}" || return 1

  cp "fatbinary/${dest_file}" "${dest_path}" || return 1
  return 0
}

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
  local workdir temp_src

  workdir="${PWD}"

  temp_src=$(mktemp -d 2>/dev/null || mktemp -d -t src-cli-XXXXX 2>/dev/null)

  pushd "${temp_src}" 1>/dev/null || return 1

  if [ -n "${ZSH_VERSION}" ]; then
    trap "popd 1>/dev/null && rm -rf \"${temp_src}\"" EXIT
  else
    trap "popd 1>/dev/null && rm -rf \"${temp_src}\"" RETURN
  fi

  cat >expected_hash <<-EOF
		8f3e4892e924221633688d53e64c509f19fd64fb5d31bd56165a1c87b972a74a  src-cli_4.5.0_darwin_amd64.tar.gz
		61436fa145e549e3cdc41bf95c47963fc5294277f01c397b7fde767cf348a2a6  src-cli_4.5.0_darwin_arm64.tar.gz
	EOF
  curl -fsSLO https://github.com/sourcegraph/src-cli/releases/download/4.5.0/src-cli_4.5.0_darwin_amd64.tar.gz
  curl -fsSLO https://github.com/sourcegraph/src-cli/releases/download/4.5.0/src-cli_4.5.0_darwin_arm64.tar.gz
  sha256sum -c expected_hash >/dev/null || {
    echo "corrupt download!" 1>&2
    return 1
  }
  mkdir amd64 && tar -xzf src-cli_4.5.0_darwin_amd64.tar.gz -C amd64
  mkdir arm64 && tar -xzf src-cli_4.5.0_darwin_arm64.tar.gz -C arm64

  make_fat_binary src-universal-4.5.0 amd64/src arm64/src || return 1
  echo "${PWD}/src-universal-4.5.0"
  return 0
}

build_from_release
