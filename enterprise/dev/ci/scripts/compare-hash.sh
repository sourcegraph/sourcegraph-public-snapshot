#!/usr/bin/env bash

# This script exits 0 if the hash script output has changed against the previous revision,
# indicating a new build should be created. If nothing has changed, a special exit code
# 222 is returned.

HASH_SCRIPT=$1

set -ex -o pipefail

current_commit=$(git rev-parse HEAD)

checkout() {
  git checkout -f "$1"

  # Re-run asdf to ensure we have the correct set of utilities to
  # run the currently checked out version of the Go unit tests.
  echo "--- asdf install checked out tools"
  ./dev/ci/asdf-install.sh
}

restore() {
  checked_out_commit=$(git rev-parse HEAD)
  if [ "$current_commit" != "$checked_out_commit" ]; then
    echo "Restoring correct commit"
    checkout -
  else
    echo "Already on correct commit"
  fi
}
trap restore EXIT

# Build previous
checkout HEAD^

commit=$(git rev-parse HEAD)
echo "--- compare-hash.sh: running $HASH_SCRIPT against $commit"
if test -f "$HASH_SCRIPT"; then
  previous_hash=$($HASH_SCRIPT)
else
  echo "+++ Previous revision does not have a hash script at $HASH_SCRIPT"
  exit 0
fi

# Build current
checkout -
echo "--- compare-hash.sh: running $HASH_SCRIPT against $current_commit"
new_hash=$($HASH_SCRIPT)
if [ "$new_hash" == "$previous_hash" ]; then
  echo "+++ new_hash and previous_hash match - nothing has changed, exiting 222 soft fail"
  exit 222
else
  echo "+++ new_hash and previous_hash mismatch: '$new_hash' and '$previous_hash' respectively"
  echo 0
fi
