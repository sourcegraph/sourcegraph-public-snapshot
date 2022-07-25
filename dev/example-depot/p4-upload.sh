#!/usr/bin/env bash

cd "$(dirname "${BASH_SOURCE[0]}")"

OUTPUT=$(mktemp -d -t p4_upload_XXXXXXX)
cleanup() {
  rm -rf "$OUTPUT"
}
trap cleanup EXIT

set -euxo pipefail

export P4_USER="${P4_USER:-admin}"

export DEPOT_NAME="${DEPOT_NAME:-"integration-test"}"
export DEPOT_DESCRIPTION="${DEPOT_DESCRIPTION:-"$(printf "Created by %s for integration testing." "${P4_USER}")"}"
export DATE="$(date '+%Y/%m/%d %T')"

# delete older copy of depot if it exists
if p4 depots | awk '{print $2}' | grep -q "${DEPOT_NAME}"; then
  p4 obliterate -y "//${DEPOT_NAME}/..."
  p4 depot -d "${DEPOT_NAME}"
fi

# create depot
envsubst <./depot_template.txt | p4 depot -i

DEPOT_DIR="${OUTPUT}/depot"

# clone perforce depot
./git-p4.py clone "//${DEPOT_NAME}/..."@all "${DEPOT_DIR}"

# all files to perforce depot
cp -r ./base/ "${DEPOT_DIR}"

cd "${DEPOT_DIR}"

git checkout -b master

# commit all above files
git add --all
git commit -m "initial commit"

# TODO: Idempotently create a workspace client, have git use it, and submit all the changes (./git-p4.py submit ...)
