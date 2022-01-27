#!/bin/bash

function reorganize() {
  for file in "${1}"/*; do
    if [[ "${file}" == *.sql ]]; then
      target="${1}/$(basename "${file}" | cut -d'_' -f1)"
      mkdir -p "${target}"
      mv "${file}" "${target}"
    fi
  done

  for dir in "${1}"/*; do
    for upfile in "${dir}/"*.up.sql; do
      if [[ "${upfile}" != "${dir}/*.up.sql" ]]; then
        version=$(basename "${upfile}" | cut -d'_' -f1)
        name=$(basename "${upfile}" | cut -d'_' -f2- | cut -d'.' -f1)
        echo "name: '${name}'" >"${dir}/metadata.yaml"
        mv "${dir}/${version}_${name}.up.sql" "${dir}/up.sql"
        mv "${dir}/${version}_${name}.down.sql" "${dir}/down.sql"
      fi
    done
  done

  for dir in "${1}"/*; do
    python3 "$(dirname "${BASH_SOURCE[0]}")/reorganize.py" "${dir}/up.sql" "${dir}/metadata.yaml"
  done
}

while (("$#")); do
  echo "Reorganizing ${1}"
  reorganize "${1}"
  shift
done
