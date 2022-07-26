#!/usr/bin/env bash

cd "$(dirname "${BASH_SOURCE[0]}")"

OUTPUT=$(mktemp -d -t p4_upload_XXXXXXX)
cleanup() {
  rm -rf "$OUTPUT"
}
trap cleanup EXIT

gitp4() {
  "$(dirname "${BASH_SOURCE[0]}")"/git-p4.py "$@"
}

set -euxo pipefail

export P4USER="${P4USER:-admin}"

export P4CLIENT="${P4CLIENT:-"integration-test-client"}"
export P4_CLIENT_HOST="${P4_CLIENT_HOST:-"$(hostname)"}"

export DEPOT_NAME="${DEPOT_NAME:-"integration-test"}"
export DEPOT_DESCRIPTION="${DEPOT_DESCRIPTION:-"$(printf "Created by %s for integration testing." "${P4USER}")"}"

DATE="$(date '+%Y/%m/%d %T')"
export DATE

export DEPOT_DIR="${OUTPUT}/depot"

# delete older copy of depot if it exists
if p4 depots | awk '{print $2}' | grep -q "${DEPOT_NAME}"; then
  p4 obliterate -y "//${DEPOT_NAME}/..."
  p4 depot -d "${DEPOT_NAME}"
fi

# create depot
envsubst <./depot_template.txt | p4 depot -i

# delete older copy of client if it exists
if p4 clients | awk '{print $2}' | grep -q "${P4CLIENT}"; then
  p4 client -f -Fs -d "${P4CLIENT}"
fi

# create client
envsubst <./client_template.txt | p4 client -i

# create depot
envsubst <./depot_template.txt | p4 depot -i

# create client

# clone perforce depot
gitp4 clone "//${DEPOT_NAME}/..." "${DEPOT_DIR}"

# all files to perforce depot
cp -r ./base/ "${DEPOT_DIR}"

cd "${DEPOT_DIR}"

# git checkout -b master
gitp4 sync

# commit all above files
git add --all
git commit -m "initial commit"

cat .git/config

git remote

exa -lah --tree .

gitp4 submit "$(git rev-parse HEAD)"
