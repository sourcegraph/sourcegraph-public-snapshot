#!/usr/bin/env bash

cd "$(dirname "${BASH_SOURCE[0]}")"
set -euo pipefail

find . -name "buf.gen.yaml" -not -path ".git" | while read -r buf_yaml; do
  pushd "$(dirname "${buf_yaml}")" >/dev/null

  if ! buf generate; then
    echo "failed to generate ${buf_yaml}" >&2
    exit 1
  fi

  popd >/dev/null
done
