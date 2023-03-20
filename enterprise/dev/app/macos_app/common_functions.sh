#!/usr/bin/env bash

# shellcheck disable=SC2064

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
    trap "popd 1>/dev/null 2>&1; rm -rf \"${temp_src}\"" EXIT
  else
    trap "popd 1>/dev/null 2>&1; rm -rf \"${temp_src}\"" RETURN
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
