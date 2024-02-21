#!/usr/bin/env bash

set -euf -o pipefail

cd "$(dirname "${BASH_SOURCE[0]}")/../.."

version=$1
FLAKE_FILE="dev/backcompat/flakes.json"

# `disable_test ${path} ${prefix}` rewrites `func ${prefix}` to `func _${prefix}`
# in the given Go test file. This will return 1 if there was a matching test and
# return 0 otherwise.
function disable_test() {
  sed -i_bak "s/func ${2}/func _${2}/g" "${1}"

  local ret=1
  if diff "${1}" "${1}_bak" >/dev/null; then
    ret=0 # no diff
  fi

  rm "${1}_bak"
  return ${ret}
}

# `disable_test_file ${path} ${prefix}` rewrites `func ${prefix}` to `func _${prefix}`
# in the given Go test file. If there is no matching test, an unknown test message is
# displayed and the script is halted with exit code 1.
function disable_test_file() {
  if disable_test "${1}" "${2}"; then
    echo "Unknown test in ${1}: ${2}"
    exit 1
  fi
}

# `disable_test_dir ${path} ${prefix}` rewrites `func ${prefix}` to `func _${prefix}`
# in all Go test files under the given path.
function disable_test_dir() {
  local num_changed=0

  while read -r path; do
    if ! disable_test "${path}" "${2}"; then
      num_changed=$((num_changed + 1))
    fi
  done < <(find "${1}" -name '*_test.go' -type f)

  if [ ${num_changed} -eq 0 ]; then
    echo "Unknown test in ${1}: ${2}"
  fi
}

# `disable_test_path ${path} ${prefix}`
function disable_test_path() {
  echo "Disabling test '${2}*' in ${1}"

  if [ -d "${1}" ]; then
    disable_test_dir "${1}" "${2}"
  elif [ -f "${1}" ]; then
    echo "#### ${1}"
    disable_test_file "${1}" "${2}"
  fi
}


if [ -f "${FLAKE_FILE}" ]; then
  echo "Disabling tests listed in flakefile ${FLAKE_FILE} for tag ${version}"

  pairs=$(jq -r --arg version "${version}" 'select(.[$version] != null) | .[$version][] | "\(.path):\(.prefix)"' "${FLAKE_FILE}" )
  for pair in $pairs; do
    IFS=' ' read -ra parts <<<"${pair/:/ }"
    disable_test_path "${parts[0]}" "${parts[1]}"
  done
else
  echo "Flakefile '${FLAKE_FILE}' not found"
  exit 1
fi
