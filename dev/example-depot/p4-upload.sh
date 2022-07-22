#!/usr/bin/env bash

cd "$(dirname "${BASH_SOURCE[0]}")"

OUTPUT=$(mktemp -d -t p4_upload_XXXXXXX)
cleanup() {
  rm -rf "$OUTPUT"
}
trap cleanup EXIT

set -euxo pipefail

DEPOT_NAME="${DEPOT_NAME:-integration-test}"

if p4 depots | awk '{print $2}' | grep -q "${DEPOT_NAME}"; then
  p4 obliberate "${DEPOT_NAME}"
fi

p4 depot -t local "${DEPOT_NAME}"
