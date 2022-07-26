#!/usr/bin/env bash

SCRIPT_ROOT="$(dirname "${BASH_SOURCE[0]}")"
cd "${SCRIPT_ROOT}"

set -euxo pipefail

export P4USER="${P4USER:-"admin"}"                          # the name of the Perforce superuser that the script will use to create the depot
export P4PORT="${P4PORT:-"perforce.sgdev.org:1666"}"        # the address of the Perforce server to connect to
export P4CLIENT="${P4CLIENT:-"integration-test-client"}"    # the name of the client that the script will use while it creates the depot
export DEPOT_NAME="${DEPOT_NAME:-"integration-test-depot"}" # the name of the depot that the script will create on the server

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

delete_perforce_client() {
  if p4 clients | awk '{print $2}' | grep -q "${P4CLIENT}"; then
    p4 client -f -Fs -d "${P4CLIENT}"
  fi
}

export DEPOT_DIR="${SCRIPT_ROOT}/base"
cd "${DEPOT_DIR}"

# check to see if user is logged in
if ! p4 ping >/dev/null; then
  printf "'%s' command failed. This indicates that you might not be logged into %s@%s.\nTry using %s to generate a session ticket. See %s for more information.\n" \
    "p4 ping" \
    "${P4USER}" \
    "${P4PORT}" \
    "p4 -u ${P4USER} login -a" \
    "https://handbook.sourcegraph.com/departments/ce-support/support/process/p4-enablement/#generate-a-session-ticket"
  exit 1
fi

# delete older copy of depot if it exists
if p4 depots | awk '{print $2}' | grep -q "${DEPOT_NAME}"; then
  p4 obliterate -y "//${DEPOT_NAME}/..."
  p4 depot -df "${DEPOT_NAME}"
fi

# create new depot
# shellcheck disable=SC2034 ## (used by envsubst)
DEPOT_DESCRIPTION="$(printf "Created by %s for integration testing purposes." "${P4USER}")"
envsubst <"${SCRIPT_ROOT}/depot_template.txt" | p4 depot -i

# delete older copy of client (if it exists)
delete_perforce_client

# create new client
# shellcheck disable=SC2034 # (used by envsubst)
P4_CLIENT_HOST="$(hostname)"
envsubst <"${SCRIPT_ROOT}/client_template.txt" | p4 client -i

# ensure that we don't leave a client behind (using up one of our licenses)
trap delete_perforce_client EXIT

# discover every file in example depot and submit it to the Perforce server
find . -type f -print0 | run_parallel push_file_to_perforce {}
