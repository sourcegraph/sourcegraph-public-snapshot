#!/usr/bin/env bash

cd "$(dirname "${BASH_SOURCE[0]}")"

set -euxo pipefail

export P4USER="${P4USER:-admin}"

export P4CLIENT="${P4CLIENT:-"integration-test-client"}"
export P4_CLIENT_HOST="${P4_CLIENT_HOST:-"$(hostname)"}"

export DEPOT_NAME="${DEPOT_NAME:-"integration-test-depot"}" # the name of the depot that this script will create on the server
export DEPOT_DESCRIPTION="${DEPOT_DESCRIPTION:-"$(printf "Created by %s for integration testing." "${P4USER}")"}"

DATE="$(date '+%Y/%m/%d %T')"
export DATE

DEPOT_DIR="$(dirname "${BASH_SOURCE[0]}")/base"
export DEPOT_DIR

push_file_to_perforce() {
  local file="$1"

  local changelist_spec
  local changelist_number

  echo "pushing '$file' to server..."

  changelist_spec="$(envsubst <"$(dirname "${BASH_SOURCE[0]}")"./changelist_template.txt)"
  changelist_number=$(printf "%s\n" "$changelist_spec" | p4 change -i | awk '{print $2}')

  p4 add -c "$changelist_number" "$file"
  p4 submit -c "$changelist_number"
}
export -f push_file_to_perforce

run_parallel() {
  if ! command -v parallel &>/dev/null; then
    printf "'%s' command not installed. Please install '%s' by:\n\t- (macOS): running %s\n\t- (Linux): installing it via your distribution's package manager\nSee %s for more information.\n" \
      "parallel" \
      "GNU parallel" \
      "brew install parallel" \
      "https://www.gnu.org/software/parallel/"
    exit 1
  fi

  # Remove parallel citation log spam.
  echo 'will cite' | parallel --citation &>/dev/null

  parallel -0 --keep-order --line-buffer "$@"
}
export -f run_parallel

# delete older copy of depot if it exists
if p4 depots | awk '{print $2}' | grep -q "${DEPOT_NAME}"; then
  p4 obliterate -y "//${DEPOT_NAME}/..."
  p4 depot -df "${DEPOT_NAME}"
fi

# create depot
envsubst <./depot_template.txt | p4 depot -i

# delete older copy of client if it exists
if p4 clients | awk '{print $2}' | grep -q "${P4CLIENT}"; then
  p4 client -f -Fs -d "${P4CLIENT}"
fi

# create client
envsubst <./client_template.txt | p4 client -i

# submit every file in example depot and submit it to the Perforce server
cd "${DEPOT_DIR}"
find . -type f -print0 | run_parallel push_file_to_perforce {}
